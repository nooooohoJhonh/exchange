package utils

import (
	"gorm.io/gorm"
)

// GormPaginate GORM分页查询作用域
// 参数说明：
// - page: 页码（从1开始）
// - pageSize: 每页大小
// 返回GORM的Scope函数，可以直接在链式查询中使用
func GormPaginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 验证分页参数
		validatedPage, validatedPageSize := ValidatePageParams(int64(page), int64(pageSize))

		// 计算偏移量
		offset := CalculateOffset(validatedPage, validatedPageSize)

		return db.Offset(int(offset)).Limit(int(validatedPageSize))
	}
}

// GormPaginateWithCount GORM分页查询（带总数统计）
// 参数说明：
// - db: GORM数据库实例
// - model: 查询的模型
// - page: 页码
// - pageSize: 每页大小
// - conditions: 查询条件
// - result: 查询结果切片
// 返回：总数、错误
func GormPaginateWithCount[T any](
	db *gorm.DB,
	model T,
	page, pageSize int,
	conditions map[string]interface{},
	result *[]T,
) (int64, error) {
	var total int64

	// 构建查询
	query := db.Model(&model)

	// 添加查询条件
	for key, value := range conditions {
		if value != "" && value != nil {
			query = query.Where(key+" = ?", value)
		}
	}

	// 执行分页查询
	err := query.Count(&total).
		Scopes(GormPaginate(page, pageSize)).
		Find(result).Error

	return total, err
}

// GormPaginateWithKeyword GORM分页查询（支持关键词搜索）
// 参数说明：
// - db: GORM数据库实例
// - model: 查询的模型
// - page: 页码
// - pageSize: 每页大小
// - keyword: 搜索关键词
// - searchFields: 搜索字段列表
// - conditions: 其他查询条件
// - result: 查询结果切片
// 返回：总数、错误
func GormPaginateWithKeyword[T any](
	db *gorm.DB,
	model T,
	page, pageSize int,
	keyword string,
	searchFields []string,
	conditions map[string]interface{},
	result *[]T,
) (int64, error) {
	var total int64

	// 构建查询
	query := db.Model(&model)

	// 添加关键词搜索
	if keyword != "" && len(searchFields) > 0 {
		searchQuery := db
		for i, field := range searchFields {
			if i == 0 {
				searchQuery = searchQuery.Where(field+" LIKE ?", "%"+keyword+"%")
			} else {
				searchQuery = searchQuery.Or(field+" LIKE ?", "%"+keyword+"%")
			}
		}
		query = query.Where(searchQuery)
	}

	// 添加其他查询条件
	for key, value := range conditions {
		if value != "" && value != nil {
			query = query.Where(key+" = ?", value)
		}
	}

	// 执行分页查询
	err := query.Count(&total).
		Scopes(GormPaginate(page, pageSize)).
		Find(result).Error

	return total, err
}

// GormPaginateWithOrder GORM分页查询（支持排序）
// 参数说明：
// - db: GORM数据库实例
// - model: 查询的模型
// - page: 页码
// - pageSize: 每页大小
// - orderBy: 排序字段
// - conditions: 查询条件
// - result: 查询结果切片
// 返回：总数、错误
func GormPaginateWithOrder[T any](
	db *gorm.DB,
	model T,
	page, pageSize int,
	orderBy string,
	conditions map[string]interface{},
	result *[]T,
) (int64, error) {
	var total int64

	// 构建查询
	query := db.Model(&model)

	// 添加查询条件
	for key, value := range conditions {
		if value != "" && value != nil {
			query = query.Where(key+" = ?", value)
		}
	}

	// 添加排序
	if orderBy != "" {
		query = query.Order(orderBy)
	}

	// 执行分页查询
	err := query.Count(&total).
		Scopes(GormPaginate(page, pageSize)).
		Find(result).Error

	return total, err
}

// GormPaginateWithPreload GORM分页查询（支持预加载关联）
// 参数说明：
// - db: GORM数据库实例
// - model: 查询的模型
// - page: 页码
// - pageSize: 每页大小
// - preloads: 预加载的关联字段
// - conditions: 查询条件
// - result: 查询结果切片
// 返回：总数、错误
func GormPaginateWithPreload[T any](
	db *gorm.DB,
	model T,
	page, pageSize int,
	preloads []string,
	conditions map[string]interface{},
	result *[]T,
) (int64, error) {
	var total int64

	// 构建查询
	query := db.Model(&model)

	// 添加预加载
	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	// 添加查询条件
	for key, value := range conditions {
		if value != "" && value != nil {
			query = query.Where(key+" = ?", value)
		}
	}

	// 执行分页查询
	err := query.Count(&total).
		Scopes(GormPaginate(page, pageSize)).
		Find(result).Error

	return total, err
}
