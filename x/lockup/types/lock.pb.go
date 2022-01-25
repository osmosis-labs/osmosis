// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/lockup/lock.proto

package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	_ "google.golang.org/protobuf/types/known/durationpb"
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

type LockQueryType int32

const (
	ByDuration LockQueryType = 0
	ByTime     LockQueryType = 1
)

var LockQueryType_name = map[int32]string{
	0: "ByDuration",
	1: "ByTime",
}

var LockQueryType_value = map[string]int32{
	"ByDuration": 0,
	"ByTime":     1,
}

func (x LockQueryType) String() string {
	return proto.EnumName(LockQueryType_name, int32(x))
}

func (LockQueryType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_7e9d7527a237b489, []int{0}
}

// PeriodLock is a single unit of lock by period. It's a record of locked coin
// at a specific time. It stores owner, duration, unlock time and the amount of
// coins locked.
type PeriodLock struct {
	ID       uint64                                   `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Owner    string                                   `protobuf:"bytes,2,opt,name=owner,proto3" json:"owner,omitempty" yaml:"owner"`
	Duration time.Duration                            `protobuf:"bytes,3,opt,name=duration,proto3,stdduration" json:"duration,omitempty" yaml:"duration"`
	EndTime  time.Time                                `protobuf:"bytes,4,opt,name=end_time,json=endTime,proto3,stdtime" json:"end_time" yaml:"end_time"`
	Coins    github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,5,rep,name=coins,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"coins"`
}

func (m *PeriodLock) Reset()         { *m = PeriodLock{} }
func (m *PeriodLock) String() string { return proto.CompactTextString(m) }
func (*PeriodLock) ProtoMessage()    {}
func (*PeriodLock) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e9d7527a237b489, []int{0}
}
func (m *PeriodLock) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PeriodLock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PeriodLock.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PeriodLock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeriodLock.Merge(m, src)
}
func (m *PeriodLock) XXX_Size() int {
	return m.Size()
}
func (m *PeriodLock) XXX_DiscardUnknown() {
	xxx_messageInfo_PeriodLock.DiscardUnknown(m)
}

var xxx_messageInfo_PeriodLock proto.InternalMessageInfo

func (m *PeriodLock) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *PeriodLock) GetOwner() string {
	if m != nil {
		return m.Owner
	}
	return ""
}

func (m *PeriodLock) GetDuration() time.Duration {
	if m != nil {
		return m.Duration
	}
	return 0
}

func (m *PeriodLock) GetEndTime() time.Time {
	if m != nil {
		return m.EndTime
	}
	return time.Time{}
}

func (m *PeriodLock) GetCoins() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Coins
	}
	return nil
}

type QueryCondition struct {
	LockQueryType LockQueryType `protobuf:"varint,1,opt,name=lock_query_type,json=lockQueryType,proto3,enum=osmosis.lockup.LockQueryType" json:"lock_query_type,omitempty"`
	Denom         string        `protobuf:"bytes,2,opt,name=denom,proto3" json:"denom,omitempty"`
	// valid when query condition is ByDuration
	Duration time.Duration `protobuf:"bytes,3,opt,name=duration,proto3,stdduration" json:"duration" yaml:"duration"`
	// valid when query condition is ByTime
	Timestamp time.Time `protobuf:"bytes,4,opt,name=timestamp,proto3,stdtime" json:"timestamp" yaml:"timestamp"`
}

func (m *QueryCondition) Reset()         { *m = QueryCondition{} }
func (m *QueryCondition) String() string { return proto.CompactTextString(m) }
func (*QueryCondition) ProtoMessage()    {}
func (*QueryCondition) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e9d7527a237b489, []int{1}
}
func (m *QueryCondition) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryCondition) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryCondition.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryCondition) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryCondition.Merge(m, src)
}
func (m *QueryCondition) XXX_Size() int {
	return m.Size()
}
func (m *QueryCondition) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryCondition.DiscardUnknown(m)
}

var xxx_messageInfo_QueryCondition proto.InternalMessageInfo

