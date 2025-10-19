package datasource

import (
	"encoding/json"
	"fmt"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
)

func (s *Handler) validateConnection(raw string) error {
	var cfg telemetry.Settings
	if json.Unmarshal([]byte(raw), &cfg) != nil {
		return fmt.Errorf("%w: connectionString must contain JSON", entity.ErrInvalid)
	}
	u := cfg.ServerURL
	if u == "" {
		u = cfg.ClusterURL
	}
	if _, err := s.provider.Client.Validate(u); err != nil {
		return fmt.Errorf("%w: %v", entity.ErrInvalid, err)
	}
	return nil
}
