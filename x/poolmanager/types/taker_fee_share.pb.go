// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/poolmanager/v1beta1/taker_fee_share.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
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

// TakerFeeShareAgreement represents the agreement between the Osmosis protocol
// and a specific denom to share a certain percent of taker fees generated in
// any route that contains said denom. For example, if the agreement specifies a
// 10% skim_percent, this means 10% of the taker fees generated in a swap route
// containing the specified denom will be sent to the address specified
// in the skim_address field at the end of each epoch. These skim_percents are
// additive, so if three taker fee agreements have skim percents of 10%, 20%,
// and 30%, the total skim percent for the route will be 60%.
type TakerFeeShareAgreement struct {
	// denom is the denom that has the taker fee share agreement.
	Denom string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty" yaml:"denom"`
	// skim_percent is the percentage of taker fees that will be skimmed for the
	// denom, in the event that the denom is included in the swap route.
	SkimPercent cosmossdk_io_math.LegacyDec `protobuf:"bytes,2,opt,name=skim_percent,json=skimPercent,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"skim_percent" yaml:"skim_percent"`
	// skim_address is the address belonging to the respective denom
	// that the skimmed taker fees will be sent to at the end of each epoch.
	SkimAddress string `protobuf:"bytes,3,opt,name=skim_address,json=skimAddress,proto3" json:"skim_address,omitempty" yaml:"skim_address"`
}

func (m *TakerFeeShareAgreement) Reset()         { *m = TakerFeeShareAgreement{} }
func (m *TakerFeeShareAgreement) String() string { return proto.CompactTextString(m) }
func (*TakerFeeShareAgreement) ProtoMessage()    {}
func (*TakerFeeShareAgreement) Descriptor() ([]byte, []int) {
	return fileDescriptor_eda6ab99820fcb49, []int{0}
}
func (m *TakerFeeShareAgreement) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TakerFeeShareAgreement) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TakerFeeShareAgreement.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TakerFeeShareAgreement) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TakerFeeShareAgreement.Merge(m, src)
}
func (m *TakerFeeShareAgreement) XXX_Size() int {
	return m.Size()
}
func (m *TakerFeeShareAgreement) XXX_DiscardUnknown() {
	xxx_messageInfo_TakerFeeShareAgreement.DiscardUnknown(m)
}

var xxx_messageInfo_TakerFeeShareAgreement proto.InternalMessageInfo

func (m *TakerFeeShareAgreement) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

func (m *TakerFeeShareAgreement) GetSkimAddress() string {
	if m != nil {
		return m.SkimAddress
	}
	return ""
}

