package transport

import (
	"fmt"
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/google/uuid"
)

type Access struct{ Service *usecases.Service }

func (s *Access) Scope(r *http.Request, write bool) (string, error) {
	ns := r.Header.Get("X-Namespace")
	if _, err := uuid.Parse(ns); err != nil {
		return "", fmt.Errorf("%w: X-Namespace must be a UUID", entity.ErrInvalid)
	}
	if !auth.Principal(r.Context()).Can(ns, write) {
		return "", entity.ErrForbidden
	}
	if _, err := s.Service.Store.Get(r.Context(), usecases.Global, "namespace", ns); err != nil {
		return "", err
	}
	return ns, nil
}

func Admin(r *http.Request) error {
	if !auth.Principal(r.Context()).GlobalAdmin {
		return entity.ErrForbidden
	}
	return nil
}

func (s *Access) NamespaceAdmin(r *http.Request) (string, error) {
	ns, err := s.Scope(r, true)
	if err != nil {
		return "", err
	}
	p := auth.Principal(r.Context())
	if p.GlobalAdmin {
		return ns, nil
	}
	for _, c := range p.Claims {
		if c.NamespaceID == ns && c.Role == "Admin" {
			return ns, nil
		}
	}
	return "", entity.ErrForbidden
}
