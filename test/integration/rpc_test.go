package integration

import (
	"context"
	"testing"
	"time"

	"github.com/eason-lee/l-rpc/client"
	"github.com/eason-lee/l-rpc/registry"
	"github.com/eason-lee/l-rpc/server"
)

// 测试服务接口
type EchoService struct{}

type EchoRequest struct {
	Message string
}

type EchoResponse struct {
	Message string
}

// Echo 修改参数为指针类型
func (s *EchoService) Echo(ctx context.Context, req *EchoRequest, reply *EchoResponse) error {
	reply.Message = req.Message
	return nil
}

func TestBasicRPC(t *testing.T) {
	// 1. 启动服务端
	reg := registry.NewInMemoryRegistry()
	srv := server.NewServer()

	// 注册服务时使用指针类型
	if err := srv.Register(&EchoService{}); err != nil {
		t.Fatalf("注册服务失败: %v", err)
	}

	// 注册服务实例
	instance := &registry.ServiceInstance{
		Name:      "EchoService",
		Endpoints: []string{"127.0.0.1:8888"},
	}
	if err := reg.Register(instance); err != nil {
		t.Fatalf("注册实例失败: %v", err)
	}

	// 启动服务器
	go func() {
		if err := srv.Start(":8888"); err != nil {
			t.Errorf("服务启动失败: %v", err)
		}
	}()
	time.Sleep(time.Second) // 等待服务启动

	// 2. 创建客户端
	cli := client.NewClient(reg, registry.NewRandomBalancer())

	// 3. 执行调用测试
	t.Run("同步调用", func(t *testing.T) {
		req := &EchoRequest{Message: "hello"}
		resp := &EchoResponse{}

		err := cli.Call(context.Background(), "EchoService.Echo", req, resp)
		if err != nil {
			t.Fatalf("调用失败: %v", err)
		}

		if resp.Message != req.Message {
			t.Errorf("响应不匹配, 期望: %s, 实际: %s", req.Message, resp.Message)
		}
	})

	t.Run("异步调用", func(t *testing.T) {
		req := &EchoRequest{Message: "world"}
		resp := &EchoResponse{}

		done := make(chan *client.Call, 1)
		_ = cli.Go("EchoService.Echo", req, resp, done)

		select {
		case <-time.After(time.Second * 3):
			t.Fatal("调用超时")
		case call := <-done:
			if call.Error != nil {
				t.Fatalf("调用失败: %v", call.Error)
			}
			if resp.Message != req.Message {
				t.Errorf("响应不匹配, 期望: %s, 实际: %s", req.Message, resp.Message)
			}
		}
	})
}
