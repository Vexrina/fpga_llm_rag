// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: floatweaver.proto

package floatweaver

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	EmbedService_Embed_FullMethodName = "/floatweaver.EmbedService/Embed"
)

// EmbedServiceClient is the client API for EmbedService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EmbedServiceClient interface {
	Embed(ctx context.Context, in *EmbedRequest, opts ...grpc.CallOption) (*EmbedResponse, error)
}

type embedServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewEmbedServiceClient(cc grpc.ClientConnInterface) EmbedServiceClient {
	return &embedServiceClient{cc}
}

func (c *embedServiceClient) Embed(ctx context.Context, in *EmbedRequest, opts ...grpc.CallOption) (*EmbedResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmbedResponse)
	err := c.cc.Invoke(ctx, EmbedService_Embed_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EmbedServiceServer is the server API for EmbedService service.
// All implementations must embed UnimplementedEmbedServiceServer
// for forward compatibility.
type EmbedServiceServer interface {
	Embed(context.Context, *EmbedRequest) (*EmbedResponse, error)
	mustEmbedUnimplementedEmbedServiceServer()
}

// UnimplementedEmbedServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedEmbedServiceServer struct{}

func (UnimplementedEmbedServiceServer) Embed(context.Context, *EmbedRequest) (*EmbedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Embed not implemented")
}
func (UnimplementedEmbedServiceServer) mustEmbedUnimplementedEmbedServiceServer() {}
func (UnimplementedEmbedServiceServer) testEmbeddedByValue()                      {}

// UnsafeEmbedServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EmbedServiceServer will
// result in compilation errors.
type UnsafeEmbedServiceServer interface {
	mustEmbedUnimplementedEmbedServiceServer()
}

func RegisterEmbedServiceServer(s grpc.ServiceRegistrar, srv EmbedServiceServer) {
	// If the following call pancis, it indicates UnimplementedEmbedServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&EmbedService_ServiceDesc, srv)
}

func _EmbedService_Embed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmbedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmbedServiceServer).Embed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EmbedService_Embed_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmbedServiceServer).Embed(ctx, req.(*EmbedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// EmbedService_ServiceDesc is the grpc.ServiceDesc for EmbedService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EmbedService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "floatweaver.EmbedService",
	HandlerType: (*EmbedServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Embed",
			Handler:    _EmbedService_Embed_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "floatweaver.proto",
}
