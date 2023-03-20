// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/concentrated-liquidity/incentive_record.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
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

// IncentiveRecord is the high-level struct we use to deal with an independent
// incentive being distributed on a pool. Note that PoolId, Denom, and MinUptime
// are included in the key so we avoid storing them in state, hence the
// distinction between IncentiveRecord and IncentiveRecordBody.
type IncentiveRecord struct {
	PoolId uint64 `protobuf:"varint,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	// incentive_denom is the denom of the token being distributed as part of this
	// incentive record
	IncentiveDenom string `protobuf:"bytes,2,opt,name=incentive_denom,json=incentiveDenom,proto3" json:"incentive_denom,omitempty" yaml:"incentive_denom"`
	// incentiveCreator is the address that created the incentive record. This
	// address does not have any special privileges – it is only kept to keep
	// incentive records created by different addresses separate.
	IncentiveCreatorAddr string `protobuf:"bytes,3,opt,name=incentive_creator_addr,json=incentiveCreatorAddr,proto3" json:"incentive_creator_addr,omitempty" yaml:"incentive_creator_addr"`
	// incentive record body holds necessary
	IncentiveRecordBody IncentiveRecordBody `protobuf:"bytes,4,opt,name=incentive_record_body,json=incentiveRecordBody,proto3" json:"incentive_record_body" yaml:"incentive_record_body"`
	// min_uptime is the minimum uptime required for liquidity to qualify for this
	// incentive. It should be always be one of the supported uptimes in
	// types.SupportedUptimes
	MinUptime time.Duration `protobuf:"bytes,5,opt,name=min_uptime,json=minUptime,proto3,stdduration" json:"min_uptime" yaml:"min_uptime"`
}

func (m *IncentiveRecord) Reset()         { *m = IncentiveRecord{} }
func (m *IncentiveRecord) String() string { return proto.CompactTextString(m) }
func (*IncentiveRecord) ProtoMessage()    {}
func (*IncentiveRecord) Descriptor() ([]byte, []int) {
	return fileDescriptor_9d38bf94e42ee434, []int{0}
}
func (m *IncentiveRecord) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *IncentiveRecord) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_IncentiveRecord.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *IncentiveRecord) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IncentiveRecord.Merge(m, src)
}
func (m *IncentiveRecord) XXX_Size() int {
	return m.Size()
}
func (m *IncentiveRecord) XXX_DiscardUnknown() {
	xxx_messageInfo_IncentiveRecord.DiscardUnknown(m)
}

var xxx_messageInfo_IncentiveRecord proto.InternalMessageInfo

func (m *IncentiveRecord) GetPoolId() uint64 {
	if m != nil {
		return m.PoolId
	}
	return 0
}

func (m *IncentiveRecord) GetIncentiveDenom() string {
	if m != nil {
		return m.IncentiveDenom
	}
	return ""
}

func (m *IncentiveRecord) GetIncentiveCreatorAddr() string {
	if m != nil {
		return m.IncentiveCreatorAddr
	}
	return ""
}

func (m *IncentiveRecord) GetIncentiveRecordBody() IncentiveRecordBody {
	if m != nil {
		return m.IncentiveRecordBody
	}
	return IncentiveRecordBody{}
}

func (m *IncentiveRecord) GetMinUptime() time.Duration {
	if m != nil {
		return m.MinUptime
	}
	return 0
}

// IncentiveRecordBody represents an active perpetual incentive gauge for a pool
type IncentiveRecordBody struct {
	// remaining_amount is the total amount of incentives to be distributed
	RemainingAmount github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,1,opt,name=remaining_amount,json=remainingAmount,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"remaining_amount" yaml:"remaining_amount"`
	// emission_rate is the incentive emission rate per second
	EmissionRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,2,opt,name=emission_rate,json=emissionRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"emission_rate" yaml:"swap_fee"`
	// start_time is the time when the incentive starts distributing
	StartTime time.Time `protobuf:"bytes,3,opt,name=start_time,json=startTime,proto3,stdtime" json:"start_time" yaml:"start_time"`
}

func (m *IncentiveRecordBody) Reset()         { *m = IncentiveRecordBody{} }
func (m *IncentiveRecordBody) String() string { return proto.CompactTextString(m) }
func (*IncentiveRecordBody) ProtoMessage()    {}
func (*IncentiveRecordBody) Descriptor() ([]byte, []int) {
	return fileDescriptor_9d38bf94e42ee434, []int{1}
}
func (m *IncentiveRecordBody) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *IncentiveRecordBody) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_IncentiveRecordBody.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *IncentiveRecordBody) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IncentiveRecordBody.Merge(m, src)
}
func (m *IncentiveRecordBody) XXX_Size() int {
	return m.Size()
}
func (m *IncentiveRecordBody) XXX_DiscardUnknown() {
	xxx_messageInfo_IncentiveRecordBody.DiscardUnknown(m)
}

var xxx_messageInfo_IncentiveRecordBody proto.InternalMessageInfo

func (m *IncentiveRecordBody) GetStartTime() time.Time {
	if m != nil {
		return m.StartTime
	}
	return time.Time{}
}

func init() {
	proto.RegisterType((*IncentiveRecord)(nil), "osmosis.concentratedliquidity.v1beta1.IncentiveRecord")
	proto.RegisterType((*IncentiveRecordBody)(nil), "osmosis.concentratedliquidity.v1beta1.IncentiveRecordBody")
}

func init() {
	proto.RegisterFile("osmosis/concentrated-liquidity/incentive_record.proto", fileDescriptor_9d38bf94e42ee434)
}

var fileDescriptor_9d38bf94e42ee434 = []byte{
	// 562 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xcd, 0x6e, 0xd3, 0x4c,
	0x14, 0x8d, 0x9b, 0x7e, 0xfd, 0x94, 0xe1, 0x27, 0xe0, 0x96, 0x36, 0x8d, 0xa8, 0x1d, 0x2c, 0x40,
	0xd9, 0xc4, 0x56, 0x40, 0xdd, 0x74, 0x17, 0x37, 0x9b, 0x6c, 0x2d, 0x50, 0x11, 0x42, 0xb2, 0xc6,
	0x9e, 0x89, 0x19, 0x91, 0xf1, 0x98, 0x99, 0x71, 0x20, 0x6f, 0xd1, 0x05, 0x0b, 0x1e, 0x85, 0x47,
	0xe8, 0x06, 0xa9, 0x4b, 0xc4, 0xc2, 0xa0, 0xe4, 0x0d, 0xf2, 0x04, 0xc8, 0x63, 0xe7, 0x47, 0x2e,
	0x48, 0xb0, 0x4a, 0xee, 0x99, 0x73, 0xee, 0x3d, 0x73, 0xee, 0x24, 0xe0, 0x94, 0x09, 0xca, 0x04,
	0x11, 0x4e, 0xc8, 0xe2, 0x10, 0xc7, 0x92, 0x43, 0x89, 0x51, 0x6f, 0x42, 0xde, 0xa7, 0x04, 0x11,
	0x39, 0x73, 0x88, 0x42, 0xc9, 0x14, 0xfb, 0x1c, 0x87, 0x8c, 0x23, 0x3b, 0xe1, 0x4c, 0x32, 0xfd,
	0x49, 0x29, 0xb3, 0xb7, 0x65, 0x6b, 0x95, 0x3d, 0xed, 0x07, 0x58, 0xc2, 0x7e, 0xfb, 0x38, 0x54,
	0x3c, 0x5f, 0x89, 0x9c, 0xa2, 0x28, 0x3a, 0xb4, 0x0f, 0x22, 0x16, 0xb1, 0x02, 0xcf, 0xbf, 0x95,
	0xa8, 0x19, 0x31, 0x16, 0x4d, 0xb0, 0xa3, 0xaa, 0x20, 0x1d, 0x3b, 0x92, 0x50, 0x2c, 0x24, 0xa4,
	0x49, 0x49, 0x30, 0xaa, 0x04, 0x94, 0x72, 0x28, 0x09, 0x8b, 0x8b, 0x73, 0xeb, 0x4b, 0x1d, 0x34,
	0x47, 0x2b, 0xcf, 0x9e, 0xb2, 0xac, 0x1f, 0x81, 0xff, 0x13, 0xc6, 0x26, 0x3e, 0x41, 0x2d, 0xad,
	0xa3, 0x75, 0x77, 0xbd, 0xbd, 0xbc, 0x1c, 0x21, 0xfd, 0x1c, 0x34, 0x37, 0xf7, 0x43, 0x38, 0x66,
	0xb4, 0xb5, 0xd3, 0xd1, 0xba, 0x0d, 0xb7, 0xbd, 0xcc, 0xcc, 0xc3, 0x19, 0xa4, 0x93, 0x33, 0xab,
	0x42, 0xb0, 0xbc, 0xbb, 0x6b, 0x64, 0x98, 0x03, 0xfa, 0x05, 0x38, 0xdc, 0x70, 0x42, 0x8e, 0xa1,
	0x64, 0xdc, 0x87, 0x08, 0xf1, 0x56, 0x5d, 0xf5, 0x7a, 0xb4, 0xcc, 0xcc, 0x93, 0x6a, 0xaf, 0x6d,
	0x9e, 0xe5, 0x1d, 0xac, 0x0f, 0xce, 0x0b, 0x7c, 0x80, 0x10, 0xd7, 0x3f, 0x69, 0xe0, 0x41, 0x35,
	0x7e, 0x3f, 0x60, 0x68, 0xd6, 0xda, 0xed, 0x68, 0xdd, 0x5b, 0xcf, 0xce, 0xec, 0xbf, 0x5a, 0x82,
	0x5d, 0x89, 0xc3, 0x65, 0x68, 0xe6, 0x3e, 0xbe, 0xca, 0xcc, 0xda, 0x32, 0x33, 0x1f, 0x56, 0x8d,
	0x6d, 0x8d, 0xb1, 0xbc, 0x7d, 0x72, 0x53, 0xaa, 0x5f, 0x00, 0x40, 0x49, 0xec, 0xa7, 0x49, 0xbe,
	0x9a, 0xd6, 0x7f, 0xca, 0xca, 0xb1, 0x5d, 0xac, 0xc5, 0x5e, 0xad, 0xc5, 0x1e, 0x96, 0x6b, 0x71,
	0x4f, 0xca, 0x49, 0xf7, 0x8b, 0x49, 0x1b, 0xa9, 0xf5, 0xf9, 0x87, 0xa9, 0x79, 0x0d, 0x4a, 0xe2,
	0x97, 0x45, 0xfd, 0x75, 0x07, 0xec, 0xff, 0xc6, 0xab, 0x2e, 0xc1, 0x3d, 0x8e, 0x29, 0x24, 0x31,
	0x89, 0x23, 0x1f, 0x52, 0x96, 0xc6, 0x52, 0xed, 0xb1, 0xe1, 0x8e, 0xf2, 0xde, 0xdf, 0x33, 0xf3,
	0x69, 0x44, 0xe4, 0xdb, 0x34, 0xb0, 0x43, 0x46, 0xcb, 0x47, 0x56, 0x7e, 0xf4, 0x04, 0x7a, 0xe7,
	0xc8, 0x59, 0x82, 0x85, 0x3d, 0xc4, 0xe1, 0x32, 0x33, 0x8f, 0x0a, 0x17, 0xd5, 0x7e, 0x96, 0xd7,
	0x5c, 0x43, 0x03, 0x85, 0xe8, 0x63, 0x70, 0x07, 0x53, 0x22, 0x04, 0x61, 0xb1, 0x9f, 0x07, 0x5b,
	0xbe, 0x8c, 0xc1, 0x3f, 0x8f, 0x6c, 0x16, 0x23, 0xc5, 0x07, 0x98, 0xf8, 0x63, 0x8c, 0x2d, 0xef,
	0xf6, 0xaa, 0xaf, 0x07, 0x25, 0xd6, 0x5f, 0x01, 0x20, 0x24, 0xe4, 0xd2, 0x57, 0x71, 0xd6, 0x55,
	0x9c, 0xed, 0x1b, 0x71, 0xbe, 0x58, 0xfd, 0x0c, 0xaa, 0x79, 0x6e, 0xb4, 0xd6, 0xa5, 0xca, 0x53,
	0x01, 0x39, 0xdd, 0x7d, 0x73, 0x35, 0x37, 0xb4, 0xeb, 0xb9, 0xa1, 0xfd, 0x9c, 0x1b, 0xda, 0xe5,
	0xc2, 0xa8, 0x5d, 0x2f, 0x8c, 0xda, 0xb7, 0x85, 0x51, 0x7b, 0xed, 0x6e, 0x99, 0x2f, 0xdf, 0x50,
	0x6f, 0x02, 0x03, 0xb1, 0x2a, 0x9c, 0x69, 0xff, 0xd4, 0xf9, 0xf8, 0xa7, 0xbf, 0x04, 0x75, 0xb9,
	0x60, 0x4f, 0x79, 0x7b, 0xfe, 0x2b, 0x00, 0x00, 0xff, 0xff, 0x56, 0x5e, 0xf4, 0xd1, 0x41, 0x04,
	0x00, 0x00,
}

func (m *IncentiveRecord) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *IncentiveRecord) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *IncentiveRecord) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n1, err1 := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.MinUptime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdDuration(m.MinUptime):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintIncentiveRecord(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x2a
	{
		size, err := m.IncentiveRecordBody.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintIncentiveRecord(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if len(m.IncentiveCreatorAddr) > 0 {
		i -= len(m.IncentiveCreatorAddr)
		copy(dAtA[i:], m.IncentiveCreatorAddr)
		i = encodeVarintIncentiveRecord(dAtA, i, uint64(len(m.IncentiveCreatorAddr)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.IncentiveDenom) > 0 {
		i -= len(m.IncentiveDenom)
		copy(dAtA[i:], m.IncentiveDenom)
		i = encodeVarintIncentiveRecord(dAtA, i, uint64(len(m.IncentiveDenom)))
		i--
		dAtA[i] = 0x12
	}
	if m.PoolId != 0 {
		i = encodeVarintIncentiveRecord(dAtA, i, uint64(m.PoolId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *IncentiveRecordBody) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *IncentiveRecordBody) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *IncentiveRecordBody) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n3, err3 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.StartTime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.StartTime):])
	if err3 != nil {
		return 0, err3
	}
	i -= n3
	i = encodeVarintIncentiveRecord(dAtA, i, uint64(n3))
	i--
	dAtA[i] = 0x1a
	{
		size := m.EmissionRate.Size()
		i -= size
		if _, err := m.EmissionRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintIncentiveRecord(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size := m.RemainingAmount.Size()
		i -= size
		if _, err := m.RemainingAmount.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintIncentiveRecord(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintIncentiveRecord(dAtA []byte, offset int, v uint64) int {
	offset -= sovIncentiveRecord(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *IncentiveRecord) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PoolId != 0 {
		n += 1 + sovIncentiveRecord(uint64(m.PoolId))
	}
	l = len(m.IncentiveDenom)
	if l > 0 {
		n += 1 + l + sovIncentiveRecord(uint64(l))
	}
	l = len(m.IncentiveCreatorAddr)
	if l > 0 {
		n += 1 + l + sovIncentiveRecord(uint64(l))
	}
	l = m.IncentiveRecordBody.Size()
	n += 1 + l + sovIncentiveRecord(uint64(l))
	l = github_com_gogo_protobuf_types.SizeOfStdDuration(m.MinUptime)
	n += 1 + l + sovIncentiveRecord(uint64(l))
	return n
}

func (m *IncentiveRecordBody) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.RemainingAmount.Size()
	n += 1 + l + sovIncentiveRecord(uint64(l))
	l = m.EmissionRate.Size()
	n += 1 + l + sovIncentiveRecord(uint64(l))
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.StartTime)
	n += 1 + l + sovIncentiveRecord(uint64(l))
	return n
}

func sovIncentiveRecord(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozIncentiveRecord(x uint64) (n int) {
	return sovIncentiveRecord(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *IncentiveRecord) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowIncentiveRecord
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
			return fmt.Errorf("proto: IncentiveRecord: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IncentiveRecord: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolId", wireType)
			}
			m.PoolId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIncentiveRecord
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
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IncentiveDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIncentiveRecord
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
				return ErrInvalidLengthIncentiveRecord
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIncentiveRecord
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IncentiveDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IncentiveCreatorAddr", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIncentiveRecord
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
				return ErrInvalidLengthIncentiveRecord
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIncentiveRecord
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IncentiveCreatorAddr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IncentiveRecordBody", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIncentiveRecord
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
				return ErrInvalidLengthIncentiveRecord
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthIncentiveRecord
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.IncentiveRecordBody.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinUptime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIncentiveRecord
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
				return ErrInvalidLengthIncentiveRecord
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthIncentiveRecord
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&m.MinUptime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipIncentiveRecord(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthIncentiveRecord
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
func (m *IncentiveRecordBody) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowIncentiveRecord
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
			return fmt.Errorf("proto: IncentiveRecordBody: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IncentiveRecordBody: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RemainingAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIncentiveRecord
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
				return ErrInvalidLengthIncentiveRecord
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIncentiveRecord
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.RemainingAmount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EmissionRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIncentiveRecord
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
				return ErrInvalidLengthIncentiveRecord
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIncentiveRecord
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.EmissionRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StartTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIncentiveRecord
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
				return ErrInvalidLengthIncentiveRecord
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthIncentiveRecord
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.StartTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipIncentiveRecord(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthIncentiveRecord
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
func skipIncentiveRecord(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowIncentiveRecord
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
					return 0, ErrIntOverflowIncentiveRecord
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
					return 0, ErrIntOverflowIncentiveRecord
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
				return 0, ErrInvalidLengthIncentiveRecord
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupIncentiveRecord
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthIncentiveRecord
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthIncentiveRecord        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowIncentiveRecord          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupIncentiveRecord = fmt.Errorf("proto: unexpected end of group")
)
