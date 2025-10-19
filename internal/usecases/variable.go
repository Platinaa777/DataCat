package usecases

import (
	"context"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Service) UpdateVariable(ctx context.Context, ns, id, placeholder, value string) error {
	if err := Required(placeholder, value); err != nil {
		return err
	}
	v, rec, err := Read[entity.Variable](ctx, s.Store, ns, "variable", id)
	if err != nil {
		return err
	}
	v.Placeholder = placeholder
	v.Value = value
	rec.Name = v.Placeholder
	if v.DashboardID != nil {
		rec.Name = *v.DashboardID + "/" + v.Placeholder
	}
	return Update(ctx, s.Store, rec, v)
}
