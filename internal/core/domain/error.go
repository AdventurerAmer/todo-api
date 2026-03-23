package domain

import "fmt"

type ResourceNotFoundError struct {
	Name string
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
	Name string
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
