package usecases

import (
	"context"
	"fmt"

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

func (s *Service) CreatePlugin(ctx context.Context, plugin entity.Plugin) (string, error) {
	if err := Required(plugin.Name, plugin.Version, plugin.Author, plugin.Endpoint); err != nil {
		return "", err
	}
	if _, err := s.Endpoints.Validate(plugin.Endpoint); err != nil {
		return "", fmt.Errorf("%w: %v", entity.ErrInvalid, err)
	}
	plugin.ID = ID()
	return plugin.ID, Create(ctx, s.Store, Global, "plugin", plugin.ID, plugin.Name, plugin)
}
