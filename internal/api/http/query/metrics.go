package query

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

func (s *Handler) metrics(r *http.Request, isRange bool) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	if requested := transport.QueryValue(r, "namespaceId"); requested != "" && requested != ns {
		return nil, entity.ErrForbidden
	}
	d, err := s.source(r, ns, transport.QueryValue(r, "dataSourceName"), "Metrics")
	if err != nil {
		return nil, err
	}
	query, err := s.service.ExpandQuery(r.Context(), ns, transport.QueryValue(r, "dashboardId"), transport.QueryValue(r, "query"))
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
		return s.plugin(r, d, operation, map[string]string{"query": query, "from": transport.QueryValue(r, "from"), "to": transport.QueryValue(r, "to"), "step": transport.QueryValue(r, "step")})
	}
	if !isRange {
		out, err := s.provider.Instant(r.Context(), d, query)
		return out, transport.Upstream(err)
	}
	from, err := time.Parse(time.RFC3339Nano, transport.QueryValue(r, "from"))
	if err != nil {
		return nil, fmt.Errorf("%w: from must be RFC3339", entity.ErrInvalid)
	}
	to, err := time.Parse(time.RFC3339Nano, transport.QueryValue(r, "to"))
	if err != nil {
		return nil, fmt.Errorf("%w: to must be RFC3339", entity.ErrInvalid)
	}
	step := entity.Duration(time.Minute)
	if raw := transport.QueryValue(r, "step"); raw != "" {
		step, err = entity.ParseDuration(raw)
		if err != nil {
			return nil, err
		}
	}
	out, err := s.provider.Range(r.Context(), d, query, from, to, time.Duration(step))
	return out, transport.Upstream(err)
}

func (s *Handler) cachedMetrics(r *http.Request, isRange bool) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	if s.cache == nil {
		return s.metrics(r, isRange)
	}
	hash := sha256.Sum256([]byte(ns + "/" + r.URL.Path + "?" + r.URL.Query().Encode()))
	key := hex.EncodeToString(hash[:])
	if b, ok := s.cache.Get(r.Context(), key); ok {
		return json.RawMessage(b), nil
	}
	out, err := s.metrics(r, isRange)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	s.cache.Set(r.Context(), key, b, 5*time.Second)
	return out, nil
}
