package secret

type putRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
