package requests

type ID struct {
	ID string `json:"id" form:"id" binding:"required"`
}

type Type struct {
	AssetType string `form:"type" binding:"required"`
}
