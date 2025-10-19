package user

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "GET", "/user/get-me", s.Me)
}
