package requests

type GoogleLogin struct {
	Token string `json:"token" binding:"required"`
}
