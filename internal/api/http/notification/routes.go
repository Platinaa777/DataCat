package notification

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "POST", "/notification-channel-group/add", s.CreateGroup)
	s.responses.Route(r, "GET", "/notification-channel-group/get-all", s.Groups)
	s.responses.Route(r, "GET", "/notification-channel-group/{id}", s.Group)
	s.responses.Route(r, "POST", "/notification-channel/add", s.CreateChannel)
	s.responses.Route(r, "GET", "/notification-channel/{id}", s.Channel)
	s.responses.Route(r, "PUT", "/notification-channel/{id}", s.UpdateChannel)
	s.responses.Route(r, "DELETE", "/notification-channel/remove/{id}", s.DeleteChannel)
}
