package usecases

import (
	"context"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Service) UpdateChannel(ctx context.Context, ns, id, destination, settings string) error {
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
