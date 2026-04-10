package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/infrastructure"
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
	MainDB         infrastructure.PostgresConfig
	TokensCache    infrastructure.RedisConfig
	MainBroker     infrastructure.RabbitMQConfig
	MailServer     MailServer
	Authentication Authentication
	Constants      Constants
}

type ServerConfig struct {
	Port                    int
	IdleTimeout             time.Duration
	ReadTimeout             time.Duration
	WriteTimeout            time.Duration
	GracefulShutdownTimeout time.Duration
	DefaultTimeout          time.Duration
	TLS                     bool
	CertFile                string
	KeyFile                 string
	TrustedOrigins          []string
}

type MailServer struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

type Authentication struct {
	JWTSecret            string
	JWTTokenExpiresAfter time.Duration
}

type Constants struct {
	NameMaxChars     int
	PasswordHashCost int
	PasswordMinChars int
	PasswordMaxChars int

	ActivationTokenExpiresAfter time.Duration

	QueueEmailTimeout      time.Duration
	QueueEmailMaxRetries   int
	QueueEmailMaxRetryTime time.Duration
	SendEmailTimeout       time.Duration

	TitleMaxChars       int
	DescriptionMaxChars int

	ContentMaxChars   int
	ListsDefaultQuery int
	ListsMaxQuery     int
	TasksDefaultQuery int
	TasksMaxQuery     int
}

func Load() (*Config, error) {
	defaultEnv := "dev"
	if testing.Testing() {
		defaultEnv = "test"
	}
	var env string
	flag.StringVar(&env, "env", defaultEnv, "server enviroment (dev, test, prod)")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("'godotenv.Load' failed: %w", err)
	}

	var err error
	cfg := &Config{
		Env: Environment(env),
	}

	cfg.Server.Port, err = loadInt("TODO_SERVER_PORT")
	cfg.Server.IdleTimeout, err = loadDuration("TODO_SERVER_IDEL_TIMEOUT")
	cfg.Server.ReadTimeout, err = loadDuration("TODO_SERVER_READ_TIMEOUT")
	cfg.Server.WriteTimeout, err = loadDuration("TODO_SERVER_WRITE_TIMEOUT")
	cfg.Server.GracefulShutdownTimeout, err = loadDuration("TODO_SERVER_GRACEFUL_SHUTDOWN_TIMEOUT")
	cfg.Server.DefaultTimeout, err = loadDuration("TODO_SERVER_DEFAULT_TIMEOUT")
	cfg.Server.TLS, err = loadBool("TODO_SERVER_TLS")
	cfg.Server.CertFile, err = loadString("TODO_SERVER_CERT_FILE")
	cfg.Server.KeyFile, err = loadString("TODO_SERVER_KEY_FILE")
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

	cfg.TokensCache.Addr, err = loadString("TODO_TOKENS_CACHE_ADDR")
	cfg.TokensCache.Password, err = loadString("TODO_TOKENS_CACHE_PASSWORD")
	cfg.TokensCache.DB, err = loadInt("TODO_TOKENS_CACHE_DB")
	cfg.TokensCache.PingTimeout, err = loadDuration("TODO_TOKENS_CACHE_TIMEOUT")

	cfg.MainBroker.Addr, err = loadString("TODO_MAIN_BROKER_ADDR")
	cfg.MainBroker.Username, err = loadString("TODO_MAIN_BROKER_USERNAME")
	cfg.MainBroker.Password, err = loadString("TODO_MAIN_BROKER_PASSWORD")
	cfg.MainBroker.ConnectionTimeout, err = loadDuration("TODO_MAIN_BROKER_CONNECTION_TIMEOUT")

	cfg.MailServer.Host, err = loadString("TODO_MAIL_SERVER_HOST")
	cfg.MailServer.Port, err = loadInt("TODO_MAIL_SERVER_PORT")
	cfg.MailServer.Username, err = loadString("TODO_MAIL_SERVER_USERNAME")
	cfg.MailServer.Password, err = loadString("TODO_MAIL_SERVER_PASSWORD")
	cfg.MailServer.Sender, err = loadString("TODO_MAIL_SERVER_SENDER")

	cfg.Authentication.JWTSecret, err = loadString("TODO_AUTHENTICATION_JWT_SECRET")
	cfg.Authentication.JWTTokenExpiresAfter, err = loadDuration("TODO_AUTHENTICATION_JWT_EXPIRES_AFTER")

	cfg.Constants.NameMaxChars, err = loadInt("TODO_NAME_MAX_CHARS")
	cfg.Constants.PasswordMinChars, err = loadInt("TODO_PASSWORD_MIN_CHARS")
	cfg.Constants.PasswordMaxChars, err = loadInt("TODO_PASSWORD_MAX_CHARS")
	cfg.Constants.PasswordHashCost, err = loadInt("TODO_PASSWORD_HASH_COST")
	cfg.Constants.ActivationTokenExpiresAfter, err = loadDuration("TODO_ACTIVATION_TOKEN_EXPIRES_AFTER")

	cfg.Constants.QueueEmailTimeout, err = loadDuration("QUEUE_EMAIL_TIMEOUT")
	cfg.Constants.QueueEmailMaxRetries, err = loadInt("QUEUE_EMAIL_MAX_RETRIES")
	cfg.Constants.QueueEmailMaxRetryTime, err = loadDuration("QUEUE_EMAIL_MAX_RETRY_TIME")
	cfg.Constants.SendEmailTimeout, err = loadDuration("SEND_EMAIL_TIMEOUT")

	cfg.Constants.TitleMaxChars, err = loadInt("TODO_TITLE_MAX_CHARS")
	cfg.Constants.DescriptionMaxChars, err = loadInt("TODO_DESCRIPTION_MAX_CHARS")
	cfg.Constants.ListsDefaultQuery, err = loadInt("TODO_LISTS_DEFAULT_QUERY")
	cfg.Constants.ListsMaxQuery, err = loadInt("TODO_LISTS_MAX_QUERY")
	cfg.Constants.TasksDefaultQuery, err = loadInt("TODO_TASKS_DEFAULT_QUERY")
	cfg.Constants.TasksMaxQuery, err = loadInt("TODO_TASKS_MAX_QUERY")
	if err != nil {
		return nil, fmt.Errorf("parsing config failed: %w", err)
	}

	return cfg, nil
}

func loadString(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("%s: %w", key, ErrKeyNotFound)
	}
	return val, nil
}

func loadInt(key string) (int, error) {
	s, ok := os.LookupEnv(key)
	if !ok {
		return 0, fmt.Errorf("%s: %w", key, ErrKeyNotFound)
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
		return 0, fmt.Errorf("%s: %w", key, ErrKeyNotFound)
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
		return false, fmt.Errorf("%s: %w", key, ErrKeyNotFound)
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
		return nil, fmt.Errorf("%s: %w", key, ErrKeyNotFound)
	}
	fields := strings.Split(s, sep)
	return fields, nil
}
