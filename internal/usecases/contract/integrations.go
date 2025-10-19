package contract

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
)

type EndpointPolicy interface {
	Validate(string) (*url.URL, error)
}
type Telemetry interface {
	Instant(context.Context, entity.DataSource, string) ([]entity.MetricPoint, error)
	Range(context.Context, entity.DataSource, string, time.Time, time.Time, time.Duration) ([]entity.TimeSeries, error)
	Logs(context.Context, entity.DataSource, entity.LogSearch) (entity.Page[entity.LogEntry], error)
	Traces(context.Context, entity.DataSource, entity.TraceSearch, string) ([]entity.TraceEntry, error)
	TraceNames(context.Context, entity.DataSource, string) ([]string, error)
}
type Plugins interface {
	Call(context.Context, entity.Plugin, entity.DataSource, string, any) (json.RawMessage, error)
}
type QueryCache interface {
	Get(context.Context, string) ([]byte, bool)
	Set(context.Context, string, []byte, time.Duration)
}
