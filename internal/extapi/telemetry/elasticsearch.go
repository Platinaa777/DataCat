package telemetry

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
)

func (p *Provider) Logs(ctx context.Context, d entity.DataSource, q entity.LogSearch) (entity.Page[entity.LogEntry], error) {
	if q.Page < 1 || q.PageSize < 1 || q.PageSize > 1000 || q.Page > 10000/q.PageSize {
		return entity.Page[entity.LogEntry]{}, fmt.Errorf("%w: log pagination exceeds 10000 results", entity.ErrInvalid)
	}
	cfg, headers, err := p.connection(d)
	if err != nil {
		return entity.Page[entity.LogEntry]{}, err
	}
	if cfg.IndexPattern == "" || strings.ContainsAny(cfg.IndexPattern, "/?#") {
		return entity.Page[entity.LogEntry]{}, fmt.Errorf("invalid Elasticsearch index pattern")
	}
	filters := []any{}
	for key, value := range map[string]string{"fields.TraceId": q.TraceID, "level": q.Severity, "fields.Host": q.ServiceName} {
		if value != "" {
			filters = append(filters, map[string]any{"term": map[string]string{key: value}})
		}
	}
	for key, value := range q.CustomFilters {
		filters = append(filters, map[string]any{"term": map[string]string{"fields." + key: value}})
	}
	bounds := map[string]string{}
	if q.From != nil {
		bounds["gte"] = q.From.Format(time.RFC3339Nano)
	}
	if q.To != nil {
		bounds["lte"] = q.To.Format(time.RFC3339Nano)
	}
	if len(bounds) > 0 {
		filters = append(filters, map[string]any{"range": map[string]any{"@timestamp": bounds}})
	}
	field := q.SortField
	if field == "" || field == "timestamp" {
		field = "@timestamp"
	}
	if field != "@timestamp" && field != "level.keyword" && field != "fields.Host.keyword" {
		return entity.Page[entity.LogEntry]{}, fmt.Errorf("%w: unsupported log sort field", entity.ErrInvalid)
	}
	order := "desc"
	if q.SortAscending {
		order = "asc"
	}
	body := map[string]any{"from": (q.Page - 1) * q.PageSize, "size": q.PageSize, "track_total_hits": true, "query": map[string]any{"bool": map[string]any{"filter": filters}}, "sort": []any{map[string]string{field: order}}}
	var response struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source struct {
					Timestamp time.Time      `json:"@timestamp"`
					Level     string         `json:"level"`
					Message   string         `json:"message"`
					Fields    map[string]any `json:"fields"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err = p.Client.Do(ctx, "POST", endpoint(cfg.ServerURL, "/"+url.PathEscape(cfg.IndexPattern)+"/_search", nil), body, headers, &response); err != nil {
		return entity.Page[entity.LogEntry]{}, err
	}
	out := make([]entity.LogEntry, 0, len(response.Hits.Hits))
	for _, hit := range response.Hits.Hits {
		d := hit.Source
		host, _ := d.Fields["Host"].(string)
		trace, _ := d.Fields["TraceId"].(string)
		if d.Fields == nil {
			d.Fields = map[string]any{}
		}
		out = append(out, entity.LogEntry{Timestamp: d.Timestamp, Message: d.Message, Severity: d.Level, ServiceName: host, TraceID: trace, AdditionalFields: d.Fields})
	}
	return entity.NewPage(out, response.Hits.Total.Value, q.Page, q.PageSize), nil
}
