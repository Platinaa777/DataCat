package variable

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

type Handler struct {
	service   *usecases.Service
	access    *transport.Access
	responses *transport.Responder
}

func New(service *usecases.Service, access *transport.Access, responses *transport.Responder) *Handler {
	return &Handler{service: service, access: access, responses: responses}
}

func (s *Handler) Create(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, true)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[createRequest](r)
	if err != nil {
		return nil, err
	}
	if input.NamespaceID != "" && input.NamespaceID != ns {
		return nil, entity.ErrForbidden
	}
	return s.service.CreateVariable(r.Context(), entity.Variable{Placeholder: input.Placeholder, Value: input.Value, NamespaceID: ns, DashboardID: input.DashboardID})
}

func (s *Handler) Get(r *http.Request) (any, error) {
	return transport.Get[entity.Variable](s.access, r, "variable")
}

func (s *Handler) ByDashboard(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	all, err := usecases.List[entity.Variable](r.Context(), s.service.Store, ns, "variable")
	out := []entity.Variable{}
	for _, v := range all {
		if v.DashboardID != nil && *v.DashboardID == transport.Param(r, "id") {
			out = append(out, v)
		}
	}
	return out, err
}

func (s *Handler) ByNamespace(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	if transport.Param(r, "id") != ns {
		return nil, entity.ErrForbidden
	}
	return usecases.List[entity.Variable](r.Context(), s.service.Store, ns, "variable")
}

func (s *Handler) Delete(r *http.Request) (any, error) { return s.access.Remove(r, "variable") }

func (s *Handler) Update(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, true)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[updateRequest](r)
	if err != nil {
		return nil, err
	}
	return nil, s.service.UpdateVariable(r.Context(), ns, transport.Param(r, "id"), input.Placeholder, input.Value)
}
