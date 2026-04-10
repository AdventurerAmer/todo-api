package v1

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AdventurerAmer/todo-api/infrastructure"
	"github.com/AdventurerAmer/todo-api/internal/brokers"
	"github.com/AdventurerAmer/todo-api/internal/config"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/AdventurerAmer/todo-api/internal/core/services/listssrv"
	"github.com/AdventurerAmer/todo-api/internal/core/services/taskssrv"
	"github.com/AdventurerAmer/todo-api/internal/core/services/tokenauthsrv"
	"github.com/AdventurerAmer/todo-api/internal/core/services/tokenssrv"
	"github.com/AdventurerAmer/todo-api/internal/core/services/userssrv"
	"github.com/AdventurerAmer/todo-api/internal/repositories/listsrepo"
	"github.com/AdventurerAmer/todo-api/internal/repositories/tasksrepo"
	"github.com/AdventurerAmer/todo-api/internal/repositories/tokensrepo"
	"github.com/AdventurerAmer/todo-api/internal/repositories/usersrepo"
	"github.com/AdventurerAmer/todo-api/web"
)

const version = "1.0.0"

type application struct {
	web.App

	config *config.Config

	usersService     ports.UsersService
	listsService     ports.ListsService
	tasksService     ports.TasksService
	tokensService    ports.TokensService
	tokenAuthService ports.TokenAuthService

	mainBroker ports.Broker
}

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config loading failed", "error", err)
		return 1
	}

	db, err := infrastructure.ConnectToPostgres(cfg.MainDB)
	if err != nil {
		slog.Error("main database connection failed", "error", err)
		return 1
	}

	slog.Info("connected to database")

	tokensCache, err := infrastructure.ConnectToRedis(cfg.TokensCache)
	if err != nil {
		slog.Error("tokens cache connection failed", "error", err)
		return 1
	}

	slog.Info("connected to cache")

	conn, err := infrastructure.ConnectToRabbitMQ(cfg.MainBroker)
	if err != nil {
		slog.Error("main broker connection failed", "error", err)
		return 1
	}

	slog.Info("connected to main broker")

	mainBroker, err := brokers.NewRabbitMQ(conn)
	if err != nil {
		slog.Error("main broker creation failed", "error", err)
		return 1
	}

	usersRepo := usersrepo.NewPostgres(db)

	tokensRepo := tokensrepo.NewRedis(tokensCache)

	tokensServiceCfg := tokenssrv.Config{
		ActivationTokenExpiresAfter: cfg.Constants.ActivationTokenExpiresAfter,
	}
	tokensService := tokenssrv.New(usersRepo, tokensRepo, mainBroker, tokensServiceCfg)

	usersServiceConfig := userssrv.Config{
		NameMaxChars:     cfg.Constants.NameMaxChars,
		PasswordHashCost: cfg.Constants.PasswordHashCost,
	}
	usersService := userssrv.New(usersRepo, tokensService, usersServiceConfig)

	listsRepo := listsrepo.NewPostgres(db)
	listsServiceConfig := listssrv.Config{
		TitleMaxChars:       cfg.Constants.TitleMaxChars,
		DescriptionMaxChars: cfg.Constants.DescriptionMaxChars,
	}
	listsService := listssrv.New(listsRepo, listsServiceConfig)

	tasksRepo := tasksrepo.NewPostgres(db)
	tasksServiceConfig := taskssrv.Config{
		ContentMaxChars: cfg.Constants.ContentMaxChars,
	}
	tasksService := taskssrv.New(tasksRepo, tasksServiceConfig)

	tokenauthsrvCfg := tokenauthsrv.JWTConfig{
		Secret:            cfg.Authentication.JWTSecret,
		TokenExpiresAfter: cfg.Authentication.JWTTokenExpiresAfter,
	}
	tokenAuthService := tokenauthsrv.NewJWT(usersRepo, tokenauthsrvCfg)

	authHandler := func(r *http.Request, token string) (context.Context, error) {
		dctx, cancel := context.WithTimeout(r.Context(), cfg.Server.DefaultTimeout)
		defer cancel()

		resp, err := tokenAuthService.Check(dctx, ports.CheckAuthTokenRequest{Token: token})
		if err != nil {
			err := fmt.Errorf("'tokenAuthService.Check' failed: %w", err)
			return nil, err
		}

		user := resp.User
		return context.WithValue(r.Context(), userContextKey, user), nil
	}

	app := &application{
		App: web.App{
			TrustedOrigins:   cfg.Server.TrustedOrigins,
			TokenAuthHandler: authHandler,
		},
		config:           cfg,
		usersService:     usersService,
		listsService:     listsService,
		tasksService:     tasksService,
		tokensService:    tokensService,
		tokenAuthService: tokenAuthService,

		mainBroker: mainBroker,
	}

	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		MaxVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		IdleTimeout:  cfg.Server.IdleTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		Handler:      composeRoutes(app),
	}

	if cfg.Server.TLS {
		srv.TLSConfig = tlsConfig
	}

	shutDownCh := make(chan error)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
		defer cancel()

		err := srv.Shutdown(ctx)
		shutDownCh <- err
	}()

	slog.Info("Server Started", "env", cfg.Env, "port", cfg.Server.Port)
	defer slog.Info("Server Stopped")

	if cfg.Server.TLS {
		if err := srv.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("graceful shutdown failed", "error", err)
		}
	} else {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("graceful shutdown failed", "error", err)
		}
	}

	if err := <-shutDownCh; err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		return 1
	}

	return 0
}
