package adapters

// stringRequest is the JSON body for POST /v1/string/{key}.
type stringRequest struct {
	Value      string `json:"value"`
	TTLSeconds int    `json:"ttl_seconds,omitempty"`
}

// listRequest is the JSON body for POST /v1/list/{key}/push.
type listRequest struct {
	Items []string `json:"items"`
}
