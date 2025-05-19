package client

// stringRequest matches your server’s DTO.
type stringRequest struct {
	Value      string `json:"value"`
	TTLSeconds int    `json:"ttl_seconds,omitempty"`
}

// listRequest matches your server’s DTO.
type listRequest struct {
	Items      []string `json:"items"`
	TTLSeconds int      `json:"ttl_seconds,omitempty"`
}

// stringResponse matches {"value":"..."}.
type stringResponse struct {
	Value string `json:"value"`
}
