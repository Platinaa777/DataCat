package plugin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

type Handler struct {
	service   *usecases.Service
	access    *transport.Access
	responses *transport.Responder
	provider  *telemetry.Provider
}

func New(service *usecases.Service, access *transport.Access, responses *transport.Responder, provider *telemetry.Provider) *Handler {
	return &Handler{service: service, access: access, responses: responses, provider: provider}
}

func (s *Handler) Create(r *http.Request) (any, error) {
	if err := transport.Admin(r); err != nil {
		return nil, err
	}
	input, err := transport.Body[createRequest](r)
	if err != nil {
		return nil, err
	}
	if err = usecases.Required(input.Name, input.Version, input.Author, input.Endpoint); err != nil {
		return nil, err
	}
	if _, err = s.provider.Client.Validate(input.Endpoint); err != nil {
		return nil, fmt.Errorf("%w: %v", entity.ErrInvalid, err)
	}
	p := entity.Plugin{ID: usecases.ID(), Name: input.Name, Version: input.Version, Author: input.Author, Description: input.Description, Settings: input.Settings, Endpoint: input.Endpoint}
	return p.ID, usecases.Create(r.Context(), s.service.Store, usecases.Global, "plugin", p.ID, p.Name, p)
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
	p, _, err := usecases.Read[entity.Plugin](r.Context(), s.service.Store, usecases.Global, "plugin", transport.Param(r, "id"))
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
