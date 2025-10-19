//go:build integration

package httpapi_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/bootstrap"
	"github.com/Platinaa777/DataCat/internal/config"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/httpclient"
	"github.com/Platinaa777/DataCat/internal/extapi/notifications"
	"github.com/Platinaa777/DataCat/internal/extapi/plugins"
	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
	"github.com/Platinaa777/DataCat/internal/pkg/secrets"
	"github.com/Platinaa777/DataCat/internal/repository/postgres"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/Platinaa777/DataCat/internal/worker"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

type issuer struct {
	server *httptest.Server
	signer jose.Signer
}

func newIssuer(t *testing.T) *issuer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "test"))
	if err != nil {
		t.Fatal(err)
	}
	i := &issuer{signer: signer}
	i.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(map[string]any{"issuer": i.server.URL, "authorization_endpoint": i.server.URL + "/auth", "token_endpoint": i.server.URL + "/token", "jwks_uri": i.server.URL + "/keys", "id_token_signing_alg_values_supported": []string{"RS256"}})
		case "/keys":
			json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "test", Algorithm: "RS256", Use: "sig"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(i.server.Close)
	return i
}

func (i *issuer) token(t *testing.T, roles map[string][]string, global bool, audience string, expiry time.Time) string {
	t.Helper()
	realm := []string{}
	if global {
		realm = append(realm, "datacat-admin")
	}
	raw, err := jwt.Signed(i.signer).Claims(jwt.Claims{Issuer: i.server.URL, Subject: "test-user", Audience: jwt.Audience{audience}, Expiry: jwt.NewNumericDate(expiry), IssuedAt: jwt.NewNumericDate(time.Now().Add(-time.Minute))}).Claims(map[string]any{"namespace_roles": roles, "realm_access": map[string]any{"roles": realm}}).CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestAPIAndWorkerWithPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is required")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(parsed.Path, "_test") {
		t.Fatal("integration database name must end in _test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	store, err := postgres.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err = store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Pool().Exec(ctx, "CREATE TABLE datacat.metric_batches (id bigint)"); err != nil {
		t.Fatal(err)
	}
	if err = store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	var removed bool
	if err = store.Pool().QueryRow(ctx, "SELECT to_regclass('datacat.metric_batches') IS NULL").Scan(&removed); err != nil || !removed {
		t.Fatalf("retired metric table still exists: %v", err)
	}
	if _, err = store.Pool().Exec(ctx, "TRUNCATE datacat.resources,datacat.outbox RESTART IDENTITY"); err != nil {
		t.Fatal(err)
	}
	var deliveries atomic.Int32
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/query":
			w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up"},"value":[1735689600,"1"]}]}}`))
		case "/api/v1/query_range":
			w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up"},"values":[[1735689600,"1"]]}]}}`))
		case "/logs-*/_search":
			w.Write([]byte(`{"hits":{"total":{"value":1},"hits":[{"_source":{"@timestamp":"2025-01-01T00:00:00Z","message":"hello","level":"info","fields":{"Host":"api","TraceId":"abc"}}}]}}`))
		case "/api/services":
			w.Write([]byte(`{"data":["api"]}`))
		case "/api/services/api/operations":
			w.Write([]byte(`{"data":["GET /health"]}`))
		case "/api/traces", "/api/traces/abc":
			w.Write([]byte(`{"data":[{"traceID":"abc","spans":[{"traceID":"abc","spanID":"s1","operationName":"GET /health","startTime":1735689600000000,"duration":1000,"tags":[],"references":[],"processID":"p1"}],"processes":{"p1":{"serviceName":"api","tags":[]}}}]}`))
		case "/webhook":
			deliveries.Add(1)
			w.Write([]byte(`{}`))
		case "/plugin":
			w.Write([]byte(`[{"metricName":"plugin_metric","value":1,"timestamp":"2025-01-01T00:00:00Z","labels":{}}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer providerServer.Close()
	upstreamURL, _ := url.Parse(providerServer.URL)
	client := httpclient.New([]string{upstreamURL.Host})
	cipher, _ := secrets.New(bytes.Repeat([]byte{1}, 32))
	service := usecases.New(store, cipher, client)
	if err = service.Seed(ctx); err != nil {
		t.Fatal(err)
	}
	idp := newIssuer(t)
	identity, err := auth.New(ctx, config.Config{Issuer: idp.server.URL, ClientID: "datacat", RedirectURL: "http://localhost/callback"}, service, cipher)
	if err != nil {
		t.Fatal(err)
	}
	provider := &telemetry.Provider{Client: client}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := usecases.NewQueries(service, provider, &plugins.Client{HTTP: client}, nil)
	server := bootstrap.NewRouter(service, identity, queries, config.Config{AllowedOrigin: "https://ui.example.test"}, log)
	api := httptest.NewServer(server.Handler())
	defer api.Close()
	adminToken := idp.token(t, nil, true, "datacat", time.Now().Add(time.Hour))
	call := func(method, path, token, ns string, input any, want int) []byte {
		t.Helper()
		var data []byte
		if input != nil {
			var err error
			data, err = json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
		}
		req, err := http.NewRequestWithContext(ctx, method, api.URL+"/api/v1"+path, bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if ns != "" {
			req.Header.Set("X-Namespace", ns)
		}
		req.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		b, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, response.StatusCode, want, b)
		}
		return b
	}
	id := func(b []byte) string {
		t.Helper()
		var id string
		if err := json.Unmarshal(b, &id); err != nil {
			t.Fatalf("expected ID: %s", b)
		}
		return id
	}
	call("GET", "/namespace/get-available", "", "", nil, 401)
	call("GET", "/namespace/get-available", idp.token(t, nil, true, "wrong", time.Now().Add(time.Hour)), "", nil, 401)
	call("GET", "/namespace/get-available", idp.token(t, nil, true, "datacat", time.Now().Add(-time.Minute)), "", nil, 401)
	ns := id(call("POST", "/namespace", adminToken, "", map[string]any{"name": "production"}, 200))
	other := id(call("POST", "/namespace", adminToken, "", map[string]any{"name": "other"}, 200))
	viewer := idp.token(t, map[string][]string{ns: {"Viewer"}}, false, "datacat", time.Now().Add(time.Hour))
	editor := idp.token(t, map[string][]string{ns: {"Editor"}}, false, "datacat", time.Now().Add(time.Hour))
	call("POST", "/dashboard/add", viewer, ns, map[string]any{"name": "denied"}, 403)
	call("POST", "/dashboard/search", viewer, other, map[string]any{}, 403)
	call("POST", "/dashboard/add", editor, ns, map[string]any{"name": "denied", "namespaceId": other}, 403)
	dashboard := id(call("POST", "/dashboard/add", editor, ns, map[string]any{"name": "API", "namespaceId": ns, "tags": []string{"prod", "prod"}}, 200))
	call("POST", "/dashboard/add", editor, ns, map[string]any{"name": "API"}, 409)
	call("GET", "/dashboard/"+dashboard, adminToken, other, nil, 404)
	call("PUT", "/dashboard/update/"+dashboard, editor, ns, map[string]any{"name": "API updated", "description": "Production overview"}, 200)
	var updated entity.Dashboard
	if err := json.Unmarshal(call("GET", "/dashboard/"+dashboard, viewer, ns, nil, 200), &updated); err != nil || updated.Name != "API updated" {
		t.Fatalf("dashboard update was not persisted: %+v %v", updated, err)
	}
	call("POST", "/dashboard/search?page=0", viewer, ns, map[string]any{}, 400)
	call("POST", "/dashboard/search", viewer, ns, map[string]any{"filters": []map[string]any{{"key": "Tags", "value": []string{"prod"}, "matchMode": "contains", "fieldType": "array"}}}, 200)
	call("POST", "/secret/add", editor, ns, map[string]any{"key": "token", "value": "hidden"}, 403)
	call("POST", "/secret/add", adminToken, ns, map[string]any{"key": "token", "value": "hidden"}, 200)
	call("GET", "/secret/token", adminToken, other, nil, 404)
	if string(call("GET", "/secret/token", adminToken, ns, nil, 200)) != "\"hidden\"\n" {
		t.Fatal("secret roundtrip failed")
	}
	var rawSecret string
	if err = store.Pool().QueryRow(ctx, "SELECT data::text FROM datacat.resources WHERE namespace=$1 AND kind='secret'", ns).Scan(&rawSecret); err != nil || strings.Contains(rawSecret, "hidden") {
		t.Fatal("secret was not encrypted")
	}
	connection := fmt.Sprintf(`{"serverUrl":%q}`, providerServer.URL)
	source := id(call("POST", "/data-source/add", adminToken, ns, map[string]any{"uniqueName": "prom", "dataSourceType": "prometheus", "connectionString": connection, "purpose": "metrics"}, 200))
	call("PUT", "/data-source/update-connection-string/"+source, adminToken, ns, map[string]any{"connectionString": connection}, 200)
	if dataSource, err := service.DataSource(ctx, ns, source); err != nil || dataSource.ConnectionString != connection {
		t.Fatalf("connection update failed: %+v %v", dataSource, err)
	}
	call("POST", "/data-source/add", editor, ns, map[string]any{"uniqueName": "denied"}, 403)
	call("GET", "/metrics/query?DataSourceName=prom&Query=up&NamespaceId="+ns, viewer, ns, nil, 200)
	call("GET", "/metrics/query?DataSourceName=prom&Query=up&NamespaceId="+other, viewer, ns, nil, 403)
	call("GET", "/metrics/query-range?DataSourceName=prom&Query=up&From=2025-01-01T00:00:00Z&To=2025-01-01T00:01:00Z&Step=00:01:00", viewer, ns, nil, 200)
	call("POST", "/dashboard/search", viewer, ns, map[string]any{"sort": map[string]any{"fieldName": "name", "sortOrder": "asc"}}, 200)
	call("POST", "/data-source/search", viewer, ns, map[string]any{"filters": []map[string]any{{"key": "purpose", "value": "Metrics", "matchMode": "equals", "fieldType": "string"}}}, 200)
	panel := id(call("POST", "/panel/add", editor, ns, map[string]any{"title": "Requests", "type": 1, "rawQuery": "up", "dataSourceId": source, "layout": "{}", "dashboardId": dashboard, "styleConfiguration": "{}"}, 200))
	call("GET", "/panel/"+panel, viewer, ns, nil, 200)
	call("POST", "/panel/add", adminToken, other, map[string]any{"title": "denied", "type": 1, "rawQuery": "up", "dataSourceId": source, "layout": "{}", "dashboardId": dashboard}, 404)
	variable := id(call("POST", "/variable/add", editor, ns, map[string]any{"placeholder": "service", "value": "api", "namespaceId": ns, "dashboardId": dashboard}, 200))
	call("PUT", "/variable/update/"+variable, editor, ns, map[string]any{"placeholder": "service", "value": "web"}, 200)
	call("GET", "/variables/dashboard/"+dashboard, viewer, ns, nil, 200)
	call("GET", "/variables/namespace/"+other, viewer, ns, nil, 403)
	var updatedVariable entity.Variable
	if err := json.Unmarshal(call("GET", "/variable/"+variable, viewer, ns, nil, 200), &updatedVariable); err != nil || updatedVariable.Value != "web" {
		t.Fatalf("variable update was not persisted: %+v %v", updatedVariable, err)
	}
	call("GET", "/dashboard/"+dashboard+"/full", viewer, ns, nil, 200)
	group := id(call("POST", "/notification-channel-group/add", editor, ns, map[string]any{"name": "ops"}, 200))
	call("POST", "/notification-channel/add", adminToken, ns, map[string]any{"notificationChannelGroupName": "ops", "destinationName": "webhook", "settings": fmt.Sprintf(`{"Url":%q}`, providerServer.URL+"/webhook")}, 200)
	call("GET", "/notification-channel-group/"+group, viewer, ns, nil, 200)
	alert := id(call("POST", "/alert/add", editor, ns, map[string]any{"template": "{.status}", "rawQuery": "up", "dataSourceId": source, "notificationChannelGroupName": "ops", "waitTimeBeforeAlerting": "00:00:00", "repeatInterval": "00:01:00"}, 200))
	call("POST", "/alert/search", viewer, ns, map[string]any{"filters": []map[string]any{{"key": "dataSourceId", "value": source, "matchMode": "equals", "fieldType": "string"}}}, 200)
	call("PUT", "/alert/update/"+alert, editor, ns, map[string]any{"description": "Availability", "template": "{.status}", "rawQuery": "up", "dataSourceId": source}, 200)
	call("GET", "/alert/"+alert, viewer, ns, nil, 200)
	call("DELETE", "/data-source/remove/"+source, adminToken, ns, nil, 409)
	workerCtx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	background := &worker.Worker{Store: store, Service: service, Telemetry: provider, Sender: &notifications.Sender{Client: client, Service: service}, Log: log}
	go func() { defer close(done); background.Run(workerCtx) }()
	timer := time.NewTimer(10 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	waiting := true
	for waiting {
		select {
		case <-timer.C:
			stop()
			<-done
			t.Fatal("notification was not delivered")
		case <-ticker.C:
			waiting = deliveries.Load() == 0
		}
	}
	timer.Stop()
	ticker.Stop()
	stop()
	<-done
	call("POST", "/alert/mute/"+alert+"?nextExecutionTime=00:10:00", editor, ns, nil, 200)
	if a, _, err := usecases.Read[entity.Alert](ctx, store, ns, "alert", alert); err != nil || a.Status != "Muted" {
		t.Fatalf("mute failed: %+v %v", a, err)
	}
	for _, test := range []struct{ name, kind, settings string }{{"logs", "elasticsearch", fmt.Sprintf(`{"clusterUrl":%q,"indexPattern":"logs-*"}`, providerServer.URL)}, {"traces", "jaeger", connection}} {
		call("POST", "/data-source/add", adminToken, ns, map[string]any{"uniqueName": test.name, "dataSourceType": test.kind, "connectionString": test.settings, "purpose": test.name}, 200)
	}
	call("POST", "/logs/search", viewer, ns, map[string]any{"dataSourceName": "logs", "page": 1, "pageSize": 10}, 200)
	call("GET", "/traces/services?dataSourceName=traces", viewer, ns, nil, 200)
	call("GET", "/traces/operations?dataSourceName=traces&serviceName=api", viewer, ns, nil, 200)
	call("GET", "/traces/abc?dataSourceName=traces", viewer, ns, nil, 200)
	call("POST", "/traces/search", viewer, ns, map[string]any{"dataSourceName": "traces", "service": "api", "start": "2025-01-01T00:00:00Z", "end": "2025-01-01T00:01:00Z", "limit": 10}, 200)
	plugin := id(call("POST", "/plugin/add", adminToken, "", map[string]any{"name": "custom", "version": "1", "author": "test", "endpoint": providerServer.URL + "/plugin"}, 200))
	call("PUT", "/plugin/update/"+plugin, adminToken, "", map[string]any{"description": "Custom provider", "settings": "{}"}, 200)
	call("POST", "/plugin/"+plugin+"?isActive=true", adminToken, "", nil, 200)
	call("POST", "/data-source-type/add", adminToken, "", map[string]any{"name": "custom"}, 200)
	call("POST", "/data-source/add", adminToken, ns, map[string]any{"uniqueName": "custom-metrics", "dataSourceType": "custom", "connectionString": connection, "purpose": "metrics"}, 200)
	call("GET", "/metrics/query?DataSourceName=custom-metrics&Query=up", viewer, ns, nil, 200)
	call("DELETE", "/debug/rollback/10", adminToken, "", nil, 404)
	call("GET", "/meta/database-schema", adminToken, "", nil, 404)
	call("GET", "/meta/sql-queries", adminToken, "", nil, 404)
	call("DELETE", "/dashboard/remove/"+dashboard, editor, ns, nil, 200)
	call("GET", "/panel/"+panel, viewer, ns, nil, 404)
	call("GET", "/variable/"+variable, viewer, ns, nil, 404)
}
