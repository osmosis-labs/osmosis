// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/downtimedetector/v1beta1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/codec/types"
	proto "github.com/cosmos/gogoproto/proto"
	_ "github.com/cosmos/gogoproto/types"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	
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

type GenesisDowntimeEntry struct {
	Duration     Downtime  `protobuf:"varint,1,opt,name=duration,proto3,enum=osmosis.downtimedetector.v1beta1.Downtime" json:"duration,omitempty" yaml:"duration"`
	LastDowntime time.Time `protobuf:"bytes,2,opt,name=last_downtime,json=lastDowntime,proto3,stdtime" json:"last_downtime" yaml:"last_downtime"`
}

func (m *GenesisDowntimeEntry) Reset()         { *m = GenesisDowntimeEntry{} }
func (m *GenesisDowntimeEntry) String() string { return proto.CompactTextString(m) }
func (*GenesisDowntimeEntry) ProtoMessage()    {}
func (*GenesisDowntimeEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_3d44d4cc05d2cb13, []int{0}
}
func (m *GenesisDowntimeEntry) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisDowntimeEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisDowntimeEntry.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisDowntimeEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisDowntimeEntry.Merge(m, src)
}
func (m *GenesisDowntimeEntry) XXX_Size() int {
	return m.Size()
}
func (m *GenesisDowntimeEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisDowntimeEntry.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisDowntimeEntry proto.InternalMessageInfo

func (m *GenesisDowntimeEntry) GetDuration() Downtime {
	if m != nil {
		return m.Duration
	}
	return Downtime_DURATION_30S
}

func (m *GenesisDowntimeEntry) GetLastDowntime() time.Time {
	if m != nil {
		return m.LastDowntime
	}
	return time.Time{}
}

// GenesisState defines the twap module's genesis state.
type GenesisState struct {
	Downtimes     []GenesisDowntimeEntry `protobuf:"bytes,1,rep,name=downtimes,proto3" json:"downtimes"`
	LastBlockTime time.Time              `protobuf:"bytes,2,opt,name=last_block_time,json=lastBlockTime,proto3,stdtime" json:"last_block_time" yaml:"last_block_time"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_3d44d4cc05d2cb13, []int{1}
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

func (m *GenesisState) GetDowntimes() []GenesisDowntimeEntry {
	if m != nil {
		return m.Downtimes
	}
	return nil
}

func (m *GenesisState) GetLastBlockTime() time.Time {
	if m != nil {
		return m.LastBlockTime
	}
	return time.Time{}
}

func init() {
	proto.RegisterType((*GenesisDowntimeEntry)(nil), "osmosis.downtimedetector.v1beta1.GenesisDowntimeEntry")
	proto.RegisterType((*GenesisState)(nil), "osmosis.downtimedetector.v1beta1.GenesisState")
}

func init() {
	proto.RegisterFile("osmosis/downtimedetector/v1beta1/genesis.proto", fileDescriptor_3d44d4cc05d2cb13)
}

var fileDescriptor_3d44d4cc05d2cb13 = []byte{
	// 406 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xc1, 0x6a, 0xe2, 0x40,
	0x1c, 0xc6, 0x33, 0xbb, 0xcb, 0xb2, 0x1b, 0xdd, 0x15, 0xb2, 0xb2, 0xb8, 0x1e, 0x92, 0x90, 0x93,
	0x2c, 0x38, 0x53, 0x53, 0x28, 0xa5, 0xd0, 0x4b, 0x68, 0xe9, 0xdd, 0x16, 0x0a, 0xf6, 0x10, 0x26,
	0x71, 0x4c, 0x43, 0x93, 0x8c, 0x64, 0x46, 0xdb, 0xbc, 0x85, 0x8f, 0xe5, 0x51, 0x7a, 0x28, 0x3d,
	0xd9, 0xa2, 0x6f, 0xe0, 0x13, 0x94, 0x24, 0x33, 0x5a, 0xa5, 0x60, 0x6f, 0xf9, 0xe7, 0xff, 0x7d,
	0x1f, 0xdf, 0xef, 0xcf, 0xa8, 0x90, 0xb2, 0x98, 0xb2, 0x90, 0xa1, 0x3e, 0xbd, 0x4f, 0x78, 0x18,
	0x93, 0x3e, 0xe1, 0xc4, 0xe7, 0x34, 0x45, 0xe3, 0x8e, 0x47, 0x38, 0xee, 0xa0, 0x80, 0x24, 0x84,
	0x85, 0x0c, 0x0e, 0x53, 0xca, 0xa9, 0x66, 0x0a, 0x3d, 0xdc, 0xd5, 0x43, 0xa1, 0x6f, 0xd6, 0x03,
	0x1a, 0xd0, 0x42, 0x8c, 0xf2, 0xaf, 0xd2, 0xd7, 0xfc, 0x17, 0x50, 0x1a, 0x44, 0x04, 0x15, 0x93,
	0x37, 0x1a, 0x20, 0x9c, 0x64, 0x72, 0xe5, 0x17, 0x99, 0x6e, 0xe9, 0x29, 0x07, 0xb1, 0xd2, 0x77,
	0x5d, 0xfd, 0x51, 0x8a, 0x79, 0x48, 0x13, 0xb1, 0x37, 0x76, 0xf7, 0x79, 0x23, 0xc6, 0x71, 0x3c,
	0x14, 0x82, 0xe3, 0xbd, 0x78, 0x72, 0xe1, 0x6e, 0x47, 0x5b, 0x4f, 0x40, 0xad, 0x5f, 0x94, 0xe8,
	0x67, 0x42, 0x72, 0x9e, 0xf0, 0x34, 0xd3, 0x6e, 0xd4, 0x1f, 0x52, 0xda, 0x00, 0x26, 0x68, 0xfd,
	0xb6, 0xff, 0xc3, 0x7d, 0x47, 0x81, 0x32, 0xc2, 0xf9, 0xb3, 0x9a, 0x1b, 0xb5, 0x0c, 0xc7, 0xd1,
	0x89, 0x25, 0x53, 0xac, 0xee, 0x3a, 0x50, 0xc3, 0xea, 0xaf, 0x08, 0x33, 0xee, 0xca, 0xa0, 0xc6,
	0x17, 0x13, 0xb4, 0x2a, 0x76, 0x13, 0x96, 0xa0, 0x50, 0x82, 0xc2, 0x2b, 0x09, 0xea, 0x98, 0xd3,
	0xb9, 0xa1, 0xac, 0xe6, 0x46, 0xbd, 0x4c, 0xdd, 0xb2, 0x5b, 0x93, 0x17, 0x03, 0x74, 0xab, 0xf9,
	0x3f, 0xd9, 0xc0, 0x7a, 0x04, 0x6a, 0x55, 0x80, 0x5d, 0x72, 0xcc, 0x89, 0xd6, 0x53, 0x7f, 0x4a,
	0x3d, 0x6b, 0x00, 0xf3, 0x6b, 0xab, 0x62, 0x1f, 0xed, 0x27, 0xfa, 0xe8, 0x36, 0xce, 0xb7, 0xbc,
	0x4b, 0x77, 0x13, 0xa7, 0x0d, 0xd4, 0x5a, 0x51, 0xc8, 0x8b, 0xa8, 0x7f, 0xe7, 0x7e, 0x92, 0xc8,
	0x12, 0x44, 0x7f, 0xdf, 0x11, 0x6d, 0x02, 0x4a, 0xa6, 0xe2, 0x4c, 0x4e, 0xfe, 0x33, 0xf7, 0x39,
	0xd7, 0xd3, 0x85, 0x0e, 0x66, 0x0b, 0x1d, 0xbc, 0x2e, 0x74, 0x30, 0x59, 0xea, 0xca, 0x6c, 0xa9,
	0x2b, 0xcf, 0x4b, 0x5d, 0xe9, 0x9d, 0x06, 0x21, 0xbf, 0x1d, 0x79, 0xd0, 0xa7, 0x31, 0x12, 0x50,
	0xed, 0x08, 0x7b, 0x4c, 0x0e, 0x68, 0x6c, 0x1f, 0xa0, 0x87, 0xf5, 0x33, 0x68, 0xaf, 0x1f, 0x08,
	0xcf, 0x86, 0x84, 0x79, 0xdf, 0x8b, 0x7e, 0x87, 0x6f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x26, 0x9f,
	0x28, 0x6b, 0x28, 0x03, 0x00, 0x00,
}

func (m *GenesisDowntimeEntry) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisDowntimeEntry) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisDowntimeEntry) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n1, err1 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.LastDowntime, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LastDowntime):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintGenesis(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x12
	if m.Duration != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Duration))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
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
	n2, err2 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.LastBlockTime, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LastBlockTime):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintGenesis(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x12
	if len(m.Downtimes) > 0 {
		for iNdEx := len(m.Downtimes) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Downtimes[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
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
func (m *GenesisDowntimeEntry) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Duration != 0 {
		n += 1 + sovGenesis(uint64(m.Duration))
	}
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LastDowntime)
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Downtimes) > 0 {
		for _, e := range m.Downtimes {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LastBlockTime)
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisDowntimeEntry) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: GenesisDowntimeEntry: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisDowntimeEntry: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Duration", wireType)
			}
			m.Duration = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Duration |= Downtime(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastDowntime", wireType)
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
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.LastDowntime, dAtA[iNdEx:postIndex]); err != nil {
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
				return fmt.Errorf("proto: wrong wireType = %d for field Downtimes", wireType)
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
			m.Downtimes = append(m.Downtimes, GenesisDowntimeEntry{})
			if err := m.Downtimes[len(m.Downtimes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastBlockTime", wireType)
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
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.LastBlockTime, dAtA[iNdEx:postIndex]); err != nil {
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
