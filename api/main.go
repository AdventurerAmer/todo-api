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

	"github.com/AdventurerAmer/todo-api/internal/config"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/AdventurerAmer/todo-api/internal/core/services/listssrv"
	"github.com/AdventurerAmer/todo-api/internal/core/services/taskssrv"
	"github.com/AdventurerAmer/todo-api/internal/core/services/userssrv"
	"github.com/AdventurerAmer/todo-api/internal/repositories/listsrepo"
	"github.com/AdventurerAmer/todo-api/internal/repositories/tasksrepo"
	"github.com/AdventurerAmer/todo-api/internal/repositories/usersrepo"
	"github.com/AdventurerAmer/todo-api/internal/utils"
)

const version = "1.0.0"

type application struct {
	config  *config.Config
	mailer  *utils.Mailer
	storage *storage

	usersRepo    ports.UsersRepository // TODO: remove this from here.
	usersService ports.UsersService
	listsService ports.ListsService
	tasksService ports.TasksService
}

func main() {
	// var cfg config.Config
	// flag.StringVar(&cfg.env, "env", "dev", "Environment [dev|test|prod]")

	// flag.IntVar(&cfg.port, "port", 3000, "Server Port")

	// flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DB_DSN"), "PostgreSQL DSN")
	// flag.IntVar(&cfg.db.maxOpenConnections, "db-max-open-conns", 25, "PostgreSQL max open connections")
	// flag.IntVar(&cfg.db.maxIdelConnections, "db-max-idel-conns", 25, "PostgreSQL max idel connections")
	// var maxIdelTime string
	// flag.StringVar(&maxIdelTime, "db-max-idel-time", "15m", "PostgreSQL max connection idel time")

	// flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")

	// smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// flag.IntVar(&cfg.smtp.port, "smtp-port", smtpPort, "SMTP port")
	// flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP host")
	// flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	// flag.StringVar(&cfg.smtp.sender, "smtp-sender", os.Getenv("SMTP_SENDER"), "SMTP sender")

	// flag.StringVar(&cfg.jwt.secret, "jwt-secret", os.Getenv("JWT_SECRET"), "JWT secret")

	// flag.Float64Var(&cfg.limiter.maxRequestPerSecond, "limiter-max-rps", 2, "Rate Limiter max requests per second")
	// flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate Limiter max burst")
	// flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// var trustedOrigins string
	// flag.StringVar(&trustedOrigins, "cors-trusted-origins", "*", "Trusted CORS origins saperated by space")
	// flag.Parse()

	// d, err := time.ParseDuration(maxIdelTime)
	// if err != nil {
	// 	cfg.db.maxIdelTime = 15 * time.Minute
	// 	log.Printf(`invalid value %s for flag "db-max-idel-time" defaulting to %s`, maxIdelTime, cfg.db.maxIdelTime)
	// } else {
	// 	cfg.db.maxIdelTime = d
	// }

	// cfg.cors.trustedOrigins = strings.Fields(trustedOrigins)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config loading failed", "error", err)
		os.Exit(1)
	}

	db, err := openDB(cfg.MainDB)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to database")

	mailer :=
		utils.NewMailer(cfg.MailServer.Host, cfg.MailServer.Port, cfg.MailServer.Username, cfg.MailServer.Password, cfg.MailServer.Sender)

	usersRepo := usersrepo.NewPostgres(db)
	usersService := userssrv.New(usersRepo, templates, mailer, userssrv.DefaultConfig())

	listsRepo := listsrepo.NewPostgres(db)
	listsService := listssrv.New(listsRepo, listssrv.DefaultConfig())

	tasksRepo := tasksrepo.NewPostgres(db)
	tasksService := taskssrv.New(tasksRepo, taskssrv.DefaultConfig())

	app := &application{
		config:       cfg,
		mailer:       mailer,
		storage:      newStorage(),
		usersRepo:    usersRepo,
		usersService: usersService,
		listsService: listsService,
		tasksService: tasksService,
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
