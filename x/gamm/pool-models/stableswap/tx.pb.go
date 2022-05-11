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
	ScalingFactorGovernor string   `protobuf:"bytes,1,opt,name=scaling_factor_governor,json=scalingFactorGovernor,proto3" json:"scaling_factor_governor,omitempty" yaml:"scaling_factor_governor"`
	PoolID                uint64   `protobuf:"varint,2,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	ScalingFactors        []uint64 `protobuf:"varint,3,rep,packed,name=scaling_factors,json=scalingFactors,proto3" json:"scaling_factors,omitempty" yaml:"stableswap_scaling_factor"`
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

func (m *MsgStableSwapAdjustScalingFactors) GetScalingFactors() []uint64 {
	if m != nil {
		return m.ScalingFactors
	}
	return nil
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
	// 601 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0xcd, 0x6e, 0xd3, 0x4c,
	0x14, 0x8d, 0x9b, 0x2a, 0x9f, 0xbe, 0xa9, 0x00, 0x61, 0x85, 0x36, 0x04, 0xc9, 0x0e, 0x66, 0x93,
	0xaa, 0xd4, 0x43, 0xcb, 0x82, 0x9f, 0x55, 0x71, 0x51, 0x51, 0x11, 0x95, 0x5a, 0x77, 0xd7, 0x4d,
	0x34, 0xb6, 0xa7, 0x66, 0xc0, 0xf6, 0x18, 0xdf, 0x49, 0x7f, 0x76, 0xf0, 0x06, 0x88, 0xc7, 0x60,
	0xc5, 0x63, 0x74, 0x83, 0xd4, 0x25, 0x2b, 0x83, 0x92, 0x37, 0xc8, 0x86, 0x2d, 0xb2, 0xc7, 0xce,
	0x8f, 0x94, 0x34, 0x15, 0xac, 0x32, 0x39, 0x39, 0xf7, 0x9c, 0x7b, 0xef, 0x99, 0x0c, 0x7a, 0xc8,
	0x21, 0xe4, 0xc0, 0x00, 0xfb, 0x24, 0x0c, 0x71, 0xcc, 0x79, 0xb0, 0x1e, 0x72, 0x8f, 0x06, 0x80,
	0x41, 0x10, 0x27, 0xa0, 0x70, 0x4a, 0x62, 0x2c, 0xce, 0xcc, 0x38, 0xe1, 0x82, 0xab, 0x7a, 0xc1,
	0x36, 0x33, 0xb6, 0x39, 0x62, 0x98, 0x27, 0x1b, 0x0e, 0x15, 0x64, 0xa3, 0xa9, 0xb9, 0x39, 0x03,
	0x3b, 0x04, 0x28, 0x2e, 0x40, 0xec, 0x72, 0x16, 0x49, 0x81, 0x66, 0xdd, 0xe7, 0x3e, 0xcf, 0x8f,
	0x38, 0x3b, 0x15, 0xe8, 0xb3, 0xeb, 0x34, 0x31, 0x3a, 0x76, 0x32, 0x86, 0x2c, 0x35, 0x3e, 0x56,
	0xd1, 0xca, 0x1e, 0xf8, 0xdb, 0x09, 0x25, 0x82, 0x1e, 0x0e, 0x29, 0xfb, 0x9c, 0x07, 0xea, 0x2a,
	0xaa, 0x01, 0x8d, 0x3c, 0x9a, 0x34, 0x94, 0x96, 0xd2, 0xfe, 0xdf, 0xba, 0x3d, 0x48, 0xf5, 0x1b,
	0xe7, 0x24, 0x0c, 0x9e, 0x1b, 0x12, 0x37, 0xec, 0x82, 0xa0, 0xba, 0x08, 0x65, 0xa2, 0xfb, 0x24,
	0x21, 0x21, 0x34, 0x16, 0x5a, 0x4a, 0x7b, 0x69, 0x73, 0xcd, 0x9c, 0x33, 0xad, 0xb9, 0x3f, 0x2c,
	0xb1, 0x96, 0x07, 0xa9, 0xae, 0x4a, 0xed, 0x4c, 0xa8, 0x13, 0xe7, 0xb0, 0x61, 0x8f, 0xc9, 0xaa,
	0x9f, 0x14, 0xb4, 0xcc, 0x22, 0x26, 0x18, 0x09, 0xf2, 0x11, 0x3a, 0x01, 0xfb, 0xd0, 0x65, 0x1e,
	0x13, 0xe7, 0x8d, 0x6a, 0xab, 0xda, 0x5e, 0xda, 0xbc, 0x6b, 0xca, 0xf5, 0x99, 0xd9, 0xfa, 0x86,
	0x2e, 0xdb, 0x9c, 0x45, 0xd6, 0xa3, 0x8b, 0x54, 0xaf, 0x7c, 0xfd, 0xa9, 0xb7, 0x7d, 0x26, 0xde,
	0x76, 0x1d, 0xd3, 0xe5, 0x21, 0x2e, 0x76, 0x2d, 0x3f, 0xd6, 0xc1, 0x7b, 0x8f, 0xc5, 0x79, 0x4c,
	0x21, 0x2f, 0x00, 0xbb, 0x5e, 0x58, 0x65, 0x4d, 0xbe, 0x29, 0x8d, 0xd4, 0x03, 0x54, 0x3f, 0xee,
	0x8a, 0x6e, 0x42, 0x65, 0x07, 0x3e, 0x3f, 0xa1, 0x49, 0xc4, 0x93, 0xc6, 0x62, 0xbe, 0x21, 0x7d,
	0x90, 0xea, 0xf7, 0xe4, 0x14, 0xd3, 0x58, 0x86, 0xad, 0x4a, 0x38, 0xd3, 0x7c, 0x55, 0x82, 0x3b,
	0x48, 0x9f, 0x91, 0x80, 0x4d, 0x21, 0xe6, 0x11, 0x50, 0xf5, 0x01, 0xfa, 0x2f, 0x17, 0x62, 0x5e,
	0x1e, 0xc5, 0xa2, 0x85, 0x7a, 0xa9, 0x5e, 0xcb, 0x28, 0xbb, 0x2f, 0xed, 0x5a, 0xf6, 0xd3, 0xae,
	0x67, 0xfc, 0x56, 0xd0, 0xfd, 0x3d, 0xf0, 0xa5, 0xc4, 0xe1, 0x29, 0x89, 0x5f, 0x78, 0xef, 0xba,
	0x20, 0x0e, 0x5d, 0x12, 0xb0, 0xc8, 0xdf, 0x21, 0xae, 0xe0, 0x09, 0xa8, 0x47, 0x68, 0x05, 0x24,
	0xd2, 0x39, 0xce, 0xa1, 0xd1, 0x0c, 0x32, 0x65, 0x63, 0x90, 0xea, 0x5a, 0x91, 0xf2, 0x74, 0xa2,
	0x61, 0xdf, 0x81, 0x71, 0xd1, 0x72, 0x92, 0xf1, 0x36, 0x17, 0x66, 0xb5, 0xa9, 0x1e, 0xa0, 0x5b,
	0x93, 0xba, 0x90, 0xa7, 0xb7, 0x68, 0xb5, 0xb3, 0x88, 0x06, 0xa9, 0xde, 0x2a, 0xcc, 0x47, 0xf7,
	0x75, 0x92, 0x6f, 0xd8, 0x37, 0x27, 0xec, 0xc1, 0x58, 0x43, 0xab, 0x73, 0x07, 0x2f, 0x77, 0xb9,
	0xf9, 0x7d, 0x01, 0x55, 0xf7, 0xc0, 0x57, 0xbf, 0x28, 0xa8, 0x3e, 0xf5, 0xda, 0x3f, 0x9d, 0x7b,
	0x6f, 0x67, 0xc4, 0xd5, 0xdc, 0xfa, 0xdb, 0xca, 0x61, 0xd0, 0xdf, 0x14, 0xa4, 0xcd, 0x09, 0xd0,
	0xba, 0x8e, 0xc9, 0xd5, 0x1a, 0xcd, 0xd7, 0xff, 0xae, 0x51, 0xb6, 0x6c, 0x1d, 0x5d, 0xf4, 0x34,
	0xe5, 0xb2, 0xa7, 0x29, 0xbf, 0x7a, 0x9a, 0xf2, 0xb9, 0xaf, 0x55, 0x2e, 0xfb, 0x5a, 0xe5, 0x47,
	0x5f, 0xab, 0x1c, 0x6d, 0x8d, 0xfd, 0xd7, 0x0a, 0xbf, 0xf5, 0x80, 0x38, 0x50, 0x7e, 0xc1, 0x27,
	0x4f, 0xf0, 0xd9, 0x55, 0x6f, 0x96, 0x53, 0xcb, 0x1f, 0xa9, 0xc7, 0x7f, 0x02, 0x00, 0x00, 0xff,
	0xff, 0x8b, 0x50, 0xad, 0x65, 0x66, 0x05, 0x00, 0x00,
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
	StableSwapAdjustScalingFactors(ctx context.Context, in *MsgStableSwapAdjustScalingFactors, opts ...grpc.CallOption) (*MsgStableSwapAdjustScalingFactorsResponse, error)
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

func (c *msgClient) StableSwapAdjustScalingFactors(ctx context.Context, in *MsgStableSwapAdjustScalingFactors, opts ...grpc.CallOption) (*MsgStableSwapAdjustScalingFactorsResponse, error) {
	out := new(MsgStableSwapAdjustScalingFactorsResponse)
	err := c.cc.Invoke(ctx, "/osmosis.gamm.stableswap.v1beta1.Msg/StableSwapAdjustScalingFactors", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	CreateStableswapPool(context.Context, *MsgCreateStableswapPool) (*MsgCreateStableswapPoolResponse, error)
	StableSwapAdjustScalingFactors(context.Context, *MsgStableSwapAdjustScalingFactors) (*MsgStableSwapAdjustScalingFactorsResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) CreateStableswapPool(ctx context.Context, req *MsgCreateStableswapPool) (*MsgCreateStableswapPoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateStableswapPool not implemented")
}
func (*UnimplementedMsgServer) StableSwapAdjustScalingFactors(ctx context.Context, req *MsgStableSwapAdjustScalingFactors) (*MsgStableSwapAdjustScalingFactorsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StableSwapAdjustScalingFactors not implemented")
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

func _Msg_StableSwapAdjustScalingFactors_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgStableSwapAdjustScalingFactors)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).StableSwapAdjustScalingFactors(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.gamm.stableswap.v1beta1.Msg/StableSwapAdjustScalingFactors",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).StableSwapAdjustScalingFactors(ctx, req.(*MsgStableSwapAdjustScalingFactors))
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
			MethodName: "StableSwapAdjustScalingFactors",
			Handler:    _Msg_StableSwapAdjustScalingFactors_Handler,
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
	if len(m.ScalingFactors) > 0 {
		dAtA3 := make([]byte, len(m.ScalingFactors)*10)
		var j2 int
		for _, num := range m.ScalingFactors {
			for num >= 1<<7 {
				dAtA3[j2] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j2++
			}
			dAtA3[j2] = uint8(num)
			j2++
		}
		i -= j2
		copy(dAtA[i:], dAtA3[:j2])
		i = encodeVarintTx(dAtA, i, uint64(j2))
		i--
		dAtA[i] = 0x1a
	}
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
	if len(m.ScalingFactors) > 0 {
		l = 0
		for _, e := range m.ScalingFactors {
			l += sovTx(uint64(e))
		}
		n += 1 + sovTx(uint64(l)) + l
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
		case 3:
			if wireType == 0 {
				var v uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTx
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.ScalingFactors = append(m.ScalingFactors, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTx
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthTx
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthTx
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA[iNdEx:postIndex] {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.ScalingFactors) == 0 {
					m.ScalingFactors = make([]uint64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowTx
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.ScalingFactors = append(m.ScalingFactors, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field ScalingFactors", wireType)
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
