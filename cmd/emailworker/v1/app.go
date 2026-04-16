package v1

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/infrastructure"
	"github.com/AdventurerAmer/todo-api/internal/brokers"
	"github.com/AdventurerAmer/todo-api/internal/config"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/AdventurerAmer/todo-api/internal/repositories/usersrepo"
	"github.com/AdventurerAmer/todo-api/internal/utils"
)

const Version = "1.0.0"

func Run(templatesFS embed.FS) int {
	userActivationTmpl, err := template.ParseFS(templatesFS, "templates/user_activation.gotmpl")
	if err != nil {
		slog.Error("parsing templates failed", "error", err)
		return 1
	}

	templates := map[domain.Template]*template.Template{
		domain.UserActivationTemplate: userActivationTmpl,
	}

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

	conn, err := infrastructure.ConnectToRabbitMQ(cfg.MainBroker)
	if err != nil {
		slog.Error("main broker connection failed", "error", err)
		return 1
	}

	mainBroker, err := brokers.NewRabbitMQ(conn)
	if err != nil {
		slog.Error("main broker creation failed", "error", err)
		return 1
	}

	slog.Info("connected to main broker")

	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	mailer := utils.NewMailer(cfg.MailServer.Host, cfg.MailServer.Port, cfg.MailServer.Username, cfg.MailServer.Password, cfg.MailServer.Sender)

	usersRepo := usersrepo.NewPostgres(db)

	handler := func(ctx context.Context, contentType string, data []byte) (bool, error) {
		if contentType != "application/json" {
			return false, &failures.UnsupportedMediaTypeError{Type: contentType}
		}
		var req ports.SendEmailRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return false, fmt.Errorf("'json.Unmarshal' failed: %w", err)
		}

		if req.UserID == "" {
			return false, fmt.Errorf("'userID' is empty")
		}

		tmpl, ok := templates[req.Template]
		if !ok {
			return false, fmt.Errorf("unsupported template %q", req.Template)
		}

		dctx, cancel := context.WithTimeout(ctx, cfg.Constants.SendEmailTimeout)
		defer cancel()

		user, err := usersRepo.Get(dctx, req.UserID)
		if err != nil {
			if errors.Is(err, ports.ErrUserNotFound) {
				return false, err
			}
			return true, fmt.Errorf("'usersRepo.Get' failed: %w", err)
		}

		if err := mailer.Send(user.Email, tmpl, data); err != nil {
			err := fmt.Errorf("'mailer.Send' failed: %w", err)
			return true, err
		}

		return false, nil
	}
	if err := mainBroker.Consume(sigCtx, ports.EmailQueue, handler); err != nil {
		slog.Error("consume failed", "error", err)
		return 1
	}

	return 0
}
