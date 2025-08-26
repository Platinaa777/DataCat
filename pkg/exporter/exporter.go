// Package exporter sends OpenTelemetry metrics to DataCat.
package exporter

import (
	"context"
	"fmt"
	"sync/atomic"

	metricsv1 "github.com/Platinaa777/DataCat/api/gen/metrics/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Exporter struct {
	client           metricsv1.DataCatMetricExporterClient
	namespace, token string
	closed           atomic.Bool
}

func New(conn grpc.ClientConnInterface, namespace, token string) *Exporter {
	return &Exporter{client: metricsv1.NewDataCatMetricExporterClient(conn), namespace: namespace, token: token}
}
func (e *Exporter) Temporality(metric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}
func (e *Exporter) Aggregation(kind metric.InstrumentKind) metric.Aggregation {
	return metric.DefaultAggregationSelector(kind)
}
func (e *Exporter) ForceFlush(context.Context) error { return nil }
func (e *Exporter) Shutdown(context.Context) error   { e.closed.Store(true); return nil }
func (e *Exporter) Export(ctx context.Context, resource *metricdata.ResourceMetrics) error {
	if e.closed.Load() {
		return fmt.Errorf("exporter is shut down")
	}
	if resource == nil {
		return fmt.Errorf("resource metrics are required")
	}
	resourceTags := []*metricsv1.Tag{}
	if resource.Resource != nil {
		resourceTags = attributes(*resource.Resource.Set())
	}
	batch := []*metricsv1.Metric{}
	for _, scope := range resource.ScopeMetrics {
		for _, m := range scope.Metrics {
			base := append([]*metricsv1.Tag(nil), resourceTags...)
			base = append(base, &metricsv1.Tag{Key: "__name__", Value: m.Name}, &metricsv1.Tag{Key: "unit", Value: m.Unit}, &metricsv1.Tag{Key: "otel.scope.name", Value: scope.Scope.Name})
			points, err := convert(m.Data, base)
			if err != nil {
				return fmt.Errorf("metric %s: %w", m.Name, err)
			}
			batch = append(batch, points...)
		}
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+e.token, "x-namespace", e.namespace)
	for start := 0; start < len(batch); start += 1000 {
		end := min(start+1000, len(batch))
		if _, err := e.client.SendMetrics(ctx, &metricsv1.SendMetricsRequest{Metrics: batch[start:end]}); err != nil {
			return fmt.Errorf("export metrics: %w", err)
		}
	}
	return nil
}
func tags(base []*metricsv1.Tag, set attribute.Set) []*metricsv1.Tag {
	out := append([]*metricsv1.Tag(nil), base...)
	return append(out, attributes(set)...)
}
func convert(data metricdata.Aggregation, base []*metricsv1.Tag) ([]*metricsv1.Metric, error) {
	switch d := data.(type) {
	case metricdata.Sum[int64]:
		if !d.IsMonotonic {
			return gauges(d.DataPoints, base), nil
		}
		out := make([]*metricsv1.Metric, 0, len(d.DataPoints))
		for _, p := range d.DataPoints {
			out = append(out, &metricsv1.Metric{Tags: tags(base, p.Attributes), DataPoints: &metricsv1.Metric_Counter{Counter: &metricsv1.CounterDataPoints{DataPoints: []*metricsv1.CounterDataPoint{{Value: p.Value, Timestamp: timestamppb.New(p.Time)}}}}})
		}
		return out, nil
	case metricdata.Sum[float64]:
		return gauges(d.DataPoints, base), nil
	case metricdata.Gauge[int64]:
		return gauges(d.DataPoints, base), nil
	case metricdata.Gauge[float64]:
		return gauges(d.DataPoints, base), nil
	case metricdata.Histogram[int64]:
		return histograms(d.DataPoints, base), nil
	case metricdata.Histogram[float64]:
		return histograms(d.DataPoints, base), nil
	default:
		return nil, fmt.Errorf("unsupported aggregation")
	}
}

type number interface{ int64 | float64 }

func gauges[T number](points []metricdata.DataPoint[T], base []*metricsv1.Tag) []*metricsv1.Metric {
	out := make([]*metricsv1.Metric, 0, len(points))
	for _, p := range points {
		out = append(out, &metricsv1.Metric{Tags: tags(base, p.Attributes), DataPoints: &metricsv1.Metric_Gauge{Gauge: &metricsv1.GaugeDataPoints{DataPoints: []*metricsv1.GaugeDataPoint{{Value: float64(p.Value), Timestamp: timestamppb.New(p.Time)}}}}})
	}
	return out
}
func histograms[T number](points []metricdata.HistogramDataPoint[T], base []*metricsv1.Tag) []*metricsv1.Metric {
	out := make([]*metricsv1.Metric, 0, len(points))
	for _, p := range points {
		low, _ := p.Min.Value()
		high, _ := p.Max.Value()
		out = append(out, &metricsv1.Metric{Tags: tags(base, p.Attributes), DataPoints: &metricsv1.Metric_Histogram{Histogram: &metricsv1.HistogramDataPoints{DataPoints: []*metricsv1.HistogramDataPoint{{Count: int64(p.Count), StartTimestamp: timestamppb.New(p.StartTime), EndTimestamp: timestamppb.New(p.Time), Sum: float64(p.Sum), Min: float64(low), Max: float64(high)}}}}})
	}
	return out
}

var _ metric.Exporter = (*Exporter)(nil)
