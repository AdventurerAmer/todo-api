package userssrv

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	NameMaxChars     int
	PasswardMinChars int
	PasswardMaxChars int
	PasswordHashCost int

	TitleMaxChars       int
	DescriptionMaxChars int
}

type service struct {
	Config
	usersRepo ports.UsersRepository
}

func New(usersRepo ports.UsersRepository, config Config) ports.UsersService {
	return &service{
		Config:    config,
		usersRepo: usersRepo,
	}
}

func (srv *service) Create(ctx context.Context, req ports.CreateUserRequest) (ports.CreateUserResponse, error) {
	v := failures.NewValidator()
	srv.validateName(v, req.Name)
	v.CheckUTF8Email(req.Email)
	srv.validatePassward(v, req.Password)
	if err := v.Err(); err != nil {
		return ports.CreateUserResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), srv.PasswordHashCost)
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

	resp := ports.CreateUserResponse{
		User:    user,
		Message: "user was created successfully",
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

func (srv *service) validatePassward(v *failures.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.CheckUTF8("password", password)
	v.CheckRangeInc("password", utf8.RuneCountInString(password), srv.PasswardMinChars, srv.PasswardMaxChars)
}
