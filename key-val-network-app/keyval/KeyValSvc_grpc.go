//Generated by gRPC Go plugin
//If you make any local changes, they will be lost
//source: keyval

package keyval

import (
	flatbuffers "github.com/google/flatbuffers/go"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Client API for KeyValSvc service
type KeyValSvcClient interface {
	Put(ctx context.Context, in *flatbuffers.Builder,
		opts ...grpc.CallOption) (*PutRsp, error)
	Get(ctx context.Context, in *flatbuffers.Builder,
		opts ...grpc.CallOption) (*GetRsp, error)
}

type keyValSvcClient struct {
	cc *grpc.ClientConn
}

func NewKeyValSvcClient(cc *grpc.ClientConn) KeyValSvcClient {
	return &keyValSvcClient{cc}
}

func (c *keyValSvcClient) Put(ctx context.Context, in *flatbuffers.Builder,
	opts ...grpc.CallOption) (*PutRsp, error) {
	out := new(PutRsp)
	err := grpc.Invoke(ctx, "/keyval.KeyValSvc/Put", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyValSvcClient) Get(ctx context.Context, in *flatbuffers.Builder,
	opts ...grpc.CallOption) (*GetRsp, error) {
	out := new(GetRsp)
	err := grpc.Invoke(ctx, "/keyval.KeyValSvc/Get", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for KeyValSvc service
type KeyValSvcServer interface {
	Put(context.Context, *PutReq) (*flatbuffers.Builder, error)
	Get(context.Context, *GetReq) (*flatbuffers.Builder, error)
}

func RegisterKeyValSvcServer(s *grpc.Server, srv KeyValSvcServer) {
	s.RegisterService(&_KeyValSvc_serviceDesc, srv)
}

func _KeyValSvc_Put_Handler(srv interface{}, ctx context.Context,
	dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyValSvcServer).Put(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keyval.KeyValSvc/Put",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyValSvcServer).Put(ctx, req.(*PutReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyValSvc_Get_Handler(srv interface{}, ctx context.Context,
	dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyValSvcServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keyval.KeyValSvc/Get",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyValSvcServer).Get(ctx, req.(*GetReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _KeyValSvc_serviceDesc = grpc.ServiceDesc{
	ServiceName: "keyval.KeyValSvc",
	HandlerType: (*KeyValSvcServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Put",
			Handler:    _KeyValSvc_Put_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _KeyValSvc_Get_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}
