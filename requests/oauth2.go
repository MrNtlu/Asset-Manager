package requests

type GoogleLogin struct {
	Token string `json:"token" binding:"required"`
}

type AppleSignin struct {
	Code      string `json:"code" binding:"required"`
	IsRefresh *bool  `json:"is_refresh" binding:"required"`
}
