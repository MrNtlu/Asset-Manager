package requests

type Login struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
	Password     string `json:"password" binding:"required"`
}

type Register struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
	Currency     string `json:"currency" binding:"required,oneof=USD EUR JPY KRW GBP"`
	Password     string `json:"password" binding:"required,min=6"`
}

type ChangePassword struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangeCurrency struct {
	Currency string `json:"currency" binding:"required,oneof=USD EUR JPY KRW GBP"`
}

type ForgotPassword struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
}
