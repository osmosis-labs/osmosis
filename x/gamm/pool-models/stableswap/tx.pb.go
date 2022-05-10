// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/gamm/pool-models/stableswap/tx.proto

package stableswap

import (
	context "context"
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
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

type MsgCreateStableswapPool struct {
	Sender               string                                   `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty" yaml:"sender"`
	PoolParams           *PoolParams                              `protobuf:"bytes,2,opt,name=poolParams,proto3" json:"poolParams,omitempty" yaml:"pool_params"`
	InitialPoolLiquidity github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,3,rep,name=initial_pool_liquidity,json=initialPoolLiquidity,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"initial_pool_liquidity"`
	FuturePoolGovernor   string                                   `protobuf:"bytes,4,opt,name=future_pool_governor,json=futurePoolGovernor,proto3" json:"future_pool_governor,omitempty" yaml:"future_pool_governor"`
}

func (m *MsgCreateStableswapPool) Reset()         { *m = MsgCreateStableswapPool{} }
func (m *MsgCreateStableswapPool) String() string { return proto.CompactTextString(m) }
func (*MsgCreateStableswapPool) ProtoMessage()    {}
func (*MsgCreateStableswapPool) Descriptor() ([]byte, []int) {
	return fileDescriptor_46b7c8a0f24de97c, []int{0}
}
func (m *MsgCreateStableswapPool) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateStableswapPool) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateStableswapPool.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateStableswapPool) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateStableswapPool.Merge(m, src)
}
func (m *MsgCreateStableswapPool) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateStableswapPool) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateStableswapPool.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateStableswapPool proto.InternalMessageInfo

func (m *MsgCreateStableswapPool) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

func (m *MsgCreateStableswapPool) GetPoolParams() *PoolParams {
	if m != nil {
		return m.PoolParams
	}
	return nil
}

func (m *MsgCreateStableswapPool) GetInitialPoolLiquidity() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.InitialPoolLiquidity
	}
	return nil
}

func (m *MsgCreateStableswapPool) GetFuturePoolGovernor() string {
	if m != nil {
		return m.FuturePoolGovernor
	}
	return ""
}

type MsgCreateStableswapPoolResponse struct {
	PoolID uint64 `protobuf:"varint,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
}

func (m *MsgCreateStableswapPoolResponse) Reset()         { *m = MsgCreateStableswapPoolResponse{} }
func (m *MsgCreateStableswapPoolResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCreateStableswapPoolResponse) ProtoMessage()    {}
func (*MsgCreateStableswapPoolResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_46b7c8a0f24de97c, []int{1}
}
func (m *MsgCreateStableswapPoolResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateStableswapPoolResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateStableswapPoolResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateStableswapPoolResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateStableswapPoolResponse.Merge(m, src)
}
func (m *MsgCreateStableswapPoolResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateStableswapPoolResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateStableswapPoolResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateStableswapPoolResponse proto.InternalMessageInfo

func (m *MsgCreateStableswapPoolResponse) GetPoolID() uint64 {
	if m != nil {
		return m.PoolID
	}
	return 0
}

type MsgStableSwapAdjustScalingFactors struct {
	ScalingFactorGovernor string `protobuf:"bytes,1,opt,name=scaling_factor_governor,json=scalingFactorGovernor,proto3" json:"scaling_factor_governor,omitempty" yaml:"scaling_factor_governor"`
	PoolID                uint64 `protobuf:"varint,2,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
}

func (m *MsgStableSwapAdjustScalingFactors) Reset()         { *m = MsgStableSwapAdjustScalingFactors{} }
func (m *MsgStableSwapAdjustScalingFactors) String() string { return proto.CompactTextString(m) }
func (*MsgStableSwapAdjustScalingFactors) ProtoMessage()    {}
func (*MsgStableSwapAdjustScalingFactors) Descriptor() ([]byte, []int) {
	return fileDescriptor_46b7c8a0f24de97c, []int{2}
}
func (m *MsgStableSwapAdjustScalingFactors) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgStableSwapAdjustScalingFactors) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgStableSwapAdjustScalingFactors.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgStableSwapAdjustScalingFactors) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgStableSwapAdjustScalingFactors.Merge(m, src)
}
func (m *MsgStableSwapAdjustScalingFactors) XXX_Size() int {
	return m.Size()
}
func (m *MsgStableSwapAdjustScalingFactors) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgStableSwapAdjustScalingFactors.DiscardUnknown(m)
}

var xxx_messageInfo_MsgStableSwapAdjustScalingFactors proto.InternalMessageInfo

func (m *MsgStableSwapAdjustScalingFactors) GetScalingFactorGovernor() string {
	if m != nil {
		return m.ScalingFactorGovernor
	}
	return ""
}

func (m *MsgStableSwapAdjustScalingFactors) GetPoolID() uint64 {
	if m != nil {
		return m.PoolID
	}
	return 0
}

type MsgStableSwapAdjustScalingFactorsResponse struct {
}

func (m *MsgStableSwapAdjustScalingFactorsResponse) Reset() {
	*m = MsgStableSwapAdjustScalingFactorsResponse{}
}
func (m *MsgStableSwapAdjustScalingFactorsResponse) String() string {
	return proto.CompactTextString(m)
}
func (*MsgStableSwapAdjustScalingFactorsResponse) ProtoMessage() {}
func (*MsgStableSwapAdjustScalingFactorsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_46b7c8a0f24de97c, []int{3}
}
func (m *MsgStableSwapAdjustScalingFactorsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgStableSwapAdjustScalingFactorsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgStableSwapAdjustScalingFactorsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgStableSwapAdjustScalingFactorsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgStableSwapAdjustScalingFactorsResponse.Merge(m, src)
}
func (m *MsgStableSwapAdjustScalingFactorsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgStableSwapAdjustScalingFactorsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgStableSwapAdjustScalingFactorsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgStableSwapAdjustScalingFactorsResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgCreateStableswapPool)(nil), "osmosis.gamm.stableswap.v1beta1.MsgCreateStableswapPool")
	proto.RegisterType((*MsgCreateStableswapPoolResponse)(nil), "osmosis.gamm.stableswap.v1beta1.MsgCreateStableswapPoolResponse")
	proto.RegisterType((*MsgStableSwapAdjustScalingFactors)(nil), "osmosis.gamm.stableswap.v1beta1.MsgStableSwapAdjustScalingFactors")
	proto.RegisterType((*MsgStableSwapAdjustScalingFactorsResponse)(nil), "osmosis.gamm.stableswap.v1beta1.MsgStableSwapAdjustScalingFactorsResponse")
}

func init() {
	proto.RegisterFile("osmosis/gamm/pool-models/stableswap/tx.proto", fileDescriptor_46b7c8a0f24de97c)
}

var fileDescriptor_46b7c8a0f24de97c = []byte{
	// 571 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0x4d, 0x6e, 0xd3, 0x40,
	0x18, 0x8d, 0x93, 0x2a, 0x88, 0xa9, 0x58, 0x60, 0x85, 0x36, 0x04, 0xc9, 0x0e, 0x66, 0x93, 0xaa,
	0xc4, 0x43, 0xc3, 0x82, 0x9f, 0x55, 0x71, 0x51, 0x51, 0x11, 0x91, 0x82, 0xb3, 0xcb, 0x26, 0x1a,
	0xdb, 0x53, 0x33, 0x60, 0x7b, 0x8c, 0x67, 0x9c, 0x36, 0x3b, 0xb8, 0x01, 0xe2, 0x0c, 0x88, 0x05,
	0xd7, 0x60, 0xd3, 0x65, 0x97, 0xac, 0x0c, 0x4a, 0x6e, 0x90, 0x13, 0x20, 0x7b, 0xec, 0x24, 0x8b,
	0xa4, 0xa9, 0x60, 0x95, 0xc9, 0xcb, 0xfb, 0xde, 0xf7, 0xbe, 0xf7, 0x4d, 0x06, 0x3c, 0xa4, 0xcc,
	0xa7, 0x8c, 0x30, 0xe8, 0x22, 0xdf, 0x87, 0x21, 0xa5, 0x5e, 0xdb, 0xa7, 0x0e, 0xf6, 0x18, 0x64,
	0x1c, 0x59, 0x1e, 0x66, 0x67, 0x28, 0x84, 0xfc, 0x5c, 0x0f, 0x23, 0xca, 0xa9, 0xac, 0xe6, 0x6c,
	0x3d, 0x65, 0xeb, 0x0b, 0x86, 0x3e, 0x3a, 0xb0, 0x30, 0x47, 0x07, 0x0d, 0xc5, 0xce, 0x18, 0xd0,
	0x42, 0x0c, 0xc3, 0x1c, 0x84, 0x36, 0x25, 0x81, 0x10, 0x68, 0xd4, 0x5c, 0xea, 0xd2, 0xec, 0x08,
	0xd3, 0x53, 0x8e, 0x3e, 0xbb, 0x8e, 0x89, 0xc5, 0x71, 0x98, 0x32, 0x44, 0xa9, 0xf6, 0xa9, 0x02,
	0x76, 0xbb, 0xcc, 0x3d, 0x8a, 0x30, 0xe2, 0xb8, 0x3f, 0xa7, 0xf4, 0x28, 0xf5, 0xe4, 0x3d, 0x50,
	0x65, 0x38, 0x70, 0x70, 0x54, 0x97, 0x9a, 0x52, 0xeb, 0xa6, 0x71, 0x7b, 0x96, 0xa8, 0xb7, 0xc6,
	0xc8, 0xf7, 0x9e, 0x6b, 0x02, 0xd7, 0xcc, 0x9c, 0x20, 0xdb, 0x00, 0xa4, 0xa2, 0x3d, 0x14, 0x21,
	0x9f, 0xd5, 0xcb, 0x4d, 0xa9, 0xb5, 0xdd, 0xd9, 0xd7, 0x37, 0x4c, 0xab, 0xf7, 0xe6, 0x25, 0xc6,
	0xce, 0x2c, 0x51, 0x65, 0xa1, 0x9d, 0x0a, 0x0d, 0xc3, 0x0c, 0xd6, 0xcc, 0x25, 0x59, 0xf9, 0xb3,
	0x04, 0x76, 0x48, 0x40, 0x38, 0x41, 0x5e, 0x36, 0xc2, 0xd0, 0x23, 0x1f, 0x63, 0xe2, 0x10, 0x3e,
	0xae, 0x57, 0x9a, 0x95, 0xd6, 0x76, 0xe7, 0xae, 0x2e, 0xe2, 0xd3, 0xd3, 0xf8, 0xe6, 0x5d, 0x8e,
	0x28, 0x09, 0x8c, 0x47, 0x17, 0x89, 0x5a, 0xfa, 0xf1, 0x5b, 0x6d, 0xb9, 0x84, 0xbf, 0x8b, 0x2d,
	0xdd, 0xa6, 0x3e, 0xcc, 0xb3, 0x16, 0x1f, 0x6d, 0xe6, 0x7c, 0x80, 0x7c, 0x1c, 0x62, 0x96, 0x15,
	0x30, 0xb3, 0x96, 0xb7, 0x4a, 0x4d, 0xbe, 0x29, 0x1a, 0xc9, 0x6f, 0x41, 0xed, 0x34, 0xe6, 0x71,
	0x84, 0x85, 0x03, 0x97, 0x8e, 0x70, 0x14, 0xd0, 0xa8, 0xbe, 0x95, 0x25, 0xa4, 0xce, 0x12, 0xf5,
	0x9e, 0x98, 0x62, 0x15, 0x4b, 0x33, 0x65, 0x01, 0xa7, 0x9a, 0xaf, 0x0a, 0xf0, 0x18, 0xa8, 0x6b,
	0x36, 0x60, 0x62, 0x16, 0xd2, 0x80, 0x61, 0xf9, 0x01, 0xb8, 0x91, 0x09, 0x11, 0x27, 0x5b, 0xc5,
	0x96, 0x01, 0x26, 0x89, 0x5a, 0x4d, 0x29, 0x27, 0x2f, 0xcd, 0x6a, 0xfa, 0xd3, 0x89, 0xa3, 0x7d,
	0x93, 0xc0, 0xfd, 0x2e, 0x73, 0x85, 0x44, 0xff, 0x0c, 0x85, 0x2f, 0x9c, 0xf7, 0x31, 0xe3, 0x7d,
	0x1b, 0x79, 0x24, 0x70, 0x8f, 0x91, 0xcd, 0x69, 0xc4, 0xe4, 0x01, 0xd8, 0x65, 0x02, 0x19, 0x9e,
	0x66, 0xd0, 0x62, 0x06, 0xb1, 0x65, 0x6d, 0x96, 0xa8, 0x4a, 0xbe, 0xe5, 0xd5, 0x44, 0xcd, 0xbc,
	0xc3, 0x96, 0x45, 0x8b, 0x49, 0x96, 0x6d, 0x96, 0xd7, 0xda, 0xdc, 0x07, 0x7b, 0x1b, 0x5d, 0x16,
	0x83, 0x77, 0x7e, 0x96, 0x41, 0xa5, 0xcb, 0x5c, 0xf9, 0xab, 0x04, 0x6a, 0x2b, 0xef, 0xe8, 0xd3,
	0x8d, 0x97, 0x6c, 0x4d, 0xb6, 0x8d, 0xc3, 0x7f, 0xad, 0x9c, 0x6f, 0xe5, 0xbb, 0x04, 0x6a, 0x2b,
	0x33, 0x36, 0xae, 0x23, 0x7d, 0x75, 0x02, 0x8d, 0xd7, 0xff, 0xaf, 0x51, 0x18, 0x35, 0x06, 0x17,
	0x13, 0x45, 0xba, 0x9c, 0x28, 0xd2, 0x9f, 0x89, 0x22, 0x7d, 0x99, 0x2a, 0xa5, 0xcb, 0xa9, 0x52,
	0xfa, 0x35, 0x55, 0x4a, 0x83, 0xc3, 0xa5, 0xbf, 0x43, 0xde, 0xaf, 0xed, 0x21, 0x8b, 0x15, 0x5f,
	0xe0, 0xe8, 0x09, 0x3c, 0xbf, 0xea, 0x59, 0xb1, 0xaa, 0xd9, 0x3b, 0xf2, 0xf8, 0x6f, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x9c, 0xc5, 0xba, 0x63, 0x09, 0x05, 0x00, 0x00,
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
	CreateStableswapPool(ctx context.Context, in *MsgCreateStableswapPool, opts ...grpc.CallOption) (*MsgCreateStableswapPoolResponse, error)
	AdjustScalingFactors(ctx context.Context, in *MsgStableSwapAdjustScalingFactors, opts ...grpc.CallOption) (*MsgStableSwapAdjustScalingFactorsResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) CreateStableswapPool(ctx context.Context, in *MsgCreateStableswapPool, opts ...grpc.CallOption) (*MsgCreateStableswapPoolResponse, error) {
	out := new(MsgCreateStableswapPoolResponse)
	err := c.cc.Invoke(ctx, "/osmosis.gamm.stableswap.v1beta1.Msg/CreateStableswapPool", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) AdjustScalingFactors(ctx context.Context, in *MsgStableSwapAdjustScalingFactors, opts ...grpc.CallOption) (*MsgStableSwapAdjustScalingFactorsResponse, error) {
	out := new(MsgStableSwapAdjustScalingFactorsResponse)
	err := c.cc.Invoke(ctx, "/osmosis.gamm.stableswap.v1beta1.Msg/AdjustScalingFactors", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	CreateStableswapPool(context.Context, *MsgCreateStableswapPool) (*MsgCreateStableswapPoolResponse, error)
	AdjustScalingFactors(context.Context, *MsgStableSwapAdjustScalingFactors) (*MsgStableSwapAdjustScalingFactorsResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) CreateStableswapPool(ctx context.Context, req *MsgCreateStableswapPool) (*MsgCreateStableswapPoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateStableswapPool not implemented")
}
func (*UnimplementedMsgServer) AdjustScalingFactors(ctx context.Context, req *MsgStableSwapAdjustScalingFactors) (*MsgStableSwapAdjustScalingFactorsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AdjustScalingFactors not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_CreateStableswapPool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateStableswapPool)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateStableswapPool(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.gamm.stableswap.v1beta1.Msg/CreateStableswapPool",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateStableswapPool(ctx, req.(*MsgCreateStableswapPool))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_AdjustScalingFactors_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgStableSwapAdjustScalingFactors)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).AdjustScalingFactors(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.gamm.stableswap.v1beta1.Msg/AdjustScalingFactors",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).AdjustScalingFactors(ctx, req.(*MsgStableSwapAdjustScalingFactors))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "osmosis.gamm.stableswap.v1beta1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateStableswapPool",
			Handler:    _Msg_CreateStableswapPool_Handler,
		},
		{
			MethodName: "AdjustScalingFactors",
			Handler:    _Msg_AdjustScalingFactors_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "osmosis/gamm/pool-models/stableswap/tx.proto",
}

func (m *MsgCreateStableswapPool) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateStableswapPool) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateStableswapPool) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.FuturePoolGovernor) > 0 {
		i -= len(m.FuturePoolGovernor)
		copy(dAtA[i:], m.FuturePoolGovernor)
		i = encodeVarintTx(dAtA, i, uint64(len(m.FuturePoolGovernor)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.InitialPoolLiquidity) > 0 {
		for iNdEx := len(m.InitialPoolLiquidity) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.InitialPoolLiquidity[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if m.PoolParams != nil {
		{
			size, err := m.PoolParams.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTx(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgCreateStableswapPoolResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateStableswapPoolResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateStableswapPoolResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.PoolID != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.PoolID))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *MsgStableSwapAdjustScalingFactors) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgStableSwapAdjustScalingFactors) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgStableSwapAdjustScalingFactors) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.PoolID != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.PoolID))
		i--
		dAtA[i] = 0x10
	}
	if len(m.ScalingFactorGovernor) > 0 {
		i -= len(m.ScalingFactorGovernor)
		copy(dAtA[i:], m.ScalingFactorGovernor)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ScalingFactorGovernor)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgStableSwapAdjustScalingFactorsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgStableSwapAdjustScalingFactorsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgStableSwapAdjustScalingFactorsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgCreateStableswapPool) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.PoolParams != nil {
		l = m.PoolParams.Size()
		n += 1 + l + sovTx(uint64(l))
	}
	if len(m.InitialPoolLiquidity) > 0 {
		for _, e := range m.InitialPoolLiquidity {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	l = len(m.FuturePoolGovernor)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgCreateStableswapPoolResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PoolID != 0 {
		n += 1 + sovTx(uint64(m.PoolID))
	}
	return n
}

func (m *MsgStableSwapAdjustScalingFactors) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ScalingFactorGovernor)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.PoolID != 0 {
		n += 1 + sovTx(uint64(m.PoolID))
	}
	return n
}

func (m *MsgStableSwapAdjustScalingFactorsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgCreateStableswapPool) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgCreateStableswapPool: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateStableswapPool: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolParams", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.PoolParams == nil {
				m.PoolParams = &PoolParams{}
			}
			if err := m.PoolParams.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InitialPoolLiquidity", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InitialPoolLiquidity = append(m.InitialPoolLiquidity, types.Coin{})
			if err := m.InitialPoolLiquidity[len(m.InitialPoolLiquidity)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FuturePoolGovernor", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FuturePoolGovernor = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgCreateStableswapPoolResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgCreateStableswapPoolResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateStableswapPoolResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolID", wireType)
			}
			m.PoolID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PoolID |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgStableSwapAdjustScalingFactors) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgStableSwapAdjustScalingFactors: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgStableSwapAdjustScalingFactors: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ScalingFactorGovernor", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ScalingFactorGovernor = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolID", wireType)
			}
			m.PoolID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PoolID |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgStableSwapAdjustScalingFactorsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgStableSwapAdjustScalingFactorsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgStableSwapAdjustScalingFactorsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
