package notification

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

type Handler struct {
	service   *usecases.Service
	access    *transport.Access
	responses *transport.Responder
}

func New(service *usecases.Service, access *transport.Access, responses *transport.Responder) *Handler {
	return &Handler{service: service, access: access, responses: responses}
}

func (s *Handler) CreateGroup(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, true)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[createGroupRequest](r)
	if err != nil {
		return nil, err
	}
	return s.service.CreateNotificationGroup(r.Context(), ns, input.Name)
}

func (s *Handler) Groups(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	return s.service.GroupsFor(r.Context(), ns, auth.Principal(r.Context()))
}

func (s *Handler) Group(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	return s.service.GroupFor(r.Context(), ns, transport.Param(r, "id"), auth.Principal(r.Context()))
}

func (s *Handler) CreateChannel(r *http.Request) (any, error) {
	ns, err := s.access.NamespaceAdmin(r)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[createChannelRequest](r)
	if err != nil {
		return nil, err
	}
	return s.service.CreateChannel(r.Context(), ns, input.NotificationChannelGroupName, input.DestinationName, input.Settings)
}

func (s *Handler) Channel(r *http.Request) (any, error) {
	if _, err := s.access.NamespaceAdmin(r); err != nil {
		return nil, err
	}
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	return s.service.Channel(r.Context(), ns, transport.Param(r, "id"))
}

func (s *Handler) UpdateChannel(r *http.Request) (any, error) {
	ns, err := s.access.NamespaceAdmin(r)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[updateChannelRequest](r)
	if err != nil {
		return nil, err
	}
	return nil, s.service.UpdateChannel(r.Context(), ns, transport.Param(r, "id"), input.DestinationName, input.Settings)
}

func (s *Handler) DeleteChannel(r *http.Request) (any, error) {
	if _, err := s.access.NamespaceAdmin(r); err != nil {
		return nil, err
	}
	return s.access.Remove(r, "notification-channel")
}
