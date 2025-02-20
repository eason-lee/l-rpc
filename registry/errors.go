package registry

import "errors"

var (
	// ErrNoAvailableInstances 没有可用的服务实例
	ErrNoAvailableInstances = errors.New("no available service instances")
	
	// ErrServiceNotFound 服务未找到
	ErrServiceNotFound = errors.New("service not found")
	
	// ErrInstanceNotFound 实例未找到
	ErrInstanceNotFound = errors.New("instance not found")
	
	// ErrInvalidWeight 无效的权重值
	ErrInvalidWeight = errors.New("invalid weight value")
)