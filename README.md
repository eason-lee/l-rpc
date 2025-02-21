# L-RPC

一个轻量级、高性能的 Go RPC 框架，提供服务注册、服务发现、负载均衡等功能。

## 特性

### 已实现功能

- **服务注册与发现**
  - 内存注册中心
  - 服务实例管理
  - 服务状态监控
  - 服务订阅与通知

- **负载均衡**
  - 随机负载均衡
  - 轮询负载均衡
  - 加权随机负载均衡
  - 最小活跃数负载均衡

- **健康检查**
  - HTTP 健康检查
  - 心跳检测
  - 自动剔除不健康实例

- **传输层**
  - 连接池管理
  - 数据压缩
  - 安全传输

### 待实现功能

- [ ] 服务熔断
- [ ] 服务限流
- [ ] 服务降级
- [ ] 链路追踪
- [ ] 监控指标

## 安装

```bash
go get github.com/eason-lee/l-rpc
```

## 快速开始

### 服务端

```go
package main

import (
    "github.com/eason-lee/l-rpc/registry"
    "github.com/eason-lee/l-rpc/transport"
)

func main() {
    // 创建服务实例
    instance := &registry.ServiceInstance{
        ID:      "service-1",
        Name:    "example-service",
        Version: "1.0.0",
        Endpoints: []string{"localhost:8080"},
    }

    // 创建注册中心
    reg := registry.NewInMemoryRegistry()

    // 注册服务
    reg.Register(instance)
}
```

### 客户端

```go
package main

import (
    "github.com/eason-lee/l-rpc/registry"
)

func main() {
    // 创建注册中心客户端
    reg := registry.NewInMemoryRegistry()

    // 创建负载均衡器
    balancer := registry.NewRandomBalancer()

    // 获取服务实例
    instance, err := reg.SelectInstance("example-service", balancer)
    if err != nil {
        panic(err)
    }
}
```

## 项目结构

```
.
├── protocol/       # 协议定义和编解码
├── registry/       # 服务注册与发现
├── transport/      # 网络传输层
└── examples/       # 示例代码
```

## 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。