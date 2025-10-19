package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases/contract"
)

func (s *Service) UpdatePanel(ctx context.Context, ns, id, title, typeName, query, sourceID, layout, style string) error {
	if err := Required(title, typeName, query, layout); err != nil {
		return err
	}
	if !json.Valid([]byte(layout)) {
		return fmt.Errorf("%w: invalid layout JSON", entity.ErrInvalid)
	}
	return s.Store.Within(ctx, func(store contract.Store) error {
		p, rec, err := Read[entity.Panel](ctx, store, ns, "panel", id)
		if err != nil {
			return err
		}
		d, _, err := Read[entity.DataSource](ctx, store, ns, "data-source", sourceID)
		if err != nil {
			return err
		}
		d.ConnectionString = ""
		p.Title = title
		p.TypeName = typeName
		p.Query = entity.RawQuery{Query: query, DataSource: d}
		p.Layout = layout
		p.StyleConfiguration = style
		return Update(ctx, store, rec, p)
	})
}

func (s *Service) Groups(ctx context.Context, ns string) ([]entity.NotificationGroup, error) {
	groups, err := List[entity.NotificationGroup](ctx, s.Store, ns, "notification-channel-group")
	if err != nil {
		return nil, err
	}
	channels, err := List[entity.NotificationChannel](ctx, s.Store, ns, "notification-channel")
	if err != nil {
		return nil, err
	}
	for i := range groups {
		groups[i].NotificationChannels = []entity.NotificationChannel{}
		for _, ch := range channels {
			if ch.GroupID == groups[i].ID {
				plain, err := s.Cipher.Open(ns+"/notification-channel/"+fmt.Sprint(ch.ID), ch.Settings)
				if err != nil {
					return nil, err
				}
				ch.Settings = plain
				groups[i].NotificationChannels = append(groups[i].NotificationChannels, ch)
			}
		}
	}
	return groups, nil
}

func (s *Service) Group(ctx context.Context, ns, id string) (entity.NotificationGroup, error) {
	g, _, err := Read[entity.NotificationGroup](ctx, s.Store, ns, "notification-channel-group", id)
	if err != nil {
		return g, err
	}
	channels, err := List[entity.NotificationChannel](ctx, s.Store, ns, "notification-channel")
	g.NotificationChannels = []entity.NotificationChannel{}
	for _, ch := range channels {
		if ch.GroupID == g.ID {
			plain, err := s.Cipher.Open(ns+"/notification-channel/"+fmt.Sprint(ch.ID), ch.Settings)
			if err != nil {
				return g, err
			}
			ch.Settings = plain
			g.NotificationChannels = append(g.NotificationChannels, ch)
		}
	}
	return g, err
}

func (s *Service) CreateChannel(ctx context.Context, ns, group, destination, settings string) (int64, error) {
	if err := s.validateNotification(destination, settings); err != nil {
		return 0, err
	}
	var id int64
	err := s.Store.Within(ctx, func(store contract.Store) error {
		g, _, err := Read[entity.NotificationGroup](ctx, store, ns, "notification-channel-group", group)
		if err != nil {
			return err
		}
		if _, _, err = Read[entity.NamedType](ctx, store, Global, "notification-destination", destination); err != nil {
			return err
		}
		id, err = store.NextID(ctx)
		if err != nil {
			return err
		}
		encrypted, err := s.Cipher.Seal(ns+"/notification-channel/"+fmt.Sprint(id), settings)
		if err != nil {
			return err
		}
		settings = encrypted
		return Create(ctx, store, ns, "notification-channel", fmt.Sprint(id), "", entity.NotificationChannel{ID: id, GroupID: g.ID, NamespaceID: ns, DestinationName: destination, Settings: settings})
	})
	return id, err
}

func (s *Service) CreateAlert(ctx context.Context, a entity.Alert, sourceID, group string) (string, error) {
	if err := Required(a.Template, a.RawQuery, sourceID, group); err != nil {
		return "", err
	}
	if a.RepeatInterval < DurationSecond || a.WaitTimeBeforeAlerting < 0 {
		return "", fmt.Errorf("%w: repeat interval must be at least one second", entity.ErrInvalid)
	}
	a.ID = ID()
	a.Status = "Ok"
	a.NextExecutionTime = s.Now()
	a.PreviousExecutionTime = s.Now()
	a.Tags = Tags(a.Tags)
	err := s.Store.Within(ctx, func(store contract.Store) error {
		d, _, err := Read[entity.DataSource](ctx, store, a.NamespaceID, "data-source", sourceID)
		if err != nil {
			return err
		}
		if d.Purpose != "Metrics" {
			return fmt.Errorf("%w: alerts require a metrics source", entity.ErrInvalid)
		}
		d.ConnectionString = ""
		a.DataSource = d
		g, _, err := Read[entity.NotificationGroup](ctx, store, a.NamespaceID, "notification-channel-group", group)
		if err != nil {
			return err
		}
		g.NotificationChannels = []entity.NotificationChannel{}
		a.NotificationChannelGroup = g
		return Create(ctx, store, a.NamespaceID, "alert", a.ID, "", a)
	})
	return a.ID, err
}

const DurationSecond = entity.Duration(time.Second)

func (s *Service) Delete(ctx context.Context, ns, kind, id string) error {
	return s.Store.Within(ctx, func(store contract.Store) error {
		rec, err := store.Get(ctx, ns, kind, id)
		if err != nil {
			return err
		}
		switch kind {
		case "dashboard":
			panels, err := List[entity.Panel](ctx, store, ns, "panel")
			if err != nil {
				return err
			}
			for _, p := range panels {
				if p.DashboardID == id {
					if err = store.Delete(ctx, ns, "panel", p.ID); err != nil {
						return err
					}
				}
			}
			vars, err := List[entity.Variable](ctx, store, ns, "variable")
			if err != nil {
				return err
			}
			for _, v := range vars {
				if v.DashboardID != nil && *v.DashboardID == id {
					if err = store.Delete(ctx, ns, "variable", v.ID); err != nil {
						return err
					}
				}
			}
		case "data-source":
			panels, err := List[entity.Panel](ctx, store, ns, "panel")
			if err != nil {
				return err
			}
			for _, p := range panels {
				if p.Query.DataSource.ID == id {
					return entity.ErrConflict
				}
			}
			alerts, err := List[entity.Alert](ctx, store, ns, "alert")
			if err != nil {
				return err
			}
			for _, a := range alerts {
				if a.DataSource.ID == id {
					return entity.ErrConflict
				}
			}
		case "data-source-type":
			all, err := List[entity.DataSource](ctx, store, "*", "data-source")
			if err != nil {
				return err
			}
			for _, d := range all {
				if d.Type == rec.Name {
					return entity.ErrConflict
				}
			}
		case "notification-destination":
			all, err := List[entity.NotificationChannel](ctx, store, "*", "notification-channel")
			if err != nil {
				return err
			}
			for _, ch := range all {
				if ch.DestinationName == rec.Name {
					return entity.ErrConflict
				}
			}
		}
		return store.Delete(ctx, ns, kind, id)
	})
}

func (s *Service) Channel(ctx context.Context, ns, id string) (entity.NotificationChannel, error) {
	ch, _, err := Read[entity.NotificationChannel](ctx, s.Store, ns, "notification-channel", id)
	if err != nil {
		return ch, err
	}
	ch.Settings, err = s.Cipher.Open(ns+"/notification-channel/"+id, ch.Settings)
	return ch, err
}
