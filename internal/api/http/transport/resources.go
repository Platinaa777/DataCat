package transport

import (
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
	return usecases.SearchResources[T](r.Context(), s.Service, ns, kind, query, page, size)
}

func Get[T any](s *Access, r *http.Request, kind string) (any, error) {
	ns, err := s.Scope(r, false)
	if err != nil {
		return nil, err
	}
	v, err := usecases.Resource[T](r.Context(), s.Service, ns, kind, Param(r, "id"))
	return v, err
}

func (s *Access) Remove(r *http.Request, kind string) (any, error) {
	ns, err := s.Scope(r, true)
	if err != nil {
		return nil, err
	}
	return nil, s.Service.Delete(r.Context(), ns, kind, Param(r, "id"))
}
