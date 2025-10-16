package exporter

import (
	"context"
	"testing"
	"time"

	metricsv1 "github.com/Platinaa777/DataCat/api/gen/metrics/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type receiver struct {
	batches  []*metricsv1.SendMetricsRequest
	metadata metadata.MD
}

func (r *receiver) SendMetrics(ctx context.Context, request *metricsv1.SendMetricsRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	r.batches = append(r.batches, request)
	r.metadata, _ = metadata.FromOutgoingContext(ctx)
	return &emptypb.Empty{}, nil
}
func TestExportKeepsAttributeSeriesSeparate(t *testing.T) {
	r := &receiver{}
	e := &Exporter{client: r, namespace: "tenant-a", token: "token"}
	now := time.Now()
	input := &metricdata.ResourceMetrics{ScopeMetrics: []metricdata.ScopeMetrics{{Metrics: []metricdata.Metrics{{Name: "requests", Data: metricdata.Sum[int64]{IsMonotonic: true, DataPoints: []metricdata.DataPoint[int64]{{Value: 2, Time: now, Attributes: attribute.NewSet(attribute.String("service", "api"))}, {Value: 3, Time: now, Attributes: attribute.NewSet(attribute.String("service", "worker"))}}}}}}}}
	if err := e.Export(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if len(r.batches) != 1 || len(r.batches[0].Metrics) != 2 {
		t.Fatalf("attribute series were merged: %+v", r.batches)
	}
	for i, want := range []string{"api", "worker"} {
		m := r.batches[0].Metrics[i]
		found := 0
		for _, tag := range m.Tags {
			if tag.Key == "service" {
				found++
				if tag.Value != want {
					t.Fatal("wrong series label")
				}
			}
		}
		if found != 1 || m.GetCounter().DataPoints[0].Value != int64(i+2) {
			t.Fatal("counter or label changed")
		}
	}
	if r.metadata.Get("x-namespace")[0] != "tenant-a" || r.metadata.Get("authorization")[0] != "Bearer token" {
		t.Fatal("missing authentication metadata")
	}
	e.Shutdown(context.Background())
	if err := e.Export(context.Background(), input); err == nil {
		t.Fatal("export after shutdown succeeded")
	}
}
