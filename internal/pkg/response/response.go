package response

type ApiResponse[T any] struct {
	Data    *T     `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}
