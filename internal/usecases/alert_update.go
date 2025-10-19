package usecases

import (
	"context"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Service) UpdateAlert(ctx context.Context, ns, id, description, template, query, sourceID string) error {
	if err := Required(template, query, sourceID); err != nil {
		return err
	}
	a, rec, err := Read[entity.Alert](ctx, s.Store, ns, "alert", id)
	if err != nil {
		return err
	}
	d, _, err := Read[entity.DataSource](ctx, s.Store, ns, "data-source", sourceID)
	if err != nil {
		return err
	}
	if d.Purpose != "Metrics" {
		return entity.ErrInvalid
	}
	d.ConnectionString = ""
	a.Description = description
	a.Template = template
	a.RawQuery = query
	a.DataSource = d
	a.PendingSince = nil
	a.LastNotifiedAt = nil
	return Update(ctx, s.Store, rec, a)
}

func (s *Service) MuteAlert(ctx context.Context, ns, id string, duration time.Duration) (string, error) {
	if duration <= 0 {
		return "", entity.ErrInvalid
	}
	alert, record, err := Read[entity.Alert](ctx, s.Store, ns, "alert", id)
	if err != nil {
		return "", err
	}
	alert.Status = "Muted"
	alert.NextExecutionTime = s.Now().Add(duration)
	alert.PendingSince = nil
	return alert.ID, Update(ctx, s.Store, record, alert)
}
