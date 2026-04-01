package tokenssrv

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"time"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/AdventurerAmer/todo-api/internal/utils"
	"github.com/google/uuid"
)

type service struct {
	usersRepo  ports.UsersRepository
	tokensRepo ports.TokensRepository
	templates  fs.FS
	mailer     *utils.Mailer
}

func New(usersRepo ports.UsersRepository,
	tokensRepo ports.TokensRepository,
	templates fs.FS,
	mailer *utils.Mailer) ports.TokensService {
	return &service{
		usersRepo:  usersRepo,
		tokensRepo: tokensRepo,
		templates:  templates,
		mailer:     mailer,
	}
}

func (srv *service) ActivateViaEmail(ctx context.Context, user domain.User) (ports.ActivateViaEmailResponse, error) {
	if user.IsActivated {
		resp := ports.ActivateViaEmailResponse{Message: "user is already activated"}
		return resp, nil
	}

	tmpl, err := template.ParseFS(srv.templates, "templates/*.gotmpl")
	if err != nil {
		err := fmt.Errorf("'template.ParseFS' failed: %w", err)
		return ports.ActivateViaEmailResponse{}, err
	}
	id := uuid.NewString()
	data := map[string]any{
		"code": id,
	}
	token := &domain.Token{
		ID:        id,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(10 * time.Minute), // TODO: hardcoding
		Scope:     domain.TokenScopeActivation,
	}
	if err := srv.tokensRepo.Create(ctx, token); err != nil {
		return ports.ActivateViaEmailResponse{}, fmt.Errorf("'tokensRepo.Create' failed: %w", err)
	}
	if err := srv.mailer.Send(user.Email, tmpl, data); err != nil {
		err := fmt.Errorf("'mailer.Send' failed: %w", err)
		return ports.ActivateViaEmailResponse{}, err
	}
	resp := ports.ActivateViaEmailResponse{
		Message: fmt.Sprintf("we have sent an activation code to your email: %s", user.Email),
	}
	return resp, nil
}

func (srv *service) ActivateUser(ctx context.Context, req ports.ActivateUserRequest) (ports.ActivateUserResponse, error) {
	token, err := srv.tokensRepo.Get(ctx, req.Token)
	if err != nil {
		if errors.Is(err, ports.ErrTokenNotFound) {
			return ports.ActivateUserResponse{}, &failures.ValidationError{Reason: "token is not valid or expired"}
		}
		return ports.ActivateUserResponse{}, fmt.Errorf("'tokensRepo.Get' failed: %w", err)
	}

	if time.Now().After(token.ExpiresAt) {
		return ports.ActivateUserResponse{}, &failures.ValidationError{Reason: "token is not valid or expired"}
	}

	user, err := srv.usersRepo.Get(ctx, token.UserID)
	if err != nil {
		return ports.ActivateUserResponse{}, fmt.Errorf("'usersRepo.Get' failed: %w", err)
	}

	if user.IsActivated {
		resp := ports.ActivateUserResponse{Message: "user was successfully activated"}
		return resp, nil
	}

	user.IsActivated = true
	if err := srv.usersRepo.Update(ctx, &user); err != nil {
		err := fmt.Errorf("'usersRepo.Update' failed: %w", err)
		return ports.ActivateUserResponse{}, err
	}

	resp := ports.ActivateUserResponse{Message: "user was successfully activated"}
	return resp, nil
}
