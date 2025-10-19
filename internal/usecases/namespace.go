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
