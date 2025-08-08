package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) cachedMetrics(r *http.Request, isRange bool) (any, error) {
	ns, err := s.scope(r, false)
	if err != nil {
		return nil, err
	}
	if s.Cache == nil {
		return s.metrics(r, isRange)
	}
	hash := sha256.Sum256([]byte(ns + "/" + r.URL.Path + "?" + r.URL.Query().Encode()))
	key := hex.EncodeToString(hash[:])
	if b, ok := s.Cache.Get(r.Context(), key); ok {
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
	s.Cache.Set(r.Context(), key, b, 5*time.Second)
	return out, nil
}
