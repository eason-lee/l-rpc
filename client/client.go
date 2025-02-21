package client

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	"lisen/l-rpc/protocol"
	"lisen/l-rpc/registry"
	"lisen/l-rpc/transport"
)

// Client RPC客户端
type Client struct {
	seq        uint64
	registry   registry.Registry
	balancer   registry.LoadBalancer
	transport  *transport.Client
	pendingMap sync.Map
}

// Call 表示一个待处理的调用
type Call struct {
	ServiceMethod string      // 格式: "服务.方法"
	Args         interface{} // 参数
	Reply        interface{} // 响应
	Error        error       // 错误信息
	Done         chan *Call  // 调用完成时的通知通道
}

func NewClient(reg registry.Registry, balancer registry.LoadBalancer) *Client {
	return &Client{
		registry: reg,
		balancer: balancer,
	}
}

// Call 同步调用
func (c *Client) Call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	call := c.Go(serviceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		return ctx.Err()
	case call := <-call.Done:
		return call.Error
	}
}

// Go 异步调用
func (c *Client) Go(serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call {
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:         args,
		Reply:        reply,
		Done:         done,
	}

	go c.send(call)
	return call
}

func (c *Client) send(call *Call) {
	ctx := context.Background()
	// 生成请求ID
	seq := atomic.AddUint64(&c.seq, 1)
	
	// 构造请求消息
	req := &protocol.Message{
		Header: &protocol.Header{
			ID:          seq,
			Type:        protocol.TypeRequest,
			ServiceName: call.ServiceMethod,
			MethodName:  getMethodFromServiceMethod(call.ServiceMethod),
		},
	}

	// 编码参数
	data, err := encode(call.Args)
	if err != nil {
		call.Error = err
		call.done()
		return
	}
	req.Data = data

	// 获取服务实例
	instance, err := c.registry.SelectInstance(call.ServiceMethod, c.balancer)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// 建立连接
	if c.transport == nil {
		c.transport, err = transport.NewClient("tcp", instance.Endpoints[0])
		if err != nil {
			call.Error = err
			call.done()
			return
		}
	}

	// 存储调用信息
	c.pendingMap.Store(seq, call)

	// 发送请求
	resp, err := c.transport.Send(ctx, req)
	if err != nil {
		c.pendingMap.Delete(seq)
		call.Error = err
		call.done()
		return
	}

	// 处理响应
	if resp.Header.Error != "" {
		call.Error = ErrorFromString(resp.Header.Error)
		call.done()
		return
	}

	// 解码响应
	err = decode(resp.Data, call.Reply)
	if err != nil {
		call.Error = err
	}
	call.done()
}

func (call *Call) done() {
	if call.Done != nil {
		call.Done <- call
	}
}

// 从 ServiceMethod 中解析出方法名
func getMethodFromServiceMethod(serviceMethod string) string {
	if i := strings.LastIndex(serviceMethod, "."); i >= 0 {
		return serviceMethod[i+1:]
	}
	return serviceMethod
}