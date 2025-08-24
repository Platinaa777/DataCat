package grpcapi

import (
	"context"
	"strings"
	"time"

	metricsv1 "github.com/Platinaa777/DataCat/api/gen/metrics/v1"
	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/repository/postgres"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Metrics struct {
	metricsv1.UnimplementedDataCatMetricExporterServer
	Auth  *auth.Auth
	Store *postgres.Store
}

func New(a *auth.Auth, store *postgres.Store) *grpc.Server {
	s := grpc.NewServer(grpc.MaxRecvMsgSize(4<<20), grpc.MaxConcurrentStreams(100), grpc.ConnectionTimeout(5*time.Second))
	metricsv1.RegisterDataCatMetricExporterServer(s, &Metrics{Auth: a, Store: store})
	return s
}
func (s *Metrics) SendMetrics(ctx context.Context, req *metricsv1.SendMetricsRequest) (*emptypb.Empty, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	md, _ := metadata.FromIncomingContext(ctx)
	token := first(md.Get("authorization"))
	ns := first(md.Get("x-namespace"))
	p, err := s.Auth.Verify(ctx, strings.TrimPrefix(token, "Bearer "))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	if _, err = uuid.Parse(ns); err != nil {
		return nil, status.Error(codes.InvalidArgument, "x-namespace must be a UUID")
	}
	if !p.Can(ns, true) {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}
	if len(req.Metrics) == 0 || len(req.Metrics) > 1000 {
		return nil, status.Error(codes.InvalidArgument, "batch must contain 1..1000 metrics")
	}
	for _, m := range req.Metrics {
		if m == nil || m.DataPoints == nil || len(m.Tags) > 100 {
			return nil, status.Error(codes.InvalidArgument, "invalid metric")
		}
		points := 0
		switch v := m.DataPoints.(type) {
		case *metricsv1.Metric_Counter:
			if v.Counter != nil {
				points = len(v.Counter.DataPoints)
				for _, p := range v.Counter.DataPoints {
					if p == nil || p.Timestamp.CheckValid() != nil || p.Value < 0 {
						return nil, status.Error(codes.InvalidArgument, "invalid counter point")
					}
				}
			}
		case *metricsv1.Metric_Gauge:
			if v.Gauge != nil {
				points = len(v.Gauge.DataPoints)
				for _, p := range v.Gauge.DataPoints {
					if p == nil || p.Timestamp.CheckValid() != nil {
						return nil, status.Error(codes.InvalidArgument, "invalid gauge point")
					}
				}
			}
		case *metricsv1.Metric_Histogram:
			if v.Histogram != nil {
				points = len(v.Histogram.DataPoints)
				for _, p := range v.Histogram.DataPoints {
					if p == nil || p.StartTimestamp.CheckValid() != nil || p.EndTimestamp.CheckValid() != nil || p.Count < 0 {
						return nil, status.Error(codes.InvalidArgument, "invalid histogram point")
					}
				}
			}
		}
		if points < 1 || points > 10000 {
			return nil, status.Error(codes.InvalidArgument, "metric must contain 1..10000 points")
		}
	}
	data, err := protojson.Marshal(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid metric payload")
	}
	tag, err := s.Store.Pool().Exec(ctx, `INSERT INTO datacat.metric_batches(namespace,data) SELECT $1,$2 WHERE EXISTS(SELECT 1 FROM datacat.resources WHERE namespace='global' AND kind='namespace' AND id=$1)`, ns, data)
	if err != nil {
		return nil, status.Error(codes.Internal, "metric persistence failed")
	}
	if tag.RowsAffected() != 1 {
		return nil, status.Error(codes.NotFound, "namespace not found")
	}
	return &emptypb.Empty{}, nil
}
func first(values []string) string {
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
