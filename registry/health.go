package registry

import (
    "net/http"
    "sync"
    "time"
)

// RegistryNotifier 定义注册中心通知接口
type RegistryNotifier interface {
    NotifyStatusChange(serviceName string, instance *ServiceInstance)
}

type HealthChecker struct {
    notifier    RegistryNotifier
    stopCh      chan struct{}
    mu          sync.RWMutex
    checkTasks  map[string]*checkTask
}

func NewHealthChecker(notifier RegistryNotifier) *HealthChecker {
    return &HealthChecker{
        notifier:   notifier,
        stopCh:     make(chan struct{}),
        checkTasks: make(map[string]*checkTask),
    }
}

type checkTask struct {
    instance *ServiceInstance
    stopCh   chan struct{}
}


func (h *HealthChecker) Stop() {
    close(h.stopCh)
}

func (h *HealthChecker) AddInstance(instance *ServiceInstance) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if instance.HealthCheck == nil {
        instance.HealthCheck = &HealthCheck{
            Interval: defaultHealthCheckInterval,
            Timeout:  defaultHealthCheckTimeout,
        }
    }

    task := &checkTask{
        instance: instance,
        stopCh:   make(chan struct{}),
    }
    h.checkTasks[instance.ID] = task
    go h.runCheck(task)
}

func (h *HealthChecker) RemoveInstance(instanceID string) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if task, ok := h.checkTasks[instanceID]; ok {
        close(task.stopCh)
        delete(h.checkTasks, instanceID)
    }
}

func (h *HealthChecker) runCheck(task *checkTask) {
    ticker := time.NewTicker(task.instance.HealthCheck.Interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            status := h.check(task.instance)
            if status != task.instance.Status {
                task.instance.Status = status
                h.notifier.NotifyStatusChange(task.instance.Name, task.instance)
            }
        case <-task.stopCh:
            return
        case <-h.stopCh:
            return
        }
    }
}

func (h *HealthChecker) check(instance *ServiceInstance) ServiceStatus {
    if instance.HealthCheck.URL == "" {
        // 如果没有配置健康检查URL，使用最后心跳时间判断
        if time.Since(instance.LastHeartbeat) > instance.HealthCheck.Interval*2 {
            return StatusDown
        }
        return StatusUp
    }

    client := &http.Client{
        Timeout: instance.HealthCheck.Timeout,
    }
    resp, err := client.Get(instance.HealthCheck.URL)
    if err != nil {
        return StatusDown
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return StatusDown
    }
    return StatusUp
}