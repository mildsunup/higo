package ddd

import "context"

// DomainService 领域服务标记接口
// 领域服务封装不属于任何实体或值对象的领域逻辑
type DomainService interface {
	domainService()
}

// BaseDomainService 领域服务基类
type BaseDomainService struct{}

func (BaseDomainService) domainService() {}

// ApplicationService 应用服务接口
// 应用服务编排领域对象，实现用例
type ApplicationService interface {
	applicationService()
}

// BaseApplicationService 应用服务基类
type BaseApplicationService struct {
	eventBus EventBus
}

func (BaseApplicationService) applicationService() {}

// NewBaseApplicationService 创建应用服务基类
func NewBaseApplicationService(eventBus EventBus) BaseApplicationService {
	return BaseApplicationService{eventBus: eventBus}
}

// PublishEvents 发布聚合根的领域事件
func (s *BaseApplicationService) PublishEvents(ctx context.Context, events []DomainEvent) error {
	if s.eventBus == nil || len(events) == 0 {
		return nil
	}
	return s.eventBus.Publish(ctx, events...)
}