// TakerFeeSkimAccumulator accumulates the total skimmed taker fees for each
// denom that has a taker fee share agreement.
type TakerFeeSkimAccumulator struct {
	// denom is the denom that has the taker fee share agreement.
	Denom string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty" yaml:"denom"`
	// skimmed_taker_fees is the total skimmed taker fees for the denom.
	SkimmedTakerFees github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,2,rep,name=skimmed_taker_fees,json=skimmedTakerFees,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"skimmed_taker_fees" yaml:"skimmed_taker_fees"`
}

func (m *TakerFeeSkimAccumulator) Reset()         { *m = TakerFeeSkimAccumulator{} }
func (m *TakerFeeSkimAccumulator) String() string { return proto.CompactTextString(m) }
func (*TakerFeeSkimAccumulator) ProtoMessage()    {}
func (*TakerFeeSkimAccumulator) Descriptor() ([]byte, []int) {
	return fileDescriptor_eda6ab99820fcb49, []int{1}
}
func (m *TakerFeeSkimAccumulator) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TakerFeeSkimAccumulator) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TakerFeeSkimAccumulator.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TakerFeeSkimAccumulator) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TakerFeeSkimAccumulator.Merge(m, src)
}
func (m *TakerFeeSkimAccumulator) XXX_Size() int {
	return m.Size()
}
func (m *TakerFeeSkimAccumulator) XXX_DiscardUnknown() {
	xxx_messageInfo_TakerFeeSkimAccumulator.DiscardUnknown(m)
}

var xxx_messageInfo_TakerFeeSkimAccumulator proto.InternalMessageInfo

func (m *TakerFeeSkimAccumulator) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

func (m *TakerFeeSkimAccumulator) GetSkimmedTakerFees() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.SkimmedTakerFees
	}
	return nil
}

// AlloyContractTakerFeeShareState contains the contract address of the alloyed
// asset pool, along with the adjusted taker fee share agreements for any asset
// within the alloyed asset pool that has a taker fee share agreement. If for
// instance there are two denoms, and denomA makes up 50 percent and denomB
// makes up 50 percent, and denom A has a taker fee share agreement with a skim
// percent of 10%, then the adjusted taker fee share agreement for denomA will
// be 5%.
type AlloyContractTakerFeeShareState struct {
	// contract_address is the address of the alloyed asset pool contract.
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty" yaml:"contract_address"`
	// taker_fee_share_agreements is the adjusted taker fee share agreements for
	// any asset within the alloyed asset pool that has a taker fee share
	// agreement.
	TakerFeeShareAgreements []TakerFeeShareAgreement `protobuf:"bytes,2,rep,name=taker_fee_share_agreements,json=takerFeeShareAgreements,proto3" json:"taker_fee_share_agreements" yaml:"taker_fee_share_agreements"`
}

func (m *AlloyContractTakerFeeShareState) Reset()         { *m = AlloyContractTakerFeeShareState{} }
func (m *AlloyContractTakerFeeShareState) String() string { return proto.CompactTextString(m) }
func (*AlloyContractTakerFeeShareState) ProtoMessage()    {}
func (*AlloyContractTakerFeeShareState) Descriptor() ([]byte, []int) {
	return fileDescriptor_eda6ab99820fcb49, []int{2}
}
func (m *AlloyContractTakerFeeShareState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AlloyContractTakerFeeShareState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AlloyContractTakerFeeShareState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *AlloyContractTakerFeeShareState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AlloyContractTakerFeeShareState.Merge(m, src)
}
func (m *AlloyContractTakerFeeShareState) XXX_Size() int {
	return m.Size()
}
func (m *AlloyContractTakerFeeShareState) XXX_DiscardUnknown() {
	xxx_messageInfo_AlloyContractTakerFeeShareState.DiscardUnknown(m)
}

var xxx_messageInfo_AlloyContractTakerFeeShareState proto.InternalMessageInfo

func (m *AlloyContractTakerFeeShareState) GetContractAddress() string {
	if m != nil {
		return m.ContractAddress
	}
	return ""
}

func (m *AlloyContractTakerFeeShareState) GetTakerFeeShareAgreements() []TakerFeeShareAgreement {
	if m != nil {
		return m.TakerFeeShareAgreements
	}
	return nil
}

func init() {
	proto.RegisterType((*TakerFeeShareAgreement)(nil), "osmosis.poolmanager.v1beta1.TakerFeeShareAgreement")
	proto.RegisterType((*TakerFeeSkimAccumulator)(nil), "osmosis.poolmanager.v1beta1.TakerFeeSkimAccumulator")
	proto.RegisterType((*AlloyContractTakerFeeShareState)(nil), "osmosis.poolmanager.v1beta1.AlloyContractTakerFeeShareState")
}

func init() {
	proto.RegisterFile("osmosis/poolmanager/v1beta1/taker_fee_share.proto", fileDescriptor_eda6ab99820fcb49)
}