func (m *QueryCondition) GetLockQueryType() LockQueryType {
	if m != nil {
		return m.LockQueryType
	}
	return ByDuration
}

func (m *QueryCondition) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

func (m *QueryCondition) GetDuration() time.Duration {
	if m != nil {
		return m.Duration
	}
	return 0
}

func (m *QueryCondition) GetTimestamp() time.Time {
	if m != nil {
		return m.Timestamp
	}
	return time.Time{}
}

func init() {
	proto.RegisterEnum("osmosis.lockup.LockQueryType", LockQueryType_name, LockQueryType_value)
	proto.RegisterType((*PeriodLock)(nil), "osmosis.lockup.PeriodLock")
	proto.RegisterType((*QueryCondition)(nil), "osmosis.lockup.QueryCondition")
}

func init() { proto.RegisterFile("osmosis/lockup/lock.proto", fileDescriptor_7e9d7527a237b489) }

var fileDescriptor_7e9d7527a237b489 = []byte{
	// 529 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x93, 0xbf, 0x6f, 0xd4, 0x30,
	0x14, 0xc7, 0x93, 0x6b, 0xaf, 0xb4, 0x86, 0x5e, 0x4f, 0x56, 0x87, 0xf4, 0x80, 0xe4, 0x94, 0x01,
	0x9d, 0x10, 0xb5, 0x49, 0x11, 0x0b, 0x63, 0x7a, 0x0c, 0x15, 0x0c, 0x10, 0x55, 0x0c, 0x2c, 0xa7,
	0xfc, 0x30, 0x69, 0x74, 0x49, 0x1c, 0xe2, 0xa4, 0x90, 0xff, 0x80, 0xb1, 0x23, 0xcc, 0x6c, 0xfc,
	0x25, 0x1d, 0x3b, 0x32, 0x5d, 0xd1, 0x9d, 0x58, 0x18, 0xfb, 0x17, 0x20, 0xdb, 0xf1, 0xf5, 0x07,
	0x42, 0xea, 0xe4, 0x3c, 0x7f, 0xfd, 0xbe, 0xef, 0xf9, 0xf3, 0x1c, 0xb0, 0x43, 0x59, 0x46, 0x59,
	0xc2, 0x70, 0x4a, 0xc3, 0x69, 0x5d, 0x88, 0x05, 0x15, 0x25, 0xad, 0x28, 0xec, 0xb5, 0x12, 0x92,
	0xd2, 0x60, 0x3b, 0xa6, 0x31, 0x15, 0x12, 0xe6, 0x5f, 0xf2, 0xd4, 0xc0, 0x8c, 0x29, 0x8d, 0x53,
	0x82, 0x45, 0x14, 0xd4, 0x1f, 0x70, 0x54, 0x97, 0x7e, 0x95, 0xd0, 0xbc, 0xd5, 0xad, 0x9b, 0x7a,
	0x95, 0x64, 0x84, 0x55, 0x7e, 0x56, 0x28, 0x83, 0x50, 0xd4, 0xc1, 0x81, 0xcf, 0x08, 0x3e, 0x76,
	0x02, 0x52, 0xf9, 0x0e, 0x0e, 0x69, 0xd2, 0x1a, 0xd8, 0xbf, 0x3b, 0x00, 0xbc, 0x21, 0x65, 0x42,
	0xa3, 0xd7, 0x34, 0x9c, 0xc2, 0x1e, 0xe8, 0x1c, 0x8c, 0x0d, 0x7d, 0xa8, 0x8f, 0x56, 0xbd, 0xce,
	0xc1, 0x18, 0x3e, 0x02, 0x5d, 0xfa, 0x29, 0x27, 0xa5, 0xd1, 0x19, 0xea, 0xa3, 0x0d, 0xb7, 0x7f,
	0x31, 0xb3, 0xee, 0x35, 0x7e, 0x96, 0xbe, 0xb0, 0xc5, 0xb6, 0xed, 0x49, 0x19, 0x1e, 0x81, 0x75,
	0xd5, 0x99, 0xb1, 0x32, 0xd4, 0x47, 0x77, 0xf7, 0x76, 0x90, 0x6c, 0x0d, 0xa9, 0xd6, 0xd0, 0xb8,
	0x3d, 0xe0, 0x3a, 0xa7, 0x33, 0x4b, 0xfb, 0x33, 0xb3, 0xa0, 0x4a, 0x79, 0x42, 0xb3, 0xa4, 0x22,
	0x59, 0x51, 0x35, 0x17, 0x33, 0x6b, 0x4b, 0xfa, 0x2b, 0xcd, 0xfe, 0x7a, 0x6e, 0xe9, 0xde, 0xd2,
	0x1d, 0x7a, 0x60, 0x9d, 0xe4, 0xd1, 0x84, 0xdf, 0xd3, 0x58, 0x15, 0x95, 0x06, 0xff, 0x54, 0x3a,
	0x54, 0x10, 0xdc, 0xfb, 0xbc, 0xd4, 0xa5, 0xa9, 0xca, 0xb4, 0x4f, 0xb8, 0xe9, 0x1d, 0x92, 0x47,
	0xfc, 0x28, 0xf4, 0x41, 0x97, 0x23, 0x61, 0x46, 0x77, 0xb8, 0x22, 0x5a, 0x97, 0xd0, 0x10, 0x87,
	0x86, 0x5a, 0x68, 0x68, 0x9f, 0x26, 0xb9, 0xfb, 0x94, 0xfb, 0xfd, 0x38, 0xb7, 0x46, 0x71, 0x52,
	0x1d, 0xd5, 0x01, 0x0a, 0x69, 0x86, 0x5b, 0xc2, 0x72, 0xd9, 0x65, 0xd1, 0x14, 0x57, 0x4d, 0x41,
	0x98, 0x48, 0x60, 0x9e, 0x74, 0xb6, 0xbf, 0x75, 0x40, 0xef, 0x6d, 0x4d, 0xca, 0x66, 0x9f, 0xe6,
	0x51, 0x22, 0x6e, 0xf2, 0x12, 0x6c, 0xf1, 0xd9, 0x4f, 0x3e, 0xf2, 0xed, 0x09, 0xcf, 0x11, 0xe0,
	0x7b, 0x7b, 0x0f, 0xd1, 0xf5, 0xb7, 0x81, 0xf8, 0x68, 0x44, 0xf2, 0x61, 0x53, 0x10, 0x6f, 0x33,
	0xbd, 0x1a, 0xc2, 0x6d, 0xd0, 0x8d, 0x48, 0x4e, 0x33, 0x39, 0x22, 0x4f, 0x06, 0x1c, 0xd3, 0xed,
	0x07, 0x72, 0x83, 0xd2, 0xff, 0xd0, 0xbf, 0x03, 0x1b, 0xcb, 0xe7, 0x75, 0x0b, 0xf6, 0x0f, 0x5a,
	0xd7, 0xbe, 0x74, 0x5d, 0xa6, 0x4a, 0xf8, 0x97, 0x56, 0x8f, 0x1d, 0xb0, 0x79, 0xed, 0x86, 0xb0,
	0x07, 0x80, 0xdb, 0xa8, 0xee, 0xfa, 0x1a, 0x04, 0x60, 0xcd, 0x6d, 0xb8, 0x71, 0x5f, 0x1f, 0xac,
	0x7e, 0xf9, 0x6e, 0x6a, 0xee, 0xab, 0xd3, 0xb9, 0xa9, 0x9f, 0xcd, 0x4d, 0xfd, 0xd7, 0xdc, 0xd4,
	0x4f, 0x16, 0xa6, 0x76, 0xb6, 0x30, 0xb5, 0x9f, 0x0b, 0x53, 0x7b, 0xef, 0x5c, 0x99, 0x4c, 0x8b,
	0x71, 0x37, 0xf5, 0x03, 0xa6, 0x02, 0x7c, 0xfc, 0x1c, 0x7f, 0x56, 0xff, 0xa3, 0x18, 0x54, 0xb0,
	0x26, 0x9a, 0x7f, 0xf6, 0x37, 0x00, 0x00, 0xff, 0xff, 0xeb, 0xab, 0xd1, 0xd5, 0xae, 0x03, 0x00,
	0x00,
}

