// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/concentratedliquidity/v1beta1/position.proto

// this is a legacy package that requires additional migration logic
// in order to use the correct package. Decision made to use legacy package path
// until clear steps for migration logic and the unknowns for state breaking are
// investigated for changing proto package.

package model

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	types1 "github.com/osmosis-labs/osmosis/v28/x/lockup/types"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Position contains position's id, address, pool id, lower tick, upper tick
// join time, and liquidity.
type Position struct {
	PositionId uint64                      `protobuf:"varint,1,opt,name=position_id,json=positionId,proto3" json:"position_id,omitempty" yaml:"position_id"`
	Address    string                      `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty" yaml:"address"`
	PoolId     uint64                      `protobuf:"varint,3,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty" yaml:"pool_id"`
	LowerTick  int64                       `protobuf:"varint,4,opt,name=lower_tick,json=lowerTick,proto3" json:"lower_tick,omitempty"`
	UpperTick  int64                       `protobuf:"varint,5,opt,name=upper_tick,json=upperTick,proto3" json:"upper_tick,omitempty"`
	JoinTime   time.Time                   `protobuf:"bytes,6,opt,name=join_time,json=joinTime,proto3,stdtime" json:"join_time" yaml:"join_time"`
	Liquidity  cosmossdk_io_math.LegacyDec `protobuf:"bytes,7,opt,name=liquidity,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"liquidity" yaml:"liquidity"`
}

func (m *Position) Reset()         { *m = Position{} }
func (m *Position) String() string { return proto.CompactTextString(m) }
func (*Position) ProtoMessage()    {}
func (*Position) Descriptor() ([]byte, []int) {
	return fileDescriptor_1363e25aa5179fb1, []int{0}
}
func (m *Position) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Position) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Position.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Position) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Position.Merge(m, src)
}
func (m *Position) XXX_Size() int {
	return m.Size()
}
func (m *Position) XXX_DiscardUnknown() {
	xxx_messageInfo_Position.DiscardUnknown(m)
}

var xxx_messageInfo_Position proto.InternalMessageInfo

func (m *Position) GetPositionId() uint64 {
	if m != nil {
		return m.PositionId
	}
	return 0
}

func (m *Position) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Position) GetPoolId() uint64 {
	if m != nil {
		return m.PoolId
	}
	return 0
}

func (m *Position) GetLowerTick() int64 {
	if m != nil {
		return m.LowerTick
	}
	return 0
}

func (m *Position) GetUpperTick() int64 {
	if m != nil {
		return m.UpperTick
	}
	return 0
}

func (m *Position) GetJoinTime() time.Time {
	if m != nil {
		return m.JoinTime
	}
	return time.Time{}
}

// FullPositionBreakdown returns:
// - the position itself
// - the amount the position translates in terms of asset0 and asset1
// - the amount of claimable fees
// - the amount of claimable incentives
// - the amount of incentives that would be forfeited if the position was closed
// now
type FullPositionBreakdown struct {
	Position               Position     `protobuf:"bytes,1,opt,name=position,proto3" json:"position"`
	Asset0                 types.Coin   `protobuf:"bytes,2,opt,name=asset0,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coin" json:"asset0"`
	Asset1                 types.Coin   `protobuf:"bytes,3,opt,name=asset1,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coin" json:"asset1"`
	ClaimableSpreadRewards []types.Coin `protobuf:"bytes,4,rep,name=claimable_spread_rewards,json=claimableSpreadRewards,proto3" json:"claimable_spread_rewards" yaml:"claimable_spread_rewards"`
	ClaimableIncentives    []types.Coin `protobuf:"bytes,5,rep,name=claimable_incentives,json=claimableIncentives,proto3" json:"claimable_incentives" yaml:"claimable_incentives"`
	ForfeitedIncentives    []types.Coin `protobuf:"bytes,6,rep,name=forfeited_incentives,json=forfeitedIncentives,proto3" json:"forfeited_incentives" yaml:"forfeited_incentives"`
}

