package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/pkg/secrets"
	"github.com/Platinaa777/DataCat/internal/usecases/contract"
	"github.com/google/uuid"
)

const Global = "global"

type SecretStore interface {
	Get(context.Context, string, string) (string, error)
	Put(context.Context, string, string, string) error
	Delete(context.Context, string, string) error
}
type Service struct {
	Store         contract.Transactional
	Cipher        *secrets.Cipher
	Now           func() time.Time
	RemoteSecrets SecretStore
	Endpoints     contract.EndpointPolicy
}

func New(store contract.Transactional, cipher *secrets.Cipher, endpoints contract.EndpointPolicy) *Service {
	return &Service{Store: store, Cipher: cipher, Endpoints: endpoints, Now: func() time.Time { return time.Now().UTC() }}
}

func ID() string { return uuid.NewString() }

func Required(fields ...string) error {
	for _, f := range fields {
		if strings.TrimSpace(f) == "" || len(f) > 65536 {
			return fmt.Errorf("%w: required field is empty or too long", entity.ErrInvalid)
		}
	}
	return nil
}

func Tags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" && !seen[t] {
			out = append(out, t)
			seen[t] = true
		}
	}
	sort.Strings(out)
	return out
}

func Read[T any](ctx context.Context, store contract.Store, ns, kind, id string) (T, entity.Record, error) {
	var value T
	r, err := store.Get(ctx, ns, kind, id)
	if err != nil {
		return value, r, err
	}
	err = json.Unmarshal(r.Data, &value)
	if err != nil {
		return value, r, fmt.Errorf("decode %s: %w", kind, err)
	}
	return value, r, nil
}

func List[T any](ctx context.Context, store contract.Store, ns, kind string) ([]T, error) {
	records, err := store.List(ctx, ns, kind)
	if err != nil {
		return nil, err
	}
	out := make([]T, 0, len(records))
	for _, r := range records {
		var v T
		if err = json.Unmarshal(r.Data, &v); err != nil {
			return nil, fmt.Errorf("decode %s: %w", kind, err)
		}
		out = append(out, v)
	}
	return out, nil
}

func Create[T any](ctx context.Context, store contract.Store, ns, kind, id, name string, value T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return store.Create(ctx, entity.Record{Namespace: ns, Kind: kind, ID: id, Name: name, Data: data})
}

func Update[T any](ctx context.Context, store contract.Store, record entity.Record, value T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	record.Data = data
	return store.Update(ctx, record)
}

func (s *Service) CreateNamespace(ctx context.Context, name string) (string, error) {
	if err := Required(name); err != nil {
		return "", err
	}
	id := ID()
	err := Create(ctx, s.Store, Global, "namespace", id, name, entity.Namespace{ID: id, Name: name})
	return id, err
}

func (s *Service) CreateDashboard(ctx context.Context, d entity.Dashboard) (string, error) {
	if err := Required(d.Name, d.NamespaceID); err != nil {
		return "", err
	}
	if _, _, err := Read[entity.Namespace](ctx, s.Store, Global, "namespace", d.NamespaceID); err != nil {
		return "", err
	}
	d.ID = ID()
	d.CreatedAt = s.Now()
	d.UpdatedAt = d.CreatedAt
	d.Tags = Tags(d.Tags)
	return d.ID, Create(ctx, s.Store, d.NamespaceID, "dashboard", d.ID, d.Name, d)
}

func (s *Service) DataSource(ctx context.Context, ns, id string) (entity.DataSource, error) {
	d, _, err := Read[entity.DataSource](ctx, s.Store, ns, "data-source", id)
	if err != nil {
		return d, err
	}
	d.ConnectionString, err = s.Cipher.Open(ns+"/data-source/"+d.ID, d.ConnectionString)
	return d, err
}

func (s *Service) CreateDataSource(ctx context.Context, ns string, d entity.DataSource) (string, error) {
	if err := s.validateConnection(d.ConnectionString); err != nil {
		return "", err
	}
	if err := Required(d.Name, d.Type, d.ConnectionString); err != nil {
		return "", err
	}
	if d.Purpose != "Metrics" && d.Purpose != "Logs" && d.Purpose != "Traces" {
		return "", fmt.Errorf("%w: data source purpose", entity.ErrInvalid)
	}
	d.ID = ID()
	d.NamespaceID = ns
	encrypted, err := s.Cipher.Seal(ns+"/data-source/"+d.ID, d.ConnectionString)
	if err != nil {
		return "", err
	}
	d.ConnectionString = encrypted
	err = s.Store.Within(ctx, func(store contract.Store) error {
		if _, _, err := Read[entity.Namespace](ctx, store, Global, "namespace", ns); err != nil {
			return err
		}
		sourceType, _, err := Read[entity.NamedType](ctx, store, Global, "data-source-type", d.Type)
		if err != nil {
			return err
		}
		d.TypeID = sourceType.ID
		return Create(ctx, store, ns, "data-source", d.ID, d.Name, d)
	})
	return d.ID, err
}

