// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/pool-yield/v1beta1/yield.proto

package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/duration"
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

type DistrPool struct {
	PotId string `protobuf:"bytes,1,opt,name=pot_id,json=potId,proto3" json:"pot_id,omitempty" yaml:"pot_id"`
	// Denormarized weight.
	Weight github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,2,opt,name=weight,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"weight" yaml:"denormalized_weight"`
}

func (m *DistrPool) Reset()         { *m = DistrPool{} }
func (m *DistrPool) String() string { return proto.CompactTextString(m) }
func (*DistrPool) ProtoMessage()    {}
func (*DistrPool) Descriptor() ([]byte, []int) {
	return fileDescriptor_f4b628395712b609, []int{0}
}
func (m *DistrPool) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DistrPool) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DistrPool.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DistrPool) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DistrPool.Merge(m, src)
}
func (m *DistrPool) XXX_Size() int {
	return m.Size()
}
func (m *DistrPool) XXX_DiscardUnknown() {
	xxx_messageInfo_DistrPool.DiscardUnknown(m)
}

var xxx_messageInfo_DistrPool proto.InternalMessageInfo

func (m *DistrPool) GetPotId() string {
	if m != nil {
		return m.PotId
	}
	return ""
}

type PoolYield struct {
	DistributePools []DistrPool                            `protobuf:"bytes,1,rep,name=distribute_pools,json=distributePools,proto3" json:"distribute_pools"`
	TotalWeight     github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,2,opt,name=total_weight,json=totalWeight,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"total_weight" yaml:"total_weight"`
}

func (m *PoolYield) Reset()         { *m = PoolYield{} }
func (m *PoolYield) String() string { return proto.CompactTextString(m) }
func (*PoolYield) ProtoMessage()    {}
func (*PoolYield) Descriptor() ([]byte, []int) {
	return fileDescriptor_f4b628395712b609, []int{1}
}
func (m *PoolYield) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PoolYield) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PoolYield.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PoolYield) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PoolYield.Merge(m, src)
}
func (m *PoolYield) XXX_Size() int {
	return m.Size()
}
func (m *PoolYield) XXX_DiscardUnknown() {
	xxx_messageInfo_PoolYield.DiscardUnknown(m)
}

var xxx_messageInfo_PoolYield proto.InternalMessageInfo

func (m *PoolYield) GetDistributePools() []DistrPool {
	if m != nil {
		return m.DistributePools
	}
	return nil
}

func init() {
	proto.RegisterType((*DistrPool)(nil), "osmosis.poolyield.v1beta1.DistrPool")
	proto.RegisterType((*PoolYield)(nil), "osmosis.poolyield.v1beta1.PoolYield")
}

func init() {
	proto.RegisterFile("osmosis/pool-yield/v1beta1/yield.proto", fileDescriptor_f4b628395712b609)
}

