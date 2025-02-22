package main

import (
	"context"
	"log"

	"lisen/l-rpc/examples/proto"
	"lisen/l-rpc/registry"
	"lisen/l-rpc/server"
)

// UserService 用户服务
type UserService struct{}

// GetUser 获取用户信息
func (s *UserService) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	// 模拟数据库查询
	user := &proto.User{
		ID:   req.ID,
		Name: "张三",
		Age:  25,
	}
	return &proto.GetUserResponse{User: user}, nil
}

func main() {
	// 创建注册中心
	reg := registry.NewInMemoryRegistry()

	// 创建 RPC 服务器
	srv := server.NewServer()

	// 注册服务
	if err := srv.Register(new(UserService)); err != nil {
		log.Fatalf("注册服务失败: %v", err)
	}

	// 注册服务实例
	instance := &registry.ServiceInstance{
		Name:      "UserService",
		Endpoints: []string{"127.0.0.1:8080"},
	}
	if err := reg.Register(instance); err != nil {
		log.Fatalf("注册实例失败: %v", err)
	}

	// 启动服务
	log.Println("启动 RPC 服务器...")
	if err := srv.Start(":8080"); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}