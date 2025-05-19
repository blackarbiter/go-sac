package dto

type TaskRequest struct {
	TargetURL string `json:"target_url" binding:"required,url"`
	ScanType  string `json:"scan_type" binding:"required,oneof=full quick"`
}

type TaskResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	UserID string `json:"user_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
