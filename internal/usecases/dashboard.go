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

func (s *Service) DashboardDetails(ctx context.Context, ns, id string) (entity.DashboardDetails, error) {
	dashboard, err := Resource[entity.Dashboard](ctx, s, ns, "dashboard", id)
	if err != nil {
		return entity.DashboardDetails{}, err
	}
	panels, err := List[entity.Panel](ctx, s.Store, ns, "panel")
	if err != nil {
		return entity.DashboardDetails{}, err
	}
	result := entity.DashboardDetails{Dashboard: dashboard, Panels: []entity.DashboardPanel{}}
	for _, panel := range panels {
		if panel.DashboardID == dashboard.ID {
			result.Panels = append(result.Panels, entity.DashboardPanel{ID: panel.ID, Title: panel.Title, PanelType: panel.TypeName, Query: panel.Query.Query, Layout: panel.Layout, StyleConfiguration: panel.StyleConfiguration})
		}
	}
	return result, nil
}
