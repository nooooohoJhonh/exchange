package utils

// Paginate 分页结构
type Paginate struct {
	Total    int64 `json:"total"`    // 总记录数
	Page     int64 `json:"page"`     // 当前页码
	PageSize int64 `json:"pageSize"` // 每页大小
	LastPage int64 `json:"lastPage"` // 最后一页
}

// NewPaginate 构造分页结构体
func NewPaginate(total, page, pageSize int64) Paginate {
	lastPage := (total + pageSize - 1) / pageSize
	return Paginate{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		LastPage: lastPage,
	}
}

// PageResponse 分页的返回结构
type PageResponse[T any] struct {
	Paginate Paginate `json:"paginate"` // 分页信息
	List     []T      `json:"list"`     // 数据列表
}

// ConvertPage 通用分页转换函数（支持任意类型转换）
// 参数说明：
// - src: 源数据列表
// - convertFunc: 转换函数，将源类型转换为目标类型
// - total: 总记录数
// - page: 当前页码
// - pageSize: 每页大小
func ConvertPage[T any, R any](src []T, convertFunc func(T) R, total, page, pageSize int64) *PageResponse[R] {
	result := &PageResponse[R]{
		Paginate: NewPaginate(total, page, pageSize),
		List:     make([]R, len(src)),
	}

	// 转换每个元素
	for i, v := range src {
		result.List[i] = convertFunc(v)
	}

	return result
}

// ValidatePageParams 验证分页参数
func ValidatePageParams(page, pageSize int64) (int64, int64) {
	// 页码最小为1
	if page < 1 {
		page = 1
	}

	// 每页大小最小为1，最大为100
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}

// CalculateOffset 计算数据库查询的偏移量
func CalculateOffset(page, pageSize int64) int64 {
	return (page - 1) * pageSize
}
