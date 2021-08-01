// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/txfees/query.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types/query"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

func init() { proto.RegisterFile("osmosis/txfees/query.proto", fileDescriptor_032578a1369c2e84) }

var fileDescriptor_032578a1369c2e84 = []byte{
	// 187 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x34, 0xce, 0xb1, 0x0e, 0x82, 0x30,
	0x10, 0x06, 0x60, 0x18, 0xd4, 0x84, 0xd1, 0x91, 0x98, 0x3e, 0x80, 0x89, 0xbd, 0xa0, 0x6f, 0xe0,
	0xe4, 0xea, 0xea, 0xd6, 0x92, 0x5a, 0x9b, 0x40, 0xaf, 0x72, 0xc5, 0xc0, 0x5b, 0xf8, 0x58, 0x8e,
	0x8c, 0x8e, 0x06, 0x5e, 0xc4, 0x40, 0x61, 0xbb, 0xcb, 0x7d, 0xf7, 0xe7, 0x4f, 0x52, 0xa4, 0x12,
	0xc9, 0x10, 0xf8, 0xe6, 0xae, 0x14, 0xc1, 0xb3, 0x56, 0x55, 0xcb, 0x5d, 0x85, 0x1e, 0xb7, 0xcb,
	0xad, 0x10, 0x92, 0xf8, 0x3c, 0xf3, 0xe0, 0xd2, 0x9d, 0x46, 0xd4, 0x85, 0x02, 0xe1, 0x0c, 0x08,
	0x6b, 0xd1, 0x0b, 0x6f, 0xd0, 0x52, 0xf8, 0x4c, 0xf7, 0xf9, 0xc4, 0x41, 0x0a, 0x52, 0x21, 0x12,
	0x5e, 0x99, 0x54, 0x5e, 0x64, 0xe0, 0x84, 0x36, 0x76, 0xc2, 0xc1, 0x1e, 0x37, 0xc9, 0xea, 0x3a,
	0x8a, 0xf3, 0xe5, 0xd3, 0xb3, 0xb8, 0xeb, 0x59, 0xfc, 0xeb, 0x59, 0xfc, 0x1e, 0x58, 0xd4, 0x0d,
	0x2c, 0xfa, 0x0e, 0x2c, 0xba, 0x71, 0x6d, 0xfc, 0xa3, 0x96, 0x3c, 0xc7, 0x12, 0xe6, 0x1e, 0x87,
	0xb1, 0xd4, 0xb2, 0x40, 0xb3, 0xd4, 0xf7, 0xad, 0x53, 0x24, 0xd7, 0x53, 0xf2, 0xe9, 0x1f, 0x00,
	0x00, 0xff, 0xff, 0x20, 0xf3, 0xe8, 0xb1, 0xdd, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

// QueryServer is the server API for Query service.
type QueryServer interface {
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "osmosislabs.osmosis.txfees.Query",
	HandlerType: (*QueryServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "osmosis/txfees/query.proto",
}
