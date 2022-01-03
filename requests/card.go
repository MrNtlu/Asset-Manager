package requests

type Card struct {
	Name       string `json:"name" binding:"required"`
	Last4Digit string `json:"last_digit" binding:"required"`
}

type CardUpdate struct {
	ID         string  `json:"id" binding:"required"`
	Name       *string `json:"name"`
	Last4Digit *string `json:"last_digit"`
}
