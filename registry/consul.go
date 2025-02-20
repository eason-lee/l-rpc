package registry

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/hashicorp/consul/api"
)

type ConsulRegistry struct {
	client      *api.Client
	services    sync.Map
	health      *HealthChecker
	subscribers map[string][]chan []*ServiceInstance
	mu          sync.RWMutex
}

func NewConsulRegistry(addr string) (*ConsulRegistry, error) {
	config := api.DefaultConfig()
	config.Address = addr
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	r := &ConsulRegistry{
		client:      client,
		subscribers: make(map[string][]chan []*ServiceInstance),
	}
	r.health = NewHealthChecker(r)
	return r, nil
}

func (r *ConsulRegistry) Register(instance *ServiceInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 转换为 Consul 服务注册信息
	registration := &api.AgentServiceRegistration{
		ID:      instance.ID,
		Name:    instance.Name,
		Tags:    []string{instance.Version},
		Port:    r.getPort(instance.Endpoints[0]),
		Address: r.getHost(instance.Endpoints[0]),
		Meta:    instance.Metadata,
		Check: &api.AgentServiceCheck{
			HTTP:     instance.HealthCheck.URL,
			Interval: instance.HealthCheck.Interval.String(),
			Timeout:  instance.HealthCheck.Timeout.String(),
		},
	}

	if err := r.client.Agent().ServiceRegister(registration); err != nil {
		return err
	}

	r.services.Store(instance.ID, instance)
	r.health.AddInstance(instance)
	r.notifySubscribers(instance.Name)
	return nil
}

func (r *ConsulRegistry) Deregister(instanceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.client.Agent().ServiceDeregister(instanceID); err != nil {
		return err
	}

	if instance, ok := r.services.LoadAndDelete(instanceID); ok {
		r.health.RemoveInstance(instanceID)
		r.notifySubscribers(instance.(*ServiceInstance).Name)
	}
	return nil
}

func (r *ConsulRegistry) GetService(name string) ([]*ServiceInstance, error) {
	services, _, err := r.client.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, err
	}

	var instances []*ServiceInstance
	for _, service := range services {
		instance := &ServiceInstance{
			ID:      service.Service.ID,
			Name:    service.Service.Service,
			Version: r.getVersion(service.Service.Tags),
			Metadata: service.Service.Meta,
			Endpoints: []string{fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)},
			Status:   r.convertStatus(service.Checks.AggregatedStatus()),
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func (r *ConsulRegistry) ListServices() ([]*ServiceInstance, error) {
	services, _, err := r.client.Catalog().Services(nil)
	if err != nil {
		return nil, err
	}

	var instances []*ServiceInstance
	for serviceName := range services {
		serviceInstances, err := r.GetService(serviceName)
		if err != nil {
			continue
		}
		instances = append(instances, serviceInstances...)
	}
	return instances, nil
}

func (r *ConsulRegistry) Subscribe(serviceName string) (<-chan []*ServiceInstance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan []*ServiceInstance, 1)
	r.subscribers[serviceName] = append(r.subscribers[serviceName], ch)

	// 立即推送当前实例列表
	if instances, err := r.GetService(serviceName); err == nil {
		ch <- instances
	}
	return ch, nil
}

func (r *ConsulRegistry) Unsubscribe(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.subscribers, serviceName)
	return nil
}

func (r *ConsulRegistry) NotifyStatusChange(serviceName string, instance *ServiceInstance) {
	r.notifySubscribers(serviceName)
}

func (r *ConsulRegistry) notifySubscribers(serviceName string) {
	instances, err := r.GetService(serviceName)
	if err != nil {
		return
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, ch := range r.subscribers[serviceName] {
		select {
		case ch <- instances:
		default:
		}
	}
}

// 辅助方法
func (r *ConsulRegistry) getHost(endpoint string) string {
	// 从 endpoint 解析主机地址
	// 示例: localhost:8080 -> localhost
	host, _, _ := net.SplitHostPort(endpoint)
	return host
}

func (r *ConsulRegistry) getPort(endpoint string) int {
	// 从 endpoint 解析端口
	// 示例: localhost:8080 -> 8080
	_, portStr, _ := net.SplitHostPort(endpoint)
	port, _ := strconv.Atoi(portStr)
	return port
}

func (r *ConsulRegistry) getVersion(tags []string) string {
	if len(tags) > 0 {
		return tags[0]
	}
	return ""
}

func (r *ConsulRegistry) convertStatus(status string) ServiceStatus {
	switch status {
	case "passing":
		return StatusUp
	default:
		return StatusDown
	}
}