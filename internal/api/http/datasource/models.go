package datasource

import (
	"encoding/json"
	"strings"

	"github.com/Platinaa777/DataCat/internal/entity"
)

type createRequest struct {
	UniqueName       string            `json:"uniqueName"`
	DataSourceType   string            `json:"dataSourceType"`
	ConnectionString string            `json:"connectionString"`
	Purpose          dataSourcePurpose `json:"purpose"`
}

type updateConnectionRequest struct {
	ConnectionString string `json:"connectionString"`
}

type dataSourcePurpose string

func (p *dataSourcePurpose) UnmarshalJSON(data []byte) error {
	var name string
	if json.Unmarshal(data, &name) != nil {
		var number int
		if json.Unmarshal(data, &number) != nil {
			return entity.ErrInvalid
		}
		name = map[int]string{1: "Metrics", 2: "Logs", 3: "Traces"}[number]
	}
	*p = dataSourcePurpose(map[string]string{"metrics": "Metrics", "logs": "Logs", "traces": "Traces"}[strings.ToLower(name)])
	return nil
}
