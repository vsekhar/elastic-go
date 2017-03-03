// Code generated by protoc-gen-go.
// source: remoteapi.proto
// DO NOT EDIT!

/*
Package api is a generated protocol buffer package.

It is generated from these files:
	remoteapi.proto

It has these top-level messages:
	AllocRequest
	AllocResponse
	GetRequest
	GetResponse
	SetRequest
	SetResponse
*/
package api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type AllocRequest struct {
	Size uint64 `protobuf:"varint,1,opt,name=size" json:"size,omitempty"`
}

func (m *AllocRequest) Reset()                    { *m = AllocRequest{} }
func (m *AllocRequest) String() string            { return proto.CompactTextString(m) }
func (*AllocRequest) ProtoMessage()               {}
func (*AllocRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *AllocRequest) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

type AllocResponse struct {
	Id uint64 `protobuf:"fixed64,1,opt,name=id" json:"id,omitempty"`
}

func (m *AllocResponse) Reset()                    { *m = AllocResponse{} }
func (m *AllocResponse) String() string            { return proto.CompactTextString(m) }
func (*AllocResponse) ProtoMessage()               {}
func (*AllocResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *AllocResponse) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

type GetRequest struct {
	Id uint64 `protobuf:"fixed64,1,opt,name=id" json:"id,omitempty"`
}

