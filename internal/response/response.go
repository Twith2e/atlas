package response

type APIResponse[T any] struct {
	Status  string    `json:"status"`
	Message string    `json:"message,omitempty"`
	Data    *T        `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
	Page    int       `json:"page,omitempty"`
	Limit   int       `json:"limit,omitempty"`
	Total   int64     `json:"total,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
