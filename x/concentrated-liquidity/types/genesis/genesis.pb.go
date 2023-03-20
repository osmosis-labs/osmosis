// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/concentrated-liquidity/genesis.proto

package genesis

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	model "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types1 "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
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

// FullTick contains tick index and pool id along with other tick model
// information.
type FullTick struct {
	// pool id associated with the tick.
	PoolId uint64 `protobuf:"varint,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty" yaml:"pool_id"`
	// tick's index.
	TickIndex int64 `protobuf:"varint,2,opt,name=tick_index,json=tickIndex,proto3" json:"tick_index,omitempty" yaml:"tick_index"`
	// tick's info.
	Info model.TickInfo `protobuf:"bytes,3,opt,name=info,proto3" json:"info" yaml:"tick_info"`
}

func (m *FullTick) Reset()         { *m = FullTick{} }
func (m *FullTick) String() string { return proto.CompactTextString(m) }
func (*FullTick) ProtoMessage()    {}
func (*FullTick) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c140d686ee6724a, []int{0}
}
func (m *FullTick) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FullTick) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FullTick.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FullTick) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FullTick.Merge(m, src)
}
func (m *FullTick) XXX_Size() int {
	return m.Size()
}
func (m *FullTick) XXX_DiscardUnknown() {
	xxx_messageInfo_FullTick.DiscardUnknown(m)
}

var xxx_messageInfo_FullTick proto.InternalMessageInfo

func (m *FullTick) GetPoolId() uint64 {
	if m != nil {
		return m.PoolId
	}
	return 0
}

func (m *FullTick) GetTickIndex() int64 {
	if m != nil {
		return m.TickIndex
	}
	return 0
}

func (m *FullTick) GetInfo() model.TickInfo {
	if m != nil {
		return m.Info
	}
	return model.TickInfo{}
}

// PoolData represents a serialized pool along with its ticks
// for genesis state.
type PoolData struct {
	// pool struct
	Pool *types.Any `protobuf:"bytes,1,opt,name=pool,proto3" json:"pool,omitempty"`
	// pool's ticks
	Ticks       []FullTick  `protobuf:"bytes,2,rep,name=ticks,proto3" json:"ticks" yaml:"ticks"`
	AccumObject AccumObject `protobuf:"bytes,3,opt,name=accum_object,json=accumObject,proto3" json:"accum_object" yaml:"accum_object"`
}

func (m *PoolData) Reset()         { *m = PoolData{} }
func (m *PoolData) String() string { return proto.CompactTextString(m) }
func (*PoolData) ProtoMessage()    {}
func (*PoolData) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c140d686ee6724a, []int{1}
}
func (m *PoolData) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PoolData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PoolData.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PoolData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PoolData.Merge(m, src)
}
func (m *PoolData) XXX_Size() int {
	return m.Size()
}
func (m *PoolData) XXX_DiscardUnknown() {
	xxx_messageInfo_PoolData.DiscardUnknown(m)
}

var xxx_messageInfo_PoolData proto.InternalMessageInfo

func (m *PoolData) GetPool() *types.Any {
	if m != nil {
		return m.Pool
	}
	return nil
}

func (m *PoolData) GetTicks() []FullTick {
	if m != nil {
		return m.Ticks
	}
	return nil
}

func (m *PoolData) GetAccumObject() AccumObject {
	if m != nil {
		return m.AccumObject
	}
	return AccumObject{}
}

// GenesisState defines the concentrated liquidity module's genesis state.
type GenesisState struct {
	// params are all the parameters of the module
	Params types1.Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// pool data containining serialized pool struct and ticks.
	PoolData  []PoolData       `protobuf:"bytes,2,rep,name=pool_data,json=poolData,proto3" json:"pool_data"`
	Positions []model.Position `protobuf:"bytes,3,rep,name=positions,proto3" json:"positions"`
	// incentive records to be set
	IncentiveRecords []types1.IncentiveRecord `protobuf:"bytes,4,rep,name=incentive_records,json=incentiveRecords,proto3" json:"incentive_records"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c140d686ee6724a, []int{2}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() types1.Params {
	if m != nil {
		return m.Params
	}
	return types1.Params{}
}

func (m *GenesisState) GetPoolData() []PoolData {
	if m != nil {
		return m.PoolData
	}
	return nil
}

func (m *GenesisState) GetPositions() []model.Position {
	if m != nil {
		return m.Positions
	}
	return nil
}

func (m *GenesisState) GetIncentiveRecords() []types1.IncentiveRecord {
	if m != nil {
		return m.IncentiveRecords
	}
	return nil
}

type AccumObject struct {
	// Accumulator's name (pulled from AccumulatorContent)
	Name        string                                      `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty" yaml:"name"`
	Value       github_com_cosmos_cosmos_sdk_types.DecCoins `protobuf:"bytes,2,rep,name=value,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.DecCoins" json:"value"`
	TotalShares github_com_cosmos_cosmos_sdk_types.Dec      `protobuf:"bytes,3,opt,name=total_shares,json=totalShares,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"total_shares"`
}

