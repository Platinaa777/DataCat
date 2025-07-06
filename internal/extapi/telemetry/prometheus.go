package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
)

type prometheusResult struct {
	Metric map[string]string   `json:"metric"`
	Value  []json.RawMessage   `json:"value"`
	Values [][]json.RawMessage `json:"values"`
}
type prometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string             `json:"resultType"`
		Result     []prometheusResult `json:"result"`
	} `json:"data"`
}

func point(raw []json.RawMessage, labels map[string]string) (entity.MetricPoint, error) {
	if len(raw) != 2 {
		return entity.MetricPoint{}, fmt.Errorf("invalid Prometheus sample")
	}
	var timestamp float64
	var value string
	if json.Unmarshal(raw[0], &timestamp) != nil || json.Unmarshal(raw[1], &value) != nil {
		return entity.MetricPoint{}, fmt.Errorf("invalid Prometheus sample")
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
		return entity.MetricPoint{}, fmt.Errorf("non-finite Prometheus sample")
	}
	name := labels["__name__"]
	if name == "" {
		name = "unnamed"
	}
	seconds, fraction := math.Modf(timestamp)
	return entity.MetricPoint{MetricName: name, Value: number, Timestamp: time.Unix(int64(seconds), int64(math.Round(fraction*1e9))).UTC(), Labels: labels}, nil
}
func (p *Provider) Instant(ctx context.Context, d entity.DataSource, query string) ([]entity.MetricPoint, error) {
	cfg, headers, err := p.connection(d)
	if err != nil {
		return nil, err
	}
	var response prometheusResponse
	if err = p.Client.Do(ctx, "GET", endpoint(cfg.ServerURL, "/api/v1/query", url.Values{"query": {query}}), nil, headers, &response); err != nil {
		return nil, err
	}
	if response.Status != "success" || response.Data.ResultType != "vector" {
		return nil, fmt.Errorf("Prometheus did not return a successful vector query")
	}
	out := make([]entity.MetricPoint, 0, len(response.Data.Result))
	for _, r := range response.Data.Result {
		v, e := point(r.Value, r.Metric)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, nil
}
func (p *Provider) Range(ctx context.Context, d entity.DataSource, query string, from, to time.Time, step time.Duration) ([]entity.TimeSeries, error) {
	if to.Before(from) || step <= 0 || to.Sub(from)/step > 11000 {
		return nil, fmt.Errorf("%w: invalid range or more than 11000 points per series", entity.ErrInvalid)
	}
	cfg, headers, err := p.connection(d)
	if err != nil {
		return nil, err
	}
	q := url.Values{"query": {query}, "start": {from.Format(time.RFC3339Nano)}, "end": {to.Format(time.RFC3339Nano)}, "step": {strconv.FormatFloat(step.Seconds(), 'f', -1, 64)}}
	var response prometheusResponse
	if err = p.Client.Do(ctx, "GET", endpoint(cfg.ServerURL, "/api/v1/query_range", q), nil, headers, &response); err != nil {
		return nil, err
	}
	if response.Status != "success" || response.Data.ResultType != "matrix" {
		return nil, fmt.Errorf("Prometheus did not return a successful matrix query")
	}
	out := make([]entity.TimeSeries, 0, len(response.Data.Result))
	for _, r := range response.Data.Result {
		series := entity.TimeSeries{MetricName: r.Metric["__name__"], Labels: r.Metric, Points: []entity.MetricPoint{}}
		if series.MetricName == "" {
			series.MetricName = "unnamed"
		}
		for _, raw := range r.Values {
			v, e := point(raw, r.Metric)
			if e != nil {
				return nil, e
			}
			series.Points = append(series.Points, v)
		}
		out = append(out, series)
	}
	return out, nil
}
