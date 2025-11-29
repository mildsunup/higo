// Package di 提供依赖注入容器。
//
// 核心功能：
//   - 组件注册和获取
//   - 类型安全的依赖解析
//   - 泛型 Provider 支持
//
// 使用示例：
//
//	container := di.NewContainer()
//	container.Register("db", dbInstance)
//	db := di.MustGet[*sql.DB](container, "db")
package di
