package usecases

import (
	"context"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Service) UpdateDataSourceConnection(ctx context.Context, ns, id, connection string) error {
	d, rec, err := Read[entity.DataSource](ctx, s.Store, ns, "data-source", id)
	if err != nil {
		return err
	}
	d.ConnectionString, err = s.Cipher.Seal(ns+"/data-source/"+d.ID, connection)
	if err != nil {
		return err
	}
	return Update(ctx, s.Store, rec, d)
}
