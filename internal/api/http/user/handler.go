package user

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/api/http/transport"
)

type Handler struct {
	responses *transport.Responder
}

func New(responses *transport.Responder) *Handler {
	return &Handler{responses: responses}
}

func (s *Handler) Me(r *http.Request) (any, error) {
	p := auth.Principal(r.Context())
	return map[string]any{"identityId": p.Subject, "claims": p.Claims, "isGlobalAdmin": p.GlobalAdmin}, nil
}
