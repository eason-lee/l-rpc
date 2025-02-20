package registry

import (
	"errors"
	"sync"
	"time"
)

// MemoryRegistry 基于内存的注册中心实现
type MemoryRegistry struct {
	services    map[string][]*ServiceInstance
	subscribers map[string][]chan []*ServiceInstance
	mu          sync.RWMutex
	health      *HealthChecker
}

func NewInMemoryRegistry() *MemoryRegistry {
	r := &MemoryRegistry{
		services:    make(map[string][]*ServiceInstance),
		subscribers: make(map[string][]chan []*ServiceInstance),
	}
	r.health = NewHealthChecker(r)
	return r
}

// 实现 RegistryNotifier 接口
func (r *MemoryRegistry) NotifyStatusChange(serviceName string, instance *ServiceInstance) {
	r.notifySubscribers(serviceName)
}

func (r *MemoryRegistry) Register(instance *ServiceInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance.LastHeartbeat = time.Now()
	instances := r.services[instance.Name]

	for i, inst := range instances {
		if inst.ID == instance.ID {
			instances[i] = instance
			r.health.AddInstance(instance)
			r.notifySubscribers(instance.Name)
			return nil
		}
	}

	r.services[instance.Name] = append(instances, instance)
	r.health.AddInstance(instance)
	r.notifySubscribers(instance.Name)
	return nil
}

func (r *MemoryRegistry) Deregister(instanceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, instances := range r.services {
		for i, inst := range instances {
			if inst.ID == instanceID {
				r.health.RemoveInstance(instanceID)
				r.services[name] = append(instances[:i], instances[i+1:]...)
				r.notifySubscribers(name)
				return nil
			}
		}
	}
	return errors.New("instance not found")
}

func (r *MemoryRegistry) GetService(name string) ([]*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances, ok := r.services[name]
	if !ok {
		return nil, errors.New("service not found")
	}
	return instances, nil
}

func (r *MemoryRegistry) ListServices() ([]*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ServiceInstance
	for _, instances := range r.services {
		result = append(result, instances...)
	}
	return result, nil
}

func (r *MemoryRegistry) Subscribe(serviceName string) (<-chan []*ServiceInstance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan []*ServiceInstance, 1)
	r.subscribers[serviceName] = append(r.subscribers[serviceName], ch)

	// 立即推送当前实例列表
	if instances, ok := r.services[serviceName]; ok {
		ch <- instances
	}
	return ch, nil
}

func (r *MemoryRegistry) Unsubscribe(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.subscribers, serviceName)
	return nil
}

func (r *MemoryRegistry) notifySubscribers(serviceName string) {
	instances := r.services[serviceName]
	for _, ch := range r.subscribers[serviceName] {
		select {
		case ch <- instances:
		default:
			// 如果通道已满，跳过本次通知
		}
	}
}
