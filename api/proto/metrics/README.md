# Metric protocol

`v1/metrics.proto` defines the DataCat gRPC metric receiver. Generate Go bindings
with `make generate` from the repository root. Clients send bearer authentication
and namespace metadata on every request.

A request contains 1–1000 metrics. Each metric has one counter, gauge, or
histogram payload with 1–10000 points. The server limits messages to 4 MiB.
The `__name__` tag identifies the metric. Counters use integer values; histogram
points carry count, sum, min, max, and interval timestamps.

The Go OpenTelemetry exporter is available in `pkg/exporter`. It does not own the
gRPC connection; callers configure TLS and close the connection themselves.
