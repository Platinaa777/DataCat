package query

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

func (s *Handler) source(r *http.Request, ns, name, purpose string) (entity.DataSource, error) {
	d, err := s.service.DataSource(r.Context(), ns, name)
	if err != nil {
		return d, err
	}
	if d.Purpose != purpose {
		return d, fmt.Errorf("%w: wrong data source purpose", entity.ErrInvalid)
	}
	return d, nil
}

func (s *Handler) plugin(r *http.Request, d entity.DataSource, operation string, input any) (any, error) {
	p, _, err := usecases.Read[entity.Plugin](r.Context(), s.service.Store, usecases.Global, "plugin", d.Type)
	if err != nil {
		return nil, err
	}
	if !p.IsActive {
		return nil, fmt.Errorf("%w: plugin is disabled", entity.ErrInvalid)
	}
	var result json.RawMessage
	err = s.provider.Client.Do(r.Context(), "POST", p.Endpoint, map[string]any{"version": 1, "operation": operation, "namespaceId": d.NamespaceID, "connectionSettings": json.RawMessage(d.ConnectionString), "settings": p.Settings, "input": input}, nil, &result)
	return result, transport.Upstream(err)
}
