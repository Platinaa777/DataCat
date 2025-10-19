package usecases

import (
	"context"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Service) SetPluginActive(ctx context.Context, id string, active bool) error {
	p, rec, err := Read[entity.Plugin](ctx, s.Store, Global, "plugin", id)
	if err != nil {
		return err
	}
	p.IsActive = active
	return Update(ctx, s.Store, rec, p)
}

func (s *Service) UpdatePlugin(ctx context.Context, id, description, settings string) error {
	p, rec, err := Read[entity.Plugin](ctx, s.Store, Global, "plugin", id)
	if err != nil {
		return err
	}
	p.Description = description
	p.Settings = settings
	return Update(ctx, s.Store, rec, p)
}
