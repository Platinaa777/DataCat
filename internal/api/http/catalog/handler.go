package catalog

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

type Handler struct {
	service   *usecases.Service
	responses *transport.Responder
	kind      string
}

func New(service *usecases.Service, responses *transport.Responder, kind string) *Handler {
	return &Handler{service: service, responses: responses, kind: kind}
}

func (s *Handler) Create(r *http.Request) (any, error) {
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	input, err := transport.Body[createRequest](r)
	if err != nil {
		return nil, err
	}
	return s.service.CreateNamedType(r.Context(), s.kind, input.Name)
}

func (s *Handler) List(r *http.Request) (any, error) {
	return s.service.NamedTypes(r.Context(), s.kind)
}

func (s *Handler) Get(r *http.Request) (any, error) {
	return s.service.NamedType(r.Context(), s.kind, transport.Param(r, "name"))
}

func (s *Handler) Delete(r *http.Request) (any, error) {
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	return nil, s.service.DeleteNamedType(r.Context(), s.kind, transport.Param(r, "name"))
}
