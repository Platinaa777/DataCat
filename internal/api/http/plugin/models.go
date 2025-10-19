package plugin

type createRequest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Settings    string `json:"settings"`
	Endpoint    string `json:"endpoint"`
}

type updateRequest struct {
	Description string `json:"description"`
	Settings    string `json:"settings"`
}