func (m *PeriodLock) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PeriodLock) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PeriodLock) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Coins) > 0 {
		for iNdEx := len(m.Coins) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Coins[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintLock(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	n1, err1 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.EndTime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.EndTime):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintLock(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x22
	n2, err2 := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.Duration, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdDuration(m.Duration):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintLock(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x1a
	if len(m.Owner) > 0 {
		i -= len(m.Owner)
		copy(dAtA[i:], m.Owner)
		i = encodeVarintLock(dAtA, i, uint64(len(m.Owner)))
		i--
		dAtA[i] = 0x12
	}
	if m.ID != 0 {
		i = encodeVarintLock(dAtA, i, uint64(m.ID))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryCondition) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryCondition) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryCondition) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n3, err3 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.Timestamp, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.Timestamp):])
	if err3 != nil {
		return 0, err3
	}
	i -= n3
	i = encodeVarintLock(dAtA, i, uint64(n3))
	i--
	dAtA[i] = 0x22
	n4, err4 := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.Duration, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdDuration(m.Duration):])
	if err4 != nil {
		return 0, err4
	}
	i -= n4
	i = encodeVarintLock(dAtA, i, uint64(n4))
	i--
	dAtA[i] = 0x1a
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintLock(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0x12
	}
	if m.LockQueryType != 0 {
		i = encodeVarintLock(dAtA, i, uint64(m.LockQueryType))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintLock(dAtA []byte, offset int, v uint64) int {
	offset -= sovLock(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *PeriodLock) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ID != 0 {
		n += 1 + sovLock(uint64(m.ID))
	}
	l = len(m.Owner)
	if l > 0 {
		n += 1 + l + sovLock(uint64(l))
	}
	l = github_com_gogo_protobuf_types.SizeOfStdDuration(m.Duration)
	n += 1 + l + sovLock(uint64(l))
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.EndTime)
	n += 1 + l + sovLock(uint64(l))
	if len(m.Coins) > 0 {
		for _, e := range m.Coins {
			l = e.Size()
			n += 1 + l + sovLock(uint64(l))
		}
	}
	return n
}

