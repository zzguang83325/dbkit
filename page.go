package dbkit

import "encoding/json"

// Page represents a paginated result, similar to the Java Page class.
// It uses generics to support different types of data in the List.
type Page[T any] struct {
	PageNumber int   `json:"pageNumber"` // page number
	PageSize   int   `json:"pageSize"`   // result amount of this page
	TotalPage  int   `json:"totalPage"`  // total page
	TotalRow   int64 `json:"totalRow"`   // total row
	List       []T   `json:"list"`       // list result of this page
}

// NewPage creates a new Page instance and calculates the total pages.
func NewPage[T any](list []T, pageNumber, pageSize int, totalRow int64) *Page[T] {
	totalPage := 0
	if pageSize > 0 {
		totalPage = int((totalRow + int64(pageSize) - 1) / int64(pageSize))
	}
	return &Page[T]{
		PageNumber: pageNumber,
		PageSize:   pageSize,
		TotalPage:  totalPage,
		TotalRow:   totalRow,
		List:       list,
	}
}

// IsFirstPage returns true if the current page is the first page.
func (p *Page[T]) IsFirstPage() bool {
	return p.PageNumber <= 1
}

// IsLastPage returns true if the current page is the last page.
func (p *Page[T]) IsLastPage() bool {
	return p.PageNumber >= p.TotalPage
}

// ToJson returns a JSON string representation of the Page instance.
func (p *Page[T]) ToJson() string {
	b, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(b)
}

// RecordPageToDbModelPage converts a Page[Record] to a Page[IDbModel]
func RecordPageToDbModelPage[T any](p *Page[Record]) (*Page[T], error) {
	var list []T
	if err := ToStructs(p.List, &list); err != nil {
		return nil, err
	}
	return &Page[T]{
		PageNumber: p.PageNumber,
		PageSize:   p.PageSize,
		TotalPage:  p.TotalPage,
		TotalRow:   p.TotalRow,
		List:       list,
	}, nil
}
