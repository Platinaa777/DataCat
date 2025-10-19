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
	return s.service.AvailableNamespaces(r.Context(), auth.Principal(r.Context()))
}

func (s *Handler) Get(r *http.Request) (any, error) {
	return s.service.NamespaceDetails(r.Context(), auth.Principal(r.Context()), transport.Param(r, "name"))
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
