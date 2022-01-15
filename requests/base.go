package requests

type ID struct {
	ID string `json:"id" form:"id" binding:"required"`
}
