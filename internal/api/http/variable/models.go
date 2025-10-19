package variable

type createRequest struct {
	Placeholder string  `json:"placeholder"`
	Value       string  `json:"value"`
	NamespaceID string  `json:"namespaceId"`
	DashboardID *string `json:"dashboardId"`
}

type updateRequest struct {
	Placeholder string `json:"placeholder"`
	Value       string `json:"value"`
}
