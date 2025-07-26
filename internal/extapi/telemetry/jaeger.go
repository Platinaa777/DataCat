package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
)

type jaegerTag struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}
type jaegerTrace struct {
	TraceID string `json:"traceID"`
	Spans   []struct {
		TraceID       string      `json:"traceID"`
		SpanID        string      `json:"spanID"`
		OperationName string      `json:"operationName"`
		StartTime     int64       `json:"startTime"`
		Duration      int64       `json:"duration"`
		Tags          []jaegerTag `json:"tags"`
		ProcessID     string      `json:"processID"`
		References    []struct {
			TraceID string `json:"traceID"`
			SpanID  string `json:"spanID"`
			RefType string `json:"refType"`
		} `json:"references"`
	} `json:"spans"`
	Processes map[string]struct {
		ServiceName string      `json:"serviceName"`
		Tags        []jaegerTag `json:"tags"`
	} `json:"processes"`
}

func tagMap(tags []jaegerTag) map[string]any {
	out := map[string]any{}
	for _, t := range tags {
		out[t.Key] = t.Value
	}
	return out
}
func (p *Provider) TraceNames(ctx context.Context, d entity.DataSource, service string) ([]string, error) {
	cfg, headers, err := p.connection(d)
	if err != nil {
		return nil, err
	}
	path := "/api/services"
	if service != "" {
		path = "/api/services/" + url.PathEscape(service) + "/operations"
	}
	var response struct {
		Data []string `json:"data"`
	}
	err = p.Client.Do(ctx, "GET", endpoint(cfg.ServerURL, path, nil), nil, headers, &response)
	if response.Data == nil {
		response.Data = []string{}
	}
	return response.Data, err
}
func (p *Provider) Traces(ctx context.Context, d entity.DataSource, q entity.TraceSearch, id string) ([]entity.TraceEntry, error) {
	cfg, headers, err := p.connection(d)
	if err != nil {
		return nil, err
	}
	path := "/api/traces"
	params := url.Values{}
	if id != "" {
		path += "/" + url.PathEscape(id)
	} else {
		if q.End.Before(q.Start) || q.Service == "" || q.Limit < 1 || q.Limit > 1000 {
			return nil, fmt.Errorf("%w: invalid trace search", entity.ErrInvalid)
		}
		params.Set("service", q.Service)
		params.Set("start", strconv.FormatInt(q.Start.UnixMicro(), 10))
		params.Set("end", strconv.FormatInt(q.End.UnixMicro(), 10))
		params.Set("limit", strconv.Itoa(q.Limit))
		if q.Operation != "" {
			params.Set("operation", q.Operation)
		}
		if q.MinDuration != "" {
			params.Set("minDuration", q.MinDuration)
		}
		if q.MaxDuration != "" {
			params.Set("maxDuration", q.MaxDuration)
		}
		if len(q.Tags) > 0 {
			b, e := json.Marshal(q.Tags)
			if e != nil {
				return nil, e
			}
			params.Set("tags", string(b))
		}
	}
	var response struct {
		Data []jaegerTrace `json:"data"`
	}
	if err = p.Client.Do(ctx, "GET", endpoint(cfg.ServerURL, path, params), nil, headers, &response); err != nil {
		return nil, err
	}
	out := make([]entity.TraceEntry, 0, len(response.Data))
	for _, trace := range response.Data {
		t := entity.TraceEntry{TraceID: trace.TraceID, Spans: []entity.SpanEntry{}, Processes: map[string]entity.ProcessEntry{}}
		for key, process := range trace.Processes {
			t.Processes[key] = entity.ProcessEntry{ServiceName: process.ServiceName, Tags: tagMap(process.Tags)}
		}
		for _, span := range trace.Spans {
			refs := []entity.SpanReference{}
			for _, r := range span.References {
				kind := "other"
				if r.RefType == "FOLLOWS_FROM" {
					kind = "followsFrom"
				} else if r.RefType == "CHILD_OF" {
					kind = "childOf"
				}
				refs = append(refs, entity.SpanReference{TraceID: r.TraceID, SpanID: r.SpanID, ReferenceType: kind})
			}
			t.Spans = append(t.Spans, entity.SpanEntry{TraceID: span.TraceID, SpanID: span.SpanID, OperationName: span.OperationName, StartTime: time.UnixMicro(span.StartTime).UTC(), Timestamp: span.StartTime, Duration: entity.Duration(time.Duration(span.Duration) * time.Microsecond), Tags: tagMap(span.Tags), References: refs, ProcessID: span.ProcessID})
		}
		out = append(out, t)
	}
	return out, nil
}
