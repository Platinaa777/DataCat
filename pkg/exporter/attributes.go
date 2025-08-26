package exporter

import (
	metricsv1 "github.com/Platinaa777/DataCat/api/gen/metrics/v1"
	"go.opentelemetry.io/otel/attribute"
)

func attributes(set attribute.Set) []*metricsv1.Tag {
	out := make([]*metricsv1.Tag, 0, set.Len())
	iter := set.Iter()
	for iter.Next() {
		kv := iter.Attribute()
		out = append(out, &metricsv1.Tag{Key: string(kv.Key), Value: kv.Value.Emit()})
	}
	return out
}
