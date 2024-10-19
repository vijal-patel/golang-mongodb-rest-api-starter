package models

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type GetManyResponse struct {
	Limit   int64       `json:"limit"`
	Offset  int64       `json:"offset"`
	Total   int64       `json:"total"`
	HasNext bool        `json:"hasNext"`
	Items   interface{} `json:"items"`
}

type GetManyQuery struct {
	Limit    int64
	Offset   int64
	OrderBy  string
	SortType string
}
