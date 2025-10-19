package notification

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

type Handler struct {
	service   *usecases.Service
	access    *transport.Access
	responses *transport.Responder
	provider  *telemetry.Provider
}

func New(service *usecases.Service, access *transport.Access, responses *transport.Responder, provider *telemetry.Provider) *Handler {
	return &Handler{service: service, access: access, responses: responses, provider: provider}
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
	if err = usecases.Required(input.Name); err != nil {
		return nil, err
	}
	id := usecases.ID()
	return id, usecases.Create(r.Context(), s.service.Store, ns, "notification-channel-group", id, input.Name, entity.NotificationGroup{ID: id, Name: input.Name, NamespaceID: ns, NotificationChannels: []entity.NotificationChannel{}})
}

func (s *Handler) Groups(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	groups, err := s.service.Groups(r.Context(), ns)
	if _, permissionErr := s.access.NamespaceAdmin(r); permissionErr != nil {
		for i := range groups {
			for j := range groups[i].NotificationChannels {
				groups[i].NotificationChannels[j].Settings = "{}"
			}
		}
	}
	return groups, err
}

func (s *Handler) Group(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	group, err := s.service.Group(r.Context(), ns, transport.Param(r, "id"))
	if _, permissionErr := s.access.NamespaceAdmin(r); permissionErr != nil {
		for i := range group.NotificationChannels {
			group.NotificationChannels[i].Settings = "{}"
		}
	}
	return group, err
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
	if err = s.validateNotification(input.DestinationName, input.Settings); err != nil {
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
	if err = s.validateNotification(input.DestinationName, input.Settings); err != nil {
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
