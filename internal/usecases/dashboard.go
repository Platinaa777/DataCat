package usecases

import (
	"context"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Service) UpdateDashboard(ctx context.Context, ns, id, name, description string) error {
	if err := Required(name); err != nil {
		return err
	}
	dashboard, record, err := Read[entity.Dashboard](ctx, s.Store, ns, "dashboard", id)
	if err != nil {
		return err
	}
	dashboard.Name = name
	dashboard.Description = description
	dashboard.UpdatedAt = s.Now()
	record.Name = name
	return Update(ctx, s.Store, record, dashboard)
}
