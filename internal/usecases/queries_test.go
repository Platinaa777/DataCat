package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/pkg/secrets"
	"github.com/Platinaa777/DataCat/internal/usecases/contract"
)

type applicationStore struct {
	contract.Transactional
	records []entity.Record
}

func (s *applicationStore) Get(_ context.Context, ns, kind, id string) (entity.Record, error) {
	for _, record := range s.records {
		if record.Namespace == ns && record.Kind == kind && (record.ID == id || record.Name == id) {
			return record, nil
		}
	}
	return entity.Record{}, entity.ErrNotFound
}

func (s *applicationStore) List(_ context.Context, ns, kind string) ([]entity.Record, error) {
	result := []entity.Record{}
	for _, record := range s.records {
		if record.Namespace == ns && record.Kind == kind {
			result = append(result, record)
		}
	}
	return result, nil
}

func (s *applicationStore) add(t *testing.T, ns, kind, id, name string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	s.records = append(s.records, entity.Record{Namespace: ns, Kind: kind, ID: id, Name: name, Data: data})
}

type queryProvider struct {
	contract.Telemetry
	calls int
	err   error
}

func (p *queryProvider) Instant(_ context.Context, source entity.DataSource, _ string) ([]entity.MetricPoint, error) {
	p.calls++
	return []entity.MetricPoint{{MetricName: source.NamespaceID, Value: 1}}, p.err
}

type queryPlugin struct {
	calls                int
	operation, namespace string
}

func (p *queryPlugin) Call(_ context.Context, _ entity.Plugin, source entity.DataSource, operation string, _ any) (json.RawMessage, error) {
	p.calls++
	p.operation = operation
	p.namespace = source.NamespaceID
	return json.RawMessage(`[]`), nil
}

type queryCache map[string][]byte

func (c queryCache) Get(_ context.Context, key string) ([]byte, bool) {
	value, ok := c[key]
	return value, ok
}

func (c queryCache) Set(_ context.Context, key string, value []byte, _ time.Duration) { c[key] = value }

func queryFixture(t *testing.T) (*Service, *applicationStore) {
	t.Helper()
	store := &applicationStore{}
	cipher, err := secrets.New(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	service := New(store, cipher, nil)
	for _, ns := range []string{"team-a", "team-b"} {
		connection, err := cipher.Seal(ns+"/data-source/prom", `{"serverUrl":"https://prom.example.test"}`)
		if err != nil {
			t.Fatal(err)
		}
		store.add(t, ns, "data-source", "prom", "prom", entity.DataSource{ID: "prom", NamespaceID: ns, Name: "prom", Type: "Prometheus", Purpose: "Metrics", ConnectionString: connection})
	}
	return service, store
}

func TestMetricsCacheIsolatesNamespacesAndRejectsMismatches(t *testing.T) {
	service, _ := queryFixture(t)
	provider := &queryProvider{}
	queries := NewQueries(service, provider, nil, queryCache{})
	input := MetricsQuery{DataSource: "prom", Query: "up"}
	for _, ns := range []string{"team-a", "team-b", "team-a", "team-b"} {
		result, err := queries.Metrics(context.Background(), ns, input, false)
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		var points []entity.MetricPoint
		if err := json.Unmarshal(data, &points); err != nil || len(points) != 1 || points[0].MetricName != ns {
			t.Fatalf("cached result leaked across namespaces: %s (%v)", data, err)
		}
	}
	if provider.calls != 2 {
		t.Fatalf("provider called %d times, want 2", provider.calls)
	}
	input.RequestedNamespace = "team-b"
	if _, err := queries.Metrics(context.Background(), "team-a", input, false); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("namespace mismatch accepted: %v", err)
	}
	if provider.calls != 2 {
		t.Fatal("mismatched request reached provider")
	}
}

func TestFailedQueriesAreNotCached(t *testing.T) {
	service, _ := queryFixture(t)
	provider := &queryProvider{err: errors.New("unavailable")}
	queries := NewQueries(service, provider, nil, queryCache{})
	for i := 0; i < 2; i++ {
		if _, err := queries.Metrics(context.Background(), "team-a", MetricsQuery{DataSource: "prom", Query: "up"}, false); !errors.Is(err, entity.ErrUpstream) {
			t.Fatalf("unexpected provider error: %v", err)
		}
	}
	if provider.calls != 2 {
		t.Fatal("provider error was cached")
	}
}

