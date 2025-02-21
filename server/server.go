package server

import (
	"context"
	"reflect"
	"sync"

	"lisen/l-rpc/protocol"
	"lisen/l-rpc/transport"
)

// Service 表示一个服务
type Service struct {
	name     string
	rcvr     reflect.Value
	typ      reflect.Type
	methods  map[string]*MethodType
}

// MethodType 表示一个方法
type MethodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

// Server RPC 服务端
type Server struct {
	serviceMap sync.Map
	transport  transport.Transport
}

func NewServer() *Server {
	return &Server{}
}

// Register 注册服务
func (s *Server) Register(rcvr interface{}) error {
	service := new(Service)
	service.rcvr = reflect.ValueOf(rcvr)
	service.typ = reflect.TypeOf(rcvr)
	service.name = reflect.Indirect(service.rcvr).Type().Name()
	service.methods = make(map[string]*MethodType)

	// 注册方法
	for i := 0; i < service.typ.NumMethod(); i++ {
		method := service.typ.Method(i)
		mtype := method.Type

		// 方法必须是导出的
		if method.PkgPath != "" {
			continue
		}

		// 方法必须有三个入参: receiver, *args, *reply
		if mtype.NumIn() != 3 {
			continue
		}

		// 方法必须有一个出参: error
		if mtype.NumOut() != 1 {
			continue
		}

		// 第一个参数必须是 context.Context
		if !mtype.In(1).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			continue
		}

		argType := mtype.In(1)
		replyType := mtype.In(2)

		service.methods[method.Name] = &MethodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
	}

	if len(service.methods) == 0 {
		return ErrNoAvailableMethods
	}

	s.serviceMap.Store(service.name, service)
	return nil
}

// Start 启动服务
func (s *Server) Start(addr string) error {
	server, err := transport.NewServer(addr)
	if err != nil {
		return err
	}

	return server.Accept(s.handleRequest)
}

// handleRequest 处理请求
func (s *Server) handleRequest(trans transport.Transport) {
	for {
		// 接收请求
		data, err := trans.Receive()
		if err != nil {
			return
		}

		// 解码请求
		codec := protocol.NewDefaultCodec()
		msg, err := codec.Decode(data)
		if err != nil {
			return
		}

		// 处理请求
		go s.processRequest(msg, trans)
	}
}

// processRequest 处理单个请求
func (s *Server) processRequest(req *protocol.Message, trans transport.Transport) {
	resp := &protocol.Message{
		Header: &protocol.Header{
			ID:   req.Header.ID,
			Type: protocol.TypeResponse,
		},
	}

	svc, ok := s.serviceMap.Load(req.Header.ServiceName)
	if !ok {
		resp.Header.Error = ErrServiceNotFound.Error()
		s.sendResponse(resp, trans)
		return
	}

	service := svc.(*Service)
	mtype := service.methods[req.Header.MethodName]
	if mtype == nil {
		resp.Header.Error = ErrMethodNotFound.Error()
		s.sendResponse(resp, trans)
		return
	}

	// 创建参数
	argv := reflect.New(mtype.ArgType.Elem())
	replyv := reflect.New(mtype.ReplyType.Elem())

	// 解码参数
	if err := decode(req.Data, argv.Interface()); err != nil {
		resp.Header.Error = err.Error()
		s.sendResponse(resp, trans)
		return
	}

	// 调用方法
	ctx := context.Background()
	returnValues := mtype.method.Func.Call([]reflect.Value{
		service.rcvr,
		reflect.ValueOf(ctx),
		argv,
		replyv,
	})

	// 处理返回值
	if err := returnValues[0].Interface(); err != nil {
		resp.Header.Error = err.(error).Error()
		s.sendResponse(resp, trans)
		return
	}

	// 编码响应
	resp.Data, _ = encode(replyv.Interface())
	s.sendResponse(resp, trans)
}

func (s *Server) sendResponse(resp *protocol.Message, trans transport.Transport) {
	codec := protocol.NewDefaultCodec()
	data, err := codec.Encode(resp)
	if err != nil {
		return
	}
	trans.Send(data)
}