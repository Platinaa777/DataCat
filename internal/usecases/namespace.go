package usecases

import (
	"context"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (s *Service) MapRole(ctx context.Context, input entity.RoleMapping) error {
	if input.RoleID < 1 || input.RoleID > 3 {
		return entity.ErrInvalid
	}
	if err := Required(input.ExternalRole, input.NamespaceID); err != nil {
		return err
	}
	if _, _, err := Read[entity.Namespace](ctx, s.Store, Global, "namespace", input.NamespaceID); err != nil {
		return err
	}
	return Create(ctx, s.Store, Global, "role-mapping", ID(), input.ExternalRole+"/"+input.NamespaceID, input)
}

func (s *Service) AvailableNamespaces(ctx context.Context, principal entity.Principal) ([]entity.Namespace, error) {
	all, err := List[entity.Namespace](ctx, s.Store, Global, "namespace")
	if err != nil {
		return nil, err
	}
	result := []entity.Namespace{}
	for _, ns := range all {
		if principal.Can(ns.ID, false) {
			result = append(result, ns)
		}
	}
	return result, nil
}

func (s *Service) NamespaceDetails(ctx context.Context, principal entity.Principal, name string) (entity.NamespaceDetails, error) {
	ns, err := Resource[entity.Namespace](ctx, s, Global, "namespace", name)
	if err != nil {
		return entity.NamespaceDetails{}, err
	}
	if !principal.Can(ns.ID, false) {
		return entity.NamespaceDetails{}, entity.ErrForbidden
	}
	dashboards, err := List[entity.Dashboard](ctx, s.Store, ns.ID, "dashboard")
	return entity.NamespaceDetails{Name: ns.Name, Dashboards: dashboards}, err
}
