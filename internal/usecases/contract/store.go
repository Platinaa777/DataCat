package contract

import (
	"context"

	"github.com/Platinaa777/DataCat/internal/entity"
)

type Store interface {
	Search(context.Context, string, string, entity.Search, int, int) ([]entity.Record, int, error)
	Get(context.Context, string, string, string) (entity.Record, error)
	List(context.Context, string, string) ([]entity.Record, error)
	Create(context.Context, entity.Record) error
	Update(context.Context, entity.Record) error
	Delete(context.Context, string, string, string) error
	NextID(context.Context) (int64, error)
}

type Transactional interface {
	Store
	Within(context.Context, func(Store) error) error
	Ping(context.Context) error
}
