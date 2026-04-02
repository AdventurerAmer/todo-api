package web

type AuthFunc = func(token string) (userID string, err error)

type App struct {
	TrustedOrigins []string
	AuthHandler    AuthFunc
}
