// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/concentrated-liquidity/pool-model/concentrated/tx.proto

package model

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
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

// ===================== MsgCreateConcentratedPool
type MsgCreateConcentratedPool struct {
	Sender       string                                 `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty" yaml:"sender"`
	Denom0       string                                 `protobuf:"bytes,2,opt,name=denom0,proto3" json:"denom0,omitempty" yaml:"denom0"`
	Denom1       string                                 `protobuf:"bytes,3,opt,name=denom1,proto3" json:"denom1,omitempty" yaml:"denom1"`
	TickSpacing  uint64                                 `protobuf:"varint,4,opt,name=tick_spacing,json=tickSpacing,proto3" json:"tick_spacing,omitempty" yaml:"tick_spacing"`
	SpreadFactor github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,5,opt,name=spread_factor,json=spreadFactor,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"spread_factor" yaml:"spread_factor"`
}

func (m *MsgCreateConcentratedPool) Reset()         { *m = MsgCreateConcentratedPool{} }
func (m *MsgCreateConcentratedPool) String() string { return proto.CompactTextString(m) }
func (*MsgCreateConcentratedPool) ProtoMessage()    {}
func (*MsgCreateConcentratedPool) Descriptor() ([]byte, []int) {
	return fileDescriptor_dde1ce763867060f, []int{0}
}
func (m *MsgCreateConcentratedPool) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateConcentratedPool) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateConcentratedPool.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateConcentratedPool) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateConcentratedPool.Merge(m, src)
}
func (m *MsgCreateConcentratedPool) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateConcentratedPool) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateConcentratedPool.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateConcentratedPool proto.InternalMessageInfo

func (m *MsgCreateConcentratedPool) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

func (m *MsgCreateConcentratedPool) GetDenom0() string {
	if m != nil {
		return m.Denom0
	}
	return ""
}

func (m *MsgCreateConcentratedPool) GetDenom1() string {
	if m != nil {
		return m.Denom1
	}
	return ""
}

func (m *MsgCreateConcentratedPool) GetTickSpacing() uint64 {
	if m != nil {
		return m.TickSpacing
	}
	return 0
}

// Returns a unique poolID to identify the pool with.
type MsgCreateConcentratedPoolResponse struct {
	PoolID uint64 `protobuf:"varint,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
}

