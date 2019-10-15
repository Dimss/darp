package forwarder

type Upstream struct {
	Url      string `json:"url"`
	Resource string `json:"resource"`
	Body     []byte `json:"body"`
}

type UpstreamResponse struct {
	IsAllowed bool   `json:"isAllowed" validate:"required"`
	Message   string `json:"message" validate:"required"`
}
