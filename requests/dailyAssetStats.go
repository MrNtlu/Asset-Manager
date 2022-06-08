package requests

type DailyAssetStatsInterval struct {
	Interval string `form:"interval" binding:"required,oneof=weekly monthly yearly"`
}
