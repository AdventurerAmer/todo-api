package userssrv

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"math/rand/v2"
	"unicode/utf8"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/AdventurerAmer/todo-api/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	NameMaxChars int
}

func DefaultConfig() Config {
	return Config{
		NameMaxChars: 256,
	}
}

type service struct {
	Config
	usersRepo ports.UsersRepository
	templates embed.FS
	mailer    *utils.Mailer
}

func New(usersRepo ports.UsersRepository, templates embed.FS, mailer *utils.Mailer, config Config) ports.UsersService {
	return &service{
		Config:    config,
		usersRepo: usersRepo,
		templates: templates,
		mailer:    mailer,
	}
}

func (srv *service) Create(ctx context.Context, req ports.CreateUserRequest) (ports.CreateUserResponse, error) {
	v := failures.NewValidator()
	srv.validateName(v, req.Name)
	v.CheckUTF8Email(req.Email)
	v.CheckUTF8Password(req.Password)
	if err := v.Err(); err != nil {
		return ports.CreateUserResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	// TODO: hardcoding
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 13)
	if err != nil {
		return ports.CreateUserResponse{}, fmt.Errorf("'bcrypt.GenerateFromPassword' failed: %w", err)
	}

	user := &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}
	if err := srv.usersRepo.Create(ctx, user); err != nil {
		return ports.CreateUserResponse{}, fmt.Errorf("'usersRepo.Create' failed: %w", err)
	}

	tmpl, err := template.ParseFS(srv.templates, "templates/*.gotmpl")
	if err != nil {
		return ports.CreateUserResponse{}, fmt.Errorf("' template.ParseFS' failed: %w", err)
	}
	code := uint16(rand.Uint())
	data := map[string]any{"code": code}
	if err := srv.mailer.Send(user.Email, tmpl, data); err != nil {
		return ports.CreateUserResponse{}, err
	}
	resp := ports.CreateUserResponse{
		User:    user,
		Message: fmt.Sprintf("we have sent an activation code to your email: %s", user.Email),
	}
	return resp, nil
}

func (srv *service) Get(ctx context.Context, req ports.GetUserRequest) (ports.GetUserResponse, error) {
	v := failures.NewValidator()
	v.Check(req.ID != "", "id", "must not be empty")
	if err := v.Err(); err != nil {
		return ports.GetUserResponse{}, fmt.Errorf("validation failed: %w", err)
	}
	user, err := srv.usersRepo.Get(ctx, req.ID)
	if err != nil {
		return ports.GetUserResponse{}, fmt.Errorf("'usersRepo.Get' failed: %w", err)
	}
	resp := ports.GetUserResponse{
		User: &user,
	}
	return resp, nil
}

func (srv *service) Update(ctx context.Context, user *domain.User, req ports.UpdateUserRequest) (ports.UpdateUserResponse, error) {
	v := failures.NewValidator()

	if req.Name != nil {
		srv.validateName(v, *req.Name)
	}

	if err := v.Err(); err != nil {
		return ports.UpdateUserResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if err := srv.usersRepo.Update(ctx, user); err != nil {
		return ports.UpdateUserResponse{}, fmt.Errorf("'usersRepo.Update' failed: %w", err)
	}

	return ports.UpdateUserResponse{
		User: user,
	}, nil
}

func (srv *service) Delete(ctx context.Context, req ports.DeleteUserRequest) (ports.DeleteUserResponse, error) {
	if err := srv.usersRepo.Delete(ctx, req.ID); err != nil {
		return ports.DeleteUserResponse{}, fmt.Errorf("'usersRepo.Delete' failed: %w", err)
	}

	return ports.DeleteUserResponse{Message: "user was deleted successfully"}, nil
}

func (srv *service) validateName(v *failures.Validator, name string) {
	v.Check(name != "", "name", "must be provided")
	v.CheckUTF8("name", name)
	v.CheckAtMostInc("name", utf8.RuneCountInString(name), srv.NameMaxChars, "characters long")
}
