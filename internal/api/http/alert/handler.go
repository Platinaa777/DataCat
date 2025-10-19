package alert

import (
	"net/http"
	"time"

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
	return s.service.CreateAlert(r.Context(), entity.Alert{NamespaceID: ns, Description: input.Description, Template: input.Template, RawQuery: input.RawQuery, WaitTimeBeforeAlerting: input.WaitTimeBeforeAlerting, RepeatInterval: input.RepeatInterval, Tags: input.Tags}, input.DataSourceID, input.NotificationChannelGroupName)
}

func (s *Handler) Counters(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	return s.service.AlertCounters(r.Context(), ns)
}

func (s *Handler) TemplateVariables(r *http.Request) (any, error) {
	return usecases.TemplateVariables, nil
}

func (s *Handler) Get(r *http.Request) (any, error) {
	return transport.Get[entity.Alert](s.access, r, "alert")
}

func (s *Handler) Search(r *http.Request) (any, error) {
	return transport.Search[entity.Alert](s.access, r, "alert")
}

func (s *Handler) Delete(r *http.Request) (any, error) { return s.access.Remove(r, "alert") }

func (s *Handler) Update(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, true)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[updateRequest](r)
	if err != nil {
		return nil, err
	}
	return nil, s.service.UpdateAlert(r.Context(), ns, transport.Param(r, "id"), input.Description, input.Template, input.RawQuery, input.DataSourceID)
}

func (s *Handler) Mute(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, true)
	if err != nil {
		return nil, err
	}
	duration, err := entity.ParseDuration(r.URL.Query().Get("nextExecutionTime"))
	if err != nil || duration <= 0 {
		return nil, entity.ErrInvalid
	}
	return s.service.MuteAlert(r.Context(), ns, transport.Param(r, "id"), time.Duration(duration))
}
