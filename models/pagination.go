package models

// PaginationFields represents the fields used for pagination in an API response
type PaginationFields struct {
	Count      int `json:"count"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	TotalCount int `json:"total_count"`
}