var fileDescriptor_eda6ab99820fcb49 = []byte{
	// 506 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x8e, 0x53, 0x81, 0x84, 0x5b, 0x89, 0xc8, 0x20, 0x92, 0x26, 0x92, 0x5d, 0x7c, 0x40, 0xe1,
	0xd0, 0x5d, 0xa5, 0x3d, 0x20, 0x72, 0x4b, 0x8a, 0x7a, 0x02, 0x09, 0x52, 0x4e, 0x48, 0xc8, 0x5a,
	0xaf, 0x07, 0xc7, 0x8a, 0xd7, 0x1b, 0xed, 0x6e, 0x2a, 0xf2, 0x04, 0x5c, 0x39, 0x71, 0xe1, 0x0d,
	0x78, 0x92, 0x1e, 0x7b, 0x44, 0x45, 0x32, 0x28, 0x79, 0x83, 0x3c, 0x01, 0xb2, 0x77, 0x1d, 0x92,
	0xa8, 0x20, 0x4e, 0xf6, 0xfc, 0x7c, 0xdf, 0xcc, 0x7c, 0x33, 0x6b, 0xf7, 0xb8, 0x64, 0x5c, 0x26,
	0x12, 0x4f, 0x39, 0x4f, 0x19, 0xc9, 0x48, 0x0c, 0x02, 0x5f, 0xf6, 0x42, 0x50, 0xa4, 0x87, 0x15,
	0x99, 0x80, 0x08, 0x3e, 0x00, 0x04, 0x72, 0x4c, 0x04, 0xa0, 0xa9, 0xe0, 0x8a, 0x3b, 0x1d, 0x03,
	0x41, 0x1b, 0x10, 0x64, 0x20, 0xed, 0x87, 0x31, 0x8f, 0x79, 0x99, 0x87, 0x8b, 0x3f, 0x0d, 0x69,
	0xbb, 0xb4, 0xc4, 0xe0, 0x90, 0x48, 0x58, 0xb3, 0x53, 0x9e, 0x64, 0x3a, 0xee, 0xff, 0xb0, 0xec,
	0x47, 0x6f, 0x8b, 0x62, 0xe7, 0x00, 0x17, 0x45, 0xa9, 0x41, 0x2c, 0x00, 0x18, 0x64, 0xca, 0x79,
	0x62, 0xdf, 0x89, 0x20, 0xe3, 0xac, 0x65, 0x1d, 0x59, 0xdd, 0x7b, 0xc3, 0xc6, 0x2a, 0xf7, 0x0e,
	0xe6, 0x84, 0xa5, 0x7d, 0xbf, 0x74, 0xfb, 0x23, 0x1d, 0x76, 0xde, 0xdb, 0x07, 0x72, 0x92, 0xb0,
	0x60, 0x0a, 0x82, 0x42, 0xa6, 0x5a, 0xf5, 0x32, 0xbd, 0x7f, 0x95, 0x7b, 0xb5, 0x9b, 0xdc, 0xeb,
	0xe8, 0x06, 0x64, 0x34, 0x41, 0x09, 0xc7, 0x8c, 0xa8, 0x31, 0x7a, 0x09, 0x31, 0xa1, 0xf3, 0x17,
	0x40, 0x57, 0xb9, 0xf7, 0x40, 0x33, 0x6e, 0x12, 0xf8, 0xa3, 0xfd, 0xc2, 0x7c, 0xad, 0x2d, 0xa7,
	0x6f, 0xe8, 0x49, 0x14, 0x09, 0x90, 0xb2, 0xb5, 0x57, 0xd2, 0x37, 0x77, 0xb0, 0x26, 0x6a, 0xb0,
	0x03, 0x63, 0xdd, 0x58, 0x76, 0x73, 0x3d, 0x5d, 0xe1, 0xa7, 0x74, 0xc6, 0x66, 0x29, 0x51, 0x5c,
	0xfc, 0xf7, 0x78, 0x5f, 0x2c, 0xdb, 0x29, 0x38, 0x19, 0x44, 0xc1, 0x7a, 0x2d, 0xb2, 0x55, 0x3f,
	0xda, 0xeb, 0xee, 0x9f, 0x1c, 0x22, 0x3d, 0x1e, 0x2a, 0xf4, 0xad, 0x56, 0x81, 0xce, 0x78, 0x92,
	0x0d, 0x5f, 0x15, 0x02, 0xac, 0x72, 0xef, 0xf0, 0x4f, 0x97, 0xdb, 0x14, 0xfe, 0xb7, 0x9f, 0x5e,
	0x37, 0x4e, 0xd4, 0x78, 0x16, 0x22, 0xca, 0x19, 0x36, 0x9b, 0xd2, 0x9f, 0x63, 0x19, 0x4d, 0xb0,
	0x9a, 0x4f, 0x41, 0x96, 0x6c, 0x72, 0xd4, 0x30, 0x04, 0xd5, 0x38, 0xd2, 0xff, 0x54, 0xb7, 0xbd,
	0x41, 0x9a, 0xf2, 0xf9, 0x19, 0xcf, 0x94, 0x20, 0x54, 0x6d, 0xed, 0xf1, 0x42, 0x11, 0x05, 0xce,
	0xb9, 0xdd, 0xa0, 0x26, 0xba, 0x16, 0x50, 0xcf, 0xdb, 0x59, 0xe5, 0x5e, 0x53, 0xb7, 0xb6, 0x9b,
	0xe1, 0x8f, 0xee, 0x57, 0x2e, 0x23, 0xa4, 0xf3, 0xd5, 0xb2, 0xdb, 0x3b, 0x37, 0x19, 0x90, 0xea,
	0x52, 0x2a, 0x31, 0x4e, 0xd1, 0x3f, 0xee, 0x13, 0xdd, 0x7e, 0x65, 0xc3, 0xa7, 0x46, 0xa6, 0xc7,
	0xba, 0x97, 0xbf, 0x17, 0xf1, 0x47, 0x4d, 0x75, 0x2b, 0x85, 0x1c, 0xbe, 0xb9, 0x5a, 0xb8, 0xd6,
	0xf5, 0xc2, 0xb5, 0x7e, 0x2d, 0x5c, 0xeb, 0xf3, 0xd2, 0xad, 0x5d, 0x2f, 0xdd, 0xda, 0xf7, 0xa5,
	0x5b, 0x7b, 0xf7, 0x6c, 0x43, 0x5f, 0xd3, 0xdc, 0x71, 0x4a, 0x42, 0x59, 0x19, 0xf8, 0xf2, 0xe4,
	0x39, 0xfe, 0xb8, 0xf5, 0x04, 0x4b, 0xd1, 0xc3, 0xbb, 0xe5, 0xf3, 0x38, 0xfd, 0x1d, 0x00, 0x00,
	0xff, 0xff, 0x19, 0xa9, 0x0d, 0xf1, 0xa6, 0x03, 0x00, 0x00,
}