func (m *MsgCreateConcentratedPoolResponse) Reset()         { *m = MsgCreateConcentratedPoolResponse{} }
func (m *MsgCreateConcentratedPoolResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCreateConcentratedPoolResponse) ProtoMessage()    {}
func (*MsgCreateConcentratedPoolResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_dde1ce763867060f, []int{1}
}
func (m *MsgCreateConcentratedPoolResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateConcentratedPoolResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateConcentratedPoolResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateConcentratedPoolResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateConcentratedPoolResponse.Merge(m, src)
}
func (m *MsgCreateConcentratedPoolResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateConcentratedPoolResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateConcentratedPoolResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateConcentratedPoolResponse proto.InternalMessageInfo

func (m *MsgCreateConcentratedPoolResponse) GetPoolID() uint64 {
	if m != nil {
		return m.PoolID
	}
	return 0
}

func init() {
	proto.RegisterType((*MsgCreateConcentratedPool)(nil), "osmosis.concentratedliquidity.poolmodel.concentrated.v1beta1.MsgCreateConcentratedPool")
	proto.RegisterType((*MsgCreateConcentratedPoolResponse)(nil), "osmosis.concentratedliquidity.poolmodel.concentrated.v1beta1.MsgCreateConcentratedPoolResponse")
}

func init() {
	proto.RegisterFile("osmosis/concentrated-liquidity/pool-model/concentrated/tx.proto", fileDescriptor_dde1ce763867060f)
}

var fileDescriptor_dde1ce763867060f = []byte{
	// 455 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x93, 0x41, 0x6b, 0xd4, 0x40,
	0x14, 0xc7, 0x37, 0xed, 0xba, 0xe2, 0xd8, 0x1e, 0x8c, 0x45, 0xe3, 0x1e, 0x92, 0x1a, 0x41, 0xea,
	0x61, 0x67, 0x8c, 0x9e, 0x2c, 0x82, 0x90, 0x96, 0x62, 0x0f, 0x05, 0x89, 0x07, 0x41, 0x84, 0x65,
	0x32, 0x33, 0xc6, 0x61, 0x93, 0xbc, 0x98, 0x99, 0x96, 0xee, 0xb7, 0xf0, 0x4b, 0x09, 0x3d, 0x16,
	0x4f, 0xd2, 0x43, 0x90, 0xdd, 0x6f, 0xb0, 0x9f, 0x40, 0x92, 0xc9, 0x96, 0x54, 0xcc, 0x49, 0x3c,
	0xed, 0xfe, 0xe7, 0xff, 0x7b, 0xff, 0xbc, 0xbc, 0x37, 0x41, 0x6f, 0x40, 0x65, 0xa0, 0xa4, 0x22,
	0x0c, 0x72, 0x26, 0x72, 0x5d, 0x52, 0x2d, 0xf8, 0x24, 0x95, 0x5f, 0x4f, 0x25, 0x97, 0x7a, 0x4e,
	0x0a, 0x80, 0x74, 0x92, 0x01, 0x17, 0xe9, 0x0d, 0x82, 0xe8, 0x73, 0x5c, 0x94, 0xa0, 0xc1, 0x7e,
	0xdd, 0x06, 0xe0, 0xae, 0x7d, 0x5d, 0x8f, 0xeb, 0xfa, 0xa6, 0xfc, 0x86, 0x8f, 0xcf, 0x82, 0x58,
	0x68, 0x1a, 0x8c, 0x77, 0x12, 0x48, 0xa0, 0x09, 0x22, 0xf5, 0x3f, 0x93, 0x39, 0x76, 0x59, 0x13,
	0x4a, 0x62, 0xaa, 0x04, 0x69, 0x51, 0xc2, 0x40, 0xe6, 0xc6, 0xf7, 0xbf, 0x6f, 0xa0, 0x47, 0x27,
	0x2a, 0x39, 0x28, 0x05, 0xd5, 0xe2, 0xa0, 0x93, 0xfb, 0x0e, 0x20, 0xb5, 0x9f, 0xa1, 0x91, 0x12,
	0x39, 0x17, 0xa5, 0x63, 0xed, 0x5a, 0x7b, 0x77, 0xc2, 0x7b, 0xab, 0xca, 0xdb, 0x9e, 0xd3, 0x2c,
	0xdd, 0xf7, 0xcd, 0xb9, 0x1f, 0xb5, 0x40, 0x8d, 0x72, 0x91, 0x43, 0xf6, 0xdc, 0xd9, 0xf8, 0x13,
	0x35, 0xe7, 0x7e, 0xd4, 0x02, 0xd7, 0x68, 0xe0, 0x6c, 0xfe, 0x15, 0x0d, 0xd6, 0x68, 0x60, 0xef,
	0xa3, 0x2d, 0x2d, 0xd9, 0x6c, 0xaa, 0x0a, 0xca, 0x64, 0x9e, 0x38, 0xc3, 0x5d, 0x6b, 0x6f, 0x18,
	0x3e, 0x5c, 0x55, 0xde, 0x7d, 0x53, 0xd0, 0x75, 0xfd, 0xe8, 0x6e, 0x2d, 0xdf, 0x1b, 0x65, 0xcf,
	0xd0, 0xb6, 0x2a, 0x4a, 0x41, 0xf9, 0xf4, 0x33, 0x65, 0x1a, 0x4a, 0xe7, 0x56, 0xf3, 0xb4, 0xa3,
	0x8b, 0xca, 0x1b, 0x5c, 0x55, 0xde, 0xd3, 0x44, 0xea, 0x2f, 0xa7, 0x31, 0x66, 0x90, 0x91, 0x76,
	0x48, 0xe6, 0x67, 0xa2, 0xf8, 0x8c, 0xe8, 0x79, 0x21, 0x14, 0x3e, 0x14, 0x6c, 0x55, 0x79, 0x3b,
	0xed, 0x1b, 0x77, 0xc3, 0xfc, 0x68, 0xcb, 0xe8, 0x23, 0x23, 0xdf, 0xa2, 0xc7, 0xbd, 0x63, 0x8c,
	0x84, 0x2a, 0x20, 0x57, 0xc2, 0x7e, 0x82, 0x6e, 0xd7, 0x4b, 0x9c, 0x4a, 0xde, 0xcc, 0x73, 0x18,
	0xa2, 0x45, 0xe5, 0x8d, 0x6a, 0xe4, 0xf8, 0x30, 0x1a, 0xd5, 0xd6, 0x31, 0x7f, 0x71, 0x65, 0xa1,
	0xcd, 0x13, 0x95, 0xd8, 0x3f, 0x2c, 0xf4, 0xa0, 0x67, 0x2d, 0x1f, 0xf0, 0xbf, 0xdc, 0x14, 0xdc,
	0xdb, 0xe8, 0x78, 0xfa, 0x9f, 0x82, 0xd7, 0x13, 0x08, 0x3f, 0x5d, 0x2c, 0x5c, 0xeb, 0x72, 0xe1,
	0x5a, 0xbf, 0x16, 0xae, 0xf5, 0x6d, 0xe9, 0x0e, 0x2e, 0x97, 0xee, 0xe0, 0xe7, 0xd2, 0x1d, 0x7c,
	0x0c, 0x3b, 0xeb, 0x68, 0x9b, 0x98, 0xa4, 0x34, 0x56, 0x6b, 0x41, 0xce, 0x82, 0x57, 0xe4, 0xbc,
	0xef, 0xdb, 0x6a, 0x9a, 0x8a, 0x47, 0xcd, 0x9d, 0x7e, 0xf9, 0x3b, 0x00, 0x00, 0xff, 0xff, 0x39,
	0x15, 0x7d, 0x7f, 0x8a, 0x03, 0x00, 0x00,
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
	CreateConcentratedPool(ctx context.Context, in *MsgCreateConcentratedPool, opts ...grpc.CallOption) (*MsgCreateConcentratedPoolResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) CreateConcentratedPool(ctx context.Context, in *MsgCreateConcentratedPool, opts ...grpc.CallOption) (*MsgCreateConcentratedPoolResponse, error) {
	out := new(MsgCreateConcentratedPoolResponse)
	err := c.cc.Invoke(ctx, "/osmosis.concentratedliquidity.poolmodel.concentrated.v1beta1.Msg/CreateConcentratedPool", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	CreateConcentratedPool(context.Context, *MsgCreateConcentratedPool) (*MsgCreateConcentratedPoolResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) CreateConcentratedPool(ctx context.Context, req *MsgCreateConcentratedPool) (*MsgCreateConcentratedPoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateConcentratedPool not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_CreateConcentratedPool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateConcentratedPool)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateConcentratedPool(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.concentratedliquidity.poolmodel.concentrated.v1beta1.Msg/CreateConcentratedPool",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateConcentratedPool(ctx, req.(*MsgCreateConcentratedPool))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "osmosis.concentratedliquidity.poolmodel.concentrated.v1beta1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateConcentratedPool",
			Handler:    _Msg_CreateConcentratedPool_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "osmosis/concentrated-liquidity/pool-model/concentrated/tx.proto",
}

func (m *MsgCreateConcentratedPool) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateConcentratedPool) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateConcentratedPool) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.SpreadFactor.Size()
		i -= size
		if _, err := m.SpreadFactor.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	if m.TickSpacing != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.TickSpacing))
		i--
		dAtA[i] = 0x20
	}
	if len(m.Denom1) > 0 {
		i -= len(m.Denom1)
		copy(dAtA[i:], m.Denom1)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Denom1)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Denom0) > 0 {
		i -= len(m.Denom0)
		copy(dAtA[i:], m.Denom0)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Denom0)))
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

func (m *MsgCreateConcentratedPoolResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateConcentratedPoolResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateConcentratedPoolResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
func (m *MsgCreateConcentratedPool) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Denom0)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Denom1)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.TickSpacing != 0 {
		n += 1 + sovTx(uint64(m.TickSpacing))
	}
	l = m.SpreadFactor.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgCreateConcentratedPoolResponse) Size() (n int) {
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

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgCreateConcentratedPool) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgCreateConcentratedPool: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateConcentratedPool: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field Denom0", wireType)
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
			m.Denom0 = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom1", wireType)
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
			m.Denom1 = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TickSpacing", wireType)
			}
			m.TickSpacing = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TickSpacing |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpreadFactor", wireType)
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
			if err := m.SpreadFactor.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
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
func (m *MsgCreateConcentratedPoolResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgCreateConcentratedPoolResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateConcentratedPoolResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
