package tokenauthsrv

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type JWTConfig struct {
	Secret            string
	TokenExpiresAfter time.Duration
}

type jwtService struct {
	JWTConfig
	usersRepo ports.UsersRepository
}

func NewJWT(usersRepo ports.UsersRepository, cfg JWTConfig) ports.TokenAuthService {
	return &jwtService{
		JWTConfig: cfg,
		usersRepo: usersRepo,
	}
}

func (srv *jwtService) Create(ctx context.Context, req ports.CreateAuthTokenRequest) (ports.CreateAuthTokenResponse, error) {
	v := failures.NewValidator()
	v.CheckUTF8Email(req.Email)
	v.CheckUTF8Password(req.Password)
	if err := v.Err(); err != nil {
		return ports.CreateAuthTokenResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.usersRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ports.ErrUserNotFound) {
			err := &failures.AuthenticationError{Reason: "email or password is incorrect"}
			return ports.CreateAuthTokenResponse{}, err
		}
		return ports.CreateAuthTokenResponse{}, fmt.Errorf("'usersRepo.GetByEmail' failed: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(req.Password)); err != nil {
		err := &failures.AuthenticationError{Reason: "email or password is incorrect"}
		return ports.CreateAuthTokenResponse{}, err
	}

	claims := jwt.MapClaims{
		"userID":    user.ID,
		"expiresAt": time.Now().Add(srv.TokenExpiresAfter).Format(time.RFC822),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(srv.Secret))
	if err != nil {
		err := fmt.Errorf("'token.SignedString' failed: %w", err)
		return ports.CreateAuthTokenResponse{}, err
	}

	resp := ports.CreateAuthTokenResponse{
		Token: signedToken,
	}
	return resp, nil
}

func (srv *jwtService) Check(ctx context.Context, req ports.CheckAuthTokenRequest) (ports.CheckAuthTokenResponse, error) {
	v := failures.NewValidator()
	v.CheckNotEmpty("token", req.Token)
	if err := v.Err(); err != nil {
		err := &failures.AuthenticationError{Reason: "invalid token"}
		return ports.CheckAuthTokenResponse{}, err
	}

	token, err := jwt.Parse(req.Token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(srv.Secret), nil
	})
	if err != nil {
		err := &failures.AuthenticationError{Reason: "invalid token"}
		return ports.CheckAuthTokenResponse{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		err := &failures.AuthenticationError{Reason: "invalid token"}
		return ports.CheckAuthTokenResponse{}, err
	}
	userID := claims["userId"].(string)
	expiresAtStr := claims["expiresAt"].(string)
	expiresAt, err := time.Parse(time.RFC822, expiresAtStr)
	if err != nil {
		err := &failures.AuthenticationError{Reason: "invalid token"}
		return ports.CheckAuthTokenResponse{}, err
	}

	if time.Now().After(expiresAt) {
		err := &failures.AuthenticationError{Reason: "invalid token"}
		return ports.CheckAuthTokenResponse{}, err
	}

	user, err := srv.usersRepo.Get(ctx, userID)
	if err != nil {
		return ports.CheckAuthTokenResponse{}, fmt.Errorf("'usersRepo.Get' failed: %w", err)
	}

	resp := ports.CheckAuthTokenResponse{
		User: user,
	}
	return resp, nil
}
