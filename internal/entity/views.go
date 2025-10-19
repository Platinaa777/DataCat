package entity

type DashboardPanel struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	PanelType          string `json:"panelType"`
	Query              string `json:"query"`
	Layout             string `json:"layout"`
	StyleConfiguration string `json:"styleConfiguration"`
}
type DashboardDetails struct {
	Dashboard
	Panels []DashboardPanel `json:"panels"`
}
type NamespaceDetails struct {
	Name       string      `json:"name"`
	Dashboards []Dashboard `json:"dashboards"`
}
type AlertCounter struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}
