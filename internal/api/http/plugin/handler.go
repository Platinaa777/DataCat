package plugin

import (
	"net/http"
	"strconv"

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
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	input, err := transport.Body[createRequest](r)
	if err != nil {
		return nil, err
	}
	return s.service.CreatePlugin(r.Context(), entity.Plugin{Name: input.Name, Version: input.Version, Author: input.Author, Description: input.Description, Settings: input.Settings, Endpoint: input.Endpoint})
}

func (s *Handler) Search(r *http.Request) (any, error) {
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	page, size, err := transport.Pagination(r)
	if err != nil {
		return nil, err
	}
	return transport.Page[entity.Plugin](s.access, r, usecases.Global, "plugin", entity.Search{}, page, size)
}

func (s *Handler) Get(r *http.Request) (any, error) {
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	p, err := usecases.Resource[entity.Plugin](r.Context(), s.service, usecases.Global, "plugin", transport.Param(r, "id"))
	return p, err
}

func (s *Handler) Delete(r *http.Request) (any, error) {
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	return nil, s.service.Delete(r.Context(), usecases.Global, "plugin", transport.Param(r, "id"))
}

func (s *Handler) SetActive(r *http.Request) (any, error) {
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	active, err := strconv.ParseBool(r.URL.Query().Get("isActive"))
	if err != nil {
		return nil, entity.ErrInvalid
	}
	return nil, s.service.SetPluginActive(r.Context(), transport.Param(r, "id"), active)
}

func (s *Handler) Update(r *http.Request) (any, error) {
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	input, err := transport.Body[updateRequest](r)
	if err != nil {
		return nil, err
	}
	return nil, s.service.UpdatePlugin(r.Context(), transport.Param(r, "id"), input.Description, input.Settings)
}
