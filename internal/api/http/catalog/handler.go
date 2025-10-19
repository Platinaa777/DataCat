package catalog

import (
	"net/http"
	"strconv"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/entity"
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
	if err = usecases.Required(input.Name); err != nil {
		return nil, err
	}
	id, err := s.service.Store.NextID(r.Context())
	if err != nil {
		return nil, err
	}
	return id, usecases.Create(r.Context(), s.service.Store, usecases.Global, s.kind, strconv.FormatInt(id, 10), input.Name, entity.NamedType{ID: id, Name: input.Name})
}

func (s *Handler) List(r *http.Request) (any, error) {
	return usecases.List[entity.NamedType](r.Context(), s.service.Store, usecases.Global, s.kind)
}

func (s *Handler) Get(r *http.Request) (any, error) {
	v, _, err := usecases.Read[entity.NamedType](r.Context(), s.service.Store, usecases.Global, s.kind, transport.Param(r, "name"))
	return v, err
}

func (s *Handler) Delete(r *http.Request) (any, error) {
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	_, rec, err := usecases.Read[entity.NamedType](r.Context(), s.service.Store, usecases.Global, s.kind, transport.Param(r, "name"))
	if err != nil {
		return nil, err
	}
	return nil, s.service.Delete(r.Context(), usecases.Global, s.kind, rec.ID)
}
