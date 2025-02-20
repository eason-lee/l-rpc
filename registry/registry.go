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
}

type ServiceStatus int

const (
	StatusUp ServiceStatus = iota
	StatusDown
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
	services   map[string][]*ServiceInstance  // 服务实例列表
	subscribers map[string][]chan []*ServiceInstance  // 服务变更订阅者
	mu         sync.RWMutex
}

func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		services:    make(map[string][]*ServiceInstance),
		subscribers: make(map[string][]chan []*ServiceInstance),
	}
}