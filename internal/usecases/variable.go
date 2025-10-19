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

func (s *Service) DashboardVariables(ctx context.Context, ns, dashboardID string) ([]entity.Variable, error) {
	all, err := List[entity.Variable](ctx, s.Store, ns, "variable")
	if err != nil {
		return nil, err
	}
	result := []entity.Variable{}
	for _, v := range all {
		if v.DashboardID != nil && *v.DashboardID == dashboardID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (s *Service) NamespaceVariables(ctx context.Context, ns, requested string) ([]entity.Variable, error) {
	if ns != requested {
		return nil, entity.ErrForbidden
	}
	return List[entity.Variable](ctx, s.Store, ns, "variable")
}
