package userssrv

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"math/rand/v2"

	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/AdventurerAmer/todo-api/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	EmailMaxChars    int
	NameMaxChars     int
	PasswordMaxChars int
}

func DefaultConfig() Config {
	return Config{
		EmailMaxChars:    320,
		NameMaxChars:     256,
		PasswordMaxChars: 126,
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
	v := utils.NewValidator()
	v.CheckCond(req.Name != "", "name", "must be provided")
	v.CheckCond(len(req.Name) <= 255, "name", "must be atmost 255 characters")
	v.CheckEmail(req.Email)
	v.CheckPassword(req.Password)
	if v.HasErrors() {
		err := v.ToError()
		return ports.CreateUserResponse{}, err
	}

	if _, err := srv.usersRepo.GetByEmail(ctx, req.Email); err != nil {
		return ports.CreateUserResponse{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 13)
	if err != nil {
		return ports.CreateUserResponse{}, err
	}

	user := &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}
	if err := srv.usersRepo.Create(ctx, user); err != nil {
		return ports.CreateUserResponse{}, err
	}

	tmpl, err := template.ParseFS(srv.templates, "templates/*.gotmpl")
	if err != nil {
		return ports.CreateUserResponse{}, err
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
	return ports.GetUserResponse{}, nil
}

func (srv *service) Update(ctx context.Context, user *domain.User, req ports.UpdateUserRequest) (ports.UpdateUserResponse, error) {
	v := utils.NewValidator()

	if req.Name != "" {
		v.CheckCond(req.Name != "", "name", "must be provided")
		v.CheckCond(len(req.Name) <= 255, "name", "must be atmost 255 characters")
	}

	if v.HasErrors() {
		err := v.ToError()
		return ports.UpdateUserResponse{}, err
	}

	if req.Name != "" {
		user.Name = req.Name
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
