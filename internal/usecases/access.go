package usecases

import (
	"context"
	"fmt"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/google/uuid"
)

func (s *Service) AuthorizeNamespace(ctx context.Context, principal entity.Principal, ns string, write bool) error {
	if _, err := uuid.Parse(ns); err != nil {
		return fmt.Errorf("%w: namespace must be a UUID", entity.ErrInvalid)
	}
	if !principal.Can(ns, write) {
		return entity.ErrForbidden
	}
	_, err := s.Store.Get(ctx, Global, "namespace", ns)
	return err
}

func CanManageNamespace(principal entity.Principal, ns string) bool {
	if principal.GlobalAdmin {
		return true
	}
	for _, claim := range principal.Claims {
		if claim.NamespaceID == ns && claim.Role == "Admin" {
			return true
		}
	}
	return false
}

func (s *Service) AuthorizeNamespaceAdmin(ctx context.Context, principal entity.Principal, ns string) error {
	if err := s.AuthorizeNamespace(ctx, principal, ns, true); err != nil {
		return err
	}
	if !CanManageNamespace(principal, ns) {
		return entity.ErrForbidden
	}
	return nil
}

func (s *Service) ResolvePrincipal(ctx context.Context, subject string, realmRoles []string, namespaceRoles map[string][]string) (entity.Principal, error) {
	p := entity.Principal{Subject: subject, Claims: []entity.Claim{}}
	roles := map[string]bool{}
	for _, role := range realmRoles {
		roles[role] = true
	}
	p.GlobalAdmin = roles["datacat-admin"]
	for ns, rs := range namespaceRoles {
		for _, role := range rs {
			if role == "Admin" || role == "Editor" || role == "Viewer" {
				p.Claims = append(p.Claims, entity.Claim{Role: role, NamespaceID: ns})
			}
		}
	}
	mappings, err := List[entity.RoleMapping](ctx, s.Store, Global, "role-mapping")
	if err != nil {
		return entity.Principal{}, err
	}
	for _, m := range mappings {
		if roles[m.ExternalRole] {
			role := map[int]string{1: "Admin", 2: "Viewer", 3: "Editor"}[m.RoleID]
			if role != "" {
				p.Claims = append(p.Claims, entity.Claim{Role: role, NamespaceID: m.NamespaceID})
			}
		}
	}
	manual, _, err := Read[[]entity.Claim](ctx, s.Store, Global, "user-claims", p.Subject)
	if err != nil && err != entity.ErrNotFound {
		return entity.Principal{}, err
	}
	p.Claims = append(p.Claims, manual...)
	return p, nil
}