func (m *FullPositionBreakdown) Reset()         { *m = FullPositionBreakdown{} }
func (m *FullPositionBreakdown) String() string { return proto.CompactTextString(m) }
func (*FullPositionBreakdown) ProtoMessage()    {}
func (*FullPositionBreakdown) Descriptor() ([]byte, []int) {
	return fileDescriptor_1363e25aa5179fb1, []int{1}
}
func (m *FullPositionBreakdown) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FullPositionBreakdown) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FullPositionBreakdown.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FullPositionBreakdown) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FullPositionBreakdown.Merge(m, src)
}
func (m *FullPositionBreakdown) XXX_Size() int {
	return m.Size()
}
func (m *FullPositionBreakdown) XXX_DiscardUnknown() {
	xxx_messageInfo_FullPositionBreakdown.DiscardUnknown(m)
}

var xxx_messageInfo_FullPositionBreakdown proto.InternalMessageInfo

func (m *FullPositionBreakdown) GetPosition() Position {
	if m != nil {
		return m.Position
	}
	return Position{}
}

func (m *FullPositionBreakdown) GetAsset0() types.Coin {
	if m != nil {
		return m.Asset0
	}
	return types.Coin{}
}

func (m *FullPositionBreakdown) GetAsset1() types.Coin {
	if m != nil {
		return m.Asset1
	}
	return types.Coin{}
}

func (m *FullPositionBreakdown) GetClaimableSpreadRewards() []types.Coin {
	if m != nil {
		return m.ClaimableSpreadRewards
	}
	return nil
}

func (m *FullPositionBreakdown) GetClaimableIncentives() []types.Coin {
	if m != nil {
		return m.ClaimableIncentives
	}
	return nil
}

func (m *FullPositionBreakdown) GetForfeitedIncentives() []types.Coin {
	if m != nil {
		return m.ForfeitedIncentives
	}
	return nil
}

type PositionWithPeriodLock struct {
	Position Position          `protobuf:"bytes,1,opt,name=position,proto3" json:"position"`
	Locks    types1.PeriodLock `protobuf:"bytes,2,opt,name=locks,proto3" json:"locks"`
}

func (m *PositionWithPeriodLock) Reset()         { *m = PositionWithPeriodLock{} }
func (m *PositionWithPeriodLock) String() string { return proto.CompactTextString(m) }
func (*PositionWithPeriodLock) ProtoMessage()    {}
func (*PositionWithPeriodLock) Descriptor() ([]byte, []int) {
	return fileDescriptor_1363e25aa5179fb1, []int{2}
}
func (m *PositionWithPeriodLock) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PositionWithPeriodLock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PositionWithPeriodLock.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PositionWithPeriodLock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PositionWithPeriodLock.Merge(m, src)
}
func (m *PositionWithPeriodLock) XXX_Size() int {
	return m.Size()
}
func (m *PositionWithPeriodLock) XXX_DiscardUnknown() {
	xxx_messageInfo_PositionWithPeriodLock.DiscardUnknown(m)
}

var xxx_messageInfo_PositionWithPeriodLock proto.InternalMessageInfo

func (m *PositionWithPeriodLock) GetPosition() Position {
	if m != nil {
		return m.Position
	}
	return Position{}
}

func (m *PositionWithPeriodLock) GetLocks() types1.PeriodLock {
	if m != nil {
		return m.Locks
	}
	return types1.PeriodLock{}
}

func init() {
	proto.RegisterType((*Position)(nil), "osmosis.concentratedliquidity.v1beta1.Position")
	proto.RegisterType((*FullPositionBreakdown)(nil), "osmosis.concentratedliquidity.v1beta1.FullPositionBreakdown")
	proto.RegisterType((*PositionWithPeriodLock)(nil), "osmosis.concentratedliquidity.v1beta1.PositionWithPeriodLock")
}

func init() {
	proto.RegisterFile("osmosis/concentratedliquidity/v1beta1/position.proto", fileDescriptor_1363e25aa5179fb1)
}

