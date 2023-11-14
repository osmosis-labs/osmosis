// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/cosmwasmpool/v1beta1/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

func init() {
	proto.RegisterFile("osmosis/cosmwasmpool/v1beta1/tx.proto", fileDescriptor_93020a1fe95e1be8)
}

var fileDescriptor_93020a1fe95e1be8 = []byte{
	// 179 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0xcd, 0x2f, 0xce, 0xcd,
	0x2f, 0xce, 0x2c, 0xd6, 0x4f, 0xce, 0x2f, 0xce, 0x2d, 0x4f, 0x2c, 0xce, 0x2d, 0xc8, 0xcf, 0xcf,
	0xd1, 0x2f, 0x33, 0x4c, 0x4a, 0x2d, 0x49, 0x34, 0xd4, 0x2f, 0xa9, 0xd0, 0x2b, 0x28, 0xca, 0x2f,
	0xc9, 0x17, 0x92, 0x81, 0x2a, 0xd3, 0x43, 0x56, 0xa6, 0x07, 0x55, 0x26, 0x25, 0x92, 0x9e, 0x9f,
	0x9e, 0x0f, 0x56, 0xa8, 0x0f, 0x62, 0x41, 0xf4, 0x48, 0xc9, 0x25, 0x83, 0x35, 0xe9, 0x27, 0x25,
	0x16, 0xa7, 0xc2, 0x4d, 0x4c, 0xce, 0xcf, 0xcc, 0x83, 0xc8, 0x1b, 0xb1, 0x72, 0x31, 0xfb, 0x16,
	0xa7, 0x3b, 0x05, 0x9d, 0x78, 0x24, 0xc7, 0x78, 0xe1, 0x91, 0x1c, 0xe3, 0x83, 0x47, 0x72, 0x8c,
	0x13, 0x1e, 0xcb, 0x31, 0x5c, 0x78, 0x2c, 0xc7, 0x70, 0xe3, 0xb1, 0x1c, 0x43, 0x94, 0x45, 0x7a,
	0x66, 0x49, 0x46, 0x69, 0x92, 0x5e, 0x72, 0x7e, 0xae, 0x3e, 0xd4, 0x7e, 0xdd, 0x9c, 0xc4, 0xa4,
	0x62, 0x18, 0x47, 0xbf, 0xcc, 0xc8, 0x40, 0xbf, 0x02, 0xd5, 0xe5, 0x25, 0x95, 0x05, 0xa9, 0xc5,
	0x49, 0x6c, 0x60, 0x1b, 0x8c, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x05, 0x92, 0xb8, 0xaf, 0xde,
	0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "osmosis.cosmwasmpool.v1beta1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "osmosis/cosmwasmpool/v1beta1/tx.proto",
}
