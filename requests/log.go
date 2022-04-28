package requests

type CreateLog struct {
	Log     string `json:"log" binding:"required"`
	LogType int    `json:"log_type" binding:"required"`
}