func TestRangeValidationRunsWithoutHTTP(t *testing.T) {
	service, _ := queryFixture(t)
	queries := NewQueries(service, &queryProvider{}, nil, nil)
	if _, err := queries.Metrics(context.Background(), "team-a", MetricsQuery{DataSource: "prom", Query: "up", From: "invalid"}, true); !errors.Is(err, entity.ErrInvalid) {
		t.Fatalf("invalid range accepted: %v", err)
	}
}

func TestApplicationDispatchesOnlyActivePlugins(t *testing.T) {
	service, store := queryFixture(t)
	connection, err := service.Cipher.Seal("team-a/data-source/logs", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	store.add(t, "team-a", "data-source", "logs", "logs", entity.DataSource{ID: "logs", NamespaceID: "team-a", Type: "custom", Purpose: "Logs", ConnectionString: connection})
	store.add(t, Global, "plugin", "custom", "custom", entity.Plugin{ID: "custom", IsActive: false})
	plugin := &queryPlugin{}
	queries := NewQueries(service, nil, plugin, nil)
	if _, err := queries.Logs(context.Background(), "team-a", entity.LogSearch{DataSourceName: "logs"}); !errors.Is(err, entity.ErrInvalid) {
		t.Fatalf("disabled plugin accepted: %v", err)
	}
	if plugin.calls != 0 {
		t.Fatal("disabled plugin was called")
	}
	store.records[len(store.records)-1].Data = json.RawMessage(`{"id":"custom","isActive":true}`)
	if _, err := queries.Logs(context.Background(), "team-a", entity.LogSearch{DataSourceName: "logs"}); err != nil {
		t.Fatal(err)
	}
	if plugin.calls != 1 || plugin.operation != "logs.search" || plugin.namespace != "team-a" {
		t.Fatalf("wrong plugin dispatch: %+v", plugin)
	}
}

func TestApplicationViewsRespectCredentialPermissions(t *testing.T) {
	service, store := queryFixture(t)
	ctx := context.Background()
	viewer := entity.Principal{}
	admin := entity.Principal{GlobalAdmin: true}
	for _, test := range []struct {
		principal   entity.Principal
		credentials bool
	}{{viewer, false}, {admin, true}} {
		source, err := service.DataSourceFor(ctx, "team-a", "prom", test.principal)
		if err != nil {
			t.Fatal(err)
		}
		if (source.ConnectionString != "") != test.credentials {
			t.Fatalf("incorrect connection visibility: %+v", source)
		}
	}
	store.add(t, "team-a", "notification-channel-group", "ops", "ops", entity.NotificationGroup{ID: "ops"})
	settings, err := service.Cipher.Seal("team-a/notification-channel/1", `{"token":"sensitive"}`)
	if err != nil {
		t.Fatal(err)
	}
	store.add(t, "team-a", "notification-channel", "1", "", entity.NotificationChannel{ID: 1, GroupID: "ops", Settings: settings})
	for _, test := range []struct {
		principal entity.Principal
		want      string
	}{{viewer, "{}"}, {admin, `{"token":"sensitive"}`}} {
		group, err := service.GroupFor(ctx, "team-a", "ops", test.principal)
		if err != nil {
			t.Fatal(err)
		}
		if len(group.NotificationChannels) != 1 || group.NotificationChannels[0].Settings != test.want {
			t.Fatalf("incorrect channel visibility: %+v", group)
		}
	}
}

func TestResolvePrincipalCombinesOnlyAssignedRoles(t *testing.T) {
	service, store := queryFixture(t)
	store.add(t, Global, "role-mapping", "mapping", "readers", entity.RoleMapping{ExternalRole: "readers", NamespaceID: "mapped", RoleID: 2})
	store.add(t, Global, "user-claims", "alice", "", []entity.Claim{{NamespaceID: "manual", Role: "Editor"}})
	principal, err := service.ResolvePrincipal(context.Background(), "alice", []string{"readers"}, map[string][]string{"direct": {"Viewer"}, "invalid": {"Owner"}})
	if err != nil {
		t.Fatal(err)
	}
	if principal.GlobalAdmin || !principal.Can("direct", false) || principal.Can("direct", true) || !principal.Can("mapped", false) || principal.Can("mapped", true) || !principal.Can("manual", true) || principal.Can("invalid", false) {
		t.Fatalf("incorrect resolved permissions: %+v", principal)
	}
}
