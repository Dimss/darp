package forwarder

type Upstream struct {
	Url      string `json:"url"`
	Resource string `json:"resource"`
	Body     []byte `json:"body"`
}

type UpstreamResponse struct {
	IsAllowed bool   `json:"isAllowed"`
	Message   string `json:"message"`
}
