package requests

type BankAccountCreate struct {
	Name          string `json:"name" binding:"required"`
	Iban          string `json:"iban" binding:"required"`
	AccountHolder string `json:"account_holder" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
}

type BankAccountUpdate struct {
	ID            string  `json:"id" binding:"required"`
	Name          *string `json:"name"`
	Iban          *string `json:"iban"`
	AccountHolder *string `json:"account_holder"`
	Currency      *string `json:"currency"`
}
