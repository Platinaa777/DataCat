package httpapi

import (
	"net/http"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/go-chi/chi/v5"
)

func (s *Server) alerts(r chi.Router) {
	s.route(r, "POST", "/alert/add", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, true)
		if err != nil {
			return nil, err
		}
		input, err := body[struct {
			Description, Template, RawQuery, DataSourceID, NotificationChannelGroupName string
			WaitTimeBeforeAlerting, RepeatInterval                                      entity.Duration
			Tags                                                                        []string
		}](r)
		if err != nil {
			return nil, err
		}
		return s.Service.CreateAlert(r.Context(), entity.Alert{NamespaceID: ns, Description: input.Description, Template: input.Template, RawQuery: input.RawQuery, WaitTimeBeforeAlerting: input.WaitTimeBeforeAlerting, RepeatInterval: input.RepeatInterval, Tags: input.Tags}, input.DataSourceID, input.NotificationChannelGroupName)
	})
	s.route(r, "GET", "/alert/get-counters", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, false)
		if err != nil {
			return nil, err
		}
		all, err := usecases.List[entity.Alert](r.Context(), s.Service.Store, ns, "alert")
		if err != nil {
			return nil, err
		}
		counts := map[string]int{"Ok": 0, "Firing": 0, "Error": 0, "Muted": 0}
		for _, a := range all {
			counts[a.Status]++
		}
		out := []map[string]any{}
		for _, status := range []string{"Ok", "Firing", "Error", "Muted"} {
			out = append(out, map[string]any{"status": status, "count": counts[status]})
		}
		return out, nil
	})
	s.route(r, "GET", "/alert/template-vars", func(r *http.Request) (any, error) { return usecases.TemplateVariables, nil })
	s.route(r, "GET", "/alert/{id}", func(r *http.Request) (any, error) { return get[entity.Alert](s, r, "alert") })
	s.route(r, "POST", "/alert/search", func(r *http.Request) (any, error) { return search[entity.Alert](s, r, "alert") })
	s.route(r, "DELETE", "/alert/remove/{id}", func(r *http.Request) (any, error) { return s.remove(r, "alert") })
	s.route(r, "PUT", "/alert/update/{id}", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, true)
		if err != nil {
			return nil, err
		}
		input, err := body[struct{ Description, Template, RawQuery, DataSourceID string }](r)
		if err != nil {
			return nil, err
		}
		if err = usecases.Required(input.Template, input.RawQuery, input.DataSourceID); err != nil {
			return nil, err
		}
		a, rec, err := usecases.Read[entity.Alert](r.Context(), s.Service.Store, ns, "alert", param(r, "id"))
		if err != nil {
			return nil, err
		}
		d, _, err := usecases.Read[entity.DataSource](r.Context(), s.Service.Store, ns, "data-source", input.DataSourceID)
		if err != nil {
			return nil, err
		}
		if d.Purpose != "Metrics" {
			return nil, entity.ErrInvalid
		}
		d.ConnectionString = ""
		a.Description = input.Description
		a.Template = input.Template
		a.RawQuery = input.RawQuery
		a.DataSource = d
		a.PendingSince = nil
		a.LastNotifiedAt = nil
		return nil, usecases.Update(r.Context(), s.Service.Store, rec, a)
	})
	s.route(r, "POST", "/alert/mute/{id}", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, true)
		if err != nil {
			return nil, err
		}
		duration, err := entity.ParseDuration(r.URL.Query().Get("nextExecutionTime"))
		if err != nil || duration <= 0 {
			return nil, entity.ErrInvalid
		}
		a, rec, err := usecases.Read[entity.Alert](r.Context(), s.Service.Store, ns, "alert", param(r, "id"))
		if err != nil {
			return nil, err
		}
		a.Status = "Muted"
		a.NextExecutionTime = s.Service.Now().Add(time.Duration(duration))
		a.PendingSince = nil
		return a.ID, usecases.Update(r.Context(), s.Service.Store, rec, a)
	})
}
