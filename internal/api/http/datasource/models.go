package datasource

import "encoding/json"

type createRequest struct {
	UniqueName       string          `json:"uniqueName"`
	DataSourceType   string          `json:"dataSourceType"`
	ConnectionString string          `json:"connectionString"`
	Purpose          json.RawMessage `json:"purpose"`
}

type updateConnectionRequest struct {
	ConnectionString string `json:"connectionString"`
}
