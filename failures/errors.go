package failures

import (
	"fmt"
	"reflect"
)

type ResourceNotFoundError struct {
	Name string `json:"name"`
}

func (r *ResourceNotFoundError) Error() string {
	return fmt.Sprintf("%s not found", r.Name)
}

func (r *ResourceNotFoundError) Is(target error) bool {
	t, ok := target.(*ResourceNotFoundError)
	if !ok {
		return false
	}
	return r.Name == t.Name
}

type ResourceAlreadyExistsError struct {
	Name string `json:"name"`
}

func (r *ResourceAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists", r.Name)
}

func (r *ResourceAlreadyExistsError) Is(target error) bool {
	t, ok := target.(*ResourceAlreadyExistsError)
	if !ok {
		return false
	}
	return r.Name == t.Name
}

type ValidationError struct {
	Reason string `json:"reason"`
}

func (v *ValidationError) Error() string {
	return v.Reason
}

type ValidationsError struct {
	Errors map[string]string `json:"errors"`
}

func (v *ValidationsError) Error() string {
	return fmt.Sprintf("%+v", v.Errors)
}

func (r *ValidationsError) Is(target error) bool {
	t, ok := target.(*ValidationsError)
	if !ok {
		return false
	}
	return reflect.DeepEqual(r.Errors, t.Errors)
}

type AuthenticationError struct {
	Reason string `json:"reason"`
}

func (v *AuthenticationError) Error() string {
	return v.Reason
}

func (v *AuthenticationError) Is(target error) bool {
	t, ok := target.(*AuthenticationError)
	if !ok {
		return false
	}
	return t.Reason == v.Reason
}

type AuthorizationError struct {
	Reason string `json:"reason"`
}

func (v *AuthorizationError) Error() string {
	return v.Reason
}

func (v *AuthorizationError) Is(target error) bool {
	t, ok := target.(*AuthorizationError)
	if !ok {
		return false
	}
	return t.Reason == v.Reason
}
