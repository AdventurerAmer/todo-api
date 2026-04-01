package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/AdventurerAmer/todo-api/failures"
)

const InternalServerError = `{"error": "internal server error"}`

func ReadJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var (
			synatxErr        *json.SyntaxError
			unmarshalTypeErr *json.UnmarshalTypeError
		)

		switch {
		case errors.Is(err, io.ErrUnexpectedEOF):
			err = &failures.ValidationError{Reason: "body contains malformed JSON"}
		case errors.Is(err, io.EOF):
			err = &failures.ValidationError{Reason: "body is empty"}
		case errors.As(err, &synatxErr):
			err = &failures.ValidationError{Reason: fmt.Sprintf("body contains malformed JSON at character %d", synatxErr.Offset)}
		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				err = &failures.ValidationError{Reason: fmt.Sprintf("body contains incorrect JSON type for field %q", unmarshalTypeErr.Field)}
			} else {
				err = &failures.ValidationError{
					Reason: fmt.Sprintf("body contains malformed JSON at character %d", unmarshalTypeErr.Offset),
				}
			}
		}

		return fmt.Errorf("'dec.Decode' failed: %w", err)
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, resp any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		slog.Error("response body write failed", "error", err)
		if _, err := w.Write([]byte(InternalServerError)); err != nil {
			slog.Error("response body write failed", "error", err)
		}
		return
	}

	if _, err := w.Write(buf.Bytes()); err != nil {
		slog.Error("response body write failed", "error", err)
	}
}

func WriteError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	var (
		notFoundErr       *failures.ResourceNotFoundError
		alreadyExistsErr  *failures.ResourceAlreadyExistsError
		validationErr     *failures.ValidationError
		validationsErr    *failures.ValidationsError
		authenticationErr *failures.AuthenticationError
		authorizationErr  *failures.AuthorizationError
	)
	statusCode := http.StatusInternalServerError
	resp := struct {
		TraceID string `json:"traceID"`
		Error   any    `json:"error,omitempty"`
		Errors  any    `json:"errors,omitempty"`
	}{}
	switch {
	case errors.As(err, &notFoundErr):
		resp.Error = err.Error()
		statusCode = http.StatusNotFound
	case errors.As(err, &alreadyExistsErr):
		resp.Error = err.Error()
		statusCode = http.StatusConflict
	case errors.As(err, &validationErr):
		resp.Error = err.Error()
		statusCode = http.StatusBadRequest
	case errors.As(err, &validationsErr):
		resp.Errors = validationsErr.Errors
		statusCode = http.StatusBadRequest
	case errors.As(err, &authenticationErr):
		resp.Error = err.Error()
		statusCode = http.StatusUnauthorized
	case errors.As(err, &authorizationErr):
		resp.Error = err.Error()
		statusCode = http.StatusForbidden
	default:
		resp.Error = "internal server error"
	}
	WriteJSON(w, resp, statusCode)
}
