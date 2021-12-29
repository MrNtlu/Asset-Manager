package requests

type ID struct {
	ID string `json:"id" binding:"required"`
}
