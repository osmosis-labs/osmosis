// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/gamm/v1beta1/shared.proto

package migration

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/cosmos/cosmos-sdk/types"
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

// MigrationRecords contains all the links between balancer and concentrated
// pools
type MigrationRecords struct {
	BalancerToConcentratedPoolLinks []BalancerToConcentratedPoolLink `protobuf:"bytes,1,rep,name=balancer_to_concentrated_pool_links,json=balancerToConcentratedPoolLinks,proto3" json:"balancer_to_concentrated_pool_links"`
}

func (m *MigrationRecords) Reset()         { *m = MigrationRecords{} }
func (m *MigrationRecords) String() string { return proto.CompactTextString(m) }
func (*MigrationRecords) ProtoMessage()    {}
func (*MigrationRecords) Descriptor() ([]byte, []int) {
	return fileDescriptor_ecfa270aa4b2c075, []int{0}
}
func (m *MigrationRecords) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MigrationRecords) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MigrationRecords.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MigrationRecords) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MigrationRecords.Merge(m, src)
}
func (m *MigrationRecords) XXX_Size() int {
	return m.Size()
}
func (m *MigrationRecords) XXX_DiscardUnknown() {
	xxx_messageInfo_MigrationRecords.DiscardUnknown(m)
}

var xxx_messageInfo_MigrationRecords proto.InternalMessageInfo

func (m *MigrationRecords) GetBalancerToConcentratedPoolLinks() []BalancerToConcentratedPoolLink {
	if m != nil {
		return m.BalancerToConcentratedPoolLinks
	}
	return nil
}

// BalancerToConcentratedPoolLink defines a single link between a single
// balancer pool and a single concentrated liquidity pool. This link is used to
// allow a balancer pool to migrate to a single canonical full range
// concentrated liquidity pool position
// A balancer pool can be linked to a maximum of one cl pool, and a cl pool can
// be linked to a maximum of one balancer pool.
type BalancerToConcentratedPoolLink struct {
	BalancerPoolId uint64 `protobuf:"varint,1,opt,name=balancer_pool_id,json=balancerPoolId,proto3" json:"balancer_pool_id,omitempty"`
	ClPoolId       uint64 `protobuf:"varint,2,opt,name=cl_pool_id,json=clPoolId,proto3" json:"cl_pool_id,omitempty"`
}

func (m *BalancerToConcentratedPoolLink) Reset()         { *m = BalancerToConcentratedPoolLink{} }
func (m *BalancerToConcentratedPoolLink) String() string { return proto.CompactTextString(m) }
func (*BalancerToConcentratedPoolLink) ProtoMessage()    {}
func (*BalancerToConcentratedPoolLink) Descriptor() ([]byte, []int) {
	return fileDescriptor_ecfa270aa4b2c075, []int{1}
}
func (m *BalancerToConcentratedPoolLink) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BalancerToConcentratedPoolLink) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BalancerToConcentratedPoolLink.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BalancerToConcentratedPoolLink) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BalancerToConcentratedPoolLink.Merge(m, src)
}
func (m *BalancerToConcentratedPoolLink) XXX_Size() int {
	return m.Size()
}
func (m *BalancerToConcentratedPoolLink) XXX_DiscardUnknown() {
	xxx_messageInfo_BalancerToConcentratedPoolLink.DiscardUnknown(m)
}

var xxx_messageInfo_BalancerToConcentratedPoolLink proto.InternalMessageInfo

func (m *BalancerToConcentratedPoolLink) GetBalancerPoolId() uint64 {
	if m != nil {
		return m.BalancerPoolId
	}
	return 0
}

func (m *BalancerToConcentratedPoolLink) GetClPoolId() uint64 {
	if m != nil {
		return m.ClPoolId
	}
	return 0
}

func init() {
	proto.RegisterType((*MigrationRecords)(nil), "osmosis.gamm.v1beta1.MigrationRecords")
	proto.RegisterType((*BalancerToConcentratedPoolLink)(nil), "osmosis.gamm.v1beta1.BalancerToConcentratedPoolLink")
}

