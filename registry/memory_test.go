package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RegistryTestSuite struct {
	suite.Suite
	registry *MemoryRegistry
}

func (s *RegistryTestSuite) SetupTest() {
	s.registry = NewInMemoryRegistry()
}

func (s *RegistryTestSuite) TestRegisterNewInstance() {
	instance := &ServiceInstance{
		ID:        "instance-1",
		Name:      "test-service",
		Version:   "1.0.0",
		Endpoints: []string{"localhost:8080"},
		Status:    StatusUp,
	}

	// 测试注册
	err := s.registry.Register(instance)
	assert.NoError(s.T(), err)

	// 验证注册结果
	services, err := s.registry.GetService(instance.Name)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), services, 1)
	assert.Equal(s.T(), instance.ID, services[0].ID)
}

func (s *RegistryTestSuite) TestRegisterExistingInstance() {
	instance := &ServiceInstance{
		ID:        "instance-1",
		Name:      "test-service",
		Version:   "1.0.0",
		Endpoints: []string{"localhost:8080"},
	}

	// 先注册一个实例
	err := s.registry.Register(instance)
	assert.NoError(s.T(), err)

	// 更新实例
	updatedInstance := &ServiceInstance{
		ID:        "instance-1",
		Name:      "test-service",
		Version:   "1.0.1",
		Endpoints: []string{"localhost:8081"},
	}
	err = s.registry.Register(updatedInstance)
	assert.NoError(s.T(), err)

	// 验证更新结果
	services, err := s.registry.GetService(instance.Name)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), services, 1)
	assert.Equal(s.T(), "1.0.1", services[0].Version)
}

func (s *RegistryTestSuite) TestSubscription() {
	ch, err := s.registry.Subscribe("test-service")
	assert.NoError(s.T(), err)

	instance := &ServiceInstance{
		ID:      "instance-1",
		Name:    "test-service",
		Version: "1.0.0",
	}

	// 注册服务，应该触发通知
	err = s.registry.Register(instance)
	assert.NoError(s.T(), err)

	// 验证是否收到通知
	select {
	case services := <-ch:
		assert.Len(s.T(), services, 1)
		assert.Equal(s.T(), instance.ID, services[0].ID)
	case <-time.After(time.Second):
		s.T().Error("Timeout waiting for notification")
	}
}

func (s *RegistryTestSuite) TestDeregister() {
	// 准备测试数据
	instance1 := &ServiceInstance{
		ID:        "instance-1",
		Name:      "test-service-1",
		Version:   "1.0.0",
		Endpoints: []string{"localhost:8080"},
	}
	instance2 := &ServiceInstance{
		ID:        "instance-2",
		Name:      "test-service-1",
		Version:   "1.0.0",
		Endpoints: []string{"localhost:8081"},
	}

	s.Run("注销不存在的实例", func() {
		err := s.registry.Deregister("non-exist")
		s.Error(err)
		s.Equal("instance not found", err.Error())
	})

	s.Run("注销已存在的实例", func() {
		// 注册实例
		err := s.registry.Register(instance1)
		s.NoError(err)

		// 创建订阅以验证通知
		ch, err := s.registry.Subscribe(instance1.Name)
		s.NoError(err)
		// 验证订阅者收到通知
		select {
		case services := <-ch:
			s.NotEmpty(services)
		case <-time.After(time.Second):
			s.Fail("未收到注销通知")
		}

		// 执行注销
		err = s.registry.Deregister(instance1.ID)
		s.NoError(err)

		// 验证实例已被删除
		services, err := s.registry.GetService(instance1.Name)
		s.NoError(err)
		s.Empty(services)

		// 验证订阅者收到通知
		select {
		case services := <-ch:
			s.Empty(services)
		case <-time.After(time.Second):
			s.Fail("未收到注销通知")
		}

		// 再次注销应该返回错误
		err = s.registry.Deregister(instance1.ID)
		s.Error(err)
	})

	s.Run("注销多个实例中的一个", func() {
		// 注册两个实例
		err := s.registry.Register(instance1)
		s.NoError(err)
		err = s.registry.Register(instance2)
		s.NoError(err)

		// 创建订阅
		ch, err := s.registry.Subscribe(instance1.Name)
		s.NoError(err)
		s.NotEmpty(<-ch)

		// 注销第一个实例
		err = s.registry.Deregister(instance1.ID)
		s.NoError(err)

		// 验证剩余实例
		services, err := s.registry.GetService(instance1.Name)
		s.NoError(err)
		s.Len(services, 1)
		s.Equal(instance2.ID, services[0].ID)

		// 验证订阅通知
		select {
		case services := <-ch:
			s.Len(services, 1)
			s.Equal(instance2.ID, services[0].ID)
		case <-time.After(time.Second):
			s.Fail("未收到注销通知")
		}
	})
}

func TestRegistrySuite(t *testing.T) {
	suite.Run(t, new(RegistryTestSuite))
}
