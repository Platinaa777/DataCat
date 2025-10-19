package dashboard

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
	return s.service.CreateDashboard(r.Context(), entity.Dashboard{Name: input.Name, Description: input.Description, NamespaceID: ns, Tags: input.Tags})
}

func (s *Handler) Get(r *http.Request) (any, error) {
	return transport.Get[entity.Dashboard](s.access, r, "dashboard")
}

func (s *Handler) Full(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	return s.service.DashboardDetails(r.Context(), ns, transport.Param(r, "id"))
}

func (s *Handler) Search(r *http.Request) (any, error) {
	return transport.Search[entity.Dashboard](s.access, r, "dashboard")
}

func (s *Handler) Update(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, true)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[updateRequest](r)
	if err != nil {
		return nil, err
	}
	return nil, s.service.UpdateDashboard(r.Context(), ns, transport.Param(r, "id"), input.Name, input.Description)
}

func (s *Handler) Delete(r *http.Request) (any, error) { return s.access.Remove(r, "dashboard") }
