package requests

type ID struct {
	ID string `json:"id" form:"id" binding:"required"`
}

type Type struct {
	Type string `form:"type" binding:"required"`
}