func init() { proto.RegisterFile("osmosis/gamm/v1beta1/shared.proto", fileDescriptor_ecfa270aa4b2c075) }

var fileDescriptor_ecfa270aa4b2c075 = []byte{
	// 345 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0xb1, 0x4e, 0x32, 0x41,
	0x14, 0x85, 0x77, 0xfe, 0x9f, 0x18, 0xb3, 0x26, 0x86, 0x6c, 0x28, 0x90, 0x98, 0x01, 0xb1, 0xa1,
	0x71, 0x47, 0xd0, 0x8a, 0x12, 0x2b, 0x13, 0x4d, 0x0c, 0xa1, 0xb2, 0xd9, 0xcc, 0xcc, 0x8e, 0xcb,
	0x84, 0xd9, 0xb9, 0x64, 0x67, 0x20, 0xf2, 0x06, 0x96, 0xbe, 0x80, 0x89, 0x8f, 0x43, 0x49, 0x69,
	0x65, 0x0c, 0x34, 0x3e, 0x86, 0x61, 0x67, 0x97, 0x58, 0x18, 0xba, 0x3d, 0x7b, 0xbe, 0x7b, 0x72,
	0xee, 0x1d, 0xff, 0x0c, 0x4c, 0x0a, 0x46, 0x1a, 0x92, 0xd0, 0x34, 0x25, 0xf3, 0x2e, 0x13, 0x96,
	0x76, 0x89, 0x19, 0xd3, 0x4c, 0xc4, 0xe1, 0x34, 0x03, 0x0b, 0x41, 0xad, 0x40, 0xc2, 0x2d, 0x12,
	0x16, 0x48, 0xa3, 0x96, 0x40, 0x02, 0x39, 0x40, 0xb6, 0x5f, 0x8e, 0x6d, 0x9c, 0x24, 0x00, 0x89,
	0x12, 0x24, 0x57, 0x6c, 0xf6, 0x44, 0xa8, 0x5e, 0x94, 0x16, 0xcf, 0x73, 0x22, 0x37, 0xe3, 0x44,
	0x61, 0x61, 0xa7, 0x08, 0xa3, 0x46, 0xec, 0x3a, 0x70, 0x90, 0xda, 0xf9, 0xed, 0x37, 0xe4, 0x57,
	0xef, 0x65, 0x92, 0x51, 0x2b, 0x41, 0x0f, 0x05, 0x87, 0x2c, 0x36, 0xc1, 0x0b, 0xf2, 0xcf, 0x19,
	0x55, 0x54, 0x73, 0x91, 0x45, 0x16, 0x22, 0x0e, 0x9a, 0x0b, 0x6d, 0x33, 0x6a, 0x45, 0x1c, 0x4d,
	0x01, 0x54, 0xa4, 0xa4, 0x9e, 0x98, 0x3a, 0x6a, 0xfd, 0xef, 0x1c, 0xf5, 0xae, 0xc3, 0xbf, 0xb6,
	0x08, 0x07, 0x45, 0xc0, 0x08, 0x6e, 0x7e, 0x8d, 0x3f, 0x00, 0xa8, 0x3b, 0xa9, 0x27, 0x83, 0xca,
	0xf2, 0xb3, 0xe9, 0x0d, 0x9b, 0x6c, 0x2f, 0x65, 0xda, 0xda, 0xc7, 0xfb, 0x83, 0x82, 0x8e, 0x5f,
	0xdd, 0x75, 0xcd, 0xbb, 0xc9, 0xb8, 0x8e, 0x5a, 0xa8, 0x53, 0x19, 0x1e, 0x97, 0xff, 0xb7, 0xec,
	0x6d, 0x1c, 0x9c, 0xfa, 0x3e, 0x57, 0x3b, 0xe6, 0x5f, 0xce, 0x1c, 0x72, 0xe5, 0xdc, 0x7e, 0xe5,
	0xfb, 0xbd, 0x89, 0x06, 0xa3, 0xe5, 0x1a, 0xa3, 0xd5, 0x1a, 0xa3, 0xaf, 0x35, 0x46, 0xaf, 0x1b,
	0xec, 0xad, 0x36, 0xd8, 0xfb, 0xd8, 0x60, 0xef, 0xb1, 0x9f, 0x48, 0x3b, 0x9e, 0xb1, 0x90, 0x43,
	0x4a, 0x8a, 0x85, 0x2f, 0x14, 0x65, 0xa6, 0x14, 0x64, 0xde, 0xbb, 0x24, 0xcf, 0xee, 0xb1, 0xed,
	0x62, 0x2a, 0x0c, 0x49, 0xcb, 0xbb, 0xb2, 0x83, 0xfc, 0xd8, 0x57, 0x3f, 0x01, 0x00, 0x00, 0xff,
	0xff, 0x93, 0x33, 0x43, 0x3f, 0x13, 0x02, 0x00, 0x00,
}

