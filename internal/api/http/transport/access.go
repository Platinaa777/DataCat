package transport

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

type Access struct{ Service *usecases.Service }

func (s *Access) Scope(r *http.Request, write bool) (string, error) {
	ns := r.Header.Get("X-Namespace")
	return ns, s.Service.AuthorizeNamespace(r.Context(), auth.Principal(r.Context()), ns, write)
}

func Admin(r *http.Request) error {
	if !auth.Principal(r.Context()).GlobalAdmin {
		return entity.ErrForbidden
	}
	return nil
}

func (s *Access) NamespaceAdmin(r *http.Request) (string, error) {
	ns := r.Header.Get("X-Namespace")
	return ns, s.Service.AuthorizeNamespaceAdmin(r.Context(), auth.Principal(r.Context()), ns)
}
