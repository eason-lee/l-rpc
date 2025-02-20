package registry

import (
	"math"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	Select(instances []*ServiceInstance) (*ServiceInstance, error)
}

// RandomBalancer 随机负载均衡
type RandomBalancer struct {
	rand *rand.Rand
	mu   sync.Mutex
}

func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (b *RandomBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoAvailableInstances
	}
	
	b.mu.Lock()
	index := b.rand.Intn(len(instances))
	b.mu.Unlock()
	
	return instances[index], nil
}

// RoundRobinBalancer 轮询负载均衡
type RoundRobinBalancer struct {
	counter uint64
}

func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

func (b *RoundRobinBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoAvailableInstances
	}
	
	count := atomic.AddUint64(&b.counter, 1)
	index := int(count % uint64(len(instances)))
	return instances[index], nil
}

// WeightedRandomBalancer 加权随机负载均衡
type WeightedRandomBalancer struct {
	rand *rand.Rand
	mu   sync.Mutex
}

func NewWeightedRandomBalancer() *WeightedRandomBalancer {
	return &WeightedRandomBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (b *WeightedRandomBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoAvailableInstances
	}

	// 计算总权重
	var totalWeight int
	for _, inst := range instances {
		if weight, ok := inst.Metadata["weight"]; ok {
			w, _ := strconv.Atoi(weight)
			totalWeight += w
		} else {
			totalWeight += 1 // 默认权重为1
		}
	}

	// 随机选择
	b.mu.Lock()
	target := b.rand.Intn(totalWeight)
	b.mu.Unlock()

	currentWeight := 0
	for _, inst := range instances {
		weight := 1
		if w, ok := inst.Metadata["weight"]; ok {
			weight, _ = strconv.Atoi(w)
		}
		currentWeight += weight
		if target < currentWeight {
			return inst, nil
		}
	}

	return instances[0], nil
}

// LeastActiveBalancer 最小活跃数负载均衡
type LeastActiveBalancer struct {
	activeCount sync.Map
}

func NewLeastActiveBalancer() *LeastActiveBalancer {
	return &LeastActiveBalancer{}
}

func (b *LeastActiveBalancer) IncrementActive(instanceID string) {
	count := int64(1)
	if v, ok := b.activeCount.Load(instanceID); ok {
		count = v.(int64) + 1
	}
	b.activeCount.Store(instanceID, count)
}

func (b *LeastActiveBalancer) DecrementActive(instanceID string) {
	if v, ok := b.activeCount.Load(instanceID); ok {
		count := v.(int64) - 1
		if count <= 0 {
			b.activeCount.Delete(instanceID)
		} else {
			b.activeCount.Store(instanceID, count)
		}
	}
}

func (b *LeastActiveBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoAvailableInstances
	}

	var minActive int64 = math.MaxInt64
	var selectedInst *ServiceInstance

	for _, inst := range instances {
		active := int64(0)
		if v, ok := b.activeCount.Load(inst.ID); ok {
			active = v.(int64)
		}
		if active < minActive {
			minActive = active
			selectedInst = inst
		}
	}

	return selectedInst, nil
}