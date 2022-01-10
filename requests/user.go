package requests

type Login struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
	Password     string `json:"password" binding:"required"`
}

type Register struct {
	Username     string `json:"username" binding:"required"`
	EmailAddress string `json:"email_address" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
}

type ChangePassword struct {
	OldPassword string `json:"old_password" binding:"required,min=8"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ForgotPassword struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
}
