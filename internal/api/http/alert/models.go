package alert

import "github.com/Platinaa777/DataCat/internal/entity"

type createRequest struct {
	Description                  string          `json:"description"`
	Template                     string          `json:"template"`
	RawQuery                     string          `json:"rawQuery"`
	DataSourceID                 string          `json:"dataSourceId"`
	NotificationChannelGroupName string          `json:"notificationChannelGroupName"`
	WaitTimeBeforeAlerting       entity.Duration `json:"waitTimeBeforeAlerting"`
	RepeatInterval               entity.Duration `json:"repeatInterval"`
	Tags                         []string        `json:"tags"`
}

type updateRequest struct {
	Description  string `json:"description"`
	Template     string `json:"template"`
	RawQuery     string `json:"rawQuery"`
	DataSourceID string `json:"dataSourceId"`
}
