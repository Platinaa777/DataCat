package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
)

type MetricsQuery struct {
	DataSource, Dashboard, Query string
	RequestedNamespace           string
	From, To, Step               string
}

func (q *Queries) Metrics(ctx context.Context, ns string, input MetricsQuery, isRange bool) (any, error) {
	if input.RequestedNamespace != "" && input.RequestedNamespace != ns {
		return nil, entity.ErrForbidden
	}
	request, err := json.Marshal(struct {
		Namespace string
		Range     bool
		Input     MetricsQuery
	}{ns, isRange, input})
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(request)
	key := hex.EncodeToString(hash[:])
	if q.cache != nil {
		if data, ok := q.cache.Get(ctx, key); ok {
			return json.RawMessage(data), nil
		}
	}
	result, err := q.metrics(ctx, ns, input, isRange)
	if err != nil {
		return nil, err
	}
	if q.cache != nil {
		data, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		q.cache.Set(ctx, key, data, 5*time.Second)
	}
	return result, nil
}

func (q *Queries) metrics(ctx context.Context, ns string, input MetricsQuery, isRange bool) (any, error) {
	source, err := q.source(ctx, ns, input.DataSource, "Metrics")
	if err != nil {
		return nil, err
	}
	query, err := q.service.ExpandQuery(ctx, ns, input.Dashboard, input.Query)
	if err != nil {
		return nil, err
	}
	if err = Required(query); err != nil {
		return nil, err
	}
	if !strings.EqualFold(source.Type, "Prometheus") {
		operation := "metrics.query"
		if isRange {
			operation = "metrics.queryRange"
		}
		return q.plugin(ctx, source, operation, map[string]string{"query": query, "from": input.From, "to": input.To, "step": input.Step})
	}
	if !isRange {
		result, err := q.provider.Instant(ctx, source, query)
		return result, upstream(err)
	}
	from, err := time.Parse(time.RFC3339Nano, input.From)
	if err != nil {
		return nil, fmt.Errorf("%w: from must be RFC3339", entity.ErrInvalid)
	}
	to, err := time.Parse(time.RFC3339Nano, input.To)
	if err != nil {
		return nil, fmt.Errorf("%w: to must be RFC3339", entity.ErrInvalid)
	}
	step := entity.Duration(time.Minute)
	if input.Step != "" {
		step, err = entity.ParseDuration(input.Step)
		if err != nil {
			return nil, err
		}
	}
	result, err := q.provider.Range(ctx, source, query, from, to, time.Duration(step))
	return result, upstream(err)
}
