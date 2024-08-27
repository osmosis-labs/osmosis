// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/poolincentives/v1beta1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	_ "google.golang.org/protobuf/types/known/durationpb"
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

// GenesisState defines the pool incentives module's genesis state.
type GenesisState struct {
	// params defines all the parameters of the module.
	Params            Params          `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	LockableDurations []time.Duration `protobuf:"bytes,2,rep,name=lockable_durations,json=lockableDurations,proto3,stdduration" json:"lockable_durations" yaml:"lockable_durations"`
	DistrInfo         *DistrInfo      `protobuf:"bytes,3,opt,name=distr_info,json=distrInfo,proto3" json:"distr_info,omitempty" yaml:"distr_info"`
	// any_pool_to_internal_gauges defines the gauges for any pool to internal
	// pool. For every pool type (e.g. LP, Concentrated, etc), there is one such
	// link
	AnyPoolToInternalGauges *AnyPoolToInternalGauges `protobuf:"bytes,4,opt,name=any_pool_to_internal_gauges,json=anyPoolToInternalGauges,proto3" json:"any_pool_to_internal_gauges,omitempty" yaml:"internal_pool_to_gauges"`
	// concentrated_pool_to_no_lock_gauges defines the no lock gauges for
	// concentrated pool. This only exists between concentrated pool and no lock
	// gauges. Both external and internal gauges are included.
	ConcentratedPoolToNoLockGauges *ConcentratedPoolToNoLockGauges `protobuf:"bytes,5,opt,name=concentrated_pool_to_no_lock_gauges,json=concentratedPoolToNoLockGauges,proto3" json:"concentrated_pool_to_no_lock_gauges,omitempty" yaml:"concentrated_pool_to_no_lock_gauges"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_4400dc3495d1d6ad, []int{0}
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

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetLockableDurations() []time.Duration {
	if m != nil {
		return m.LockableDurations
	}
	return nil
}

func (m *GenesisState) GetDistrInfo() *DistrInfo {
	if m != nil {
		return m.DistrInfo
	}
	return nil
}

func (m *GenesisState) GetAnyPoolToInternalGauges() *AnyPoolToInternalGauges {
	if m != nil {
		return m.AnyPoolToInternalGauges
	}
	return nil
}

func (m *GenesisState) GetConcentratedPoolToNoLockGauges() *ConcentratedPoolToNoLockGauges {
	if m != nil {
		return m.ConcentratedPoolToNoLockGauges
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "osmosis.poolincentives.v1beta1.GenesisState")
}

func init() {
	proto.RegisterFile("osmosis/poolincentives/v1beta1/genesis.proto", fileDescriptor_4400dc3495d1d6ad)
}

var fileDescriptor_4400dc3495d1d6ad = []byte{
	// 468 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x4d, 0x6b, 0x13, 0x41,
	0x1c, 0xc6, 0x33, 0xf6, 0x05, 0xdc, 0x7a, 0xe9, 0x22, 0x98, 0x54, 0x98, 0x2d, 0x2b, 0x96, 0x2a,
	0x76, 0x96, 0x46, 0x50, 0x50, 0x10, 0x8c, 0x81, 0x52, 0x10, 0x29, 0xab, 0x5e, 0xbc, 0x2c, 0xb3,
	0x9b, 0xc9, 0x3a, 0x74, 0x32, 0xff, 0xb0, 0x33, 0x09, 0xe6, 0x3b, 0x78, 0xf0, 0xe8, 0xc5, 0xcf,
	0xe2, 0x35, 0xc7, 0x1e, 0x3d, 0x45, 0x49, 0xbe, 0x41, 0x3f, 0x81, 0xec, 0xbc, 0xc4, 0x88, 0xd2,
	0xf4, 0xb6, 0xcb, 0x3c, 0x2f, 0xbf, 0x19, 0x9e, 0xe0, 0x11, 0xa8, 0x01, 0x28, 0xae, 0x92, 0x21,
	0x80, 0xe0, 0xb2, 0x60, 0x52, 0xf3, 0x31, 0x53, 0xc9, 0xf8, 0x38, 0x67, 0x9a, 0x1e, 0x27, 0x25,
	0x93, 0x4c, 0x71, 0x45, 0x86, 0x15, 0x68, 0x08, 0xb1, 0x53, 0x93, 0xbf, 0xd5, 0xc4, 0xa9, 0xf7,
	0x6e, 0x97, 0x50, 0x82, 0x91, 0x26, 0xf5, 0x97, 0x75, 0xed, 0xe1, 0x12, 0xa0, 0x14, 0x2c, 0x31,
	0x7f, 0xf9, 0xa8, 0x9f, 0xf4, 0x46, 0x15, 0xd5, 0x1c, 0xa4, 0x3b, 0x4f, 0xd6, 0x30, 0xac, 0x14,
	0x19, 0x43, 0xfc, 0x79, 0x2b, 0xb8, 0x75, 0x62, 0xc1, 0xde, 0x6a, 0xaa, 0x59, 0xd8, 0x0d, 0xb6,
	0x87, 0xb4, 0xa2, 0x03, 0xd5, 0x44, 0xfb, 0xe8, 0x70, 0xa7, 0x7d, 0x40, 0xae, 0x06, 0x25, 0x67,
	0x46, 0xdd, 0xd9, 0x9c, 0xce, 0xa2, 0x46, 0xea, 0xbc, 0x21, 0x04, 0xa1, 0x80, 0xe2, 0x9c, 0xe6,
	0x82, 0x65, 0x1e, 0x51, 0x35, 0x6f, 0xec, 0x6f, 0x1c, 0xee, 0xb4, 0x5b, 0xc4, 0x5e, 0x82, 0xf8,
	0x4b, 0x90, 0xae, 0x53, 0x74, 0xee, 0xd7, 0x21, 0x97, 0xb3, 0xa8, 0x35, 0xa1, 0x03, 0xf1, 0x2c,
	0xfe, 0x37, 0x22, 0xfe, 0xfa, 0x33, 0x42, 0xe9, 0xae, 0x3f, 0xf0, 0x46, 0x15, 0x16, 0x41, 0xd0,
	0xe3, 0x4a, 0x57, 0x19, 0x97, 0x7d, 0x68, 0x6e, 0x18, 0xf4, 0x07, 0xeb, 0xd0, 0xbb, 0xb5, 0xe3,
	0x54, 0xf6, 0xa1, 0xd3, 0x9a, 0xce, 0x22, 0x74, 0x39, 0x8b, 0x76, 0x6d, 0xf1, 0x9f, 0xa8, 0x38,
	0xbd, 0xd9, 0xf3, 0xaa, 0xf0, 0x1b, 0x0a, 0xee, 0x52, 0x39, 0xc9, 0xea, 0xb8, 0x4c, 0x43, 0xc6,
	0xa5, 0x66, 0x95, 0xa4, 0x22, 0x2b, 0xe9, 0xa8, 0x64, 0xaa, 0xb9, 0x69, 0x6a, 0x9f, 0xae, 0xab,
	0x7d, 0x29, 0x27, 0x67, 0x00, 0xe2, 0x1d, 0x9c, 0x3a, 0xff, 0x89, 0xb1, 0x77, 0x0e, 0x1c, 0x04,
	0xb6, 0x10, 0xcb, 0x74, 0x5f, 0x67, 0x5b, 0xe2, 0xf4, 0x0e, 0xfd, 0x7f, 0x40, 0xf8, 0x1d, 0x05,
	0xf7, 0x0a, 0x30, 0x85, 0x15, 0xd5, 0xac, 0xb7, 0x74, 0x4a, 0xc8, 0xea, 0x27, 0xf3, 0x9c, 0x5b,
	0x86, 0xf3, 0xc5, 0x3a, 0xce, 0x57, 0x2b, 0x51, 0xb6, 0xef, 0x0d, 0xbc, 0x86, 0xe2, 0xdc, 0xe1,
	0xb6, 0x1d, 0xee, 0x43, 0x8b, 0x7b, 0x8d, 0xe2, 0x38, 0xc5, 0xc5, 0xd5, 0x99, 0xef, 0xa7, 0x73,
	0x8c, 0x2e, 0xe6, 0x18, 0xfd, 0x9a, 0x63, 0xf4, 0x65, 0x81, 0x1b, 0x17, 0x0b, 0xdc, 0xf8, 0xb1,
	0xc0, 0x8d, 0x0f, 0xcf, 0x4b, 0xae, 0x3f, 0x8e, 0x72, 0x52, 0xc0, 0xc0, 0x8f, 0xfc, 0x48, 0xd0,
	0x5c, 0x2d, 0x17, 0x3f, 0x6e, 0x3f, 0x49, 0x3e, 0x99, 0xdd, 0x1f, 0xad, 0x0c, 0x5f, 0x4f, 0x86,
	0x4c, 0xe5, 0xdb, 0x66, 0x6a, 0x8f, 0x7f, 0x07, 0x00, 0x00, 0xff, 0xff, 0xf2, 0x95, 0xcc, 0x60,
	0xa3, 0x03, 0x00, 0x00,
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
	if m.ConcentratedPoolToNoLockGauges != nil {
		{
			size, err := m.ConcentratedPoolToNoLockGauges.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	if m.AnyPoolToInternalGauges != nil {
		{
			size, err := m.AnyPoolToInternalGauges.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x22
	}
	if m.DistrInfo != nil {
		{
			size, err := m.DistrInfo.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if len(m.LockableDurations) > 0 {
		for iNdEx := len(m.LockableDurations) - 1; iNdEx >= 0; iNdEx-- {
			n, err := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.LockableDurations[iNdEx], dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.LockableDurations[iNdEx]):])
			if err != nil {
				return 0, err
			}
			i -= n
			i = encodeVarintGenesis(dAtA, i, uint64(n))
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
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
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.LockableDurations) > 0 {
		for _, e := range m.LockableDurations {
			l = github_com_cosmos_gogoproto_types.SizeOfStdDuration(e)
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if m.DistrInfo != nil {
		l = m.DistrInfo.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.AnyPoolToInternalGauges != nil {
		l = m.AnyPoolToInternalGauges.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.ConcentratedPoolToNoLockGauges != nil {
		l = m.ConcentratedPoolToNoLockGauges.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
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
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
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
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LockableDurations", wireType)
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
			m.LockableDurations = append(m.LockableDurations, time.Duration(0))
			if err := github_com_cosmos_gogoproto_types.StdDurationUnmarshal(&(m.LockableDurations[len(m.LockableDurations)-1]), dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DistrInfo", wireType)
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
			if m.DistrInfo == nil {
				m.DistrInfo = &DistrInfo{}
			}
			if err := m.DistrInfo.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AnyPoolToInternalGauges", wireType)
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
			if m.AnyPoolToInternalGauges == nil {
				m.AnyPoolToInternalGauges = &AnyPoolToInternalGauges{}
			}
			if err := m.AnyPoolToInternalGauges.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConcentratedPoolToNoLockGauges", wireType)
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
			if m.ConcentratedPoolToNoLockGauges == nil {
				m.ConcentratedPoolToNoLockGauges = &ConcentratedPoolToNoLockGauges{}
			}
			if err := m.ConcentratedPoolToNoLockGauges.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
