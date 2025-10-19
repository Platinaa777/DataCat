package panel

type panelInput struct {
	Title              string `json:"title"`
	Type               int    `json:"type"`
	RawQuery           string `json:"rawQuery"`
	DataSourceID       string `json:"dataSourceId"`
	Layout             string `json:"layout"`
	DashboardID        string `json:"dashboardId"`
	StyleConfiguration string `json:"styleConfiguration"`
}
