package main

import (
	"context"
	"log"
	"time"

	"github.com/eason-lee/l-rpc/client"
	"github.com/eason-lee/l-rpc/examples/proto"
	"github.com/eason-lee/l-rpc/registry"
)

func main() {
	// 创建注册中心
	reg := registry.NewInMemoryRegistry()

	// 创建负载均衡器
	balancer := registry.NewRandomBalancer()

	// 创建 RPC 客户端
	c := client.NewClient(reg, balancer)

	// 同步调用示例
	{
		req := &proto.GetUserRequest{ID: 1}
		resp := &proto.GetUserResponse{}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		err := c.Call(ctx, "UserService.GetUser", req, resp)
		if err != nil {
			log.Fatalf("调用失败: %v", err)
		}
		log.Printf("同步调用结果: %+v", resp.User)
	}

	// 异步调用示例
	{
		req := &proto.GetUserRequest{ID: 2}
		resp := &proto.GetUserResponse{}

		done := make(chan *client.Call, 1)
		_ = c.Go("UserService.GetUser", req, resp, done)

		// 等待响应
		select {
		case <-time.After(time.Second * 3):
			log.Fatal("调用超时")
		case call := <-done:
			if call.Error != nil {
				log.Fatalf("调用失败: %v", call.Error)
			}
			log.Printf("异步调用结果: %+v", resp.User)
		}
	}
}