func (m *AccumObject) Reset()         { *m = AccumObject{} }
func (m *AccumObject) String() string { return proto.CompactTextString(m) }
func (*AccumObject) ProtoMessage()    {}
func (*AccumObject) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c140d686ee6724a, []int{3}
}
func (m *AccumObject) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AccumObject) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AccumObject.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *AccumObject) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccumObject.Merge(m, src)
}
func (m *AccumObject) XXX_Size() int {
	return m.Size()
}
func (m *AccumObject) XXX_DiscardUnknown() {
	xxx_messageInfo_AccumObject.DiscardUnknown(m)
}

var xxx_messageInfo_AccumObject proto.InternalMessageInfo

func (m *AccumObject) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *AccumObject) GetValue() github_com_cosmos_cosmos_sdk_types.DecCoins {
	if m != nil {
		return m.Value
	}
	return nil
}

func init() {
	proto.RegisterType((*FullTick)(nil), "osmosis.concentratedliquidity.v1beta1.FullTick")
	proto.RegisterType((*PoolData)(nil), "osmosis.concentratedliquidity.v1beta1.PoolData")
	proto.RegisterType((*GenesisState)(nil), "osmosis.concentratedliquidity.v1beta1.GenesisState")
	proto.RegisterType((*AccumObject)(nil), "osmosis.concentratedliquidity.v1beta1.AccumObject")
}

func init() {
	proto.RegisterFile("osmosis/concentrated-liquidity/genesis.proto", fileDescriptor_5c140d686ee6724a)
}

