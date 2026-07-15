package dto

// PageResp 表示分页响应。
type PageResp[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
}

// NewPageResp 创建分页响应。
func NewPageResp[T any](list []T, total int64) *PageResp[T] {
	if list == nil {
		list = make([]T, 0)
	}

	return &PageResp[T]{
		List:  list,
		Total: total,
	}
}
