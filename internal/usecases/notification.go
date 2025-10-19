package usecases

import (
	"context"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Service) UpdateChannel(ctx context.Context, ns, id, destination, settings string) error {
	if err := s.validateNotification(destination, settings); err != nil {
		return err
	}
	ch, rec, err := Read[entity.NotificationChannel](ctx, s.Store, ns, "notification-channel", id)
	if err != nil {
		return err
	}
	if _, _, err = Read[entity.NamedType](ctx, s.Store, Global, "notification-destination", destination); err != nil {
		return err
	}
	ch.DestinationName = destination
	ch.Settings, err = s.Cipher.Seal(ns+"/notification-channel/"+id, settings)
	if err != nil {
		return err
	}
	return Update(ctx, s.Store, rec, ch)
}

func (s *Service) CreateNotificationGroup(ctx context.Context, ns, name string) (string, error) {
	if err := Required(name); err != nil {
		return "", err
	}
	id := ID()
	return id, Create(ctx, s.Store, ns, "notification-channel-group", id, name, entity.NotificationGroup{ID: id, Name: name, NamespaceID: ns, NotificationChannels: []entity.NotificationChannel{}})
}

func (s *Service) GroupsFor(ctx context.Context, ns string, principal entity.Principal) ([]entity.NotificationGroup, error) {
	groups, err := s.Groups(ctx, ns)
	if err != nil {
		return nil, err
	}
	if !CanManageNamespace(principal, ns) {
		for i := range groups {
			hideChannelSettings(&groups[i])
		}
	}
	return groups, nil
}

func (s *Service) GroupFor(ctx context.Context, ns, id string, principal entity.Principal) (entity.NotificationGroup, error) {
	group, err := s.Group(ctx, ns, id)
	if err != nil {
		return group, err
	}
	if !CanManageNamespace(principal, ns) {
		hideChannelSettings(&group)
	}
	return group, nil
}

func hideChannelSettings(group *entity.NotificationGroup) {
	for i := range group.NotificationChannels {
		group.NotificationChannels[i].Settings = "{}"
	}
}
