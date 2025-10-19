package usecases

import (
	"context"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Service) UpdateDataSourceConnection(ctx context.Context, ns, id, connection string) error {
	if err := s.validateConnection(connection); err != nil {
		return err
	}
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

func (s *Service) DataSourceFor(ctx context.Context, ns, id string, principal entity.Principal) (entity.DataSource, error) {
	source, err := s.DataSource(ctx, ns, id)
	if !CanManageNamespace(principal, ns) {
		source.ConnectionString = ""
	}
	return source, err
}

func (s *Service) SearchDataSources(ctx context.Context, ns string, query entity.Search, page, size int) (entity.Page[entity.DataSource], error) {
	result, err := SearchResources[entity.DataSource](ctx, s, ns, "data-source", query, page, size)
	if err != nil {
		return result, err
	}
	for i := range result.Items {
		result.Items[i].ConnectionString = ""
	}
	return result, nil
}
