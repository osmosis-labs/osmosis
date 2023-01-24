// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/concentrated-liquidity/pool.proto

// This is a legacy package that requires additional migration logic
// in order to use the correct package. Decision made to use legacy package path
// until clear steps for migration logic and the unknowns for state breaking are
// investigated for changing proto package.

package model

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
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

type Pool struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty" yaml:"address"`
	Id      uint64 `protobuf:"varint,2,opt,name=id,proto3" json:"id,omitempty"`
	// Amount of total liquidity
	Liquidity        github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,3,opt,name=liquidity,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"liquidity" yaml:"liquidity"`
	Token0           string                                 `protobuf:"bytes,4,opt,name=token0,proto3" json:"token0,omitempty"`
	Token1           string                                 `protobuf:"bytes,5,opt,name=token1,proto3" json:"token1,omitempty"`
	CurrentSqrtPrice github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,6,opt,name=current_sqrt_price,json=currentSqrtPrice,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"current_sqrt_price" yaml:"spot_price"`
	CurrentTick      github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,7,opt,name=current_tick,json=currentTick,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"current_tick" yaml:"current_tick"`
	// tick_spacing must be one of the authorized_tick_spacing values set in the
	// concentrated-liquidity parameters
	TickSpacing               uint64                                 `protobuf:"varint,8,opt,name=tick_spacing,json=tickSpacing,proto3" json:"tick_spacing,omitempty" yaml:"tick_spacing"`
	PrecisionFactorAtPriceOne github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,9,opt,name=precision_factor_at_price_one,json=precisionFactorAtPriceOne,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"precision_factor_at_price_one" yaml:"precision_factor_at_price_one"`
	// swap_fee is the ratio that is charged on the amount of token in.
	SwapFee github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,10,opt,name=swap_fee,json=swapFee,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"swap_fee" yaml:"swap_fee"`
}

func (m *Pool) Reset()      { *m = Pool{} }
func (*Pool) ProtoMessage() {}
func (*Pool) Descriptor() ([]byte, []int) {
	return fileDescriptor_3526ea5373d96c9a, []int{0}
}
func (m *Pool) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Pool) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Pool.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Pool) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Pool.Merge(m, src)
}
func (m *Pool) XXX_Size() int {
	return m.Size()
}
func (m *Pool) XXX_DiscardUnknown() {
	xxx_messageInfo_Pool.DiscardUnknown(m)
}

var xxx_messageInfo_Pool proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Pool)(nil), "osmosis.concentratedliquidity.v1beta1.Pool")
}

func init() {
	proto.RegisterFile("osmosis/concentrated-liquidity/pool.proto", fileDescriptor_3526ea5373d96c9a)
}

var fileDescriptor_3526ea5373d96c9a = []byte{
	// 522 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0x3f, 0x6f, 0xd3, 0x40,
	0x1c, 0xb5, 0x43, 0xfe, 0x34, 0xd7, 0xaa, 0xb4, 0x07, 0x82, 0x6b, 0x25, 0xec, 0xca, 0x02, 0x54,
	0x24, 0x12, 0x13, 0xc1, 0x94, 0xad, 0x01, 0x2a, 0x75, 0xa2, 0x72, 0x11, 0x03, 0xaa, 0x64, 0x9c,
	0xf3, 0x35, 0x3d, 0xc5, 0xf1, 0x39, 0x77, 0x97, 0x42, 0xbe, 0x01, 0x23, 0x63, 0xc7, 0x7e, 0x08,
	0x3e, 0x44, 0xc5, 0xd4, 0x11, 0x31, 0x58, 0x28, 0xf9, 0x06, 0x59, 0x59, 0x90, 0x7d, 0x17, 0xd7,
	0x0b, 0x48, 0x99, 0x7c, 0xef, 0xfd, 0x7e, 0xf7, 0xde, 0xb3, 0x7f, 0xfe, 0x81, 0x67, 0x4c, 0x8c,
	0x98, 0xa0, 0xc2, 0xc5, 0x2c, 0xc6, 0x24, 0x96, 0x3c, 0x90, 0x24, 0x6c, 0x45, 0x74, 0x3c, 0xa1,
	0x21, 0x95, 0x53, 0x37, 0x61, 0x2c, 0x6a, 0x27, 0x9c, 0x49, 0x06, 0x9f, 0xe8, 0xd6, 0x76, 0xb9,
	0xb5, 0xe8, 0x6c, 0x5f, 0x74, 0xfa, 0x44, 0x06, 0x9d, 0xdd, 0x1d, 0x9c, 0xf7, 0xf9, 0xf9, 0x25,
	0x57, 0x01, 0xa5, 0xb0, 0x7b, 0x7f, 0xc0, 0x06, 0x4c, 0xf1, 0xd9, 0x49, 0xb1, 0xce, 0x9f, 0x1a,
	0xa8, 0x1e, 0x33, 0x16, 0xc1, 0xe7, 0xa0, 0x11, 0x84, 0x21, 0x27, 0x42, 0x20, 0x73, 0xcf, 0xdc,
	0x6f, 0xf6, 0xe0, 0x22, 0xb5, 0x37, 0xa7, 0xc1, 0x28, 0xea, 0x3a, 0xba, 0xe0, 0x78, 0xcb, 0x16,
	0xb8, 0x09, 0x2a, 0x34, 0x44, 0x95, 0x3d, 0x73, 0xbf, 0xea, 0x55, 0x68, 0x08, 0x3f, 0x81, 0x66,
	0x11, 0x06, 0xdd, 0xc9, 0xef, 0xf7, 0xae, 0x53, 0xdb, 0xf8, 0x95, 0xda, 0x4f, 0x07, 0x54, 0x9e,
	0x4f, 0xfa, 0x6d, 0xcc, 0x46, 0x3a, 0x90, 0x7e, 0xb4, 0x44, 0x38, 0x74, 0xe5, 0x34, 0x21, 0xa2,
	0xfd, 0x86, 0xe0, 0x45, 0x6a, 0x6f, 0x29, 0xb7, 0x42, 0xc8, 0xf1, 0x6e, 0x45, 0xe1, 0x03, 0x50,
	0x97, 0x6c, 0x48, 0xe2, 0x17, 0xa8, 0x9a, 0xc9, 0x7b, 0x1a, 0x15, 0x7c, 0x07, 0xd5, 0x4a, 0x7c,
	0x07, 0x8e, 0x01, 0xc4, 0x13, 0xce, 0x49, 0x2c, 0x7d, 0x31, 0xe6, 0xd2, 0x4f, 0x38, 0xc5, 0x04,
	0xd5, 0xf3, 0x68, 0xaf, 0x57, 0x8e, 0xb6, 0xad, 0xa2, 0x89, 0x84, 0x69, 0x25, 0xc7, 0xdb, 0xd2,
	0xf2, 0x27, 0x63, 0x2e, 0x8f, 0x33, 0x0a, 0x9e, 0x83, 0x8d, 0xa5, 0xa5, 0xa4, 0x78, 0x88, 0x1a,
	0xb9, 0xd9, 0xdb, 0x15, 0xcc, 0x8e, 0x62, 0xb9, 0x48, 0xed, 0x7b, 0xca, 0xac, 0xac, 0xe5, 0x78,
	0xeb, 0x1a, 0xbe, 0xa7, 0x78, 0x08, 0xbb, 0x60, 0x23, 0x63, 0x7d, 0x91, 0x04, 0x98, 0xc6, 0x03,
	0xb4, 0x96, 0x0d, 0xa2, 0xf7, 0xf0, 0xf6, 0x6e, 0xb9, 0xea, 0x78, 0xeb, 0x19, 0x3c, 0x51, 0x08,
	0x5e, 0x9a, 0xe0, 0x51, 0xc2, 0x09, 0xa6, 0x82, 0xb2, 0xd8, 0x3f, 0x0b, 0xb0, 0x64, 0xdc, 0x0f,
	0xf4, 0x6b, 0xf9, 0x2c, 0x26, 0xa8, 0x99, 0xe7, 0xfe, 0xb0, 0x72, 0xee, 0xc7, 0xca, 0xfb, 0xbf,
	0xe2, 0x8e, 0xb7, 0x53, 0xd4, 0x0f, 0xf3, 0xf2, 0x81, 0xfa, 0x7a, 0xef, 0x62, 0x02, 0x4f, 0xc1,
	0x9a, 0xf8, 0x1c, 0x24, 0xfe, 0x19, 0x21, 0x08, 0xe4, 0x21, 0x0e, 0x56, 0x9e, 0xd4, 0x5d, 0x3d,
	0x29, 0xad, 0xe3, 0x78, 0x8d, 0xec, 0x78, 0x48, 0x48, 0x77, 0xfb, 0xeb, 0x95, 0x6d, 0x5c, 0x5e,
	0xd9, 0xc6, 0x8f, 0xef, 0xad, 0x5a, 0xf6, 0xcf, 0x1f, 0xf5, 0x4e, 0xaf, 0x67, 0x96, 0x79, 0x33,
	0xb3, 0xcc, 0xdf, 0x33, 0xcb, 0xfc, 0x36, 0xb7, 0x8c, 0x9b, 0xb9, 0x65, 0xfc, 0x9c, 0x5b, 0xc6,
	0xc7, 0x5e, 0xc9, 0x50, 0xaf, 0x5e, 0x2b, 0x0a, 0xfa, 0x62, 0x09, 0xdc, 0x8b, 0xce, 0x2b, 0xf7,
	0xcb, 0xbf, 0x16, 0x77, 0xc4, 0x42, 0x12, 0xf5, 0xeb, 0xf9, 0x8a, 0xbd, 0xfc, 0x1b, 0x00, 0x00,
	0xff, 0xff, 0xaf, 0x6b, 0xaf, 0x25, 0xe7, 0x03, 0x00, 0x00,
}

func (m *Pool) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Pool) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Pool) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.SwapFee.Size()
		i -= size
		if _, err := m.SwapFee.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintPool(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x52
	{
		size := m.PrecisionFactorAtPriceOne.Size()
		i -= size
		if _, err := m.PrecisionFactorAtPriceOne.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintPool(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x4a
	if m.TickSpacing != 0 {
		i = encodeVarintPool(dAtA, i, uint64(m.TickSpacing))
		i--
		dAtA[i] = 0x40
	}
	{
		size := m.CurrentTick.Size()
		i -= size
		if _, err := m.CurrentTick.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintPool(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x3a
	{
		size := m.CurrentSqrtPrice.Size()
		i -= size
		if _, err := m.CurrentSqrtPrice.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintPool(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x32
	if len(m.Token1) > 0 {
		i -= len(m.Token1)
		copy(dAtA[i:], m.Token1)
		i = encodeVarintPool(dAtA, i, uint64(len(m.Token1)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Token0) > 0 {
		i -= len(m.Token0)
		copy(dAtA[i:], m.Token0)
		i = encodeVarintPool(dAtA, i, uint64(len(m.Token0)))
		i--
		dAtA[i] = 0x22
	}
	{
		size := m.Liquidity.Size()
		i -= size
		if _, err := m.Liquidity.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintPool(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if m.Id != 0 {
		i = encodeVarintPool(dAtA, i, uint64(m.Id))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintPool(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintPool(dAtA []byte, offset int, v uint64) int {
	offset -= sovPool(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Pool) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovPool(uint64(l))
	}
	if m.Id != 0 {
		n += 1 + sovPool(uint64(m.Id))
	}
	l = m.Liquidity.Size()
	n += 1 + l + sovPool(uint64(l))
	l = len(m.Token0)
	if l > 0 {
		n += 1 + l + sovPool(uint64(l))
	}
	l = len(m.Token1)
	if l > 0 {
		n += 1 + l + sovPool(uint64(l))
	}
	l = m.CurrentSqrtPrice.Size()
	n += 1 + l + sovPool(uint64(l))
	l = m.CurrentTick.Size()
	n += 1 + l + sovPool(uint64(l))
	if m.TickSpacing != 0 {
		n += 1 + sovPool(uint64(m.TickSpacing))
	}
	l = m.PrecisionFactorAtPriceOne.Size()
	n += 1 + l + sovPool(uint64(l))
	l = m.SwapFee.Size()
	n += 1 + l + sovPool(uint64(l))
	return n
}

func sovPool(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPool(x uint64) (n int) {
	return sovPool(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Pool) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPool
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
			return fmt.Errorf("proto: Pool: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Pool: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
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
				return ErrInvalidLengthPool
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPool
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Liquidity", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
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
				return ErrInvalidLengthPool
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPool
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Liquidity.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Token0", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
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
				return ErrInvalidLengthPool
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPool
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Token0 = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Token1", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
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
				return ErrInvalidLengthPool
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPool
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Token1 = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CurrentSqrtPrice", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
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
				return ErrInvalidLengthPool
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPool
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.CurrentSqrtPrice.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CurrentTick", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
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
				return ErrInvalidLengthPool
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPool
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.CurrentTick.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TickSpacing", wireType)
			}
			m.TickSpacing = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
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
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PrecisionFactorAtPriceOne", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
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
				return ErrInvalidLengthPool
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPool
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.PrecisionFactorAtPriceOne.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SwapFee", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
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
				return ErrInvalidLengthPool
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPool
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.SwapFee.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPool(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPool
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
func skipPool(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPool
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
					return 0, ErrIntOverflowPool
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
					return 0, ErrIntOverflowPool
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
				return 0, ErrInvalidLengthPool
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPool
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPool
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPool        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPool          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPool = fmt.Errorf("proto: unexpected end of group")
)
