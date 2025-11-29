package ddd

import "context"

// Repository 泛型仓储接口
type Repository[T any, ID Identifier] interface {
	// FindByID 根据ID查找
	FindByID(ctx context.Context, id ID) (*T, error)
	// Save 保存（新增或更新）
	Save(ctx context.Context, entity *T) error
	// Delete 删除
	Delete(ctx context.Context, id ID) error
	// Exists 判断是否存在
	Exists(ctx context.Context, id ID) (bool, error)
}

// ReadRepository 只读仓储接口（CQRS 查询端）
type ReadRepository[T any, ID Identifier] interface {
	FindByID(ctx context.Context, id ID) (*T, error)
	FindAll(ctx context.Context, opts ...QueryOption) ([]*T, error)
	Count(ctx context.Context, opts ...QueryOption) (int64, error)
}

// QueryOptions 查询选项
type QueryOptions struct {
	Offset  int
	Limit   int
	OrderBy string
	Desc    bool
	Filters map[string]any
}

// QueryOption 查询选项函数
type QueryOption func(*QueryOptions)

// WithPagination 分页
func WithPagination(offset, limit int) QueryOption {
	return func(o *QueryOptions) {
		o.Offset = offset
		o.Limit = limit
	}
}

// WithOrderBy 排序
func WithOrderBy(field string, desc bool) QueryOption {
	return func(o *QueryOptions) {
		o.OrderBy = field
		o.Desc = desc
	}
}

// WithFilter 过滤条件
func WithFilter(key string, val any) QueryOption {
	return func(o *QueryOptions) {
		if o.Filters == nil {
			o.Filters = make(map[string]any)
		}
		o.Filters[key] = val
	}
}

// ApplyQueryOptions 应用查询选项
func ApplyQueryOptions(opts ...QueryOption) QueryOptions {
	o := QueryOptions{Limit: 100}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
