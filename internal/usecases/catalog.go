package usecases

import (
	"context"
	"strconv"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Service) CreateNamedType(ctx context.Context, kind, name string) (int64, error) {
	if err := Required(name); err != nil {
		return 0, err
	}
	id, err := s.Store.NextID(ctx)
	if err != nil {
		return 0, err
	}
	return id, Create(ctx, s.Store, Global, kind, strconv.FormatInt(id, 10), name, entity.NamedType{ID: id, Name: name})
}

func (s *Service) NamedTypes(ctx context.Context, kind string) ([]entity.NamedType, error) {
	return List[entity.NamedType](ctx, s.Store, Global, kind)
}

func (s *Service) NamedType(ctx context.Context, kind, name string) (entity.NamedType, error) {
	return Resource[entity.NamedType](ctx, s, Global, kind, name)
}

func (s *Service) DeleteNamedType(ctx context.Context, kind, name string) error {
	_, record, err := Read[entity.NamedType](ctx, s.Store, Global, kind, name)
	if err != nil {
		return err
	}
	return s.Delete(ctx, Global, kind, record.ID)
}