var fileDescriptor_1363e25aa5179fb1 = []byte{
	// 695 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x54, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x8e, 0xc9, 0x4f, 0xdb, 0x8d, 0x84, 0x90, 0x5b, 0x2a, 0x37, 0x85, 0x38, 0x32, 0x42, 0x8d,
	0x04, 0xf5, 0x92, 0x80, 0xa8, 0xc4, 0x31, 0x20, 0xa4, 0x4a, 0x3d, 0x14, 0x53, 0x84, 0x84, 0x90,
	0xa2, 0xb5, 0x77, 0x9b, 0x2e, 0xb1, 0xb3, 0xae, 0x77, 0xd3, 0x12, 0x89, 0x27, 0xe0, 0xd4, 0x13,
	0x2f, 0xc0, 0x8d, 0x27, 0xe9, 0xb1, 0x47, 0xc4, 0x21, 0x45, 0xed, 0x1b, 0xf4, 0x09, 0xd0, 0xae,
	0xd7, 0x76, 0x40, 0xe5, 0xef, 0xd0, 0x93, 0x77, 0xe7, 0xdb, 0xef, 0xfb, 0xc6, 0x33, 0xa3, 0x01,
	0x8f, 0x18, 0x8f, 0x18, 0xa7, 0x1c, 0x06, 0x6c, 0x14, 0x90, 0x91, 0x48, 0x90, 0x20, 0x38, 0xa4,
	0xfb, 0x63, 0x8a, 0xa9, 0x98, 0xc0, 0x83, 0x8e, 0x4f, 0x04, 0xea, 0xc0, 0x98, 0x71, 0x2a, 0x28,
	0x1b, 0xb9, 0x71, 0xc2, 0x04, 0x33, 0xef, 0x6a, 0x96, 0x7b, 0x29, 0xcb, 0xd5, 0xac, 0x86, 0x3d,
	0x60, 0x6c, 0x10, 0x12, 0xa8, 0x48, 0xfe, 0x78, 0x17, 0x0a, 0x1a, 0x11, 0x2e, 0x50, 0x14, 0xa7,
	0x3a, 0x8d, 0xa5, 0x01, 0x1b, 0x30, 0x75, 0x84, 0xf2, 0xa4, 0xa3, 0xcd, 0x40, 0xc9, 0x43, 0x1f,
	0x71, 0x92, 0x67, 0x10, 0x30, 0xaa, 0xdd, 0x1b, 0x2b, 0x59, 0xce, 0x21, 0x0b, 0x86, 0xe3, 0x58,
	0x7d, 0x52, 0xc8, 0xf9, 0x58, 0x06, 0xf3, 0xdb, 0x3a, 0x57, 0x73, 0x03, 0xd4, 0xb3, 0xbc, 0xfb,
	0x14, 0x5b, 0x46, 0xcb, 0x68, 0x57, 0x7a, 0xcb, 0x17, 0x53, 0xdb, 0x9c, 0xa0, 0x28, 0x7c, 0xe2,
	0xcc, 0x80, 0x8e, 0x07, 0xb2, 0xdb, 0x26, 0x36, 0xef, 0x83, 0x39, 0x84, 0x71, 0x42, 0x38, 0xb7,
	0xae, 0xb5, 0x8c, 0xf6, 0x42, 0xcf, 0xbc, 0x98, 0xda, 0xd7, 0x53, 0x92, 0x06, 0x1c, 0x2f, 0x7b,
	0x62, 0xde, 0x03, 0x73, 0x31, 0x63, 0xa1, 0xb4, 0x28, 0x2b, 0x8b, 0x99, 0xd7, 0x1a, 0x70, 0xbc,
	0x9a, 0x3c, 0x6d, 0x62, 0xf3, 0x36, 0x00, 0x21, 0x3b, 0x24, 0x49, 0x5f, 0xd0, 0x60, 0x68, 0x55,
	0x5a, 0x46, 0xbb, 0xec, 0x2d, 0xa8, 0xc8, 0x0e, 0x0d, 0x86, 0x12, 0x1e, 0xc7, 0x71, 0x06, 0x57,
	0x53, 0x58, 0x45, 0x14, 0xfc, 0x0a, 0x2c, 0xbc, 0x63, 0x74, 0xd4, 0x97, 0x75, 0xb4, 0x6a, 0x2d,
	0xa3, 0x5d, 0xef, 0x36, 0xdc, 0xb4, 0xc8, 0x6e, 0x56, 0x64, 0x77, 0x27, 0x2b, 0x72, 0xef, 0xd6,
	0xf1, 0xd4, 0x2e, 0x5d, 0x4c, 0xed, 0x1b, 0x69, 0x32, 0x39, 0xd5, 0x39, 0x3a, 0xb5, 0x0d, 0x6f,
	0x5e, 0xde, 0xe5, 0x63, 0x29, 0x9b, 0x37, 0xcf, 0x9a, 0x53, 0x7f, 0xbc, 0x21, 0xa9, 0xdf, 0xa6,
	0xf6, 0x6a, 0xda, 0x0b, 0x8e, 0x87, 0x2e, 0x65, 0x30, 0x42, 0x62, 0xcf, 0xdd, 0x22, 0x03, 0x14,
	0x4c, 0x9e, 0x91, 0xa0, 0x50, 0xce, 0xd9, 0x8e, 0x57, 0x28, 0x39, 0x9f, 0xaa, 0xe0, 0xe6, 0xf3,
	0x71, 0x18, 0x66, 0x0d, 0xe9, 0x25, 0x04, 0x0d, 0x31, 0x3b, 0x1c, 0x99, 0x2f, 0xc0, 0x7c, 0x56,
	0x6e, 0xd5, 0x96, 0x7a, 0x17, 0xba, 0xff, 0x34, 0x52, 0x6e, 0xae, 0x55, 0x91, 0x09, 0x7a, 0xb9,
	0x8c, 0xe9, 0x83, 0x1a, 0xe2, 0x9c, 0x88, 0x07, 0xaa, 0x65, 0xf5, 0xee, 0x8a, 0x9b, 0x66, 0xee,
	0xca, 0x29, 0xca, 0xe9, 0x4f, 0x19, 0x1d, 0xf5, 0xa0, 0xa4, 0x7e, 0x39, 0xb5, 0xd7, 0x06, 0x54,
	0xec, 0x8d, 0x7d, 0x37, 0x60, 0x11, 0xd4, 0x23, 0x97, 0x7e, 0xd6, 0x39, 0x1e, 0x42, 0x31, 0x89,
	0x09, 0x57, 0x04, 0x4f, 0x2b, 0xe7, 0x1e, 0x1d, 0xd5, 0xe8, 0xab, 0xf0, 0xe8, 0x98, 0x1f, 0x80,
	0x15, 0x84, 0x88, 0x46, 0xc8, 0x0f, 0x49, 0x9f, 0xc7, 0x09, 0x41, 0xb8, 0x9f, 0x90, 0x43, 0x94,
	0x60, 0x6e, 0x55, 0x5a, 0xe5, 0x3f, 0xbb, 0xae, 0xe9, 0x86, 0xdb, 0x69, 0x5b, 0x7e, 0x27, 0xe4,
	0x78, 0xcb, 0x39, 0xf4, 0x52, 0x21, 0x5e, 0x0a, 0x98, 0xfb, 0x60, 0xa9, 0x20, 0x51, 0xd5, 0x08,
	0x7a, 0x40, 0xb8, 0x55, 0xfd, 0x9b, 0xf3, 0x1d, 0xed, 0xbc, 0xfa, 0xab, 0x73, 0x21, 0xe2, 0x78,
	0x8b, 0x79, 0x78, 0x33, 0x8f, 0x4a, 0xcb, 0x5d, 0x96, 0xec, 0x12, 0x2a, 0x08, 0x9e, 0xb5, 0xac,
	0xfd, 0xa7, 0xe5, 0x65, 0x22, 0x8e, 0xb7, 0x98, 0x87, 0x0b, 0x4b, 0xe7, 0xb3, 0x01, 0x96, 0xb3,
	0x41, 0x7a, 0x4d, 0xc5, 0xde, 0x36, 0x49, 0x28, 0xc3, 0x5b, 0x2c, 0x18, 0x5e, 0xc5, 0x64, 0x3e,
	0x06, 0x55, 0xb9, 0xa1, 0xb8, 0x1e, 0xcc, 0x46, 0xae, 0x97, 0xae, 0x2f, 0xb7, 0x70, 0xd7, 0xd4,
	0xf4, 0x79, 0xef, 0xed, 0xf1, 0x59, 0xd3, 0x38, 0x39, 0x6b, 0x1a, 0xdf, 0xcf, 0x9a, 0xc6, 0xd1,
	0x79, 0xb3, 0x74, 0x72, 0xde, 0x2c, 0x7d, 0x3d, 0x6f, 0x96, 0xde, 0xf4, 0x66, 0x86, 0x4a, 0x8b,
	0xad, 0x87, 0xc8, 0xe7, 0xd9, 0x05, 0x1e, 0x74, 0x37, 0xe0, 0xfb, 0x9f, 0x56, 0xfa, 0x7a, 0xb1,
	0xd3, 0x23, 0x86, 0x49, 0xe8, 0xd7, 0xd4, 0xbe, 0x78, 0xf8, 0x23, 0x00, 0x00, 0xff, 0xff, 0x02,
	0x77, 0x39, 0xa3, 0x01, 0x06, 0x00, 0x00,
}

