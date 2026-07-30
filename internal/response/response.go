package response

// APIResponse is the envelope for a successful response carrying a payload.
type APIResponse[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    *T     `json:"data,omitempty"`
}

// PaginatedResponse is the envelope for a successful response carrying a page
// of results.
type PaginatedResponse[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    *T     `json:"data,omitempty"`
	Page    int    `json:"page,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Total   int64  `json:"total,omitempty"`
}

// MessageResponse is the envelope for a successful response with no payload.
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// ErrorResponse is the envelope for a failed response. Success and error
// envelopes are separate types so neither documents fields the other owns.
type ErrorResponse struct {
	Status string    `json:"status"`
	Error  *APIError `json:"error"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
