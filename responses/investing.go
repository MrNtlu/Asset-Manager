package responses

type InvestingResponse struct {
	Name   string `bson:"name" json:"name"`
	Symbol string `bson:"symbol" json:"symbol"`
}

type InvestingListResponse struct {
	Data       []InvestingResponse `bson:"data" json:"data"`
	Pagination PaginationMetaData  `bson:"pagination" json:"pagination"`
}

type PaginationResponse struct {
	Metadata PaginationMetaData `bson:"metadata" json:"metadata"`
}

type PaginationMetaData struct {
	Count int64 `bson:"count" json:"count"`
	Total int64 `bson:"total" json:"total"`
	Page  int64 `bson:"page" json:"page"`
}
