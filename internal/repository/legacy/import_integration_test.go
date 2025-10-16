//go:build integration

package legacy

import (
	"bytes"
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/pkg/secrets"
	"github.com/Platinaa777/DataCat/internal/repository/postgres"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestImportPreservesLegacyConfiguration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is required")
	}
	u, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("test database must end in _test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(context.Background())
	name := "datacat_legacy_" + strings.ReplaceAll(uuid.NewString(), "-", "") + "_test"
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec(context.Background(), "DROP DATABASE "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)")
	u.Path = "/" + name
	store, err := postgres.Open(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	fixture, err := os.ReadFile("testdata/legacy.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.Pool().Exec(ctx, string(fixture)); err != nil {
		t.Fatal(err)
	}
	if err = store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	cipher, _ := secrets.New(bytes.Repeat([]byte{3}, 32))
	counts, err := Import(ctx, u.String(), store, cipher)
	if err != nil {
		t.Fatal(err)
	}
	if counts["dashboard"] != 1 || counts["alert"] != 1 {
		t.Fatalf("bad import counts: %+v", counts)
	}
	service := usecases.New(store, cipher)
	d, err := service.DataSource(ctx, defaultNamespace, "source-1")
	if err != nil || !strings.Contains(d.ConnectionString, "localhost:9090") {
		t.Fatalf("connection import: %+v %v", d, err)
	}
	alert, _, err := usecases.Read[entity.Alert](ctx, store, defaultNamespace, "alert", "alert-1")
	if err != nil || alert.RepeatInterval != entity.Duration(time.Minute) || alert.DataSource.ID != "source-1" || alert.Status != "Muted" {
		t.Fatalf("alert import: %+v %v", alert, err)
	}
	roles, _, err := usecases.Read[[]entity.Claim](ctx, store, usecases.Global, "user-claims", "oidc-subject")
	if err != nil || len(roles) != 1 || roles[0].Role != "Editor" {
		t.Fatalf("manual role import: %+v %v", roles, err)
	}
	if _, err = Import(ctx, u.String(), store, cipher); err == nil {
		t.Fatal("nonempty target accepted")
	}
	var n int
	if err = store.Pool().QueryRow(ctx, "SELECT count(*) FROM public.alerts").Scan(&n); err != nil || n != 1 {
		t.Fatal("legacy source changed")
	}
}
