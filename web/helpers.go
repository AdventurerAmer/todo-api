package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/microcosm-cc/bluemonday"

	"golang.org/x/text/secure/precis"
)

const InternalServerError = `{"error": "internal server error"}`

func ReadJSON(r *http.Request, v any) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return &failures.UnsupportedMediaTypeError{Type: r.Header.Get("Content-Type")}
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var (
			synatxErr        *json.SyntaxError
			unmarshalTypeErr *json.UnmarshalTypeError
			maxBytesErr      *http.MaxBytesError
		)

		switch {
		case errors.Is(err, io.ErrUnexpectedEOF):
			err = &failures.ValidationError{Reason: "body contains malformed JSON"}
		case errors.Is(err, io.EOF):
			err = &failures.ValidationError{Reason: "body is empty"}
		case errors.As(err, &synatxErr):
			err = &failures.ValidationError{Reason: fmt.Sprintf("body contains malformed JSON at character %d", synatxErr.Offset)}
		case errors.As(err, &maxBytesErr):
			err = &failures.ValidationError{Reason: "body is too large"}
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

	if dec.More() {
		return &failures.ValidationError{Reason: "request body must contain only one JSON object"}
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
		notFoundErr             *failures.ResourceNotFoundError
		alreadyExistsErr        *failures.ResourceAlreadyExistsError
		validationErr           *failures.ValidationError
		validationsErr          *failures.ValidationsError
		authenticationErr       *failures.AuthenticationError
		authorizationErr        *failures.AuthorizationError
		unsupportedMediaTypeErr *failures.UnsupportedMediaTypeError
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
	case errors.As(err, &unsupportedMediaTypeErr):
		resp.Error = err.Error()
		statusCode = http.StatusUnsupportedMediaType
	default:
		resp.Error = "internal server error"
	}
	if statusCode == http.StatusInternalServerError {
		slog.Error("internal server error", "error", err)
	}
	WriteJSON(w, resp, statusCode)
}

var policy = bluemonday.StrictPolicy()

func canonicalize(input string) (string, error) {
	var (
		err   error
		prev  string
		count int
	)

	for prev != input {
		if count > 10 {
			return "", fmt.Errorf("too many escape layers")
		}
		prev = input
		if input, err = url.QueryUnescape(input); err != nil {
			return "", err
		}
	}

	return input, nil
}

func Path(r *http.Request, v *failures.Validator, key string) string {
	var err error
	val := r.PathValue(key)
	if !utf8.ValidString(val) {
		v.Check(false, key, "invalid utf-8 value")
		return ""
	}
	if val, err = canonicalize(val); err != nil {
		v.Check(false, key, err.Error())
		return ""
	}
	if val, err = precis.UsernameCasePreserved.String(val); err != nil {
		v.Check(false, key, err.Error())
		return ""
	}
	policy.Sanitize(val)
	return val
}

func Query(r *http.Request, v *failures.Validator, key string, defaultVal string) string {
	if !r.URL.Query().Has(key) {
		return defaultVal
	}
	val := r.URL.Query().Get(key)
	if !utf8.ValidString(val) {
		v.Check(false, key, "invalid utf-8 value")
		return ""
	}
	var err error
	if val, err = canonicalize(val); err != nil {
		v.Check(false, key, err.Error())
		return ""
	}
	if val, err = precis.UsernameCasePreserved.String(val); err != nil {
		v.Check(false, key, err.Error())
		return ""
	}
	return policy.Sanitize(val)
}

func QueryInt(r *http.Request, v *failures.Validator, key string, defaultVal int) int {
	if !r.URL.Query().Has(key) {
		return defaultVal
	}
	val := r.URL.Query().Get(key)
	n, err := strconv.Atoi(val)
	if err != nil {
		v.Check(false, key, err.Error())
		return 0
	}
	return n
}

func QueryBool(r *http.Request, v *failures.Validator, key string) *bool {
	if !r.URL.Query().Has(key) {
		return nil
	}
	val := strings.ToLower(r.URL.Query().Get(key))
	switch val {
	case "true":
		t := true
		return &t
	case "false":
		t := false
		return &t
	}
	v.Check(false, key, "invalid bool value")
	return nil
}
