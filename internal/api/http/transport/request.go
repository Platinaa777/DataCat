package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/go-chi/chi/v5"
)

func Body[T any](r *http.Request) (T, error) {
	var value T
	err := DecodeInto(r, &value)
	return value, err
}

func Param(r *http.Request, key string) string { return chi.URLParam(r, key) }

func Pagination(r *http.Request) (int, int, error) {
	page, size := 1, 10
	var err error
	if v := r.URL.Query().Get("page"); v != "" {
		page, err = strconv.Atoi(v)
		if err != nil {
			return 0, 0, entity.ErrInvalid
		}
	}
	if v := r.URL.Query().Get("pageSize"); v != "" {
		size, err = strconv.Atoi(v)
		if err != nil {
			return 0, 0, entity.ErrInvalid
		}
	}
	return page, size, nil
}

func QueryValue(r *http.Request, key string) string {
	for k, vs := range r.URL.Query() {
		if strings.EqualFold(k, key) && len(vs) > 0 {
			return vs[0]
		}
	}
	return ""
}

func DecodeInto(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%w: invalid JSON body", entity.ErrInvalid)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("%w: only one JSON value is allowed", entity.ErrInvalid)
	}
	return nil
}
