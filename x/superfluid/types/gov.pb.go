// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/superfluid/v1beta1/gov.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	proto "github.com/cosmos/gogoproto/proto"
	_ "github.com/gogo/protobuf/gogoproto"
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

// SetSuperfluidAssetsProposal is a gov Content type to update the superfluid
// assets
type SetSuperfluidAssetsProposal struct {
	Title       string            `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description string            `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Assets      []SuperfluidAsset `protobuf:"bytes,3,rep,name=assets,proto3" json:"assets"`
}

func (m *SetSuperfluidAssetsProposal) Reset()      { *m = SetSuperfluidAssetsProposal{} }
func (*SetSuperfluidAssetsProposal) ProtoMessage() {}
func (*SetSuperfluidAssetsProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_2e37d6a8d0e42294, []int{0}
}
func (m *SetSuperfluidAssetsProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SetSuperfluidAssetsProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SetSuperfluidAssetsProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SetSuperfluidAssetsProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SetSuperfluidAssetsProposal.Merge(m, src)
}
func (m *SetSuperfluidAssetsProposal) XXX_Size() int {
	return m.Size()
}
func (m *SetSuperfluidAssetsProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_SetSuperfluidAssetsProposal.DiscardUnknown(m)
}

var xxx_messageInfo_SetSuperfluidAssetsProposal proto.InternalMessageInfo

// RemoveSuperfluidAssetsProposal is a gov Content type to remove the superfluid
// assets by denom
type RemoveSuperfluidAssetsProposal struct {
	Title                 string   `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description           string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	SuperfluidAssetDenoms []string `protobuf:"bytes,3,rep,name=superfluid_asset_denoms,json=superfluidAssetDenoms,proto3" json:"superfluid_asset_denoms,omitempty"`
}

func (m *RemoveSuperfluidAssetsProposal) Reset()      { *m = RemoveSuperfluidAssetsProposal{} }
func (*RemoveSuperfluidAssetsProposal) ProtoMessage() {}
func (*RemoveSuperfluidAssetsProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_2e37d6a8d0e42294, []int{1}
}
func (m *RemoveSuperfluidAssetsProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RemoveSuperfluidAssetsProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RemoveSuperfluidAssetsProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RemoveSuperfluidAssetsProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoveSuperfluidAssetsProposal.Merge(m, src)
}
func (m *RemoveSuperfluidAssetsProposal) XXX_Size() int {
	return m.Size()
}
func (m *RemoveSuperfluidAssetsProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoveSuperfluidAssetsProposal.DiscardUnknown(m)
}

var xxx_messageInfo_RemoveSuperfluidAssetsProposal proto.InternalMessageInfo

// UpdateUnpoolWhiteListProposal is a gov Content type to update the
// allowed list of pool ids.
type UpdateUnpoolWhiteListProposal struct {
	Title       string   `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Ids         []uint64 `protobuf:"varint,3,rep,packed,name=ids,proto3" json:"ids,omitempty"`
	IsOverwrite bool     `protobuf:"varint,4,opt,name=is_overwrite,json=isOverwrite,proto3" json:"is_overwrite,omitempty"`
}

