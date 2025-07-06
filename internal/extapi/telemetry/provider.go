package telemetry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/httpclient"
)

type Provider struct{ Client *httpclient.Client }
type Settings struct {
	ServerURL    string `json:"serverUrl"`
	ClusterURL   string `json:"clusterUrl"`
	IndexPattern string `json:"indexPattern"`
	AuthType     string `json:"authType"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	AuthToken    string `json:"authToken"`
}

func (p *Provider) connection(d entity.DataSource) (Settings, map[string]string, error) {
	var cfg Settings
	if err := json.Unmarshal([]byte(d.ConnectionString), &cfg); err != nil {
		return cfg, nil, fmt.Errorf("invalid connection settings")
	}
	if cfg.ServerURL == "" {
		cfg.ServerURL = cfg.ClusterURL
	}
	if _, err := p.Client.Validate(cfg.ServerURL); err != nil {
		return cfg, nil, err
	}
	headers := map[string]string{}
	switch strings.ToLower(cfg.AuthType) {
	case "basic":
		if cfg.Username == "" || cfg.Password == "" {
			return cfg, nil, fmt.Errorf("basic credentials are required")
		}
		headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(cfg.Username+":"+cfg.Password))
	case "bearer":
		if cfg.AuthToken == "" {
			return cfg, nil, fmt.Errorf("bearer token is required")
		}
		headers["Authorization"] = "Bearer " + cfg.AuthToken
	case "", "none":
		if cfg.ClusterURL != "" && cfg.Username != "" {
			headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(cfg.Username+":"+cfg.Password))
		}
	default:
		return cfg, nil, fmt.Errorf("unsupported authentication type")
	}
	return cfg, headers, nil
}
func endpoint(base, path string, q url.Values) string {
	return strings.TrimRight(base, "/") + path + "?" + q.Encode()
}
