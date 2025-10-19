package namespace

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

type Handler struct {
	service   *usecases.Service
	responses *transport.Responder
}

func New(service *usecases.Service, responses *transport.Responder) *Handler {
	return &Handler{service: service, responses: responses}
}

func (s *Handler) Create(r *http.Request) (any, error) {
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	input, err := transport.Body[createRequest](r)
	if err != nil {
		return nil, err
	}
	return s.service.CreateNamespace(r.Context(), input.Name)
}

func (s *Handler) Available(r *http.Request) (any, error) {
	all, err := usecases.List[entity.Namespace](r.Context(), s.service.Store, usecases.Global, "namespace")
	if err != nil {
		return nil, err
	}
	out := []entity.Namespace{}
	for _, ns := range all {
		if auth.Principal(r.Context()).Can(ns.ID, false) {
			out = append(out, ns)
		}
	}
	return out, nil
}

func (s *Handler) Get(r *http.Request) (any, error) {
	ns, _, err := usecases.Read[entity.Namespace](r.Context(), s.service.Store, usecases.Global, "namespace", transport.Param(r, "name"))
	if err != nil {
		return nil, err
	}
	if !auth.Principal(r.Context()).Can(ns.ID, false) {
		return nil, entity.ErrForbidden
	}
	dashboards, err := usecases.List[entity.Dashboard](r.Context(), s.service.Store, ns.ID, "dashboard")
	return map[string]any{"name": ns.Name, "dashboards": dashboards}, err
}

func (s *Handler) MapRole(r *http.Request) (any, error) {
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	input, err := transport.Body[entity.RoleMapping](r)
	if err != nil {
		return nil, err
	}
	return nil, s.service.MapRole(r.Context(), input)
}
