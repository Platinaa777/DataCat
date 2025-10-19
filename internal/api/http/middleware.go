package httpapi

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Server) headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Responses.Requests.Add(1)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		origin := r.Header.Get("Origin")
		if origin != "" {
			if origin != s.Config.AllowedOrigin {
				transport.Problem(w, r, entity.ErrForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Namespace, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" && origin == "" && r.Header.Get("Authorization") == "" {
			if _, err := r.Cookie("access_token"); err == nil {
				transport.Problem(w, r, entity.ErrForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