func (m *QueryCondition) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.LockQueryType != 0 {
		n += 1 + sovLock(uint64(m.LockQueryType))
	}
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovLock(uint64(l))
	}
	l = github_com_gogo_protobuf_types.SizeOfStdDuration(m.Duration)
	n += 1 + l + sovLock(uint64(l))
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.Timestamp)
	n += 1 + l + sovLock(uint64(l))
	return n
}

func sovLock(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozLock(x uint64) (n int) {
	return sovLock(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PeriodLock) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLock
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
			return fmt.Errorf("proto: PeriodLock: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PeriodLock: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			m.ID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLock
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ID |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Owner", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLock
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
				return ErrInvalidLengthLock
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthLock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Owner = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Duration", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLock
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
				return ErrInvalidLengthLock
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&m.Duration, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EndTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLock
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
				return ErrInvalidLengthLock
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.EndTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Coins", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLock
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
				return ErrInvalidLengthLock
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Coins = append(m.Coins, types.Coin{})
			if err := m.Coins[len(m.Coins)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLock(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLock
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
func (m *QueryCondition) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLock
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
			return fmt.Errorf("proto: QueryCondition: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryCondition: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LockQueryType", wireType)
			}
			m.LockQueryType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLock
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LockQueryType |= LockQueryType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLock
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
				return ErrInvalidLengthLock
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthLock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Duration", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLock
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
				return ErrInvalidLengthLock
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&m.Duration, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLock
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
				return ErrInvalidLengthLock
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.Timestamp, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLock(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLock
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
func skipLock(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowLock
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
					return 0, ErrIntOverflowLock
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
					return 0, ErrIntOverflowLock
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
				return 0, ErrInvalidLengthLock
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupLock
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthLock
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthLock        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLock          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupLock = fmt.Errorf("proto: unexpected end of group")
)
