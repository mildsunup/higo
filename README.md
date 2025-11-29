# Higo

企业级 Go 应用脚手架，基于清晰的架构边界和职责分离原则构建。

## 架构设计准则

### 1. 分层隔离
- **基础设施层**：与外部系统交互（数据库、缓存、消息队列）
- **领域层**：业务逻辑和规则，不依赖基础设施
- **应用层**：协调领域对象和基础设施
- **接口层**：HTTP/gRPC 服务端点

### 2. 依赖倒置
- 高层模块定义接口，低层模块实现接口
- 通过 `di` 容器管理依赖注入
- 避免循环依赖，保持单向依赖流

### 3. 关注点分离
- 每个包只负责一个明确的职责
- 横切关注点（日志、监控、追踪）通过 AOP 和中间件实现
- 配置、运行时、业务逻辑严格分离

### 4. 可测试性
- 面向接口编程，便于 Mock
- 纯函数优先，减少副作用
- 依赖注入支持测试替身

## 核心包职责边界

### 运行时与生命周期

#### `runtime`
**职责**：应用生命周期管理  
**边界**：
- 管理组件启动/停止顺序（按优先级）
- 提供生命周期钩子（BeforeStart/AfterStart/BeforeStop/AfterStop）
- 信号处理和优雅关闭
- **不涉及**：具体业务逻辑、依赖注入

#### `di`
**职责**：依赖注入容器  
**边界**：
- 组件注册和获取
- 类型安全的依赖解析
- **不涉及**：组件生命周期管理（由 `runtime` 负责）

#### `config`
**职责**：配置加载和管理  
**边界**：
- 支持多源配置（文件、环境变量、远程配置中心）
- 配置热更新和监听
- 结构化配置映射
- **不涉及**：配置的业务语义解释

---

### 基础设施

#### `storage`
**职责**：数据存储抽象层  
**边界**：
- 统一的存储接口（MySQL/PostgreSQL/MongoDB/ClickHouse/Elasticsearch）
- 连接池管理、健康检查、重连机制
- 链路追踪和指标采集
- **不涉及**：具体的 ORM 操作和业务查询逻辑

#### `cache`
**职责**：缓存抽象层  
**边界**：
- 多级缓存（内存 + Redis）
- 缓存穿透/击穿/雪崩防护
- 序列化策略（JSON/MessagePack）
- 统计信息（命中率、键数量）
- **不涉及**：缓存键的业务语义

#### `mq`
**职责**：消息队列抽象层  
**边界**：
- 统一的生产者/消费者接口（Kafka/RabbitMQ/Memory）
- 消息发布/订阅、异步处理
- 链路追踪和指标采集
- **不涉及**：消息的业务处理逻辑

#### `lock`
**职责**：分布式锁  
**边界**：
- 基于 Redis/内存的分布式锁实现
- 自动续期、超时释放
- **不涉及**：锁保护的业务逻辑

---

### 领域驱动设计

#### `ddd`
**职责**：DDD 战术模式基础设施  
**边界**：
- 实体（Entity）、值对象（ValueObject）、聚合根（AggregateRoot）
- 领域事件（DomainEvent）
- 仓储接口（Repository）、规约模式（Specification）
- 工作单元（UnitOfWork）、领域服务（DomainService）
- **不涉及**：具体的业务领域实现

#### `eventbus`
**职责**：进程内事件总线  
**边界**：
- 发布/订阅模式
- 同步/异步事件分发
- **不涉及**：跨进程消息传递（使用 `mq`）

---

### 服务端点

#### `server`
**职责**：HTTP/gRPC 服务器抽象  
**边界**：
- HTTP 服务器（基于 Gin）
- gRPC 服务器
- 多路复用（同端口同时支持 HTTP/gRPC）
- 服务器组管理
- **不涉及**：具体的路由和处理器实现

#### `middleware`
**职责**：HTTP/gRPC 中间件  
**边界**：
- HTTP：CORS、认证、限流、超时、请求 ID、日志、恢复、幂等性
- gRPC：拦截器、链路追踪、指标采集
- **不涉及**：业务逻辑

