package forwarder

type UpstreamRequest struct {
	Upstream Upstream `json:"upstream"`
	Payload  *[]byte  `json:"body"`
}

type UpstreamResponse struct {
	IsAllowed   *bool  `json:"isAllowed" validate:"required"`
	Message     string `json:"message" validate:"required"`
	UpstreamUrl string `json:"upstreamUrl"`
}

type Upstream struct {
	Url      string `json:"url"`
	Resource string `json:"resource"`
}

