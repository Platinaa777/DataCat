package usecases

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/Platinaa777/DataCat/internal/entity"
)

type rejectEndpoint struct{}

func (rejectEndpoint) Validate(string) (*url.URL, error) { return nil, errors.New("host not allowed") }

func TestApplicationValidatesEndpointsBeforePersistence(t *testing.T) {
	service := &Service{Endpoints: rejectEndpoint{}}
	ctx := context.Background()
	for _, test := range []struct {
		name string
		run  func() error
	}{
		{"create source", func() error {
			_, err := service.CreateDataSource(ctx, "team", entity.DataSource{Name: "prom", Type: "prometheus", Purpose: "Metrics", ConnectionString: `{"serverUrl":"https://blocked.test"}`})
			return err
		}},
		{"update source", func() error {
			return service.UpdateDataSourceConnection(ctx, "team", "prom", `{"serverUrl":"https://blocked.test"}`)
		}},
		{"create channel", func() error {
			_, err := service.CreateChannel(ctx, "team", "ops", "webhook", `{"url":"https://blocked.test"}`)
			return err
		}},
		{"update channel", func() error {
			return service.UpdateChannel(ctx, "team", "1", "webhook", `{"url":"https://blocked.test"}`)
		}},
		{"create plugin", func() error {
			_, err := service.CreatePlugin(ctx, entity.Plugin{Name: "custom", Version: "1", Author: "test", Endpoint: "https://blocked.test"})
			return err
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); !errors.Is(err, entity.ErrInvalid) {
				t.Fatalf("endpoint validation was bypassed: %v", err)
			}
		})
	}
}