---

### 横切关注点

#### `aop`
**职责**：面向切面编程  
**边界**：
- 拦截器链（Interceptor Chain）
- 方法调用拦截（Invocation）
- 泛型处理器包装
- **不涉及**：具体的切面逻辑（如日志、事务）

#### `observability`
**职责**：可观测性  
**边界**：
- 链路追踪（OpenTelemetry）
- 指标采集（Prometheus）
- 自动 Span 注入
- **不涉及**：日志记录（由 `logger` 负责）

#### `logger`
**职责**：结构化日志  
**边界**：
- 基于 Zap 的高性能日志
- 日志级别、格式、输出配置
- 链路追踪上下文注入
- **不涉及**：日志的业务语义

---

### 安全与认证

#### `auth`
**职责**：认证抽象  
**边界**：
- JWT 令牌生成和验证
- Bcrypt 密码哈希
- 简单认证器接口
- **不涉及**：授权策略（RBAC/ABAC）

#### `security`
**职责**：安全工具  
**边界**：
- 限流器（令牌桶/漏桶）
- 密码强度验证
- **不涉及**：具体的业务安全规则

#### `idempotent`
**职责**：幂等性保证  
**边界**：
- 基于 Redis 的幂等性令牌
- 防止重复提交
- **不涉及**：业务操作的幂等性实现

---

### 弹性与容错

#### `resilience`
**职责**：弹性模式  
**边界**：
- 熔断器（Circuit Breaker）
- 重试策略（Retry）
- 限流器（Rate Limiter）
- **不涉及**：具体的业务降级逻辑

---

### 工具与辅助

#### `pool`
**职责**：对象池  
**边界**：
- Worker Pool（协程池）
- Buffer Pool（字节缓冲池）
- Object Pool（通用对象池）
- **不涉及**：池化对象的业务逻辑

#### `utils`
**职责**：通用工具函数  
**边界**：
- 字符串、切片、Map 操作
- 时间、数字处理
- 指针辅助、异步工具
- ID 生成（UUID/Snowflake）
- **不涉及**：业务相关的工具函数

#### `errors`
**职责**：错误处理  
**边界**：
- 错误码定义（HTTP/gRPC）
- 错误构造器
- 错误响应转换
- **不涉及**：具体的业务错误

#### `response`
**职责**：统一响应格式  
**边界**：
- 成功/失败响应封装
- 分页响应
- **不涉及**：响应数据的业务逻辑

---

## 依赖关系

```
应用层
  ↓
领域层 (ddd)
  ↓
基础设施层 (storage, cache, mq, lock)
  ↓
横切关注点 (aop, middleware, observability, logger)
  ↓
运行时 (runtime, di, config)
  ↓
工具层 (utils, errors, pool)
```

## 使用示例

```go
package main

import (
    "context"
    "github.com/mildsunup/higo/runtime"
    "github.com/mildsunup/higo/di"
    "github.com/mildsunup/higo/config"
    "github.com/mildsunup/higo/logger"
    "github.com/mildsunup/higo/server"
)

func main() {
    // 1. 加载配置
    cfg := config.MustLoad("config.yaml")
    
    // 2. 初始化日志
    log := logger.New(cfg.Logger)
    
    // 3. 创建应用
    app := runtime.New(cfg.App, log)
    
    // 4. 创建 DI 容器
    container := di.NewContainer()
    
    // 5. 注册组件
    httpServer := server.NewHTTP(cfg.Server.HTTP)
    app.Register(httpServer, 100)
    
    // 6. 启动应用
    if err := app.Run(context.Background()); err != nil {
        log.Fatal("app failed", "error", err)
    }
}
```

## 设计原则总结

1. **单一职责**：每个包只做一件事
2. **开闭原则**：对扩展开放，对修改关闭（通过接口和插件）
3. **里氏替换**：接口实现可互换（如 Cache、Storage、MQ）
4. **接口隔离**：小而专注的接口（如 Producer/Consumer 分离）
5. **依赖倒置**：依赖抽象而非具体实现

## License

Apache License 2.0
