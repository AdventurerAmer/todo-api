package domain

type Template = string

const (
	UserActivationTemplate Template = "UserActivation"
)

type UserActivationTemplateData struct {
	Code string `json:"code"`
}
