package dashboard

type createRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	NamespaceID string   `json:"namespaceId"`
	Tags        []string `json:"tags"`
}

type updateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
