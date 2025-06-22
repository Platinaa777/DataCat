package entity

import "time"

type MetricPoint struct {
	MetricName string            `json:"metricName"`
	Value      float64           `json:"value"`
	Timestamp  time.Time         `json:"timestamp"`
	Labels     map[string]string `json:"labels"`
}
type TimeSeries struct {
	MetricName string            `json:"metricName"`
	Labels     map[string]string `json:"labels"`
	Points     []MetricPoint     `json:"points"`
}
type LogEntry struct {
	Timestamp        time.Time      `json:"timestamp"`
	Message          string         `json:"message"`
	Severity         string         `json:"severity"`
	ServiceName      string         `json:"serviceName"`
	TraceID          string         `json:"traceId"`
	AdditionalFields map[string]any `json:"additionalFields"`
}
type LogSearch struct {
	DataSourceName string            `json:"dataSourceName"`
	TraceID        string            `json:"traceId"`
	From           *time.Time        `json:"from"`
	To             *time.Time        `json:"to"`
	Severity       string            `json:"severity"`
	ServiceName    string            `json:"serviceName"`
	CustomFilters  map[string]string `json:"customFilters"`
	PageSize       int               `json:"pageSize"`
	Page           int               `json:"page"`
	SortField      string            `json:"sortField"`
	SortAscending  bool              `json:"sortAscending"`
}
type TraceSearch struct {
	DataSourceName string         `json:"dataSourceName"`
	Service        string         `json:"service"`
	Start          time.Time      `json:"start"`
	End            time.Time      `json:"end"`
	Operation      string         `json:"operation"`
	Limit          int            `json:"limit"`
	MinDuration    string         `json:"minDuration"`
	MaxDuration    string         `json:"maxDuration"`
	Tags           map[string]any `json:"tags"`
}
type SpanReference struct {
	TraceID       string `json:"traceId"`
	SpanID        string `json:"spanId"`
	ReferenceType string `json:"referenceType"`
}
type SpanEntry struct {
	TraceID       string          `json:"traceId"`
	SpanID        string          `json:"spanId"`
	OperationName string          `json:"operationName"`
	StartTime     time.Time       `json:"startTime"`
	Timestamp     int64           `json:"timestamp"`
	Duration      Duration        `json:"duration"`
	Tags          map[string]any  `json:"tags"`
	References    []SpanReference `json:"references"`
	ProcessID     string          `json:"processId"`
}
type ProcessEntry struct {
	ServiceName string         `json:"serviceName"`
	Tags        map[string]any `json:"tags"`
}
type TraceEntry struct {
	TraceID   string                  `json:"traceId"`
	Spans     []SpanEntry             `json:"spans"`
	Processes map[string]ProcessEntry `json:"processes"`
}
