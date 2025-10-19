package usecases

import (
	"context"
	"encoding/json"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func Resource[T any](ctx context.Context, s *Service, ns, kind, id string) (T, error) {
	value, _, err := Read[T](ctx, s.Store, ns, kind, id)
	return value, err
}

func SearchResources[T any](ctx context.Context, s *Service, ns, kind string, query entity.Search, page, size int) (entity.Page[T], error) {
	records, total, err := s.Store.Search(ctx, ns, kind, query, page, size)
	if err != nil {
		return entity.Page[T]{}, err
	}
	items := make([]T, 0, len(records))
	for _, record := range records {
		var value T
		if err := json.Unmarshal(record.Data, &value); err != nil {
			return entity.Page[T]{}, err
		}
		items = append(items, value)
	}
	return entity.NewPage(items, total, page, size), nil
}

func (s *Service) Ping(ctx context.Context) error { return s.Store.Ping(ctx) }