func (m *TakerFeeShareAgreement) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TakerFeeShareAgreement) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TakerFeeShareAgreement) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SkimAddress) > 0 {
		i -= len(m.SkimAddress)
		copy(dAtA[i:], m.SkimAddress)
		i = encodeVarintTakerFeeShare(dAtA, i, uint64(len(m.SkimAddress)))
		i--
		dAtA[i] = 0x1a
	}
	{
		size := m.SkimPercent.Size()
		i -= size
		if _, err := m.SkimPercent.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintTakerFeeShare(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintTakerFeeShare(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *TakerFeeSkimAccumulator) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TakerFeeSkimAccumulator) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TakerFeeSkimAccumulator) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SkimmedTakerFees) > 0 {
		for iNdEx := len(m.SkimmedTakerFees) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.SkimmedTakerFees[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTakerFeeShare(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintTakerFeeShare(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *AlloyContractTakerFeeShareState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AlloyContractTakerFeeShareState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *AlloyContractTakerFeeShareState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.TakerFeeShareAgreements) > 0 {
		for iNdEx := len(m.TakerFeeShareAgreements) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TakerFeeShareAgreements[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTakerFeeShare(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.ContractAddress) > 0 {
		i -= len(m.ContractAddress)
		copy(dAtA[i:], m.ContractAddress)
		i = encodeVarintTakerFeeShare(dAtA, i, uint64(len(m.ContractAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintTakerFeeShare(dAtA []byte, offset int, v uint64) int {
	offset -= sovTakerFeeShare(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *TakerFeeShareAgreement) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovTakerFeeShare(uint64(l))
	}
	l = m.SkimPercent.Size()
	n += 1 + l + sovTakerFeeShare(uint64(l))
	l = len(m.SkimAddress)
	if l > 0 {
		n += 1 + l + sovTakerFeeShare(uint64(l))
	}
	return n
}

func (m *TakerFeeSkimAccumulator) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovTakerFeeShare(uint64(l))
	}
	if len(m.SkimmedTakerFees) > 0 {
		for _, e := range m.SkimmedTakerFees {
			l = e.Size()
			n += 1 + l + sovTakerFeeShare(uint64(l))
		}
	}
	return n
}

func (m *AlloyContractTakerFeeShareState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ContractAddress)
	if l > 0 {
		n += 1 + l + sovTakerFeeShare(uint64(l))
	}
	if len(m.TakerFeeShareAgreements) > 0 {
		for _, e := range m.TakerFeeShareAgreements {
			l = e.Size()
			n += 1 + l + sovTakerFeeShare(uint64(l))
		}
	}
	return n
}

func sovTakerFeeShare(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTakerFeeShare(x uint64) (n int) {
	return sovTakerFeeShare(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *TakerFeeShareAgreement) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTakerFeeShare
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
			return fmt.Errorf("proto: TakerFeeShareAgreement: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TakerFeeShareAgreement: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTakerFeeShare
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
				return ErrInvalidLengthTakerFeeShare
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTakerFeeShare
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SkimPercent", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTakerFeeShare
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
				return ErrInvalidLengthTakerFeeShare
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTakerFeeShare
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.SkimPercent.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SkimAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTakerFeeShare
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
				return ErrInvalidLengthTakerFeeShare
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTakerFeeShare
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SkimAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTakerFeeShare(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTakerFeeShare
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
func (m *TakerFeeSkimAccumulator) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTakerFeeShare
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
			return fmt.Errorf("proto: TakerFeeSkimAccumulator: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TakerFeeSkimAccumulator: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTakerFeeShare
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
				return ErrInvalidLengthTakerFeeShare
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTakerFeeShare
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SkimmedTakerFees", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTakerFeeShare
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
				return ErrInvalidLengthTakerFeeShare
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTakerFeeShare
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SkimmedTakerFees = append(m.SkimmedTakerFees, types.Coin{})
			if err := m.SkimmedTakerFees[len(m.SkimmedTakerFees)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTakerFeeShare(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTakerFeeShare
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
func (m *AlloyContractTakerFeeShareState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTakerFeeShare
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
			return fmt.Errorf("proto: AlloyContractTakerFeeShareState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AlloyContractTakerFeeShareState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTakerFeeShare
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
				return ErrInvalidLengthTakerFeeShare
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTakerFeeShare
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContractAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TakerFeeShareAgreements", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTakerFeeShare
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
				return ErrInvalidLengthTakerFeeShare
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTakerFeeShare
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TakerFeeShareAgreements = append(m.TakerFeeShareAgreements, TakerFeeShareAgreement{})
			if err := m.TakerFeeShareAgreements[len(m.TakerFeeShareAgreements)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTakerFeeShare(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTakerFeeShare
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
func skipTakerFeeShare(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTakerFeeShare
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
					return 0, ErrIntOverflowTakerFeeShare
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
					return 0, ErrIntOverflowTakerFeeShare
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
				return 0, ErrInvalidLengthTakerFeeShare
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTakerFeeShare
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTakerFeeShare
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTakerFeeShare        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTakerFeeShare          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTakerFeeShare = fmt.Errorf("proto: unexpected end of group")
)
