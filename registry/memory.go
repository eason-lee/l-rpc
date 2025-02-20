package registry

import (
	"errors"
	"sync"
	"time"
)

func (r *InMemoryRegistry) Register(instance *ServiceInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances := r.services[instance.Name]

	// 检查是否已存在
	for i, inst := range instances {
		if inst.ID == instance.ID {
			instances[i] = instance
			r.notifySubscribers(instance.Name)
			return nil
		}
	}

	// 添加新实例
	r.services[instance.Name] = append(instances, instance)
	r.notifySubscribers(instance.Name)
	return nil
}

func (r *InMemoryRegistry) Deregister(instanceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, instances := range r.services {
		for i, inst := range instances {
			if inst.ID == instanceID {
				r.services[name] = append(instances[:i], instances[i+1:]...)
				r.notifySubscribers(name)
				return nil
			}
		}
	}
	return errors.New("instance not found")
}

func (r *InMemoryRegistry) GetService(name string) ([]*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances, ok := r.services[name]
	if !ok {
		return nil, errors.New("service not found")
	}
	return instances, nil
}

func (r *InMemoryRegistry) ListServices() ([]*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ServiceInstance
	for _, instances := range r.services {
		result = append(result, instances...)
	}
	return result, nil
}

func (r *InMemoryRegistry) Subscribe(serviceName string) (<-chan []*ServiceInstance, error) {
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

func (r *InMemoryRegistry) Unsubscribe(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.subscribers, serviceName)
	return nil
}

func (r *InMemoryRegistry) notifySubscribers(serviceName string) {
	instances := r.services[serviceName]
	for _, ch := range r.subscribers[serviceName] {
		select {
		case ch <- instances:
		default:
			// 如果通道已满，跳过本次通知
		}
	}
}
