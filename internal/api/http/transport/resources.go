package transport

import (
	"encoding/json"
	"net/http"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

func Search[T any](s *Access, r *http.Request, kind string) (any, error) {
	ns, err := s.Scope(r, false)
	if err != nil {
		return nil, err
	}
	query, err := Body[entity.Search](r)
	if err != nil {
		return nil, err
	}
	page, size, err := Pagination(r)
	if err != nil {
		return nil, err
	}
	return Page[T](s, r, ns, kind, query, page, size)
}

func Page[T any](s *Access, r *http.Request, ns, kind string, query entity.Search, page, size int) (entity.Page[T], error) {
	records, total, err := s.Service.Store.Search(r.Context(), ns, kind, query, page, size)
	if err != nil {
		return entity.Page[T]{}, err
	}
	out := make([]T, 0, len(records))
	for _, record := range records {
		var value T
		if err = json.Unmarshal(record.Data, &value); err != nil {
			return entity.Page[T]{}, err
		}
		out = append(out, value)
	}
	return entity.NewPage(out, total, page, size), nil
}

func Get[T any](s *Access, r *http.Request, kind string) (any, error) {
	ns, err := s.Scope(r, false)
	if err != nil {
		return nil, err
	}
	v, _, err := usecases.Read[T](r.Context(), s.Service.Store, ns, kind, Param(r, "id"))
	return v, err
}

func (s *Access) Remove(r *http.Request, kind string) (any, error) {
	ns, err := s.Scope(r, true)
	if err != nil {
		return nil, err
	}
	return nil, s.Service.Delete(r.Context(), ns, kind, Param(r, "id"))
}
