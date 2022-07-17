package requests

type FavouriteInvestingCreate struct {
	Symbol   string `json:"symbol" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Market   string `json:"market" binding:"required"`
	Priority int    `json:"priority" binding:"required"`
}
