// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/cosmwasmpool/v1beta1/model/v3/pool_query_msg.proto

package v3

import (
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

// ===================== ShareDenomResponse
type ShareDenomResponse struct {
	// share_denom is the share denomination.
	ShareDenom string `protobuf:"bytes,1,opt,name=share_denom,json=shareDenom,proto3" json:"share_denom,omitempty"`
}

func (m *ShareDenomResponse) Reset()         { *m = ShareDenomResponse{} }
func (m *ShareDenomResponse) String() string { return proto.CompactTextString(m) }
func (*ShareDenomResponse) ProtoMessage()    {}
func (*ShareDenomResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_006bf70430f78019, []int{0}
}
func (m *ShareDenomResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ShareDenomResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ShareDenomResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ShareDenomResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ShareDenomResponse.Merge(m, src)
}
func (m *ShareDenomResponse) XXX_Size() int {
	return m.Size()
}
func (m *ShareDenomResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ShareDenomResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ShareDenomResponse proto.InternalMessageInfo

func (m *ShareDenomResponse) GetShareDenom() string {
	if m != nil {
		return m.ShareDenom
	}
	return ""
}

// ===================== TotalPoolLiquidityResponse
type TotalPoolLiquidityResponse struct {
	// total_pool_liquidity is the total liquidity in the pool denominated in
	// coins.
	TotalPoolLiquidity github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,1,rep,name=total_pool_liquidity,json=totalPoolLiquidity,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"total_pool_liquidity"`
}

func (m *TotalPoolLiquidityResponse) Reset()         { *m = TotalPoolLiquidityResponse{} }
func (m *TotalPoolLiquidityResponse) String() string { return proto.CompactTextString(m) }
func (*TotalPoolLiquidityResponse) ProtoMessage()    {}
func (*TotalPoolLiquidityResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_006bf70430f78019, []int{1}
}
func (m *TotalPoolLiquidityResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TotalPoolLiquidityResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TotalPoolLiquidityResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TotalPoolLiquidityResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TotalPoolLiquidityResponse.Merge(m, src)
}
func (m *TotalPoolLiquidityResponse) XXX_Size() int {
	return m.Size()
}
func (m *TotalPoolLiquidityResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_TotalPoolLiquidityResponse.DiscardUnknown(m)
}

var xxx_messageInfo_TotalPoolLiquidityResponse proto.InternalMessageInfo

func (m *TotalPoolLiquidityResponse) GetTotalPoolLiquidity() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.TotalPoolLiquidity
	}
	return nil
}

// ===================== AssetConfig
type AssetConfig struct {
	// denom is the asset denomination.
	Denom string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	// normalization_factor is the normalization factor for the asset.
	NormalizationFactor string `protobuf:"bytes,2,opt,name=normalization_factor,json=normalizationFactor,proto3" json:"normalization_factor,omitempty"`
}

func (m *AssetConfig) Reset()         { *m = AssetConfig{} }
func (m *AssetConfig) String() string { return proto.CompactTextString(m) }
func (*AssetConfig) ProtoMessage()    {}
func (*AssetConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_006bf70430f78019, []int{2}
}
func (m *AssetConfig) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AssetConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AssetConfig.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *AssetConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AssetConfig.Merge(m, src)
}
func (m *AssetConfig) XXX_Size() int {
	return m.Size()
}
func (m *AssetConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_AssetConfig.DiscardUnknown(m)
}

var xxx_messageInfo_AssetConfig proto.InternalMessageInfo

func (m *AssetConfig) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

func (m *AssetConfig) GetNormalizationFactor() string {
	if m != nil {
		return m.NormalizationFactor
	}
	return ""
}

// ===================== ListAssetConfigsResponse
type ListAssetConfigsResponse struct {
	// asset_configs is the list of asset configurations.
	AssetConfigs []AssetConfig `protobuf:"bytes,1,rep,name=asset_configs,json=assetConfigs,proto3" json:"asset_configs"`
}

func (m *ListAssetConfigsResponse) Reset()         { *m = ListAssetConfigsResponse{} }
func (m *ListAssetConfigsResponse) String() string { return proto.CompactTextString(m) }
func (*ListAssetConfigsResponse) ProtoMessage()    {}
func (*ListAssetConfigsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_006bf70430f78019, []int{3}
}
func (m *ListAssetConfigsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ListAssetConfigsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ListAssetConfigsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ListAssetConfigsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListAssetConfigsResponse.Merge(m, src)
}
func (m *ListAssetConfigsResponse) XXX_Size() int {
	return m.Size()
}
func (m *ListAssetConfigsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListAssetConfigsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListAssetConfigsResponse proto.InternalMessageInfo

func (m *ListAssetConfigsResponse) GetAssetConfigs() []AssetConfig {
	if m != nil {
		return m.AssetConfigs
	}
	return nil
}

func init() {
	proto.RegisterType((*ShareDenomResponse)(nil), "osmosis.cosmwasmpool.v1beta1.model.v3.ShareDenomResponse")
	proto.RegisterType((*TotalPoolLiquidityResponse)(nil), "osmosis.cosmwasmpool.v1beta1.model.v3.TotalPoolLiquidityResponse")
	proto.RegisterType((*AssetConfig)(nil), "osmosis.cosmwasmpool.v1beta1.model.v3.AssetConfig")
	proto.RegisterType((*ListAssetConfigsResponse)(nil), "osmosis.cosmwasmpool.v1beta1.model.v3.ListAssetConfigsResponse")
}

func init() {
	proto.RegisterFile("osmosis/cosmwasmpool/v1beta1/model/v3/pool_query_msg.proto", fileDescriptor_006bf70430f78019)
}

var fileDescriptor_006bf70430f78019 = []byte{
	// 419 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xbf, 0x8e, 0xd4, 0x30,
	0x10, 0xc6, 0x37, 0xfc, 0x93, 0xf0, 0x42, 0x13, 0xb6, 0x58, 0xb6, 0xc8, 0x9e, 0x22, 0x21, 0x6d,
	0x73, 0x36, 0xbb, 0x11, 0x14, 0x74, 0xe4, 0x10, 0xd5, 0x15, 0x28, 0x20, 0x0a, 0x04, 0x8a, 0x9c,
	0xc4, 0x97, 0xb3, 0x88, 0x33, 0xb9, 0x8c, 0x37, 0x10, 0x24, 0xde, 0x81, 0x9e, 0x37, 0xe0, 0x49,
	0xae, 0xbc, 0x92, 0x0a, 0xd0, 0xee, 0x8b, 0x20, 0x3b, 0xd9, 0x28, 0x27, 0x1a, 0xaa, 0x64, 0xbe,
	0xcf, 0x3f, 0xcf, 0x7c, 0xb6, 0xc9, 0x33, 0x40, 0x05, 0x28, 0x91, 0xa5, 0x80, 0xea, 0x13, 0x47,
	0x55, 0x01, 0x14, 0xac, 0x59, 0x27, 0x42, 0xf3, 0x35, 0x53, 0x90, 0x89, 0x82, 0x35, 0x01, 0x33,
	0x6a, 0x7c, 0xb1, 0x15, 0x75, 0x1b, 0x2b, 0xcc, 0x69, 0x55, 0x83, 0x06, 0xf7, 0x51, 0xcf, 0xd2,
	0x31, 0x4b, 0x7b, 0x96, 0x5a, 0x96, 0x36, 0xc1, 0x62, 0x96, 0x43, 0x0e, 0x96, 0x60, 0xe6, 0xaf,
	0x83, 0x17, 0x5e, 0x6a, 0x69, 0x96, 0x70, 0x14, 0x43, 0xbf, 0x14, 0x64, 0xd9, 0xf9, 0xfe, 0x13,
	0xe2, 0xbe, 0x3e, 0xe7, 0xb5, 0x78, 0x21, 0x4a, 0x50, 0x91, 0xc0, 0x0a, 0x4a, 0x14, 0xee, 0x92,
	0x4c, 0xd1, 0xa8, 0x71, 0x66, 0xe4, 0xb9, 0x73, 0xe4, 0xac, 0xee, 0x46, 0x04, 0x87, 0x85, 0xfe,
	0x77, 0x87, 0x2c, 0xde, 0x80, 0xe6, 0xc5, 0x2b, 0x80, 0xe2, 0x54, 0x5e, 0x6c, 0x65, 0x26, 0x75,
	0x3b, 0xf0, 0x5f, 0xc9, 0x4c, 0x1b, 0x37, 0xb6, 0x81, 0x8a, 0x83, 0x3f, 0x77, 0x8e, 0x6e, 0xae,
	0xa6, 0x9b, 0x87, 0xb4, 0x1b, 0x8a, 0x9a, 0xa1, 0x86, 0x20, 0x27, 0x20, 0xcb, 0xf0, 0xf1, 0xe5,
	0xaf, 0xe5, 0xe4, 0xc7, 0xef, 0xe5, 0x2a, 0x97, 0xfa, 0x7c, 0x9b, 0xd0, 0x14, 0x14, 0xeb, 0x13,
	0x74, 0x9f, 0x63, 0xcc, 0x3e, 0x32, 0xdd, 0x56, 0x02, 0x2d, 0x80, 0x91, 0xab, 0xff, 0x19, 0xc3,
	0x7f, 0x4b, 0xa6, 0xcf, 0x11, 0x85, 0x3e, 0x81, 0xf2, 0x4c, 0xe6, 0xee, 0x8c, 0xdc, 0x1e, 0xe7,
	0xe8, 0x0a, 0x77, 0x4d, 0x66, 0x25, 0xd4, 0x8a, 0x17, 0xf2, 0x0b, 0xd7, 0x12, 0xca, 0xf8, 0x8c,
	0xa7, 0x1a, 0xea, 0xf9, 0x0d, 0xbb, 0xe8, 0xc1, 0x35, 0xef, 0xa5, 0xb5, 0xfc, 0x96, 0xcc, 0x4f,
	0x25, 0xea, 0xd1, 0xde, 0x38, 0x44, 0xfe, 0x40, 0xee, 0x73, 0xa3, 0xc7, 0x69, 0x67, 0xf4, 0x59,
	0x37, 0xf4, 0xbf, 0x6e, 0x8f, 0x8e, 0xf6, 0x0c, 0x6f, 0x99, 0x43, 0x88, 0xee, 0xf1, 0x51, 0x9b,
	0xf0, 0xfd, 0xe5, 0xce, 0x73, 0xae, 0x76, 0x9e, 0xf3, 0x67, 0xe7, 0x39, 0xdf, 0xf6, 0xde, 0xe4,
	0x6a, 0xef, 0x4d, 0x7e, 0xee, 0xbd, 0xc9, 0xbb, 0x70, 0x74, 0x54, 0x7d, 0xaf, 0xe3, 0x82, 0x27,
	0x78, 0x28, 0x58, 0xb3, 0x79, 0xca, 0x3e, 0x5f, 0x7f, 0x78, 0x87, 0x82, 0x29, 0xcc, 0x59, 0x13,
	0x24, 0x77, 0xec, 0x63, 0x08, 0xfe, 0x06, 0x00, 0x00, 0xff, 0xff, 0xf2, 0x0d, 0x17, 0xe1, 0xa7,
	0x02, 0x00, 0x00,
}

func (m *ShareDenomResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ShareDenomResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ShareDenomResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ShareDenom) > 0 {
		i -= len(m.ShareDenom)
		copy(dAtA[i:], m.ShareDenom)
		i = encodeVarintPoolQueryMsg(dAtA, i, uint64(len(m.ShareDenom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *TotalPoolLiquidityResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TotalPoolLiquidityResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TotalPoolLiquidityResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.TotalPoolLiquidity) > 0 {
		for iNdEx := len(m.TotalPoolLiquidity) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TotalPoolLiquidity[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintPoolQueryMsg(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *AssetConfig) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AssetConfig) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *AssetConfig) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.NormalizationFactor) > 0 {
		i -= len(m.NormalizationFactor)
		copy(dAtA[i:], m.NormalizationFactor)
		i = encodeVarintPoolQueryMsg(dAtA, i, uint64(len(m.NormalizationFactor)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintPoolQueryMsg(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *ListAssetConfigsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ListAssetConfigsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ListAssetConfigsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AssetConfigs) > 0 {
		for iNdEx := len(m.AssetConfigs) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.AssetConfigs[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintPoolQueryMsg(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintPoolQueryMsg(dAtA []byte, offset int, v uint64) int {
	offset -= sovPoolQueryMsg(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ShareDenomResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ShareDenom)
	if l > 0 {
		n += 1 + l + sovPoolQueryMsg(uint64(l))
	}
	return n
}

func (m *TotalPoolLiquidityResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.TotalPoolLiquidity) > 0 {
		for _, e := range m.TotalPoolLiquidity {
			l = e.Size()
			n += 1 + l + sovPoolQueryMsg(uint64(l))
		}
	}
	return n
}

func (m *AssetConfig) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovPoolQueryMsg(uint64(l))
	}
	l = len(m.NormalizationFactor)
	if l > 0 {
		n += 1 + l + sovPoolQueryMsg(uint64(l))
	}
	return n
}

func (m *ListAssetConfigsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.AssetConfigs) > 0 {
		for _, e := range m.AssetConfigs {
			l = e.Size()
			n += 1 + l + sovPoolQueryMsg(uint64(l))
		}
	}
	return n
}

func sovPoolQueryMsg(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPoolQueryMsg(x uint64) (n int) {
	return sovPoolQueryMsg(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ShareDenomResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPoolQueryMsg
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
			return fmt.Errorf("proto: ShareDenomResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ShareDenomResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ShareDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPoolQueryMsg
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
				return ErrInvalidLengthPoolQueryMsg
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPoolQueryMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ShareDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPoolQueryMsg(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPoolQueryMsg
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
func (m *TotalPoolLiquidityResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPoolQueryMsg
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
			return fmt.Errorf("proto: TotalPoolLiquidityResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TotalPoolLiquidityResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalPoolLiquidity", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPoolQueryMsg
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
				return ErrInvalidLengthPoolQueryMsg
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPoolQueryMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TotalPoolLiquidity = append(m.TotalPoolLiquidity, types.Coin{})
			if err := m.TotalPoolLiquidity[len(m.TotalPoolLiquidity)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPoolQueryMsg(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPoolQueryMsg
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
func (m *AssetConfig) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPoolQueryMsg
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
			return fmt.Errorf("proto: AssetConfig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AssetConfig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPoolQueryMsg
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
				return ErrInvalidLengthPoolQueryMsg
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPoolQueryMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NormalizationFactor", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPoolQueryMsg
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
				return ErrInvalidLengthPoolQueryMsg
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPoolQueryMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.NormalizationFactor = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPoolQueryMsg(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPoolQueryMsg
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
func (m *ListAssetConfigsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPoolQueryMsg
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
			return fmt.Errorf("proto: ListAssetConfigsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ListAssetConfigsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AssetConfigs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPoolQueryMsg
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
				return ErrInvalidLengthPoolQueryMsg
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPoolQueryMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AssetConfigs = append(m.AssetConfigs, AssetConfig{})
			if err := m.AssetConfigs[len(m.AssetConfigs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPoolQueryMsg(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPoolQueryMsg
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
func skipPoolQueryMsg(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPoolQueryMsg
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
					return 0, ErrIntOverflowPoolQueryMsg
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
					return 0, ErrIntOverflowPoolQueryMsg
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
				return 0, ErrInvalidLengthPoolQueryMsg
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPoolQueryMsg
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPoolQueryMsg
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPoolQueryMsg        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPoolQueryMsg          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPoolQueryMsg = fmt.Errorf("proto: unexpected end of group")
)