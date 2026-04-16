package web

type App struct {
	TrustedOrigins   []string
	TokenAuthHandler TokenAuthFunc
}
