package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
)

func (s *Server) live(w http.ResponseWriter, r *http.Request) {
	transport.Write(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if s.Access.Service.Store.Ping(ctx) != nil {
		transport.Write(w, 503, map[string]string{"status": "unavailable"})
		return
	}
	transport.Write(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# TYPE datacat_http_requests_total counter\ndatacat_http_requests_total %d\n# TYPE datacat_http_errors_total counter\ndatacat_http_errors_total %d\n", s.Responses.Requests.Load(), s.Responses.Failed.Load())
}
