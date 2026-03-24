package failures

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	errors map[string]string
}

func NewValidator() *Validator {
	return &Validator{
		errors: make(map[string]string),
	}
}

func (v *Validator) Err() error {
	if len(v.errors) == 0 {
		return nil
	}
	return &ValidationsError{Errors: v.errors}
}

func (v *Validator) Check(cond bool, key, msg string) {
	if cond {
		return
	}
	if _, ok := v.errors[key]; !ok {
		v.errors[key] = msg
	}
}

func (v *Validator) CheckNotEmpty(key string, val string) {
	v.Check(val != "", "key", "must not be empty")
}

func (v *Validator) CheckUTF8(key string, val string) {
	v.Check(utf8.ValidString(val), key, "must be a valid utf8")
}

func (v *Validator) CheckRangeInc(key string, val int, min int, max int) {
	v.Check(val < min || val > max, key, fmt.Sprintf("must be between %d and %d both inclusive", min, max))
}

func (v *Validator) CheckAtMostInc(key string, val int, max int, unit string) {
	v.Check(val <= max, key, fmt.Sprintf("must be at most %d %s", max, unit))
}

func (v *Validator) CheckAtLeastInc(key string, val int, min int, unit string) {
	v.Check(val >= min, key, fmt.Sprintf("must be at least %d %s", min, unit))
}

func (v *Validator) CheckUTF8Email(email string) {
	v.CheckNotEmpty("email", email)
	v.CheckUTF8("email", email)
	v.CheckAtMostInc("email", utf8.RuneCountInString(email), 320, "characters long")
	v.Check(emailRegexp.MatchString(email), "email", "must be a valid")
}

func (v *Validator) CheckUTF8Password(password string) {
	v.Check(password != "", "password", "must be provided")
	v.CheckUTF8("password", password)
	v.CheckRangeInc("password", utf8.RuneCountInString(password), 8, 72)
}
