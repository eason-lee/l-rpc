package integration

import (
	"context"
	"testing"
	"time"

	"github.com/eason-lee/l-rpc/client"
	"github.com/eason-lee/l-rpc/registry"
	"github.com/eason-lee/l-rpc/server"
)

func BenchmarkRPC(b *testing.B) {
	// 启动服务端
	reg := registry.NewInMemoryRegistry()
	srv := server.NewServer()

	if err := srv.Register(new(EchoService)); err != nil {
		b.Fatalf("注册服务失败: %v", err)
	}

	instance := &registry.ServiceInstance{
		Name:      "EchoService",
		Endpoints: []string{"127.0.0.1:8889"},
	}
	if err := reg.Register(instance); err != nil {
		b.Fatalf("注册实例失败: %v", err)
	}

	go srv.Start(":8889")
	time.Sleep(time.Second)

	// 创建客户端
	cli := client.NewClient(reg, registry.NewRandomBalancer())

	// 准备请求参数
	req := &EchoRequest{Message: "hello"}
	resp := &EchoResponse{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := cli.Call(context.Background(), "EchoService.Echo", req, resp)
			if err != nil {
				b.Errorf("调用失败: %v", err)
			}
		}
	})
}
