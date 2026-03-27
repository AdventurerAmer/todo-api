package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/joho/godotenv"
)

var ErrKeyNotFound = &failures.ResourceNotFoundError{Name: "key"}

type Environment string

const (
	EnvDev  Environment = "dev"
	EnvTest Environment = "test"
	EnvProd Environment = "prod"
)

type Config struct {
	Env            Environment
	Server         ServerConfig
	MainDB         MainDB
	MailServer     MailServer
	Authentication Authentication
}

type ServerConfig struct {
	Port                    int
	IdleTimeout             time.Duration
	ReadTimeout             time.Duration
	WriteTimeout            time.Duration
	GracefulShutdownTimeout time.Duration
	TLS                     bool
	TrustedOrigins          []string
}

type MainDB struct {
	Username           string
	Password           string
	Host               string
	Port               int
	Name               string
	SSLMode            string
	MaxOpenConnections int
	MaxIdelConnections int
	MaxIdelTime        time.Duration
	StartupTimeout     time.Duration
	PingTimeout        time.Duration
}

type MailServer struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

type Authentication struct {
	JWTSecret string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("'godotenv.Load' failed: %w", err)
	}

	var err error
	cfg := &Config{}

	cfg.Server.Port, err = loadInt("TODO_SERVER_PORT")
	cfg.Server.IdleTimeout, err = loadDuration("TODO_SERVER_IDEL_TIMEOUT")
	cfg.Server.ReadTimeout, err = loadDuration("TODO_SERVER_READ_TIMEOUT")
	cfg.Server.WriteTimeout, err = loadDuration("TODO_SERVER_WRITE_TIMEOUT")
	cfg.Server.GracefulShutdownTimeout, err = loadDuration("TODO_SERVER_GRACEFUL_SHUTDOWN_TIMEOUT")
	cfg.Server.TLS, err = loadBool("TODO_SERVER_TLS")
	cfg.Server.TrustedOrigins, err = loadStringSlice("TODO_SERVER_TRUSTED_ORIGINS", ",")

	cfg.MainDB.Username, err = loadString("TODO_MAIN_DB_USERNAME")
	cfg.MainDB.Password, err = loadString("TODO_MAIN_DB_PASSWORD")
	cfg.MainDB.Host, err = loadString("TODO_MAIN_DB_HOST")
	cfg.MainDB.Port, err = loadInt("TODO_MAIN_DB_PORT")
	cfg.MainDB.Name, err = loadString("TODO_MAIN_DB_NAME")
	cfg.MainDB.SSLMode, err = loadString("TODO_MAIN_DB_SSL_MODE")
	cfg.MainDB.MaxOpenConnections, err = loadInt("TODO_MAIN_DB_MAX_OPEN_CONNECTIONS")
	cfg.MainDB.MaxIdelConnections, err = loadInt("TODO_MAIN_DB_MAX_IDEL_CONNECTIONS")
	cfg.MainDB.MaxIdelTime, err = loadDuration("TODO_MAIN_DB_IDEL_TIME")
	cfg.MainDB.StartupTimeout, err = loadDuration("TODO_MAIN_DB_STARTUP_TIMEOUT")
	cfg.MainDB.PingTimeout, err = loadDuration("TODO_MAIN_DB_PING_TIMEOUT")

	cfg.MailServer.Host, err = loadString("TODO_MAIL_SERVER_HOST")
	cfg.MailServer.Port, err = loadInt("TODO_MAIL_SERVER_PORT")
	cfg.MailServer.Username, err = loadString("TODO_MAIL_SERVER_USERNAME")
	cfg.MailServer.Password, err = loadString("TODO_MAIL_SERVER_PASSWORD")
	cfg.MailServer.Sender, err = loadString("TODO_MAIL_SERVER_SENDER")

	cfg.Authentication.JWTSecret, err = loadString("TODO_AUTHENTICATION_JWT_SECRET")

	if err != nil {
		return nil, fmt.Errorf("parsing config failed: %w", err)
	}

	return cfg, nil
}

func loadString(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", ErrKeyNotFound
	}
	return val, nil
}

func loadInt(key string) (int, error) {
	s, ok := os.LookupEnv(key)
	if !ok {
		return 0, ErrKeyNotFound
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("'strconv.Atoi' failed: %w", err)
	}
	return val, nil
}

func loadDuration(key string) (time.Duration, error) {
	s, ok := os.LookupEnv(key)
	if !ok {
		return 0, ErrKeyNotFound
	}
	val, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("'strconv.Atoi' failed: %w", err)
	}
	return val, nil
}

func loadBool(key string) (bool, error) {
	s, ok := os.LookupEnv(key)
	if !ok {
		return false, ErrKeyNotFound
	}
	val := strings.ToLower(s)
	switch val {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	return false, fmt.Errorf("invalid bool value expecting true, or false")
}

func loadStringSlice(key string, sep string) ([]string, error) {
	s, ok := os.LookupEnv(key)
	if !ok {
		return nil, ErrKeyNotFound
	}
	fields := strings.Split(s, sep)
	return fields, nil
}