func (m *GetRequest) Reset()                    { *m = GetRequest{} }
func (m *GetRequest) String() string            { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()               {}
func (*GetRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *GetRequest) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

type GetResponse struct {
	Value []byte `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *GetResponse) Reset()                    { *m = GetResponse{} }
func (m *GetResponse) String() string            { return proto.CompactTextString(m) }
func (*GetResponse) ProtoMessage()               {}
func (*GetResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *GetResponse) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type SetRequest struct {
	Id    uint64 `protobuf:"fixed64,1,opt,name=id" json:"id,omitempty"`
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *SetRequest) Reset()                    { *m = SetRequest{} }
func (m *SetRequest) String() string            { return proto.CompactTextString(m) }
func (*SetRequest) ProtoMessage()               {}
func (*SetRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *SetRequest) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *SetRequest) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type SetResponse struct {
}

func (m *SetResponse) Reset()                    { *m = SetResponse{} }
func (m *SetResponse) String() string            { return proto.CompactTextString(m) }
func (*SetResponse) ProtoMessage()               {}
func (*SetResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func init() {
	proto.RegisterType((*AllocRequest)(nil), "api.AllocRequest")
	proto.RegisterType((*AllocResponse)(nil), "api.AllocResponse")
	proto.RegisterType((*GetRequest)(nil), "api.GetRequest")
	proto.RegisterType((*GetResponse)(nil), "api.GetResponse")
	proto.RegisterType((*SetRequest)(nil), "api.SetRequest")
	proto.RegisterType((*SetResponse)(nil), "api.SetResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for RemoteRuntime service

type RemoteRuntimeClient interface {
	Alloc(ctx context.Context, in *AllocRequest, opts ...grpc.CallOption) (*AllocResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	Set(ctx context.Context, in *SetRequest, opts ...grpc.CallOption) (*SetResponse, error)
}

type remoteRuntimeClient struct {
	cc *grpc.ClientConn
}

func NewRemoteRuntimeClient(cc *grpc.ClientConn) RemoteRuntimeClient {
	return &remoteRuntimeClient{cc}
}

func (c *remoteRuntimeClient) Alloc(ctx context.Context, in *AllocRequest, opts ...grpc.CallOption) (*AllocResponse, error) {
	out := new(AllocResponse)
	err := grpc.Invoke(ctx, "/api.RemoteRuntime/Alloc", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remoteRuntimeClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := grpc.Invoke(ctx, "/api.RemoteRuntime/Get", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remoteRuntimeClient) Set(ctx context.Context, in *SetRequest, opts ...grpc.CallOption) (*SetResponse, error) {
	out := new(SetResponse)
	err := grpc.Invoke(ctx, "/api.RemoteRuntime/Set", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for RemoteRuntime service

type RemoteRuntimeServer interface {
	Alloc(context.Context, *AllocRequest) (*AllocResponse, error)
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Set(context.Context, *SetRequest) (*SetResponse, error)
}

func RegisterRemoteRuntimeServer(s *grpc.Server, srv RemoteRuntimeServer) {
	s.RegisterService(&_RemoteRuntime_serviceDesc, srv)
}

func _RemoteRuntime_Alloc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AllocRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemoteRuntimeServer).Alloc(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.RemoteRuntime/Alloc",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemoteRuntimeServer).Alloc(ctx, req.(*AllocRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RemoteRuntime_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemoteRuntimeServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.RemoteRuntime/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemoteRuntimeServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RemoteRuntime_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemoteRuntimeServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.RemoteRuntime/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemoteRuntimeServer).Set(ctx, req.(*SetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _RemoteRuntime_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.RemoteRuntime",
	HandlerType: (*RemoteRuntimeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Alloc",
			Handler:    _RemoteRuntime_Alloc_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _RemoteRuntime_Get_Handler,
		},
		{
			MethodName: "Set",
			Handler:    _RemoteRuntime_Set_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "remoteapi.proto",
}

func init() { proto.RegisterFile("remoteapi.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 217 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x74, 0x90, 0xb1, 0x4e, 0x86, 0x30,
	0x14, 0x85, 0x03, 0x3f, 0x30, 0x5c, 0x40, 0xf4, 0xc6, 0xc1, 0x10, 0x13, 0x4d, 0x5d, 0x98, 0x18,
	0xf0, 0x09, 0x9c, 0xd8, 0xdb, 0x27, 0x40, 0xb9, 0x43, 0x13, 0xa0, 0x95, 0x16, 0x07, 0x5f, 0xc3,
	0x17, 0x36, 0x16, 0xb4, 0x0d, 0x89, 0x5b, 0x7b, 0xef, 0x77, 0x4e, 0xcf, 0x29, 0x54, 0x2b, 0xcd,
	0xca, 0xd2, 0xa0, 0x65, 0xab, 0x57, 0x65, 0x15, 0x5e, 0x06, 0x2d, 0x19, 0x83, 0xe2, 0x65, 0x9a,
	0xd4, 0x1b, 0xa7, 0xf7, 0x8d, 0x8c, 0x45, 0x84, 0xc4, 0xc8, 0x4f, 0xba, 0x8b, 0x1e, 0xa3, 0x26,
	0xe1, 0xee, 0xcc, 0x1e, 0xa0, 0x3c, 0x18, 0xa3, 0xd5, 0x62, 0x08, 0xaf, 0x20, 0x96, 0xa3, 0x43,
	0x32, 0x1e, 0xcb, 0x91, 0xdd, 0x03, 0xf4, 0x64, 0x7f, 0x2d, 0xce, 0xdb, 0x27, 0xc8, 0xdd, 0xf6,
	0x10, 0xdf, 0x42, 0xfa, 0x31, 0x4c, 0xdb, 0xfe, 0x44, 0xc1, 0xf7, 0x0b, 0xeb, 0x00, 0xc4, 0xbf,
	0x16, 0x5e, 0x13, 0x87, 0x9a, 0x12, 0x72, 0xe1, 0x8d, 0xbb, 0xaf, 0x08, 0x4a, 0xee, 0x3a, 0xf2,
	0x6d, 0xb1, 0x72, 0x26, 0x6c, 0x21, 0x75, 0xc1, 0xf1, 0xa6, 0xfd, 0xa9, 0x1d, 0x16, 0xad, 0x31,
	0x1c, 0x1d, 0xd1, 0x1a, 0xb8, 0xf4, 0x64, 0xb1, 0x72, 0x2b, 0xdf, 0xa8, 0xbe, 0xf6, 0x03, 0x4f,
	0x8a, 0x3f, 0x52, 0x9c, 0xc9, 0x20, 0xd5, 0x6b, 0xe6, 0x3e, 0xfb, 0xf9, 0x3b, 0x00, 0x00, 0xff,
	0xff, 0x3b, 0x0f, 0x40, 0x69, 0x7f, 0x01, 0x00, 0x00,
}
