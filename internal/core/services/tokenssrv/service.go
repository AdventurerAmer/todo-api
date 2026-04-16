package tokenssrv

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/google/uuid"
)

type Config struct {
	ActivationTokenExpiresAfter time.Duration
}

type service struct {
	Config
	usersRepo   ports.UsersRepository
	tokensRepo  ports.TokensRepository
	emailSender *ports.EmailSender
}

func New(usersRepo ports.UsersRepository,
	tokensRepo ports.TokensRepository,
	emailSender *ports.EmailSender, cfg Config) ports.TokensService {
	return &service{
		Config:      cfg,
		usersRepo:   usersRepo,
		tokensRepo:  tokensRepo,
		emailSender: emailSender,
	}
}

func (srv *service) ActivateViaEmail(ctx context.Context, user domain.User) (ports.ActivateViaEmailResponse, error) {
	if user.IsActivated {
		resp := ports.ActivateViaEmailResponse{Message: "user is already activated"}
		return resp, nil
	}

	id := uuid.NewString()
	token := &domain.Token{
		ID:        id,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(srv.ActivationTokenExpiresAfter),
		Scope:     domain.TokenScopeActivation,
	}
	if err := srv.tokensRepo.Create(ctx, token); err != nil {
		return ports.ActivateViaEmailResponse{}, fmt.Errorf("'tokensRepo.Create' failed: %w", err)
	}

	sendEmailMessage := ports.SendEmailRequest{
		UserID:   user.ID,
		Template: domain.UserActivationTemplate,
		Data: domain.UserActivationTemplateData{
			Code: id,
		},
	}
	srv.emailSender.SendAsync(sendEmailMessage)

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