var fileDescriptor_5c140d686ee6724a = []byte{
	// 699 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xcd, 0x6e, 0xd3, 0x4a,
	0x14, 0x8e, 0x93, 0xb4, 0xb7, 0x99, 0x44, 0xf7, 0xb6, 0x73, 0x7b, 0xa5, 0xdc, 0x82, 0xec, 0xc8,
	0xa8, 0xa8, 0x52, 0x89, 0xad, 0xb6, 0x94, 0x05, 0xbb, 0xba, 0x15, 0x28, 0x6c, 0x28, 0x6e, 0x57,
	0x20, 0x14, 0x8d, 0xed, 0x89, 0x3b, 0xd4, 0xf1, 0x04, 0xcf, 0x24, 0x6a, 0xde, 0xa2, 0xcf, 0xc1,
	0x9a, 0x87, 0xa8, 0x10, 0x8b, 0x2e, 0x11, 0x8b, 0x80, 0xda, 0x05, 0x3b, 0x16, 0x79, 0x02, 0x34,
	0x3f, 0x6e, 0x52, 0x24, 0x94, 0x74, 0x65, 0xcf, 0x9c, 0xef, 0xe7, 0xcc, 0x39, 0x33, 0x07, 0x3c,
	0xa2, 0xac, 0x4b, 0x19, 0x61, 0x6e, 0x48, 0xd3, 0x10, 0xa7, 0x3c, 0x43, 0x1c, 0x47, 0xcd, 0x84,
	0xbc, 0xef, 0x93, 0x88, 0xf0, 0xa1, 0x1b, 0xe3, 0x14, 0x33, 0xc2, 0x9c, 0x5e, 0x46, 0x39, 0x85,
	0xeb, 0x1a, 0xed, 0x4c, 0xa3, 0x6f, 0xc0, 0xce, 0x60, 0x2b, 0xc0, 0x1c, 0x6d, 0xad, 0xad, 0xc6,
	0x34, 0xa6, 0x92, 0xe1, 0x8a, 0x3f, 0x45, 0x5e, 0xfb, 0x3f, 0x94, 0xec, 0xb6, 0x0a, 0xa8, 0x85,
	0x0e, 0x99, 0x6a, 0xe5, 0x06, 0x88, 0x61, 0x57, 0xab, 0xb8, 0x21, 0x25, 0x69, 0x4e, 0x8d, 0x29,
	0x8d, 0x13, 0xec, 0xca, 0x55, 0xd0, 0xef, 0xb8, 0x28, 0x1d, 0xea, 0xd0, 0xe6, 0x8c, 0x03, 0xf4,
	0x50, 0x86, 0xba, 0xb9, 0x4f, 0x73, 0x16, 0x98, 0x32, 0xc2, 0x09, 0x4d, 0xe7, 0x84, 0x73, 0x12,
	0x9e, 0xb6, 0xd2, 0x4e, 0x7e, 0xc0, 0xdd, 0x19, 0x70, 0x22, 0x77, 0xc9, 0x00, 0xb7, 0x33, 0x1c,
	0xd2, 0x2c, 0x52, 0x34, 0xfb, 0xb3, 0x01, 0x96, 0x9e, 0xf5, 0x93, 0xe4, 0x98, 0x84, 0xa7, 0x70,
	0x13, 0xfc, 0xd5, 0xa3, 0x34, 0x69, 0x93, 0xa8, 0x6e, 0x34, 0x8c, 0x8d, 0xb2, 0x07, 0xc7, 0x23,
	0xeb, 0xef, 0x21, 0xea, 0x26, 0x4f, 0x6d, 0x1d, 0xb0, 0xfd, 0x45, 0xf1, 0xd7, 0x8a, 0xe0, 0x63,
	0x00, 0x44, 0x0a, 0x6d, 0x92, 0x46, 0xf8, 0xac, 0x5e, 0x6c, 0x18, 0x1b, 0x25, 0xef, 0xbf, 0xf1,
	0xc8, 0x5a, 0x51, 0xf8, 0x49, 0xcc, 0xf6, 0x2b, 0x2a, 0xd7, 0x08, 0x9f, 0xc1, 0xb7, 0xa0, 0x4c,
	0xd2, 0x0e, 0xad, 0x97, 0x1a, 0xc6, 0x46, 0x75, 0xdb, 0x75, 0xe6, 0xea, 0xa9, 0x73, 0xac, 0xcf,
	0xea, 0xd5, 0x2f, 0x46, 0x56, 0x61, 0x3c, 0xb2, 0x96, 0x6f, 0x99, 0x74, 0xa8, 0xed, 0x4b, 0x59,
	0xfb, 0xbc, 0x08, 0x96, 0x0e, 0x29, 0x4d, 0x0e, 0x10, 0x47, 0x70, 0x07, 0x94, 0x45, 0xae, 0xf2,
	0x2c, 0xd5, 0xed, 0x55, 0x47, 0xf5, 0xd1, 0xc9, 0xfb, 0xe8, 0xec, 0xa5, 0x43, 0xaf, 0xf2, 0xe9,
	0x63, 0x73, 0x41, 0x30, 0x5a, 0xbe, 0x04, 0xc3, 0x37, 0x60, 0x41, 0xa8, 0xb2, 0x7a, 0xb1, 0x51,
	0xba, 0x43, 0x86, 0x79, 0x0d, 0xbd, 0x55, 0x9d, 0x61, 0x6d, 0x92, 0x21, 0xb3, 0x7d, 0xa5, 0x09,
	0x33, 0x50, 0x43, 0x61, 0xd8, 0xef, 0xb6, 0x69, 0xf0, 0x0e, 0x87, 0x5c, 0x57, 0x61, 0x7b, 0x4e,
	0x8f, 0x3d, 0x41, 0x7d, 0x29, 0x99, 0xde, 0x3d, 0x6d, 0xf3, 0xaf, 0xb2, 0x99, 0x56, 0xb5, 0xfd,
	0x2a, 0x9a, 0x20, 0xed, 0x1f, 0x45, 0x50, 0x7b, 0xae, 0x1e, 0xd2, 0x11, 0x47, 0x1c, 0xc3, 0x7d,
	0xb0, 0xa8, 0xee, 0xa5, 0x2e, 0xcc, 0xfa, 0x0c, 0xfb, 0x43, 0x09, 0xf6, 0xca, 0xc2, 0xd1, 0xd7,
	0x54, 0xe8, 0x83, 0x8a, 0xbc, 0x11, 0x11, 0xe2, 0xe8, 0x8e, 0xa5, 0xca, 0xfb, 0xa3, 0x15, 0x97,
	0x7a, 0x79, 0xbf, 0x8e, 0x84, 0xa6, 0x7a, 0x03, 0xac, 0x5e, 0xba, 0xa3, 0xa6, 0xe2, 0x69, 0xcd,
	0x89, 0x0e, 0x24, 0x60, 0xe5, 0xf7, 0xab, 0xcf, 0xea, 0x65, 0x29, 0xfe, 0x64, 0x4e, 0xf1, 0x56,
	0xce, 0xf7, 0x25, 0x5d, 0x7b, 0x2c, 0x93, 0xdb, 0xdb, 0xcc, 0xfe, 0x69, 0x80, 0xea, 0x54, 0x8f,
	0xe0, 0x03, 0x50, 0x4e, 0x51, 0x17, 0xcb, 0x32, 0x57, 0xbc, 0x7f, 0xc6, 0x23, 0xab, 0xaa, 0xba,
	0x25, 0x76, 0x6d, 0x5f, 0x06, 0x61, 0x0c, 0x16, 0x06, 0x28, 0xe9, 0x63, 0x5d, 0xc4, 0xfb, 0x8e,
	0x9e, 0x4d, 0x62, 0x1a, 0xdd, 0x64, 0x70, 0x80, 0xc3, 0x7d, 0x4a, 0x52, 0x6f, 0x47, 0x38, 0x7f,
	0xf8, 0x66, 0x6d, 0xc6, 0x84, 0x9f, 0xf4, 0x03, 0x27, 0xa4, 0x5d, 0x3d, 0xcb, 0xf4, 0xa7, 0xc9,
	0xa2, 0x53, 0x97, 0x0f, 0x7b, 0x98, 0xe5, 0x1c, 0xe6, 0x2b, 0x7d, 0xf8, 0x0a, 0xd4, 0x38, 0xe5,
	0x28, 0x69, 0xb3, 0x13, 0x94, 0x61, 0x26, 0xef, 0x5e, 0xc5, 0x73, 0x84, 0xe2, 0xd7, 0x91, 0xf5,
	0x70, 0x3e, 0x45, 0xbf, 0x2a, 0x35, 0x8e, 0xa4, 0x84, 0x17, 0x5d, 0x5c, 0x99, 0xc6, 0xe5, 0x95,
	0x69, 0x7c, 0xbf, 0x32, 0x8d, 0xf3, 0x6b, 0xb3, 0x70, 0x79, 0x6d, 0x16, 0xbe, 0x5c, 0x9b, 0x85,
	0xd7, 0x2f, 0xa6, 0xe4, 0x74, 0x91, 0x9b, 0x09, 0x0a, 0x58, 0xbe, 0x70, 0x07, 0x5b, 0xbb, 0xee,
	0xd9, 0x1f, 0x47, 0x9b, 0xb0, 0xcb, 0xa7, 0x7f, 0xb0, 0x28, 0x1f, 0xec, 0xce, 0xaf, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x11, 0x7b, 0x99, 0x9d, 0x2e, 0x06, 0x00, 0x00,
}

