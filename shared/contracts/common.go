package contracts

type Empty struct{}

type StatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
