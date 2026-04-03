package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AdventurerAmer/todo-api/infrastructure"
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
	"github.com/AdventurerAmer/todo-api/internal/utils"
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
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config loading failed", "error", err)
		os.Exit(1)
	}

	db, err := infrastructure.ConnectToPostgres(cfg.MainDB)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to database")

	tokensCache, err := infrastructure.ConnectToRedis(cfg.TokensCache)
	if err != nil {
		slog.Error("tokens cache connection failed", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to cache")

	mailer :=
		utils.NewMailer(cfg.MailServer.Host, cfg.MailServer.Port, cfg.MailServer.Username, cfg.MailServer.Password, cfg.MailServer.Sender)

	usersRepo := usersrepo.NewPostgres(db)
	usersServiceConfig := userssrv.Config{
		NameMaxChars:     cfg.Constants.NameMaxChars,
		PasswordHashCost: cfg.Constants.PasswordHashCost,
	}
	usersService := userssrv.New(usersRepo, templates, mailer, usersServiceConfig)

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

	tokensRepo := tokensrepo.NewRedis(tokensCache)

	tokensService := tokenssrv.New(usersRepo, tokensRepo, templates, mailer)

	tokenauthsrvCfg := tokenauthsrv.JWTConfig{
		Secret:            cfg.Authentication.JWTSecret,
		TokenExpiresAfter: cfg.Authentication.JWTTokenExpiresAfter,
	}
	tokenAuthService := tokenauthsrv.NewJWT(usersRepo, tokenauthsrvCfg)

	app := &application{
		App: web.App{
			TrustedOrigins: cfg.Server.TrustedOrigins,
		},
		config:           cfg,
		usersService:     usersService,
		listsService:     listsService,
		tasksService:     tasksService,
		tokensService:    tokensService,
		tokenAuthService: tokenAuthService,
	}

	// tlsConfig := &tls.Config{
	// 	MinVersion:       tls.VersionTLS12,
	// 	MaxVersion:       tls.VersionTLS13,
	// 	CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	// 	CipherSuites: []uint16{
	// 		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	// 		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	// 		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	// 		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	// 		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	// 		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	// 	},
	// }

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		IdleTimeout:  cfg.Server.IdleTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		// TLSConfig:    tlsConfig,
		Handler: composeRoutes(app),
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

	slog.Info("Starting server", "env", cfg.Env, "port", cfg.Server.Port)
	defer slog.Info("Server Stopped")

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("graceful shutdown failed", "error", err)
	}

	if err := <-shutDownCh; err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}
