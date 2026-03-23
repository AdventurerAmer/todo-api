package utils

import (
	"encoding/json"
	"errors"
	"regexp"
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

func (v *Validator) ToError() error {
	if v == nil {
		return errors.New("")
	}
	data, err := json.Marshal(v.errors)
	if err != nil {
		return err
	}
	return errors.New(string(data))
}

func (v *Validator) HasErrors() bool {
	return len(v.errors) != 0
}

func (v *Validator) CheckCond(cond bool, key, msg string) {
	if cond {
		return
	}
	if _, ok := v.errors[key]; !ok {
		v.errors[key] = msg
	}
}

func (v *Validator) CheckEmail(email string) {
	v.CheckCond(email != "", "email", "must be provided")
	v.CheckCond(emailRegexp.Match([]byte(email)), "email", "must be a valid email address")
}

func (v *Validator) CheckPassword(password string) {
	v.CheckCond(password != "", "password", "must be provided")
	v.CheckCond(len(password) >= 8, "password", "must be atleast 8 characters long")
	v.CheckCond(len(password) <= 72, "password", "must be atmost 72 characters long")
}
