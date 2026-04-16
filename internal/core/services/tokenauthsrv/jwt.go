package tokenauthsrv

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
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

type CustomCalims struct {
	jwt.RegisteredClaims
	UserID string `json:"userID"`
}

func (calims *CustomCalims) Valid() error {
	if err := calims.RegisteredClaims.Valid(); err != nil {
		return err
	}
	if calims.UserID == "" {
		return fmt.Errorf("'userID' is empty")
	}
	return nil
}

func (srv *jwtService) Create(ctx context.Context, req ports.CreateAuthTokenRequest) (ports.CreateAuthTokenResponse, error) {
	v := failures.NewValidator()
	v.CheckUTF8Email(req.Email)
	v.CheckNotEmpty("passward", req.Password)
	v.CheckUTF8("passward", req.Password)
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

	signedToken, expiresAt, err := srv.signToken(user)
	if err != nil {
		err := fmt.Errorf("'signToken' failed: %w", err)
		return ports.CreateAuthTokenResponse{}, err
	}

	resp := ports.CreateAuthTokenResponse{
		Token:     signedToken,
		ExpiresAt: expiresAt,
	}
	return resp, nil
}

func (srv *jwtService) Refresh(ctx context.Context, user domain.User) (ports.RefreshAuthTokenResponse, error) {
	token, expiresAt, err := srv.signToken(user)
	if err != nil {
		err := fmt.Errorf("'signToken' failed: %w", err)
		return ports.RefreshAuthTokenResponse{}, err
	}

	resp := ports.RefreshAuthTokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
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

	keyFunc := func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(srv.Secret), nil
	}

	token, err := jwt.ParseWithClaims(req.Token, &CustomCalims{}, keyFunc)
	if err != nil {
		err := &failures.AuthenticationError{Reason: "invalid token"}
		return ports.CheckAuthTokenResponse{}, err
	}

	claims, ok := token.Claims.(*CustomCalims)
	if !token.Valid || !ok {
		err := &failures.AuthenticationError{Reason: "invalid token"}
		return ports.CheckAuthTokenResponse{}, err
	}
	userID := claims.UserID
	user, err := srv.usersRepo.Get(ctx, userID)
	if err != nil {
		return ports.CheckAuthTokenResponse{}, fmt.Errorf("'usersRepo.Get' failed: %w", err)
	}

	resp := ports.CheckAuthTokenResponse{
		User: user,
	}
	return resp, nil
}

func (srv *jwtService) signToken(user domain.User) (string, time.Time, error) {
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(srv.TokenExpiresAfter)
	calims := &CustomCalims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID: user.ID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, calims)
	signedToken, err := token.SignedString([]byte(srv.Secret))
	if err != nil {
		err := fmt.Errorf("'token.SignedString' failed: %w", err)
		return "", time.Time{}, err
	}
	return signedToken, expiresAt, nil
}
