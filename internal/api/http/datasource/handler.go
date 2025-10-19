package datasource

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/auth"
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
	ns, err := s.access.NamespaceAdmin(r)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[createRequest](r)
	if err != nil {
		return nil, err
	}
	return s.service.CreateDataSource(r.Context(), ns, entity.DataSource{Name: input.UniqueName, Type: input.DataSourceType, ConnectionString: input.ConnectionString, Purpose: string(input.Purpose)})
}

func (s *Handler) Get(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	return s.service.DataSourceFor(r.Context(), ns, transport.Param(r, "id"), auth.Principal(r.Context()))
}

func (s *Handler) Search(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[entity.Search](r)
	if err != nil {
		return nil, err
	}
	page, size, err := transport.Pagination(r)
	if err != nil {
		return nil, err
	}
	return s.service.SearchDataSources(r.Context(), ns, input, page, size)
}

func (s *Handler) UpdateConnection(r *http.Request) (any, error) {
	ns, err := s.access.NamespaceAdmin(r)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[updateConnectionRequest](r)
	if err != nil {
		return nil, err
	}
	return nil, s.service.UpdateDataSourceConnection(r.Context(), ns, transport.Param(r, "id"), input.ConnectionString)
}

func (s *Handler) Delete(r *http.Request) (any, error) {
	if _, err := s.access.NamespaceAdmin(r); err != nil {
		return nil, err
	}
	return s.access.Remove(r, "data-source")
}