func (s *Service) CreatePanel(ctx context.Context, p entity.Panel, sourceID string) (string, error) {
	if err := Required(p.Title, p.Layout, p.Query.Query, p.TypeName, p.DashboardID); err != nil {
		return "", err
	}
	if !json.Valid([]byte(p.Layout)) {
		return "", fmt.Errorf("%w: layout must contain JSON", entity.ErrInvalid)
	}
	p.ID = ID()
	err := s.Store.Within(ctx, func(store contract.Store) error {
		if _, _, err := Read[entity.Dashboard](ctx, store, p.NamespaceID, "dashboard", p.DashboardID); err != nil {
			return err
		}
		d, _, err := Read[entity.DataSource](ctx, store, p.NamespaceID, "data-source", sourceID)
		if err != nil {
			return err
		}
		d.ConnectionString = ""
		p.Query.DataSource = d
		return Create(ctx, store, p.NamespaceID, "panel", p.ID, "", p)
	})
	return p.ID, err
}

func (s *Service) CreateVariable(ctx context.Context, v entity.Variable) (string, error) {
	if err := Required(v.Placeholder, v.Value); err != nil {
		return "", err
	}
	v.ID = ID()
	err := s.Store.Within(ctx, func(store contract.Store) error {
		if v.DashboardID != nil {
			if _, _, err := Read[entity.Dashboard](ctx, store, v.NamespaceID, "dashboard", *v.DashboardID); err != nil {
				return err
			}
		}
		name := v.Placeholder
		if v.DashboardID != nil {
			name = *v.DashboardID + "/" + name
		}
		return Create(ctx, store, v.NamespaceID, "variable", v.ID, name, v)
	})
	return v.ID, err
}

func (s *Service) ExpandQuery(ctx context.Context, ns, dashboardID, query string) (string, error) {
	vars, err := List[entity.Variable](ctx, s.Store, ns, "variable")
	if err != nil {
		return "", err
	}
	values := map[string]string{}
	for _, v := range vars {
		if v.DashboardID == nil {
			values[v.Placeholder] = v.Value
		}
	}
	for _, v := range vars {
		if v.DashboardID != nil && *v.DashboardID == dashboardID {
			values[v.Placeholder] = v.Value
		}
	}
	return replaceQueryVariables(query, values), nil
}

func (s *Service) PutSecret(ctx context.Context, ns, key, value string) error {
	if err := Required(key, value); err != nil {
		return err
	}
	if s.RemoteSecrets != nil {
		return s.RemoteSecrets.Put(ctx, ns, key, value)
	}
	ciphertext, err := s.Cipher.Seal(ns+"/secret/"+key, value)
	if err != nil {
		return err
	}
	_, record, err := Read[string](ctx, s.Store, ns, "secret", key)
	if err == entity.ErrNotFound {
		return Create(ctx, s.Store, ns, "secret", key, "", ciphertext)
	}
	if err != nil {
		return err
	}
	return Update(ctx, s.Store, record, ciphertext)
}

func (s *Service) Secret(ctx context.Context, ns, key string) (string, error) {
	if s.RemoteSecrets != nil {
		return s.RemoteSecrets.Get(ctx, ns, key)
	}
	value, _, err := Read[string](ctx, s.Store, ns, "secret", key)
	if err != nil {
		return "", err
	}
	return s.Cipher.Open(ns+"/secret/"+key, value)
}

func (s *Service) Seed(ctx context.Context) error {
	return s.Store.Within(ctx, func(store contract.Store) error {
		if _, err := store.Get(ctx, Global, "namespace", "00000000-0000-0000-0000-000000000000"); err == entity.ErrNotFound {
			if err = Create(ctx, store, Global, "namespace", "00000000-0000-0000-0000-000000000000", "datacat_system", entity.Namespace{ID: "00000000-0000-0000-0000-000000000000", Name: "datacat_system"}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		for kind, names := range map[string][]string{"data-source-type": {"prometheus", "elasticsearch", "jaeger"}, "notification-destination": {"telegram", "email", "webhook"}} {
			for _, name := range names {
				if _, err := store.Get(ctx, Global, kind, name); err == nil {
					continue
				} else if err != entity.ErrNotFound {
					return err
				}
				id, err := store.NextID(ctx)
				if err != nil {
					return err
				}
				if err = Create(ctx, store, Global, kind, fmt.Sprint(id), name, entity.NamedType{ID: id, Name: name}); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *Service) DeleteSecret(ctx context.Context, ns, key string) error {
	if s.RemoteSecrets != nil {
		return s.RemoteSecrets.Delete(ctx, ns, key)
	}
	return s.Store.Delete(ctx, ns, "secret", key)
}

var queryVariablePattern = regexp.MustCompile(`\{\s*\.(\w+)\s*\}|\$\{(\w+)\}|\{\{(\w+)\}\}`)

func replaceQueryVariables(query string, values map[string]string) string {
	return queryVariablePattern.ReplaceAllStringFunc(query, func(token string) string {
		groups := queryVariablePattern.FindStringSubmatch(token)
		for _, key := range groups[1:] {
			if key != "" {
				if value, ok := values[key]; ok {
					return value
				}
				break
			}
		}
		return token
	})
}
