package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/go-chi/chi/v5"
)

func (s *Server) integrations(r chi.Router) {
	s.route(r, "GET", "/metrics/query", func(r *http.Request) (any, error) { return s.cachedMetrics(r, false) })
	s.route(r, "GET", "/metrics/query-range", func(r *http.Request) (any, error) { return s.cachedMetrics(r, true) })
	s.route(r, "POST", "/logs/search", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, false)
		if err != nil {
			return nil, err
		}
		input := entity.LogSearch{Page: 1, PageSize: 100}
		if err = decodeInto(r, &input); err != nil {
			return nil, err
		}
		d, err := s.source(r, ns, input.DataSourceName, "Logs")
		if err != nil {
			return nil, err
		}
		if !strings.EqualFold(d.Type, "Elasticsearch") {
			return s.plugin(r, d, "logs.search", input)
		}
		out, err := s.Telemetry.Logs(r.Context(), d, input)
		return out, upstream(err)
	})
	s.route(r, "GET", "/traces/services", func(r *http.Request) (any, error) { return s.traceNames(r, "") })
	s.route(r, "GET", "/traces/operations", func(r *http.Request) (any, error) {
		service := queryValue(r, "serviceName")
		if service == "" {
			return nil, entity.ErrInvalid
		}
		return s.traceNames(r, service)
	})
	s.route(r, "GET", "/traces/{id}", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, false)
		if err != nil {
			return nil, err
		}
		d, err := s.source(r, ns, queryValue(r, "dataSourceName"), "Traces")
		if err != nil {
			return nil, err
		}
		if !strings.EqualFold(d.Type, "Jaeger") {
			return s.plugin(r, d, "traces.get", map[string]string{"traceId": param(r, "id")})
		}
		traces, err := s.Telemetry.Traces(r.Context(), d, entity.TraceSearch{}, param(r, "id"))
		if err != nil {
			return nil, upstream(err)
		}
		if len(traces) == 0 {
			return nil, entity.ErrNotFound
		}
		return traces[0], nil
	})
	s.route(r, "POST", "/traces/search", func(r *http.Request) (any, error) {
		ns, err := s.scope(r, false)
		if err != nil {
			return nil, err
		}
		input := entity.TraceSearch{Limit: 100}
		if err = decodeInto(r, &input); err != nil {
			return nil, err
		}
		d, err := s.source(r, ns, input.DataSourceName, "Traces")
		if err != nil {
			return nil, err
		}
		if !strings.EqualFold(d.Type, "Jaeger") {
			return s.plugin(r, d, "traces.search", input)
		}
		out, err := s.Telemetry.Traces(r.Context(), d, input, "")
		return out, upstream(err)
	})
	s.route(r, "POST", "/secret/add", func(r *http.Request) (any, error) {
		ns, err := s.namespaceAdmin(r)
		if err != nil {
			return nil, err
		}
		input, err := body[struct{ Key, Value string }](r)
		if err != nil {
			return nil, err
		}
		return "Secret added", s.Service.PutSecret(r.Context(), ns, input.Key, input.Value)
	})
	s.route(r, "GET", "/secret/{key}", func(r *http.Request) (any, error) {
		ns, err := s.namespaceAdmin(r)
		if err != nil {
			return nil, err
		}
		return s.Service.Secret(r.Context(), ns, param(r, "key"))
	})
	s.route(r, "DELETE", "/secret/remove/{key}", func(r *http.Request) (any, error) {
		ns, err := s.namespaceAdmin(r)
		if err != nil {
			return nil, err
		}
		return "Secret removed", s.Service.DeleteSecret(r.Context(), ns, param(r, "key"))
	})
	s.plugins(r)
	s.route(r, "GET", "/meta/database-schema", func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		return []map[string]any{{"schema": "datacat", "tables": []string{"resources", "outbox"}}}, nil
	})
	s.route(r, "GET", "/meta/sql-queries", func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		return []string{}, nil
	})
	// Destructive migrations are available only through the database toolchain.
	s.route(r, "DELETE", "/debug/rollback/{steps}", func(r *http.Request) (any, error) { return nil, entity.ErrForbidden })
}
func decodeInto(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%w: invalid JSON body", entity.ErrInvalid)
	}
	return nil
}
func upstream(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", entity.ErrUpstream, err)
}
func (s *Server) source(r *http.Request, ns, name, purpose string) (entity.DataSource, error) {
	d, err := s.Service.DataSource(r.Context(), ns, name)
	if err != nil {
		return d, err
	}
	if d.Purpose != purpose {
		return d, fmt.Errorf("%w: wrong data source purpose", entity.ErrInvalid)
	}
	return d, nil
}
func (s *Server) metrics(r *http.Request, isRange bool) (any, error) {
	ns, err := s.scope(r, false)
	if err != nil {
		return nil, err
	}
	if requested := queryValue(r, "namespaceId"); requested != "" && requested != ns {
		return nil, entity.ErrForbidden
	}
	d, err := s.source(r, ns, queryValue(r, "dataSourceName"), "Metrics")
	if err != nil {
		return nil, err
	}
	query, err := s.Service.ExpandQuery(r.Context(), ns, queryValue(r, "dashboardId"), queryValue(r, "query"))
	if err != nil {
		return nil, err
	}
	if err = usecases.Required(query); err != nil {
		return nil, err
	}
	if !strings.EqualFold(d.Type, "Prometheus") {
		operation := "metrics.query"
		if isRange {
			operation = "metrics.queryRange"
		}
		return s.plugin(r, d, operation, map[string]string{"query": query, "from": queryValue(r, "from"), "to": queryValue(r, "to"), "step": queryValue(r, "step")})
	}
	if !isRange {
		out, err := s.Telemetry.Instant(r.Context(), d, query)
		return out, upstream(err)
	}
	from, err := time.Parse(time.RFC3339Nano, queryValue(r, "from"))
	if err != nil {
		return nil, fmt.Errorf("%w: from must be RFC3339", entity.ErrInvalid)
	}
	to, err := time.Parse(time.RFC3339Nano, queryValue(r, "to"))
	if err != nil {
		return nil, fmt.Errorf("%w: to must be RFC3339", entity.ErrInvalid)
	}
	step := entity.Duration(time.Minute)
	if raw := queryValue(r, "step"); raw != "" {
		step, err = entity.ParseDuration(raw)
		if err != nil {
			return nil, err
		}
	}
	out, err := s.Telemetry.Range(r.Context(), d, query, from, to, time.Duration(step))
	return out, upstream(err)
}
func (s *Server) traceNames(r *http.Request, service string) (any, error) {
	ns, err := s.scope(r, false)
	if err != nil {
		return nil, err
	}
	d, err := s.source(r, ns, queryValue(r, "dataSourceName"), "Traces")
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(d.Type, "Jaeger") {
		return s.plugin(r, d, "traces.names", map[string]string{"service": service})
	}
	out, err := s.Telemetry.TraceNames(r.Context(), d, service)
	return out, upstream(err)
}
func (s *Server) validateConnection(raw string) error {
	var cfg telemetry.Settings
	if json.Unmarshal([]byte(raw), &cfg) != nil {
		return fmt.Errorf("%w: connectionString must contain JSON", entity.ErrInvalid)
	}
	u := cfg.ServerURL
	if u == "" {
		u = cfg.ClusterURL
	}
	if _, err := s.Telemetry.Client.Validate(u); err != nil {
		return fmt.Errorf("%w: %v", entity.ErrInvalid, err)
	}
	return nil
}
func (s *Server) validateNotification(destination, raw string) error {
	var settings struct {
		URL, TelegramTokenPath, ChatID, DestinationEmail, SmtpServer, PasswordPath string
		Port                                                                       int
	}
	if json.Unmarshal([]byte(raw), &settings) != nil {
		return fmt.Errorf("%w: settings must contain JSON", entity.ErrInvalid)
	}
	switch strings.ToLower(destination) {
	case "webhook":
		if _, err := s.Telemetry.Client.Validate(settings.URL); err != nil {
			return fmt.Errorf("%w: %v", entity.ErrInvalid, err)
		}
	case "telegram":
		return usecases.Required(settings.TelegramTokenPath, settings.ChatID)
	case "email":
		if settings.Port < 1 || settings.Port > 65535 {
			return entity.ErrInvalid
		}
		if strings.ContainsAny(settings.DestinationEmail, "\r\n") {
			return entity.ErrInvalid
		}
		if _, err := s.Telemetry.Client.Validate("https://" + settings.SmtpServer + ":" + strconv.Itoa(settings.Port)); err != nil {
			return fmt.Errorf("%w: SMTP host is not allowed", entity.ErrInvalid)
		}
		return usecases.Required(settings.DestinationEmail, settings.SmtpServer, settings.PasswordPath)
	default:
		return nil
	}
	return nil
}
func (s *Server) plugins(r chi.Router) {
	s.route(r, "POST", "/plugin/add", func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		input, err := body[struct{ Name, Version, Author, Description, Settings, Endpoint string }](r)
		if err != nil {
			return nil, err
		}
		if err = usecases.Required(input.Name, input.Version, input.Author, input.Endpoint); err != nil {
			return nil, err
		}
		if _, err = s.Telemetry.Client.Validate(input.Endpoint); err != nil {
			return nil, fmt.Errorf("%w: %v", entity.ErrInvalid, err)
		}
		p := entity.Plugin{ID: usecases.ID(), Name: input.Name, Version: input.Version, Author: input.Author, Description: input.Description, Settings: input.Settings, Endpoint: input.Endpoint}
		return p.ID, usecases.Create(r.Context(), s.Service.Store, usecases.Global, "plugin", p.ID, p.Name, p)
	})
	s.route(r, "GET", "/plugin/search", func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		page, size, err := pagination(r)
		if err != nil {
			return nil, err
		}
		return resourcePage[entity.Plugin](s, r, usecases.Global, "plugin", entity.Search{}, page, size)
	})
	s.route(r, "GET", "/plugin/{id}", func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		p, _, err := usecases.Read[entity.Plugin](r.Context(), s.Service.Store, usecases.Global, "plugin", param(r, "id"))
		return p, err
	})
	s.route(r, "DELETE", "/plugin/remove/{id}", func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		return nil, s.Service.Delete(r.Context(), usecases.Global, "plugin", param(r, "id"))
	})
	s.route(r, "POST", "/plugin/{id}", func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		active, err := strconv.ParseBool(r.URL.Query().Get("isActive"))
		if err != nil {
			return nil, entity.ErrInvalid
		}
		p, rec, err := usecases.Read[entity.Plugin](r.Context(), s.Service.Store, usecases.Global, "plugin", param(r, "id"))
		if err != nil {
			return nil, err
		}
		p.IsActive = active
		return nil, usecases.Update(r.Context(), s.Service.Store, rec, p)
	})
	s.route(r, "PUT", "/plugin/update/{id}", func(r *http.Request) (any, error) {
		if err := admin(r); err != nil {
			return nil, err
		}
		input, err := body[struct{ Description, Settings string }](r)
		if err != nil {
			return nil, err
		}
		p, rec, err := usecases.Read[entity.Plugin](r.Context(), s.Service.Store, usecases.Global, "plugin", param(r, "id"))
		if err != nil {
			return nil, err
		}
		p.Description = input.Description
		p.Settings = input.Settings
		return nil, usecases.Update(r.Context(), s.Service.Store, rec, p)
	})
}
func (s *Server) plugin(r *http.Request, d entity.DataSource, operation string, input any) (any, error) {
	p, _, err := usecases.Read[entity.Plugin](r.Context(), s.Service.Store, usecases.Global, "plugin", d.Type)
	if err != nil {
		return nil, err
	}
	if !p.IsActive {
		return nil, fmt.Errorf("%w: plugin is disabled", entity.ErrInvalid)
	}
	var result json.RawMessage
	err = s.Telemetry.Client.Do(r.Context(), "POST", p.Endpoint, map[string]any{"version": 1, "operation": operation, "namespaceId": d.NamespaceID, "connectionSettings": json.RawMessage(d.ConnectionString), "settings": p.Settings, "input": input}, nil, &result)
	return result, upstream(err)
}
