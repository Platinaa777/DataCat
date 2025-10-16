//go:build integration

package grpcapi

import (
	"bytes"
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	metricsv1 "github.com/Platinaa777/DataCat/api/gen/metrics/v1"
	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/config"
	"github.com/Platinaa777/DataCat/internal/pkg/secrets"
	"github.com/Platinaa777/DataCat/internal/repository/postgres"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMetricIngestionAuthenticationAndPersistence(t *testing.T) {
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
	name := "datacat_grpc_" + strings.ReplaceAll(uuid.NewString(), "-", "") + "_test"
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
	if err = store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	cipher, _ := secrets.New(bytes.Repeat([]byte{1}, 32))
	service := usecases.New(store, cipher)
	if err = service.Seed(ctx); err != nil {
		t.Fatal(err)
	}
	token := strings.Repeat("test-token", 4)
	identity, err := auth.New(ctx, config.Config{DevToken: token}, service)
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(4 << 20)
	server := New(identity, store)
	go server.Serve(listener)
	defer server.Stop()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := metricsv1.NewDataCatMetricExporterClient(conn)
	req := &metricsv1.SendMetricsRequest{Metrics: []*metricsv1.Metric{{Tags: []*metricsv1.Tag{{Key: "__name__", Value: "temperature"}}, DataPoints: &metricsv1.Metric_Gauge{Gauge: &metricsv1.GaugeDataPoints{DataPoints: []*metricsv1.GaugeDataPoint{{Value: 23.5, Timestamp: timestamppb.Now()}}}}}}}
	if _, err = client.SendMetrics(ctx, req); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("missing authentication: %v", err)
	}
	metadataCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token, "x-namespace", "00000000-0000-0000-0000-000000000000")
	if _, err = client.SendMetrics(metadataCtx, req); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = store.Pool().QueryRow(ctx, "SELECT count(*) FROM datacat.metric_batches").Scan(&count); err != nil || count != 1 {
		t.Fatalf("batch not persisted: %d %v", count, err)
	}
	if _, err = client.SendMetrics(metadataCtx, &metricsv1.SendMetricsRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("empty batch accepted: %v", err)
	}
	unknown := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token, "x-namespace", uuid.NewString())
	if _, err = client.SendMetrics(unknown, req); status.Code(err) != codes.NotFound {
		t.Fatalf("unknown namespace accepted: %v", err)
	}
}
