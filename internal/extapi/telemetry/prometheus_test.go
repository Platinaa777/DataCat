package telemetry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/httpclient"
)

func TestPrometheusPreservesLabelsAndRejectsFailedQueries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") == "bad" {
			w.Write([]byte(`{"status":"error"}`))
			return
		}
		w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"requests","service":"api"},"value":[1735689600.125,"7"]}]}}`))
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	p := Provider{Client: httpclient.New([]string{u.Host})}
	b, _ := json.Marshal(Settings{ServerURL: server.URL})
	d := entity.DataSource{ConnectionString: string(b)}
	points, err := p.Instant(context.Background(), d, "up")
	if err != nil || len(points) != 1 {
		t.Fatalf("query: %v %v", points, err)
	}
	if points[0].Value != 7 || points[0].MetricName != "requests" || points[0].Labels["service"] != "api" || points[0].Timestamp.UnixMilli() != 1735689600125 {
		t.Fatalf("bad sample: %+v", points[0])
	}
	if _, err = p.Instant(context.Background(), d, "bad"); err == nil {
		t.Fatal("query error was swallowed")
	}
}
