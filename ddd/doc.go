// Package ddd 提供领域驱动设计（DDD）战术模式基础设施。
//
// 核心概念：
//   - Entity（实体）
//   - ValueObject（值对象）
//   - AggregateRoot（聚合根）
//   - DomainEvent（领域事件）
//   - Repository（仓储）
//   - Specification（规约）
//   - UnitOfWork（工作单元）
//   - DomainService（领域服务）
//
// 使用示例：
//
//	type User struct {
//	    ddd.AggregateRoot[string]
//	    Name string
//	}
//	user := User{AggregateRoot: ddd.NewAggregateRoot("id")}
//	user.RaiseEvent(UserCreatedEvent{})
package ddd
