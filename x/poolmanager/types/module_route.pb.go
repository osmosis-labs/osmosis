// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/poolmanager/v1beta1/module_route.proto

package types

import (
	fmt "fmt"
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

// PoolType is an enumeration of all supported pool types.
type PoolType int32

const (
	// Balancer is the standard xy=k curve. Its pool model is defined in x/gamm.
	Balancer PoolType = 0
	// Stableswap is the Solidly cfmm stable swap curve. Its pool model is defined
	// in x/gamm.
	Stableswap PoolType = 1
)

var PoolType_name = map[int32]string{
	0: "Balancer",
	1: "Stableswap",
}

var PoolType_value = map[string]int32{
	"Balancer":   0,
	"Stableswap": 1,
}

func (x PoolType) String() string {
	return proto.EnumName(PoolType_name, int32(x))
}

func (PoolType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_96bfcc7b6d387cee, []int{0}
}

// ModuleRouter defines a route encapsulating pool type.
// It is used as the value of a mapping from pool id to the pool type,
// allowing the pool manager to know which module to route swaps to given the
// pool id.
type ModuleRoute struct {
	// pool_type specifies the type of the pool
	PoolType PoolType `protobuf:"varint,1,opt,name=pool_type,json=poolType,proto3,enum=osmosis.poolmanager.v1beta1.PoolType" json:"pool_type,omitempty"`
}

func (m *ModuleRoute) Reset()         { *m = ModuleRoute{} }
func (m *ModuleRoute) String() string { return proto.CompactTextString(m) }
func (*ModuleRoute) ProtoMessage()    {}
func (*ModuleRoute) Descriptor() ([]byte, []int) {
	return fileDescriptor_96bfcc7b6d387cee, []int{0}
}
func (m *ModuleRoute) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ModuleRoute) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ModuleRoute.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ModuleRoute) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ModuleRoute.Merge(m, src)
}
func (m *ModuleRoute) XXX_Size() int {
	return m.Size()
}
func (m *ModuleRoute) XXX_DiscardUnknown() {
	xxx_messageInfo_ModuleRoute.DiscardUnknown(m)
}

var xxx_messageInfo_ModuleRoute proto.InternalMessageInfo

func (m *ModuleRoute) GetPoolType() PoolType {
	if m != nil {
		return m.PoolType
	}
	return Balancer
}

func init() {
	proto.RegisterEnum("osmosis.poolmanager.v1beta1.PoolType", PoolType_name, PoolType_value)
	proto.RegisterType((*ModuleRoute)(nil), "osmosis.poolmanager.v1beta1.ModuleRoute")
}

func init() {
	proto.RegisterFile("osmosis/poolmanager/v1beta1/module_route.proto", fileDescriptor_96bfcc7b6d387cee)
}

var fileDescriptor_96bfcc7b6d387cee = []byte{
	// 250 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xd2, 0xcb, 0x2f, 0xce, 0xcd,
	0x2f, 0xce, 0x2c, 0xd6, 0x2f, 0xc8, 0xcf, 0xcf, 0xc9, 0x4d, 0xcc, 0x4b, 0x4c, 0x4f, 0x2d, 0xd2,
	0x2f, 0x33, 0x4c, 0x4a, 0x2d, 0x49, 0x34, 0xd4, 0xcf, 0xcd, 0x4f, 0x29, 0xcd, 0x49, 0x8d, 0x2f,
	0xca, 0x2f, 0x2d, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x92, 0x86, 0xaa, 0xd7, 0x43,
	0x52, 0xaf, 0x07, 0x55, 0x2f, 0x25, 0x92, 0x9e, 0x9f, 0x9e, 0x0f, 0x56, 0xa7, 0x0f, 0x62, 0x41,
	0xb4, 0x28, 0x05, 0x72, 0x71, 0xfb, 0x82, 0x0d, 0x0a, 0x02, 0x99, 0x23, 0xe4, 0xc4, 0xc5, 0x09,
	0xd2, 0x1b, 0x5f, 0x52, 0x59, 0x90, 0x2a, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0x67, 0xa4, 0xaa, 0x87,
	0xc7, 0x54, 0xbd, 0x80, 0xfc, 0xfc, 0x9c, 0x90, 0xca, 0x82, 0xd4, 0x20, 0x8e, 0x02, 0x28, 0x4b,
	0x4b, 0x8f, 0x8b, 0x03, 0x26, 0x2a, 0xc4, 0xc3, 0xc5, 0xe1, 0x94, 0x98, 0x93, 0x98, 0x97, 0x9c,
	0x5a, 0x24, 0xc0, 0x20, 0xc4, 0xc7, 0xc5, 0x15, 0x5c, 0x92, 0x98, 0x94, 0x93, 0x5a, 0x5c, 0x9e,
	0x58, 0x20, 0xc0, 0x28, 0xc5, 0xd2, 0xb1, 0x58, 0x8e, 0xc1, 0x29, 0xf0, 0xc4, 0x23, 0x39, 0xc6,
	0x0b, 0x8f, 0xe4, 0x18, 0x1f, 0x3c, 0x92, 0x63, 0x9c, 0xf0, 0x58, 0x8e, 0xe1, 0xc2, 0x63, 0x39,
	0x86, 0x1b, 0x8f, 0xe5, 0x18, 0xa2, 0xcc, 0xd3, 0x33, 0x4b, 0x32, 0x4a, 0x93, 0xf4, 0x92, 0xf3,
	0x73, 0xf5, 0xa1, 0x8e, 0xd0, 0xcd, 0x49, 0x4c, 0x2a, 0x86, 0x71, 0xf4, 0xcb, 0x0c, 0x4d, 0xf4,
	0x2b, 0x50, 0x42, 0x07, 0xe4, 0xf0, 0xe2, 0x24, 0x36, 0xb0, 0xe7, 0x8c, 0x01, 0x01, 0x00, 0x00,
	0xff, 0xff, 0x98, 0x55, 0xa1, 0xc2, 0x41, 0x01, 0x00, 0x00,
}

func (m *ModuleRoute) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ModuleRoute) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ModuleRoute) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.PoolType != 0 {
		i = encodeVarintModuleRoute(dAtA, i, uint64(m.PoolType))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintModuleRoute(dAtA []byte, offset int, v uint64) int {
	offset -= sovModuleRoute(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ModuleRoute) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PoolType != 0 {
		n += 1 + sovModuleRoute(uint64(m.PoolType))
	}
	return n
}

func sovModuleRoute(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozModuleRoute(x uint64) (n int) {
	return sovModuleRoute(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ModuleRoute) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowModuleRoute
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
			return fmt.Errorf("proto: ModuleRoute: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ModuleRoute: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolType", wireType)
			}
			m.PoolType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowModuleRoute
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PoolType |= PoolType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipModuleRoute(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthModuleRoute
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
func skipModuleRoute(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowModuleRoute
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
					return 0, ErrIntOverflowModuleRoute
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
					return 0, ErrIntOverflowModuleRoute
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
				return 0, ErrInvalidLengthModuleRoute
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupModuleRoute
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthModuleRoute
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthModuleRoute        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowModuleRoute          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupModuleRoute = fmt.Errorf("proto: unexpected end of group")
)
