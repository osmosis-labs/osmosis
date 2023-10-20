// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/protorev/v1beta1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	types "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"
	
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

// GenesisState defines the protorev module's genesis state.
type GenesisState struct {
	// Parameters for the protorev module.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// Token pair arb routes for the protorev module (hot routes).
	TokenPairArbRoutes []TokenPairArbRoutes `protobuf:"bytes,2,rep,name=token_pair_arb_routes,json=tokenPairArbRoutes,proto3" json:"token_pair_arb_routes" yaml:"token_pair_arb_routes"`
	// The base denominations being used to create cyclic arbitrage routes via the
	// highest liquidity method.
	BaseDenoms []BaseDenom `protobuf:"bytes,3,rep,name=base_denoms,json=baseDenoms,proto3" json:"base_denoms" yaml:"base_denoms"`
	// The pool weights that are being used to calculate the weight (compute cost)
	// of each route.
	//
	// DEPRECATED: This field is deprecated and will be removed in the next
	// release. It is replaced by the `info_by_pool_type` field.
	PoolWeights PoolWeights `protobuf:"bytes,4,opt,name=pool_weights,json=poolWeights,proto3" json:"pool_weights" yaml:"pool_weights"`
	// The number of days since module genesis.
	DaysSinceModuleGenesis uint64 `protobuf:"varint,5,opt,name=days_since_module_genesis,json=daysSinceModuleGenesis,proto3" json:"days_since_module_genesis,omitempty" yaml:"days_since_module_genesis"`
	// The fees the developer account has accumulated over time.
	DeveloperFees []types.Coin `protobuf:"bytes,6,rep,name=developer_fees,json=developerFees,proto3" json:"developer_fees" yaml:"developer_fees"`
	// The latest block height that the module has processed.
	LatestBlockHeight uint64 `protobuf:"varint,7,opt,name=latest_block_height,json=latestBlockHeight,proto3" json:"latest_block_height,omitempty" yaml:"latest_block_height"`
	// The developer account address of the module.
	DeveloperAddress string `protobuf:"bytes,8,opt,name=developer_address,json=developerAddress,proto3" json:"developer_address,omitempty" yaml:"developer_address"`
	// Max pool points per block i.e. the maximum compute time (in ms)
	// that protorev can use per block.
	MaxPoolPointsPerBlock uint64 `protobuf:"varint,9,opt,name=max_pool_points_per_block,json=maxPoolPointsPerBlock,proto3" json:"max_pool_points_per_block,omitempty" yaml:"max_pool_points_per_block"`
	// Max pool points per tx i.e. the maximum compute time (in ms) that
	// protorev can use per tx.
	MaxPoolPointsPerTx uint64 `protobuf:"varint,10,opt,name=max_pool_points_per_tx,json=maxPoolPointsPerTx,proto3" json:"max_pool_points_per_tx,omitempty" yaml:"max_pool_points_per_tx"`
	// The number of pool points that have been consumed in the current block.
	PointCountForBlock uint64 `protobuf:"varint,11,opt,name=point_count_for_block,json=pointCountForBlock,proto3" json:"point_count_for_block,omitempty" yaml:"point_count_for_block"`
	// All of the profits that have been accumulated by the module.
	Profits []types.Coin `protobuf:"bytes,12,rep,name=profits,proto3" json:"profits" yaml:"profits"`
	// Information that is used to estimate execution time / gas
	// consumption of a swap on a given pool type.
	InfoByPoolType InfoByPoolType `protobuf:"bytes,13,opt,name=info_by_pool_type,json=infoByPoolType,proto3" json:"info_by_pool_type" yaml:"info_by_pool_type"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_3c77fc2da5752af2, []int{0}
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

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetTokenPairArbRoutes() []TokenPairArbRoutes {
	if m != nil {
		return m.TokenPairArbRoutes
	}
	return nil
}

func (m *GenesisState) GetBaseDenoms() []BaseDenom {
	if m != nil {
		return m.BaseDenoms
	}
	return nil
}

func (m *GenesisState) GetPoolWeights() PoolWeights {
	if m != nil {
		return m.PoolWeights
	}
	return PoolWeights{}
}

func (m *GenesisState) GetDaysSinceModuleGenesis() uint64 {
	if m != nil {
		return m.DaysSinceModuleGenesis
	}
	return 0
}

func (m *GenesisState) GetDeveloperFees() []types.Coin {
	if m != nil {
		return m.DeveloperFees
	}
	return nil
}

func (m *GenesisState) GetLatestBlockHeight() uint64 {
	if m != nil {
		return m.LatestBlockHeight
	}
	return 0
}

func (m *GenesisState) GetDeveloperAddress() string {
	if m != nil {
		return m.DeveloperAddress
	}
	return ""
}

func (m *GenesisState) GetMaxPoolPointsPerBlock() uint64 {
	if m != nil {
		return m.MaxPoolPointsPerBlock
	}
	return 0
}

func (m *GenesisState) GetMaxPoolPointsPerTx() uint64 {
	if m != nil {
		return m.MaxPoolPointsPerTx
	}
	return 0
}

func (m *GenesisState) GetPointCountForBlock() uint64 {
	if m != nil {
		return m.PointCountForBlock
	}
	return 0
}

func (m *GenesisState) GetProfits() []types.Coin {
	if m != nil {
		return m.Profits
	}
	return nil
}

func (m *GenesisState) GetInfoByPoolType() InfoByPoolType {
	if m != nil {
		return m.InfoByPoolType
	}
	return InfoByPoolType{}
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "osmosis.protorev.v1beta1.GenesisState")
}

func init() {
	proto.RegisterFile("osmosis/protorev/v1beta1/genesis.proto", fileDescriptor_3c77fc2da5752af2)
}

var fileDescriptor_3c77fc2da5752af2 = []byte{
	// 727 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x94, 0xcd, 0x6e, 0xd3, 0x4a,
	0x14, 0xc7, 0xe3, 0xdb, 0xde, 0xf6, 0x76, 0xd2, 0x46, 0xb7, 0xd3, 0x9b, 0xca, 0xc9, 0xa5, 0x8e,
	0x31, 0x2d, 0x64, 0x41, 0x6d, 0x5a, 0x58, 0xb1, 0x40, 0xaa, 0x8b, 0x0a, 0x08, 0x51, 0x45, 0x6e,
	0x11, 0x12, 0x48, 0x0c, 0xe3, 0x64, 0x92, 0x5a, 0xb5, 0x3d, 0x96, 0x67, 0x12, 0x92, 0x07, 0x60,
	0xcf, 0xc3, 0xf0, 0x10, 0x5d, 0x56, 0xac, 0x58, 0x45, 0xa8, 0x7d, 0x83, 0x3c, 0x01, 0xf2, 0xcc,
	0x24, 0xe9, 0x47, 0x0c, 0x3b, 0xcf, 0x39, 0xbf, 0xf3, 0xff, 0x9f, 0x33, 0x1f, 0x06, 0xf7, 0x29,
	0x8b, 0x28, 0x0b, 0x98, 0x93, 0xa4, 0x94, 0xd3, 0x94, 0xf4, 0x9c, 0xde, 0x8e, 0x4f, 0x38, 0xde,
	0x71, 0x3a, 0x24, 0x26, 0x2c, 0x60, 0xb6, 0x48, 0x40, 0x5d, 0x71, 0xf6, 0x98, 0xb3, 0x15, 0x57,
	0xfd, 0xaf, 0x43, 0x3b, 0x54, 0x44, 0x9d, 0xec, 0x4b, 0x02, 0xd5, 0x07, 0xb9, 0xba, 0x13, 0x01,
	0x09, 0x6e, 0xe5, 0x83, 0x38, 0xc5, 0x91, 0x32, 0xac, 0x56, 0x9a, 0x82, 0x43, 0xd2, 0x48, 0x2e,
	0x54, 0xca, 0x90, 0x2b, 0xc7, 0xc7, 0x8c, 0x4c, 0x8a, 0x9b, 0x34, 0x88, 0x65, 0xde, 0x1a, 0x2e,
	0x81, 0xe5, 0x17, 0x72, 0x98, 0x23, 0x8e, 0x39, 0x81, 0xcf, 0xc0, 0x82, 0xd4, 0xd6, 0x35, 0x53,
	0xab, 0x17, 0x77, 0x4d, 0x3b, 0x6f, 0x38, 0xbb, 0x21, 0x38, 0x77, 0xfe, 0x6c, 0x58, 0x2b, 0x78,
	0xaa, 0x0a, 0x7e, 0xd1, 0x40, 0x99, 0xd3, 0x53, 0x12, 0xa3, 0x04, 0x07, 0x29, 0xc2, 0xa9, 0x8f,
	0x52, 0xda, 0xe5, 0x84, 0xe9, 0x7f, 0x99, 0x73, 0xf5, 0xe2, 0xee, 0xc3, 0x7c, 0xbd, 0xe3, 0xac,
	0xac, 0x81, 0x83, 0x74, 0x2f, 0xf5, 0x3d, 0x51, 0xe3, 0x6e, 0x66, 0xda, 0xa3, 0x61, 0xed, 0xce,
	0x00, 0x47, 0xe1, 0x53, 0x6b, 0xa6, 0xb0, 0xe5, 0x41, 0x7e, 0xab, 0x12, 0x7e, 0x02, 0xc5, 0x6c,
	0x66, 0xd4, 0x22, 0x31, 0x8d, 0x98, 0x3e, 0x27, 0xcc, 0xef, 0xe5, 0x9b, 0xbb, 0x98, 0x91, 0xe7,
	0x19, 0xeb, 0x56, 0x95, 0x27, 0x94, 0x9e, 0x57, 0x54, 0x2c, 0x0f, 0xf8, 0x63, 0x8c, 0x41, 0x02,
	0x96, 0x13, 0x4a, 0x43, 0xf4, 0x99, 0x04, 0x9d, 0x13, 0xce, 0xf4, 0x79, 0xb1, 0x5f, 0x5b, 0xbf,
	0xd9, 0x2f, 0x4a, 0xc3, 0x77, 0x12, 0x76, 0xff, 0x57, 0x26, 0x6b, 0xd2, 0xe4, 0xaa, 0x90, 0xe5,
	0x15, 0x93, 0x29, 0x09, 0x11, 0xa8, 0xb4, 0xf0, 0x80, 0x21, 0x16, 0xc4, 0x4d, 0x82, 0x22, 0xda,
	0xea, 0x86, 0x04, 0xa9, 0xfb, 0xa7, 0xff, 0x6d, 0x6a, 0xf5, 0x79, 0x77, 0x73, 0x34, 0xac, 0x99,
	0x52, 0x28, 0x17, 0xb5, 0xbc, 0xf5, 0x2c, 0x77, 0x94, 0xa5, 0xde, 0x88, 0x8c, 0x3a, 0x76, 0x88,
	0x40, 0xa9, 0x45, 0x7a, 0x24, 0xa4, 0x09, 0x49, 0x51, 0x9b, 0x10, 0xa6, 0x2f, 0x88, 0xcd, 0xaa,
	0xd8, 0xea, 0x26, 0x65, 0x33, 0x4f, 0x86, 0xd8, 0xa7, 0x41, 0xec, 0x6e, 0xa8, 0xee, 0xcb, 0xca,
	0xf4, 0x5a, 0xb9, 0xe5, 0xad, 0x4c, 0x02, 0x07, 0x84, 0x30, 0x78, 0x08, 0xd6, 0x42, 0xcc, 0x09,
	0xe3, 0xc8, 0x0f, 0x69, 0xf3, 0x14, 0x9d, 0x88, 0xc9, 0xf4, 0x45, 0xd1, 0xbb, 0x31, 0x1a, 0xd6,
	0xaa, 0x52, 0x66, 0x06, 0x64, 0x79, 0xab, 0x32, 0xea, 0x66, 0xc1, 0x97, 0x22, 0x06, 0x3f, 0x80,
	0xd5, 0xa9, 0x23, 0x6e, 0xb5, 0x52, 0xc2, 0x98, 0xfe, 0x8f, 0xa9, 0xd5, 0x97, 0x5c, 0x7b, 0x34,
	0xac, 0xe9, 0x37, 0x9b, 0x52, 0x88, 0xf5, 0xfd, 0xdb, 0x76, 0x49, 0x8d, 0xb4, 0x27, 0x43, 0xde,
	0xbf, 0x13, 0x4a, 0x45, 0xe0, 0x47, 0x50, 0x89, 0x70, 0x1f, 0x89, 0x03, 0x49, 0x68, 0x10, 0x73,
	0x86, 0x32, 0x0d, 0xd1, 0x94, 0xbe, 0x74, 0x73, 0xbb, 0x73, 0x51, 0xcb, 0x2b, 0x47, 0xb8, 0x9f,
	0x9d, 0x78, 0x43, 0x64, 0x1a, 0x24, 0x15, 0x23, 0xc0, 0xb7, 0x60, 0x7d, 0x56, 0x11, 0xef, 0xeb,
	0x40, 0x88, 0xdf, 0x1d, 0x0d, 0x6b, 0x1b, 0xf9, 0xe2, 0xbc, 0x6f, 0x79, 0xf0, 0xa6, 0xf2, 0x71,
	0x1f, 0x1e, 0x81, 0xb2, 0xa0, 0x50, 0x93, 0x76, 0x63, 0x8e, 0xda, 0x74, 0xdc, 0x72, 0x51, 0xa8,
	0x9a, 0xd3, 0x37, 0x34, 0x13, 0xb3, 0x3c, 0x28, 0xe2, 0xfb, 0x59, 0xf8, 0x80, 0xaa, 0x5e, 0x5f,
	0x83, 0xc5, 0x24, 0xa5, 0xed, 0x80, 0x33, 0x7d, 0xf9, 0x4f, 0x57, 0x62, 0x5d, 0x5d, 0x89, 0x92,
	0x72, 0x91, 0x75, 0x96, 0x37, 0x56, 0x80, 0x5d, 0xb0, 0x1a, 0xc4, 0x6d, 0x8a, 0xfc, 0x81, 0x1c,
	0x8a, 0x0f, 0x12, 0xa2, 0xaf, 0x88, 0x37, 0x53, 0xcf, 0x7f, 0x33, 0xaf, 0xe2, 0x36, 0x75, 0x07,
	0xd9, 0xb4, 0xc7, 0x83, 0x84, 0xb8, 0xa6, 0x72, 0x51, 0x67, 0x7c, 0x4b, 0xd0, 0xf2, 0x4a, 0xc1,
	0xf5, 0x8a, 0xc3, 0xb3, 0x0b, 0x43, 0x3b, 0xbf, 0x30, 0xb4, 0x9f, 0x17, 0x86, 0xf6, 0xf5, 0xd2,
	0x28, 0x9c, 0x5f, 0x1a, 0x85, 0x1f, 0x97, 0x46, 0xe1, 0xfd, 0x93, 0x4e, 0xc0, 0x4f, 0xba, 0xbe,
	0xdd, 0xa4, 0x91, 0xa3, 0xfc, 0xb7, 0x43, 0xec, 0xb3, 0xf1, 0xc2, 0xe9, 0xed, 0x3e, 0x72, 0xfa,
	0xd3, 0x5f, 0x6f, 0xa6, 0xcf, 0xfc, 0x05, 0xb1, 0x7e, 0xfc, 0x2b, 0x00, 0x00, 0xff, 0xff, 0xe1,
	0xfb, 0x5d, 0x7a, 0x1c, 0x06, 0x00, 0x00,
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
	{
		size, err := m.InfoByPoolType.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x6a
	if len(m.Profits) > 0 {
		for iNdEx := len(m.Profits) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Profits[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x62
		}
	}
	if m.PointCountForBlock != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.PointCountForBlock))
		i--
		dAtA[i] = 0x58
	}
	if m.MaxPoolPointsPerTx != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxPoolPointsPerTx))
		i--
		dAtA[i] = 0x50
	}
	if m.MaxPoolPointsPerBlock != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxPoolPointsPerBlock))
		i--
		dAtA[i] = 0x48
	}
	if len(m.DeveloperAddress) > 0 {
		i -= len(m.DeveloperAddress)
		copy(dAtA[i:], m.DeveloperAddress)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.DeveloperAddress)))
		i--
		dAtA[i] = 0x42
	}
	if m.LatestBlockHeight != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.LatestBlockHeight))
		i--
		dAtA[i] = 0x38
	}
	if len(m.DeveloperFees) > 0 {
		for iNdEx := len(m.DeveloperFees) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DeveloperFees[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x32
		}
	}
	if m.DaysSinceModuleGenesis != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.DaysSinceModuleGenesis))
		i--
		dAtA[i] = 0x28
	}
	{
		size, err := m.PoolWeights.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if len(m.BaseDenoms) > 0 {
		for iNdEx := len(m.BaseDenoms) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BaseDenoms[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.TokenPairArbRoutes) > 0 {
		for iNdEx := len(m.TokenPairArbRoutes) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TokenPairArbRoutes[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.TokenPairArbRoutes) > 0 {
		for _, e := range m.TokenPairArbRoutes {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.BaseDenoms) > 0 {
		for _, e := range m.BaseDenoms {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = m.PoolWeights.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if m.DaysSinceModuleGenesis != 0 {
		n += 1 + sovGenesis(uint64(m.DaysSinceModuleGenesis))
	}
	if len(m.DeveloperFees) > 0 {
		for _, e := range m.DeveloperFees {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if m.LatestBlockHeight != 0 {
		n += 1 + sovGenesis(uint64(m.LatestBlockHeight))
	}
	l = len(m.DeveloperAddress)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.MaxPoolPointsPerBlock != 0 {
		n += 1 + sovGenesis(uint64(m.MaxPoolPointsPerBlock))
	}
	if m.MaxPoolPointsPerTx != 0 {
		n += 1 + sovGenesis(uint64(m.MaxPoolPointsPerTx))
	}
	if m.PointCountForBlock != 0 {
		n += 1 + sovGenesis(uint64(m.PointCountForBlock))
	}
	if len(m.Profits) > 0 {
		for _, e := range m.Profits {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = m.InfoByPoolType.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
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
				return fmt.Errorf("proto: wrong wireType = %d for field TokenPairArbRoutes", wireType)
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
			m.TokenPairArbRoutes = append(m.TokenPairArbRoutes, TokenPairArbRoutes{})
			if err := m.TokenPairArbRoutes[len(m.TokenPairArbRoutes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseDenoms", wireType)
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
			m.BaseDenoms = append(m.BaseDenoms, BaseDenom{})
			if err := m.BaseDenoms[len(m.BaseDenoms)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolWeights", wireType)
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
			if err := m.PoolWeights.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DaysSinceModuleGenesis", wireType)
			}
			m.DaysSinceModuleGenesis = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.DaysSinceModuleGenesis |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DeveloperFees", wireType)
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
			m.DeveloperFees = append(m.DeveloperFees, types.Coin{})
			if err := m.DeveloperFees[len(m.DeveloperFees)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LatestBlockHeight", wireType)
			}
			m.LatestBlockHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LatestBlockHeight |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DeveloperAddress", wireType)
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
			m.DeveloperAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxPoolPointsPerBlock", wireType)
			}
			m.MaxPoolPointsPerBlock = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxPoolPointsPerBlock |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 10:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxPoolPointsPerTx", wireType)
			}
			m.MaxPoolPointsPerTx = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxPoolPointsPerTx |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 11:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PointCountForBlock", wireType)
			}
			m.PointCountForBlock = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PointCountForBlock |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Profits", wireType)
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
			m.Profits = append(m.Profits, types.Coin{})
			if err := m.Profits[len(m.Profits)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 13:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InfoByPoolType", wireType)
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
			if err := m.InfoByPoolType.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