func (m *UpdateUnpoolWhiteListProposal) Reset()      { *m = UpdateUnpoolWhiteListProposal{} }
func (*UpdateUnpoolWhiteListProposal) ProtoMessage() {}
func (*UpdateUnpoolWhiteListProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_2e37d6a8d0e42294, []int{2}
}
func (m *UpdateUnpoolWhiteListProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *UpdateUnpoolWhiteListProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_UpdateUnpoolWhiteListProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *UpdateUnpoolWhiteListProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateUnpoolWhiteListProposal.Merge(m, src)
}
func (m *UpdateUnpoolWhiteListProposal) XXX_Size() int {
	return m.Size()
}
func (m *UpdateUnpoolWhiteListProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateUnpoolWhiteListProposal.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateUnpoolWhiteListProposal proto.InternalMessageInfo

func init() {
	proto.RegisterType((*SetSuperfluidAssetsProposal)(nil), "osmosis.superfluid.v1beta1.SetSuperfluidAssetsProposal")
	proto.RegisterType((*RemoveSuperfluidAssetsProposal)(nil), "osmosis.superfluid.v1beta1.RemoveSuperfluidAssetsProposal")
	proto.RegisterType((*UpdateUnpoolWhiteListProposal)(nil), "osmosis.superfluid.v1beta1.UpdateUnpoolWhiteListProposal")
}

func init() {
	proto.RegisterFile("osmosis/superfluid/v1beta1/gov.proto", fileDescriptor_2e37d6a8d0e42294)
}

var fileDescriptor_2e37d6a8d0e42294 = []byte{
	// 464 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x93, 0xcf, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0xbd, 0x24, 0x54, 0x74, 0xc3, 0x01, 0xac, 0x22, 0x42, 0x10, 0x76, 0x68, 0x11, 0x8a,
	0x90, 0x6c, 0xd3, 0x22, 0xf5, 0xd0, 0x5b, 0x0b, 0x47, 0x04, 0x95, 0xab, 0x08, 0x89, 0x4b, 0xe4,
	0xc4, 0x43, 0xba, 0x92, 0xed, 0x59, 0x79, 0xc7, 0x2e, 0xbc, 0x01, 0xe2, 0xc4, 0x91, 0x63, 0x1e,
	0x81, 0x03, 0x0f, 0x51, 0x71, 0xea, 0x91, 0x03, 0x42, 0x28, 0x39, 0x94, 0x1b, 0xaf, 0x80, 0xbc,
	0xde, 0x34, 0x01, 0x55, 0x48, 0xfc, 0xb9, 0x58, 0x3b, 0x33, 0xdf, 0x7e, 0x9f, 0x7f, 0x23, 0x2d,
	0xbf, 0x83, 0x2a, 0x45, 0x25, 0x54, 0xa0, 0x0a, 0x09, 0xf9, 0x8b, 0xa4, 0x10, 0x71, 0x50, 0x6e,
	0x0e, 0x81, 0xa2, 0xcd, 0x60, 0x8c, 0xa5, 0x2f, 0x73, 0x24, 0xb4, 0x3b, 0x46, 0xe5, 0x2f, 0x54,
	0xbe, 0x51, 0x75, 0xd6, 0xc6, 0x38, 0x46, 0x2d, 0x0b, 0xaa, 0x53, 0x7d, 0xa3, 0x73, 0x35, 0x4a,
	0x45, 0x86, 0x81, 0xfe, 0x9a, 0xd6, 0x8d, 0x91, 0x76, 0x19, 0xd4, 0xda, 0xba, 0x30, 0xa3, 0x8d,
	0x73, 0xfe, 0x62, 0x29, 0x4a, 0x8b, 0xd6, 0xbf, 0x33, 0x7e, 0xf3, 0x00, 0xe8, 0xe0, 0xac, 0xbf,
	0xab, 0x14, 0x90, 0xda, 0xcf, 0x51, 0xa2, 0x8a, 0x12, 0x7b, 0x8d, 0x5f, 0x24, 0x41, 0x09, 0xb4,
	0x59, 0x97, 0xf5, 0x56, 0xc3, 0xba, 0xb0, 0xbb, 0xbc, 0x15, 0x83, 0x1a, 0xe5, 0x42, 0x92, 0xc0,
	0xac, 0x7d, 0x41, 0xcf, 0x96, 0x5b, 0xf6, 0x2e, 0x5f, 0x89, 0xb4, 0x53, 0xbb, 0xd1, 0x6d, 0xf4,
	0x5a, 0x5b, 0x1b, 0xfe, 0x39, 0xb4, 0xbf, 0xa4, 0xee, 0x35, 0x8f, 0xbf, 0xb8, 0x56, 0x68, 0x2e,
	0xee, 0xf4, 0x5f, 0x4f, 0x5c, 0xeb, 0xdd, 0xc4, 0xb5, 0xbe, 0x4d, 0x5c, 0xf6, 0xf1, 0x83, 0xd7,
	0x31, 0x74, 0xd5, 0x06, 0xcd, 0x9e, 0xfc, 0x87, 0x98, 0x11, 0x64, 0xf4, 0xe6, 0xf4, 0xfd, 0xbd,
	0xbb, 0x67, 0xb8, 0x40, 0xde, 0x22, 0xc4, 0xab, 0xdd, 0x3c, 0x69, 0x88, 0xd6, 0x4f, 0x19, 0x77,
	0x42, 0x48, 0xb1, 0x84, 0xff, 0x0e, 0xbd, 0xcd, 0xaf, 0x2f, 0x82, 0x07, 0x3a, 0x78, 0x10, 0x43,
	0x86, 0x69, 0xbd, 0x85, 0xd5, 0xf0, 0x9a, 0xfa, 0x39, 0xf2, 0x91, 0x1e, 0xfe, 0x35, 0x69, 0x0c,
	0xc9, 0xef, 0x48, 0x3f, 0x33, 0x7e, 0xab, 0x2f, 0xe3, 0x88, 0xa0, 0x9f, 0x49, 0xc4, 0xe4, 0xd9,
	0xa1, 0x20, 0x78, 0x2c, 0x14, 0xfd, 0x33, 0xe8, 0x15, 0xde, 0x10, 0x71, 0x0d, 0xd5, 0x0c, 0xab,
	0xa3, 0x7d, 0x9b, 0x5f, 0x16, 0x6a, 0x80, 0x25, 0xe4, 0x47, 0xb9, 0x20, 0x68, 0x37, 0xbb, 0xac,
	0x77, 0x29, 0x6c, 0x09, 0xf5, 0x74, 0xde, 0xda, 0x79, 0xf2, 0x67, 0x94, 0xee, 0x9c, 0xb2, 0xd0,
	0x08, 0x5e, 0xa1, 0x19, 0xbc, 0xa3, 0x0a, 0x22, 0x11, 0x8a, 0xf6, 0xf6, 0x8f, 0xa7, 0x0e, 0x3b,
	0x99, 0x3a, 0xec, 0xeb, 0xd4, 0x61, 0x6f, 0x67, 0x8e, 0x75, 0x32, 0x73, 0xac, 0x4f, 0x33, 0xc7,
	0x7a, 0xbe, 0x3d, 0x16, 0x74, 0x58, 0x0c, 0xfd, 0x11, 0xa6, 0x81, 0x71, 0xf1, 0x92, 0x68, 0xa8,
	0xe6, 0x45, 0x50, 0x6e, 0xdd, 0x0f, 0x5e, 0x2e, 0xbf, 0x0b, 0x7a, 0x25, 0x41, 0x0d, 0x57, 0xf4,
	0x9b, 0x78, 0xf0, 0x23, 0x00, 0x00, 0xff, 0xff, 0x14, 0x6f, 0xa7, 0x44, 0xc0, 0x03, 0x00, 0x00,
}

func (this *SetSuperfluidAssetsProposal) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*SetSuperfluidAssetsProposal)
	if !ok {
		that2, ok := that.(SetSuperfluidAssetsProposal)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Title != that1.Title {
		return false
	}
	if this.Description != that1.Description {
		return false
	}
	if len(this.Assets) != len(that1.Assets) {
		return false
	}
	for i := range this.Assets {
		if !this.Assets[i].Equal(&that1.Assets[i]) {
			return false
		}
	}
	return true
}
func (this *RemoveSuperfluidAssetsProposal) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RemoveSuperfluidAssetsProposal)
	if !ok {
		that2, ok := that.(RemoveSuperfluidAssetsProposal)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Title != that1.Title {
		return false
	}
	if this.Description != that1.Description {
		return false
	}
	if len(this.SuperfluidAssetDenoms) != len(that1.SuperfluidAssetDenoms) {
		return false
	}
	for i := range this.SuperfluidAssetDenoms {
		if this.SuperfluidAssetDenoms[i] != that1.SuperfluidAssetDenoms[i] {
			return false
		}
	}
	return true
}
func (this *UpdateUnpoolWhiteListProposal) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*UpdateUnpoolWhiteListProposal)
	if !ok {
		that2, ok := that.(UpdateUnpoolWhiteListProposal)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Title != that1.Title {
		return false
	}
	if this.Description != that1.Description {
		return false
	}
	if len(this.Ids) != len(that1.Ids) {
		return false
	}
	for i := range this.Ids {
		if this.Ids[i] != that1.Ids[i] {
			return false
		}
	}
	if this.IsOverwrite != that1.IsOverwrite {
		return false
	}
	return true
}
func (m *SetSuperfluidAssetsProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SetSuperfluidAssetsProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SetSuperfluidAssetsProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Assets) > 0 {
		for iNdEx := len(m.Assets) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Assets[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGov(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintGov(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintGov(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *RemoveSuperfluidAssetsProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RemoveSuperfluidAssetsProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RemoveSuperfluidAssetsProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SuperfluidAssetDenoms) > 0 {
		for iNdEx := len(m.SuperfluidAssetDenoms) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.SuperfluidAssetDenoms[iNdEx])
			copy(dAtA[i:], m.SuperfluidAssetDenoms[iNdEx])
			i = encodeVarintGov(dAtA, i, uint64(len(m.SuperfluidAssetDenoms[iNdEx])))
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintGov(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintGov(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *UpdateUnpoolWhiteListProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *UpdateUnpoolWhiteListProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *UpdateUnpoolWhiteListProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.IsOverwrite {
		i--
		if m.IsOverwrite {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x20
	}
	if len(m.Ids) > 0 {
		dAtA2 := make([]byte, len(m.Ids)*10)
		var j1 int
		for _, num := range m.Ids {
			for num >= 1<<7 {
				dAtA2[j1] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j1++
			}
			dAtA2[j1] = uint8(num)
			j1++
		}
		i -= j1
		copy(dAtA[i:], dAtA2[:j1])
		i = encodeVarintGov(dAtA, i, uint64(j1))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintGov(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintGov(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGov(dAtA []byte, offset int, v uint64) int {
	offset -= sovGov(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *SetSuperfluidAssetsProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovGov(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovGov(uint64(l))
	}
	if len(m.Assets) > 0 {
		for _, e := range m.Assets {
			l = e.Size()
			n += 1 + l + sovGov(uint64(l))
		}
	}
	return n
}

func (m *RemoveSuperfluidAssetsProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovGov(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovGov(uint64(l))
	}
	if len(m.SuperfluidAssetDenoms) > 0 {
		for _, s := range m.SuperfluidAssetDenoms {
			l = len(s)
			n += 1 + l + sovGov(uint64(l))
		}
	}
	return n
}

func (m *UpdateUnpoolWhiteListProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovGov(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovGov(uint64(l))
	}
	if len(m.Ids) > 0 {
		l = 0
		for _, e := range m.Ids {
			l += sovGov(uint64(e))
		}
		n += 1 + sovGov(uint64(l)) + l
	}
	if m.IsOverwrite {
		n += 2
	}
	return n
}

func sovGov(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGov(x uint64) (n int) {
	return sovGov(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *SetSuperfluidAssetsProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGov
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
			return fmt.Errorf("proto: SetSuperfluidAssetsProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SetSuperfluidAssetsProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGov
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
				return ErrInvalidLengthGov
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGov
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGov
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
				return ErrInvalidLengthGov
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGov
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Assets", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGov
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
				return ErrInvalidLengthGov
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGov
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Assets = append(m.Assets, SuperfluidAsset{})
			if err := m.Assets[len(m.Assets)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGov(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGov
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
func (m *RemoveSuperfluidAssetsProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGov
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
			return fmt.Errorf("proto: RemoveSuperfluidAssetsProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RemoveSuperfluidAssetsProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGov
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
				return ErrInvalidLengthGov
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGov
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGov
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
				return ErrInvalidLengthGov
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGov
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SuperfluidAssetDenoms", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGov
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
				return ErrInvalidLengthGov
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGov
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SuperfluidAssetDenoms = append(m.SuperfluidAssetDenoms, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGov(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGov
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
func (m *UpdateUnpoolWhiteListProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGov
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
			return fmt.Errorf("proto: UpdateUnpoolWhiteListProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: UpdateUnpoolWhiteListProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGov
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
				return ErrInvalidLengthGov
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGov
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGov
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
				return ErrInvalidLengthGov
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGov
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType == 0 {
				var v uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGov
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
				m.Ids = append(m.Ids, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGov
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
					return ErrInvalidLengthGov
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthGov
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
				if elementCount != 0 && len(m.Ids) == 0 {
					m.Ids = make([]uint64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowGov
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
					m.Ids = append(m.Ids, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Ids", wireType)
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IsOverwrite", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGov
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.IsOverwrite = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipGov(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGov
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
func skipGov(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGov
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
					return 0, ErrIntOverflowGov
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
					return 0, ErrIntOverflowGov
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
				return 0, ErrInvalidLengthGov
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGov
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGov
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGov        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGov          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGov = fmt.Errorf("proto: unexpected end of group")
)
