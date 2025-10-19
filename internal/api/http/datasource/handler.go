package datasource

import (
	"encoding/json"
	"net/http"
	"strings"

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
	ns, err := s.access.NamespaceAdmin(r)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[createRequest](r)
	if err != nil {
		return nil, err
	}
	if err = s.validateConnection(input.ConnectionString); err != nil {
		return nil, err
	}
	var purposeName string
	if json.Unmarshal(input.Purpose, &purposeName) != nil {
		var purposeNumber int
		if json.Unmarshal(input.Purpose, &purposeNumber) != nil {
			return nil, entity.ErrInvalid
		}
		purposeName = map[int]string{1: "Metrics", 2: "Logs", 3: "Traces"}[purposeNumber]
	}
	purpose := map[string]string{"metrics": "Metrics", "logs": "Logs", "traces": "Traces"}[strings.ToLower(purposeName)]
	return s.service.CreateDataSource(r.Context(), ns, entity.DataSource{Name: input.UniqueName, Type: input.DataSourceType, ConnectionString: input.ConnectionString, Purpose: purpose})
}

func (s *Handler) Get(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	d, err := s.service.DataSource(r.Context(), ns, transport.Param(r, "id"))
	if err != nil {
		return nil, err
	}
	if _, err = s.access.NamespaceAdmin(r); err != nil {
		d.ConnectionString = ""
	}
	return d, nil
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
	result, err := transport.Page[entity.DataSource](s.access, r, ns, "data-source", input, page, size)
	if err != nil {
		return nil, err
	}
	for i := range result.Items {
		result.Items[i].ConnectionString = ""
	}
	return result, nil
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
	if err = s.validateConnection(input.ConnectionString); err != nil {
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
