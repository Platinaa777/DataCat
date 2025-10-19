package panel

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

func (s *Handler) Types(r *http.Request) (any, error) {
	return []entity.NamedType{{ID: 1, Name: "Graph"}, {ID: 2, Name: "Table"}, {ID: 3, Name: "Pie Chart"}}, nil
}

func (s *Handler) Create(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, true)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[panelInput](r)
	if err != nil {
		return nil, err
	}
	return s.service.CreatePanel(r.Context(), entity.Panel{NamespaceID: ns, Title: input.Title, TypeName: panelName(input.Type), Query: entity.RawQuery{Query: input.RawQuery}, Layout: input.Layout, DashboardID: input.DashboardID, StyleConfiguration: input.StyleConfiguration}, input.DataSourceID)
}

func (s *Handler) Get(r *http.Request) (any, error) {
	return transport.Get[entity.Panel](s.access, r, "panel")
}

func (s *Handler) Delete(r *http.Request) (any, error) { return s.access.Remove(r, "panel") }

func (s *Handler) Update(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, true)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[panelInput](r)
	if err != nil {
		return nil, err
	}
	if err = usecases.Required(input.Title, input.RawQuery, input.Layout, panelName(input.Type)); err != nil {
		return nil, err
	}
	return nil, s.service.UpdatePanel(r.Context(), ns, transport.Param(r, "id"), input.Title, panelName(input.Type), input.RawQuery, input.DataSourceID, input.Layout, input.StyleConfiguration)
}
