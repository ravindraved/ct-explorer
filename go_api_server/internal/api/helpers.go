package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// WriteJSON marshals data to JSON and writes it with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes a {"detail": message} error response.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"detail": message})
}

// BindQuery decodes URL query parameters into a struct using `query` struct tags.
// Supports string, int, and []string (comma-separated) field types.
// Runs go-playground/validator if `validate` tags are present.
func BindQuery(r *http.Request, dst any) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dst must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()
	q := r.URL.Query()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("query")
		if tag == "" {
			continue
		}

		raw := q.Get(tag)
		if raw == "" {
			continue
		}

		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}

		switch fv.Kind() {
		case reflect.String:
			fv.SetString(raw)
		case reflect.Int, reflect.Int64:
			n, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer for %s: %w", tag, err)
			}
			fv.SetInt(n)
		case reflect.Slice:
			if fv.Type().Elem().Kind() == reflect.String {
				parts := strings.Split(raw, ",")
				trimmed := make([]string, 0, len(parts))
				for _, p := range parts {
					s := strings.TrimSpace(p)
					if s != "" {
						trimmed = append(trimmed, s)
					}
				}
				fv.Set(reflect.ValueOf(trimmed))
			}
		}
	}

	if err := validate.Struct(dst); err != nil {
		return err
	}

	return nil
}
