package requests

type Investings struct {
	Type   string `form:"type" binding:"required"`
	Market string `form:"market" binding:"required"`
}
