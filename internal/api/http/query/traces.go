package query

import (
	"net/http"
	"strings"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
)

func (s *Handler) traceNames(r *http.Request, service string) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	d, err := s.source(r, ns, transport.QueryValue(r, "dataSourceName"), "Traces")
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(d.Type, "Jaeger") {
		return s.plugin(r, d, "traces.names", map[string]string{"service": service})
	}
	out, err := s.provider.TraceNames(r.Context(), d, service)
	return out, transport.Upstream(err)
}
