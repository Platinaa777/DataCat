package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases/contract"
)

type Queries struct {
	service  *Service
	provider contract.Telemetry
	plugins  contract.Plugins
	cache    contract.QueryCache
}

func NewQueries(service *Service, provider contract.Telemetry, plugins contract.Plugins, cache contract.QueryCache) *Queries {
	return &Queries{service: service, provider: provider, plugins: plugins, cache: cache}
}

func (q *Queries) source(ctx context.Context, ns, name, purpose string) (entity.DataSource, error) {
	source, err := q.service.DataSource(ctx, ns, name)
	if err != nil {
		return source, err
	}
	if source.Purpose != purpose {
		return source, fmt.Errorf("%w: wrong data source purpose", entity.ErrInvalid)
	}
	return source, nil
}

func (q *Queries) plugin(ctx context.Context, source entity.DataSource, operation string, input any) (json.RawMessage, error) {
	plugin, err := Resource[entity.Plugin](ctx, q.service, Global, "plugin", source.Type)
	if err != nil {
		return nil, err
	}
	if !plugin.IsActive {
		return nil, fmt.Errorf("%w: plugin is disabled", entity.ErrInvalid)
	}
	result, err := q.plugins.Call(ctx, plugin, source, operation, input)
	return result, upstream(err)
}

func (q *Queries) Logs(ctx context.Context, ns string, input entity.LogSearch) (any, error) {
	source, err := q.source(ctx, ns, input.DataSourceName, "Logs")
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(source.Type, "Elasticsearch") {
		return q.plugin(ctx, source, "logs.search", input)
	}
	result, err := q.provider.Logs(ctx, source, input)
	return result, upstream(err)
}

func (q *Queries) TraceServices(ctx context.Context, ns, source string) (any, error) {
	return q.traceNames(ctx, ns, source, "")
}

func (q *Queries) TraceOperations(ctx context.Context, ns, source, service string) (any, error) {
	if service == "" {
		return nil, entity.ErrInvalid
	}
	return q.traceNames(ctx, ns, source, service)
}

func (q *Queries) traceNames(ctx context.Context, ns, name, service string) (any, error) {
	source, err := q.source(ctx, ns, name, "Traces")
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(source.Type, "Jaeger") {
		return q.plugin(ctx, source, "traces.names", map[string]string{"service": service})
	}
	result, err := q.provider.TraceNames(ctx, source, service)
	return result, upstream(err)
}

func (q *Queries) Trace(ctx context.Context, ns, name, id string) (any, error) {
	source, err := q.source(ctx, ns, name, "Traces")
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(source.Type, "Jaeger") {
		return q.plugin(ctx, source, "traces.get", map[string]string{"traceId": id})
	}
	traces, err := q.provider.Traces(ctx, source, entity.TraceSearch{}, id)
	if err != nil {
		return nil, upstream(err)
	}
	if len(traces) == 0 {
		return nil, entity.ErrNotFound
	}
	return traces[0], nil
}

func (q *Queries) SearchTraces(ctx context.Context, ns string, input entity.TraceSearch) (any, error) {
	source, err := q.source(ctx, ns, input.DataSourceName, "Traces")
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(source.Type, "Jaeger") {
		return q.plugin(ctx, source, "traces.search", input)
	}
	result, err := q.provider.Traces(ctx, source, input, "")
	return result, upstream(err)
}

func upstream(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", entity.ErrUpstream, err)
}