func (m *FullTick) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FullTick) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FullTick) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Info.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if m.TickIndex != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.TickIndex))
		i--
		dAtA[i] = 0x10
	}
	if m.PoolId != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.PoolId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *PoolData) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PoolData) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PoolData) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.AccumObject.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.Ticks) > 0 {
		for iNdEx := len(m.Ticks) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Ticks[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if m.Pool != nil {
		{
			size, err := m.Pool.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.IncentiveRecords) > 0 {
		for iNdEx := len(m.IncentiveRecords) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.IncentiveRecords[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.Positions) > 0 {
		for iNdEx := len(m.Positions) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Positions[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.PoolData) > 0 {
		for iNdEx := len(m.PoolData) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.PoolData[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *AccumObject) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AccumObject) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *AccumObject) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.TotalShares.Size()
		i -= size
		if _, err := m.TotalShares.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.Value) > 0 {
		for iNdEx := len(m.Value) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Value[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *FullTick) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PoolId != 0 {
		n += 1 + sovGenesis(uint64(m.PoolId))
	}
	if m.TickIndex != 0 {
		n += 1 + sovGenesis(uint64(m.TickIndex))
	}
	l = m.Info.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func (m *PoolData) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Pool != nil {
		l = m.Pool.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	if len(m.Ticks) > 0 {
		for _, e := range m.Ticks {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = m.AccumObject.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.PoolData) > 0 {
		for _, e := range m.PoolData {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Positions) > 0 {
		for _, e := range m.Positions {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.IncentiveRecords) > 0 {
		for _, e := range m.IncentiveRecords {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *AccumObject) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if len(m.Value) > 0 {
		for _, e := range m.Value {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = m.TotalShares.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *FullTick) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: FullTick: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FullTick: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolId", wireType)
			}
			m.PoolId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TickIndex", wireType)
			}
			m.TickIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TickIndex |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Info", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Info.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *PoolData) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: PoolData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PoolData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pool", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Pool == nil {
				m.Pool = &types.Any{}
			}
			if err := m.Pool.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ticks", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Ticks = append(m.Ticks, FullTick{})
			if err := m.Ticks[len(m.Ticks)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccumObject", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.AccumObject.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolData", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PoolData = append(m.PoolData, PoolData{})
			if err := m.PoolData[len(m.PoolData)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Positions", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Positions = append(m.Positions, model.Position{})
			if err := m.Positions[len(m.Positions)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IncentiveRecords", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IncentiveRecords = append(m.IncentiveRecords, types1.IncentiveRecord{})
			if err := m.IncentiveRecords[len(m.IncentiveRecords)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *AccumObject) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: AccumObject: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AccumObject: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = append(m.Value, types2.DecCoin{})
			if err := m.Value[len(m.Value)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalShares", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.TotalShares.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
