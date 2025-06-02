package models

// PaginationFields represents the fields used for pagination in an API response
type PaginationFields struct {
	Count      int `json:"count"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	TotalCount int `json:"total_count"`
}

type PaginationResult[TItem any] struct {
	Items      []*TItem
	TotalCount int
}

type PaginationSuccessResult[TItem any] = SuccessResult[PaginationResult[TItem]]

func CreatePaginationSuccessResult[TItem any](items []*TItem, totalCount int) *SuccessResult[PaginationResult[TItem]] {
	paginationResult := &PaginationResult[TItem]{
		Items:      items,
		TotalCount: totalCount,
	}

	return CreateOkResult(paginationResult)
}
