package plugins

import (
	"context"
	"encoding/json"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/httpclient"
)

type Client struct{ HTTP *httpclient.Client }

func (c *Client) Call(ctx context.Context, plugin entity.Plugin, source entity.DataSource, operation string, input any) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.HTTP.Do(ctx, "POST", plugin.Endpoint, map[string]any{
		"version": 1, "operation": operation, "namespaceId": source.NamespaceID,
		"connectionSettings": json.RawMessage(source.ConnectionString), "settings": plugin.Settings, "input": input,
	}, nil, &result)
	return result, err
}