func (m *Position) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Position) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Position) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.Liquidity.Size()
		i -= size
		if _, err := m.Liquidity.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintPosition(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x3a
	n1, err1 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.JoinTime, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.JoinTime):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintPosition(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x32
	if m.UpperTick != 0 {
		i = encodeVarintPosition(dAtA, i, uint64(m.UpperTick))
		i--
		dAtA[i] = 0x28
	}
	if m.LowerTick != 0 {
		i = encodeVarintPosition(dAtA, i, uint64(m.LowerTick))
		i--
		dAtA[i] = 0x20
	}
	if m.PoolId != 0 {
		i = encodeVarintPosition(dAtA, i, uint64(m.PoolId))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintPosition(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0x12
	}
	if m.PositionId != 0 {
		i = encodeVarintPosition(dAtA, i, uint64(m.PositionId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *FullPositionBreakdown) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FullPositionBreakdown) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FullPositionBreakdown) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ForfeitedIncentives) > 0 {
		for iNdEx := len(m.ForfeitedIncentives) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ForfeitedIncentives[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintPosition(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x32
		}
	}
	if len(m.ClaimableIncentives) > 0 {
		for iNdEx := len(m.ClaimableIncentives) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ClaimableIncentives[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintPosition(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.ClaimableSpreadRewards) > 0 {
		for iNdEx := len(m.ClaimableSpreadRewards) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ClaimableSpreadRewards[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintPosition(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	{
		size, err := m.Asset1.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintPosition(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	{
		size, err := m.Asset0.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintPosition(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size, err := m.Position.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintPosition(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *PositionWithPeriodLock) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PositionWithPeriodLock) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PositionWithPeriodLock) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Locks.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintPosition(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size, err := m.Position.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintPosition(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintPosition(dAtA []byte, offset int, v uint64) int {
	offset -= sovPosition(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Position) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PositionId != 0 {
		n += 1 + sovPosition(uint64(m.PositionId))
	}
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovPosition(uint64(l))
	}
	if m.PoolId != 0 {
		n += 1 + sovPosition(uint64(m.PoolId))
	}
	if m.LowerTick != 0 {
		n += 1 + sovPosition(uint64(m.LowerTick))
	}
	if m.UpperTick != 0 {
		n += 1 + sovPosition(uint64(m.UpperTick))
	}
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.JoinTime)
	n += 1 + l + sovPosition(uint64(l))
	l = m.Liquidity.Size()
	n += 1 + l + sovPosition(uint64(l))
	return n
}

func (m *FullPositionBreakdown) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Position.Size()
	n += 1 + l + sovPosition(uint64(l))
	l = m.Asset0.Size()
	n += 1 + l + sovPosition(uint64(l))
	l = m.Asset1.Size()
	n += 1 + l + sovPosition(uint64(l))
	if len(m.ClaimableSpreadRewards) > 0 {
		for _, e := range m.ClaimableSpreadRewards {
			l = e.Size()
			n += 1 + l + sovPosition(uint64(l))
		}
	}
	if len(m.ClaimableIncentives) > 0 {
		for _, e := range m.ClaimableIncentives {
			l = e.Size()
			n += 1 + l + sovPosition(uint64(l))
		}
	}
	if len(m.ForfeitedIncentives) > 0 {
		for _, e := range m.ForfeitedIncentives {
			l = e.Size()
			n += 1 + l + sovPosition(uint64(l))
		}
	}
	return n
}

func (m *PositionWithPeriodLock) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Position.Size()
	n += 1 + l + sovPosition(uint64(l))
	l = m.Locks.Size()
	n += 1 + l + sovPosition(uint64(l))
	return n
}

func sovPosition(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPosition(x uint64) (n int) {
	return sovPosition(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Position) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPosition
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
			return fmt.Errorf("proto: Position: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Position: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PositionId", wireType)
			}
			m.PositionId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PositionId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
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
				return ErrInvalidLengthPosition
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPosition
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolId", wireType)
			}
			m.PoolId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PoolId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LowerTick", wireType)
			}
			m.LowerTick = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LowerTick |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpperTick", wireType)
			}
			m.UpperTick = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpperTick |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field JoinTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
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
				return ErrInvalidLengthPosition
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPosition
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.JoinTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Liquidity", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
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
				return ErrInvalidLengthPosition
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPosition
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Liquidity.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPosition(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPosition
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
func (m *FullPositionBreakdown) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPosition
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
			return fmt.Errorf("proto: FullPositionBreakdown: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FullPositionBreakdown: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Position", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
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
				return ErrInvalidLengthPosition
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPosition
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Position.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Asset0", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
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
				return ErrInvalidLengthPosition
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPosition
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Asset0.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Asset1", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
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
				return ErrInvalidLengthPosition
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPosition
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Asset1.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClaimableSpreadRewards", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
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
				return ErrInvalidLengthPosition
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPosition
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ClaimableSpreadRewards = append(m.ClaimableSpreadRewards, types.Coin{})
			if err := m.ClaimableSpreadRewards[len(m.ClaimableSpreadRewards)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClaimableIncentives", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
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
				return ErrInvalidLengthPosition
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPosition
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ClaimableIncentives = append(m.ClaimableIncentives, types.Coin{})
			if err := m.ClaimableIncentives[len(m.ClaimableIncentives)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ForfeitedIncentives", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
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
				return ErrInvalidLengthPosition
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPosition
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ForfeitedIncentives = append(m.ForfeitedIncentives, types.Coin{})
			if err := m.ForfeitedIncentives[len(m.ForfeitedIncentives)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPosition(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPosition
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
func (m *PositionWithPeriodLock) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPosition
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
			return fmt.Errorf("proto: PositionWithPeriodLock: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PositionWithPeriodLock: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Position", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
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
				return ErrInvalidLengthPosition
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPosition
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Position.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Locks", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPosition
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
				return ErrInvalidLengthPosition
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPosition
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Locks.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPosition(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPosition
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
func skipPosition(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPosition
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
					return 0, ErrIntOverflowPosition
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
					return 0, ErrIntOverflowPosition
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
				return 0, ErrInvalidLengthPosition
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPosition
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPosition
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPosition        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPosition          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPosition = fmt.Errorf("proto: unexpected end of group")
)
