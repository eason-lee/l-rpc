package registry

import (
	"sync"
	"time"
)

// ServiceInstance 服务实例信息
type ServiceInstance struct {
	ID        string            // 实例唯一标识
	Name      string            // 服务名称
	Version   string            // 服务版本
	Metadata  map[string]string // 元数据
	Endpoints []string          // 服务地址列表
	Status    ServiceStatus     // 服务状态
	LastHeartbeat time.Time    // 最后心跳时间
    // 新增健康检查相关字段
    HealthCheck   *HealthCheck
}

type HealthCheck struct {
    Interval time.Duration    // 健康检查间隔
    Timeout  time.Duration    // 健康检查超时时间
    URL      string          // 健康检查地址
}

type ServiceStatus int

const (
	StatusUp ServiceStatus = iota
	StatusDown
    defaultHealthCheckInterval = 10 * time.Second
    defaultHealthCheckTimeout = 5 * time.Second
)

// Registry 注册中心接口
type Registry interface {
	// Register 注册服务实例
	Register(instance *ServiceInstance) error
	
	// Deregister 注销服务实例
	Deregister(instanceID string) error
	
	// GetService 获取服务实例列表
	GetService(name string) ([]*ServiceInstance, error)
	
	// ListServices 获取所有服务
	ListServices() ([]*ServiceInstance, error)
	
	// Subscribe 订阅服务变更
	Subscribe(serviceName string) (<-chan []*ServiceInstance, error)
	
	// Unsubscribe 取消订阅
	Unsubscribe(serviceName string) error
}

// InMemoryRegistry 基于内存的注册中心实现
type InMemoryRegistry struct {
    services    map[string][]*ServiceInstance
    subscribers map[string][]chan []*ServiceInstance
    mu          sync.RWMutex
    health      *HealthChecker
}

func NewInMemoryRegistry() *InMemoryRegistry {
    r := &InMemoryRegistry{
        services:    make(map[string][]*ServiceInstance),
        subscribers: make(map[string][]chan []*ServiceInstance),
    }
    r.health = NewHealthChecker(r)
    return r
}

// 实现 RegistryNotifier 接口
func (r *InMemoryRegistry) NotifyStatusChange(serviceName string, instance *ServiceInstance) {
    r.notifySubscribers(serviceName)
}