func (this *BalancerToConcentratedPoolLink) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*BalancerToConcentratedPoolLink)
	if !ok {
		that2, ok := that.(BalancerToConcentratedPoolLink)
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
	if this.BalancerPoolId != that1.BalancerPoolId {
		return false
	}
	if this.ClPoolId != that1.ClPoolId {
		return false
	}
	return true
}
func (m *MigrationRecords) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MigrationRecords) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MigrationRecords) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.BalancerToConcentratedPoolLinks) > 0 {
		for iNdEx := len(m.BalancerToConcentratedPoolLinks) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BalancerToConcentratedPoolLinks[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintShared(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *BalancerToConcentratedPoolLink) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BalancerToConcentratedPoolLink) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BalancerToConcentratedPoolLink) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ClPoolId != 0 {
		i = encodeVarintShared(dAtA, i, uint64(m.ClPoolId))
		i--
		dAtA[i] = 0x10
	}
	if m.BalancerPoolId != 0 {
		i = encodeVarintShared(dAtA, i, uint64(m.BalancerPoolId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintShared(dAtA []byte, offset int, v uint64) int {
	offset -= sovShared(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MigrationRecords) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.BalancerToConcentratedPoolLinks) > 0 {
		for _, e := range m.BalancerToConcentratedPoolLinks {
			l = e.Size()
			n += 1 + l + sovShared(uint64(l))
		}
	}
	return n
}

func (m *BalancerToConcentratedPoolLink) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.BalancerPoolId != 0 {
		n += 1 + sovShared(uint64(m.BalancerPoolId))
	}
	if m.ClPoolId != 0 {
		n += 1 + sovShared(uint64(m.ClPoolId))
	}
	return n
}

func sovShared(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozShared(x uint64) (n int) {
	return sovShared(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MigrationRecords) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowShared
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
			return fmt.Errorf("proto: MigrationRecords: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MigrationRecords: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BalancerToConcentratedPoolLinks", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowShared
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
				return ErrInvalidLengthShared
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthShared
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BalancerToConcentratedPoolLinks = append(m.BalancerToConcentratedPoolLinks, BalancerToConcentratedPoolLink{})
			if err := m.BalancerToConcentratedPoolLinks[len(m.BalancerToConcentratedPoolLinks)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipShared(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthShared
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
func (m *BalancerToConcentratedPoolLink) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowShared
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
			return fmt.Errorf("proto: BalancerToConcentratedPoolLink: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BalancerToConcentratedPoolLink: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BalancerPoolId", wireType)
			}
			m.BalancerPoolId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowShared
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BalancerPoolId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClPoolId", wireType)
			}
			m.ClPoolId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowShared
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ClPoolId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipShared(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthShared
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
func skipShared(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowShared
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
					return 0, ErrIntOverflowShared
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
					return 0, ErrIntOverflowShared
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
				return 0, ErrInvalidLengthShared
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupShared
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthShared
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthShared        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowShared          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupShared = fmt.Errorf("proto: unexpected end of group")
)
