package ddd

import "context"

// UnitOfWork 工作单元接口
// 用于管理事务边界，确保聚合的一致性
type UnitOfWork interface {
	// Begin 开始事务
	Begin(ctx context.Context) (context.Context, error)
	// Commit 提交事务
	Commit(ctx context.Context) error
	// Rollback 回滚事务
	Rollback(ctx context.Context) error
}

// Transactional 在事务中执行函数
func Transactional(ctx context.Context, uow UnitOfWork, fn func(ctx context.Context) error) error {
	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = uow.Rollback(txCtx)
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		_ = uow.Rollback(txCtx)
		return err
	}

	return uow.Commit(txCtx)
}

// TransactionalWithResult 在事务中执行函数并返回结果
func TransactionalWithResult[T any](ctx context.Context, uow UnitOfWork, fn func(ctx context.Context) (T, error)) (T, error) {
	var result T
	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return result, err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = uow.Rollback(txCtx)
			panic(r)
		}
	}()

	result, err = fn(txCtx)
	if err != nil {
		_ = uow.Rollback(txCtx)
		return result, err
	}

	if err := uow.Commit(txCtx); err != nil {
		return result, err
	}

	return result, nil
}