var fileDescriptor_f4b628395712b609 = []byte{
	// 366 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0x3d, 0x4f, 0xc3, 0x30,
	0x10, 0x8d, 0xf9, 0xa8, 0xd4, 0x14, 0x04, 0x04, 0x86, 0xd2, 0x21, 0xa9, 0x22, 0x54, 0x75, 0x69,
	0xac, 0xc2, 0xc6, 0x58, 0xc1, 0x50, 0x89, 0x01, 0x45, 0x42, 0x08, 0x96, 0x28, 0xa9, 0x4d, 0x6a,
	0x91, 0xf4, 0xa2, 0xd8, 0x01, 0xca, 0xaf, 0x60, 0xe7, 0x0f, 0x75, 0x60, 0xe8, 0x88, 0x18, 0x22,
	0xd4, 0xfe, 0x83, 0xfe, 0x02, 0x64, 0x27, 0x85, 0x32, 0x30, 0x30, 0x39, 0xcf, 0xf7, 0xf2, 0xee,
	0xbd, 0xf3, 0xe9, 0x2d, 0xe0, 0x31, 0x70, 0xc6, 0x71, 0x02, 0x10, 0x75, 0xc6, 0x8c, 0x46, 0x04,
	0x3f, 0x74, 0x03, 0x2a, 0xfc, 0x2e, 0x56, 0xc8, 0x49, 0x52, 0x10, 0x60, 0x1c, 0x96, 0x3c, 0x47,
	0xf2, 0x8a, 0x42, 0x49, 0x6b, 0x1c, 0x84, 0x10, 0x82, 0x62, 0x61, 0xf9, 0x55, 0xfc, 0xd0, 0x30,
	0x43, 0x80, 0x30, 0xa2, 0x58, 0xa1, 0x20, 0xbb, 0xc3, 0x24, 0x4b, 0x7d, 0xc1, 0x60, 0x54, 0xd4,
	0xed, 0x57, 0xa4, 0x57, 0xcf, 0x18, 0x17, 0xe9, 0x25, 0x40, 0x64, 0xb4, 0xf5, 0x4a, 0x02, 0xc2,
	0x63, 0xa4, 0x8e, 0x9a, 0xa8, 0x5d, 0xed, 0xed, 0x2d, 0x72, 0x6b, 0x7b, 0xec, 0xc7, 0xd1, 0xa9,
	0x5d, 0xdc, 0xdb, 0xee, 0x66, 0x02, 0xa2, 0x4f, 0x0c, 0xa2, 0x57, 0x1e, 0x29, 0x0b, 0x87, 0xa2,
	0xbe, 0xa6, 0x98, 0x17, 0x93, 0xdc, 0xd2, 0x3e, 0x72, 0xab, 0x15, 0x32, 0x31, 0xcc, 0x02, 0x67,
	0x00, 0x31, 0x1e, 0x28, 0xb3, 0xe5, 0xd1, 0xe1, 0xe4, 0x1e, 0x8b, 0x71, 0x42, 0xb9, 0xd3, 0x1f,
	0x89, 0x45, 0x6e, 0x35, 0x0a, 0x5d, 0x42, 0x47, 0x90, 0xc6, 0x7e, 0xc4, 0x9e, 0x29, 0xf1, 0x0a,
	0x49, 0xdb, 0x2d, 0xb5, 0xed, 0x37, 0xa4, 0x57, 0xa5, 0xb1, 0x1b, 0x99, 0xd4, 0xb8, 0xd2, 0x77,
	0x89, 0xb4, 0xca, 0x82, 0x4c, 0x50, 0x4f, 0x4e, 0x80, 0xd7, 0x51, 0x73, 0xbd, 0x5d, 0x3b, 0x3e,
	0x72, 0xfe, 0x9c, 0x8b, 0xf3, 0x9d, 0xae, 0xb7, 0x21, 0x3d, 0xba, 0x3b, 0x3f, 0x1a, 0xf2, 0x96,
	0x1b, 0x43, 0x7d, 0x4b, 0x80, 0xf0, 0x23, 0xef, 0x57, 0xa0, 0xf3, 0x7f, 0x07, 0xda, 0x2f, 0x02,
	0xad, 0x6a, 0xd9, 0x6e, 0x4d, 0xc1, 0x6b, 0x85, 0x7a, 0xfd, 0xc9, 0xcc, 0x44, 0xd3, 0x99, 0x89,
	0x3e, 0x67, 0x26, 0x7a, 0x99, 0x9b, 0xda, 0x74, 0x6e, 0x6a, 0xef, 0x73, 0x53, 0xbb, 0xc5, 0xab,
	0x5d, 0x3a, 0xcb, 0x65, 0x58, 0x9e, 0x4f, 0xab, 0x6b, 0xa1, 0x5a, 0x06, 0x15, 0xf5, 0x7c, 0x27,
	0x5f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xf7, 0x8d, 0xaa, 0x5b, 0x39, 0x02, 0x00, 0x00,
}

func (m *DistrPool) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DistrPool) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DistrPool) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.Weight.Size()
		i -= size
		if _, err := m.Weight.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintYield(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.PotId) > 0 {
		i -= len(m.PotId)
		copy(dAtA[i:], m.PotId)
		i = encodeVarintYield(dAtA, i, uint64(len(m.PotId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *PoolYield) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PoolYield) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PoolYield) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.TotalWeight.Size()
		i -= size
		if _, err := m.TotalWeight.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintYield(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.DistributePools) > 0 {
		for iNdEx := len(m.DistributePools) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DistributePools[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintYield(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintYield(dAtA []byte, offset int, v uint64) int {
	offset -= sovYield(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *DistrPool) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.PotId)
	if l > 0 {
		n += 1 + l + sovYield(uint64(l))
	}
	l = m.Weight.Size()
	n += 1 + l + sovYield(uint64(l))
	return n
}

func (m *PoolYield) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.DistributePools) > 0 {
		for _, e := range m.DistributePools {
			l = e.Size()
			n += 1 + l + sovYield(uint64(l))
		}
	}
	l = m.TotalWeight.Size()
	n += 1 + l + sovYield(uint64(l))
	return n
}

func sovYield(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozYield(x uint64) (n int) {
	return sovYield(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *DistrPool) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowYield
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
			return fmt.Errorf("proto: DistrPool: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DistrPool: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PotId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowYield
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
				return ErrInvalidLengthYield
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthYield
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PotId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Weight", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowYield
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
				return ErrInvalidLengthYield
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthYield
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Weight.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipYield(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthYield
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthYield
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
func (m *PoolYield) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowYield
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
			return fmt.Errorf("proto: PoolYield: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PoolYield: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DistributePools", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowYield
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
				return ErrInvalidLengthYield
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthYield
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DistributePools = append(m.DistributePools, DistrPool{})
			if err := m.DistributePools[len(m.DistributePools)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalWeight", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowYield
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
				return ErrInvalidLengthYield
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthYield
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.TotalWeight.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipYield(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthYield
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthYield
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
func skipYield(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowYield
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
					return 0, ErrIntOverflowYield
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
					return 0, ErrIntOverflowYield
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
				return 0, ErrInvalidLengthYield
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupYield
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthYield
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthYield        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowYield          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupYield = fmt.Errorf("proto: unexpected end of group")
)
