package requests

type Card struct {
	Name       string `json:"name" binding:"required"`
	Last4Digit string `json:"last_digit" binding:"required"`
	CardHolder string `json:"card_holder" binding:"required"`
	Color      string `json:"color" binding:"required"`
	CardType   string `json:"type" binding:"required"`
}

type CardUpdate struct {
	ID         string  `json:"id" binding:"required"`
	Name       *string `json:"name"`
	Last4Digit *string `json:"last_digit"`
	CardHolder *string `json:"card_holder"`
	Color      *string `json:"color"`
	CardType   *string `json:"type"`
}
