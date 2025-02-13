// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/incentives/params.proto

package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
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

// Params holds parameters for the incentives module
type Params struct {
	// distr_epoch_identifier is what epoch type distribution will be triggered by
	// (day, week, etc.)
	DistrEpochIdentifier string `protobuf:"bytes,1,opt,name=distr_epoch_identifier,json=distrEpochIdentifier,proto3" json:"distr_epoch_identifier,omitempty" yaml:"distr_epoch_identifier"`
	// group_creation_fee is the fee required to create a new group
	// It is only charged to all addresses other than incentive module account
	// or addresses in the unrestricted_creator_whitelist
	GroupCreationFee github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,2,rep,name=group_creation_fee,json=groupCreationFee,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"group_creation_fee"`
	// unrestricted_creator_whitelist is a list of addresses that are
	// allowed to bypass restrictions on permissionless Group
	// creation. In the future, we might expand these to creating gauges
	// as well.
	// The goal of this is to allow a subdao to manage incentives efficiently
	// without being stopped by 5 day governance process or a high fee.
	// At the same time, it prevents spam by having a fee for all
	// other users.
	UnrestrictedCreatorWhitelist []string `protobuf:"bytes,3,rep,name=unrestricted_creator_whitelist,json=unrestrictedCreatorWhitelist,proto3" json:"unrestricted_creator_whitelist,omitempty" yaml:"unrestricted_creator_whitelist"`
	// internal_uptime is the uptime used for internal incentives on pools that
	// use NoLock gauges (currently only Concentrated Liquidity pools).
	//
	// Since Group gauges route through internal gauges, this parameter affects
	// the uptime of those incentives as well (i.e. distributions through volume
	// splitting incentives will use this uptime).
	InternalUptime time.Duration `protobuf:"bytes,4,opt,name=internal_uptime,json=internalUptime,proto3,stdduration" json:"internal_uptime" yaml:"internal_uptime"`
	// min_value_for_distribution is the minimum amount a token must be worth
	// in order to be eligible for distribution. If the token is worth
	// less than this amount (or the route between the two denoms is not
	// registered), it will not be distributed and is forfeited to the remaining
	// distributees that are eligible.
	MinValueForDistribution types.Coin `protobuf:"bytes,5,opt,name=min_value_for_distribution,json=minValueForDistribution,proto3" json:"min_value_for_distribution"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_1cc8b460d089f845, []int{0}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetDistrEpochIdentifier() string {
	if m != nil {
		return m.DistrEpochIdentifier
	}
	return ""
}

func (m *Params) GetGroupCreationFee() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.GroupCreationFee
	}
	return nil
}

func (m *Params) GetUnrestrictedCreatorWhitelist() []string {
	if m != nil {
		return m.UnrestrictedCreatorWhitelist
	}
	return nil
}

func (m *Params) GetInternalUptime() time.Duration {
	if m != nil {
		return m.InternalUptime
	}
	return 0
}

func (m *Params) GetMinValueForDistribution() types.Coin {
	if m != nil {
		return m.MinValueForDistribution
	}
	return types.Coin{}
}

func init() {
	proto.RegisterType((*Params)(nil), "osmosis.incentives.Params")
}

func init() { proto.RegisterFile("osmosis/incentives/params.proto", fileDescriptor_1cc8b460d089f845) }

var fileDescriptor_1cc8b460d089f845 = []byte{
	// 475 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x52, 0xdd, 0x8a, 0xd3, 0x40,
	0x18, 0x6d, 0x6c, 0x5d, 0xd8, 0x08, 0x2a, 0x61, 0x59, 0x63, 0xd1, 0xa4, 0x06, 0x84, 0x7a, 0xb1,
	0x19, 0x77, 0x05, 0x41, 0x2f, 0xd3, 0x75, 0xc1, 0xbb, 0x25, 0xa0, 0x0b, 0x22, 0x84, 0xfc, 0x7c,
	0x49, 0x3f, 0x4c, 0xf2, 0x85, 0x99, 0x49, 0xb5, 0x6f, 0x21, 0x78, 0xe3, 0x33, 0xf8, 0x24, 0x7b,
	0xb9, 0x97, 0x5e, 0x75, 0xa5, 0x7d, 0x83, 0x3e, 0x81, 0x64, 0x92, 0x68, 0x11, 0xf5, 0x2a, 0x99,
	0x73, 0xce, 0x37, 0x67, 0xce, 0x9c, 0xd1, 0x6d, 0x12, 0x05, 0x09, 0x14, 0x0c, 0xcb, 0x18, 0x4a,
	0x89, 0x0b, 0x10, 0xac, 0x0a, 0x79, 0x58, 0x08, 0xb7, 0xe2, 0x24, 0xc9, 0x30, 0x3a, 0x81, 0xfb,
	0x5b, 0x30, 0x3e, 0xc8, 0x28, 0x23, 0x45, 0xb3, 0xe6, 0xaf, 0x55, 0x8e, 0xad, 0x58, 0x49, 0x59,
	0x14, 0x0a, 0x60, 0x8b, 0xe3, 0x08, 0x64, 0x78, 0xcc, 0x62, 0xc2, 0xb2, 0xe7, 0x33, 0xa2, 0x2c,
	0x07, 0xa6, 0x56, 0x51, 0x9d, 0xb2, 0xa4, 0xe6, 0xa1, 0x44, 0xea, 0x78, 0xe7, 0xcb, 0x48, 0xdf,
	0x3b, 0x57, 0xd6, 0xc6, 0x85, 0x7e, 0x98, 0xa0, 0x90, 0x3c, 0x80, 0x8a, 0xe2, 0x79, 0x80, 0x49,
	0xe3, 0x9c, 0x22, 0x70, 0x53, 0x9b, 0x68, 0xd3, 0x7d, 0xef, 0xd1, 0x76, 0x65, 0x3f, 0x5c, 0x86,
	0x45, 0xfe, 0xd2, 0xf9, 0xbb, 0xce, 0xf1, 0x0f, 0x14, 0xf1, 0xaa, 0xc1, 0x5f, 0xff, 0x82, 0x8d,
	0xa5, 0x6e, 0x64, 0x9c, 0xea, 0x2a, 0x88, 0x39, 0x28, 0xef, 0x20, 0x05, 0x30, 0x6f, 0x4c, 0x86,
	0xd3, 0x5b, 0x27, 0xf7, 0xdd, 0x36, 0x80, 0xdb, 0x04, 0x70, 0xbb, 0x00, 0xee, 0x8c, 0xb0, 0xf4,
	0x9e, 0x5e, 0xae, 0xec, 0xc1, 0xb7, 0x6b, 0x7b, 0x9a, 0xa1, 0x9c, 0xd7, 0x91, 0x1b, 0x53, 0xc1,
	0xba, 0xb4, 0xed, 0xe7, 0x48, 0x24, 0x1f, 0x98, 0x5c, 0x56, 0x20, 0xd4, 0x80, 0xf0, 0xef, 0x2a,
	0x9b, 0x59, 0xe7, 0x72, 0x06, 0x60, 0x90, 0x6e, 0xd5, 0x25, 0x07, 0x21, 0x39, 0xc6, 0x12, 0x92,
	0xf6, 0x04, 0xc4, 0x83, 0x8f, 0x73, 0x94, 0x90, 0xa3, 0x90, 0xe6, 0x70, 0x32, 0x9c, 0xee, 0x7b,
	0x4f, 0xb6, 0x2b, 0xfb, 0x71, 0x9b, 0xed, 0xff, 0x7a, 0xc7, 0x7f, 0xb0, 0x2b, 0x98, 0xb5, 0xfc,
	0x45, 0x4f, 0x1b, 0xa9, 0x7e, 0x07, 0x4b, 0x09, 0xbc, 0x0c, 0xf3, 0xa0, 0xae, 0x24, 0x16, 0x60,
	0x8e, 0x26, 0x9a, 0x0a, 0xda, 0x36, 0xe1, 0xf6, 0x4d, 0xb8, 0xa7, 0x5d, 0x13, 0x9e, 0xd3, 0x04,
	0xdd, 0xae, 0xec, 0xc3, 0xf6, 0x00, 0x7f, 0xcc, 0x3b, 0x5f, 0xaf, 0x6d, 0xcd, 0xbf, 0xdd, 0xa3,
	0x6f, 0x14, 0x68, 0xbc, 0xd7, 0xc7, 0x05, 0x96, 0xc1, 0x22, 0xcc, 0x6b, 0x08, 0x52, 0xe2, 0x81,
	0xba, 0x79, 0x8c, 0xea, 0x66, 0x47, 0xf3, 0x66, 0x67, 0xf9, 0xcf, 0xbb, 0x1d, 0x35, 0x96, 0xfe,
	0xbd, 0x02, 0xcb, 0xb7, 0xcd, 0x0e, 0x67, 0xc4, 0x4f, 0x77, 0xe6, 0xbd, 0xf3, 0xcb, 0xb5, 0xa5,
	0x5d, 0xad, 0x2d, 0xed, 0xc7, 0xda, 0xd2, 0x3e, 0x6f, 0xac, 0xc1, 0xd5, 0xc6, 0x1a, 0x7c, 0xdf,
	0x58, 0x83, 0x77, 0xcf, 0x77, 0xca, 0xe8, 0x1e, 0xe9, 0x51, 0x1e, 0x46, 0xa2, 0x5f, 0xb0, 0xc5,
	0xc9, 0x0b, 0xf6, 0x69, 0xf7, 0x61, 0xab, 0x82, 0xa2, 0x3d, 0x15, 0xfb, 0xd9, 0xcf, 0x00, 0x00,
	0x00, 0xff, 0xff, 0xd8, 0x39, 0x35, 0x44, 0xfb, 0x02, 0x00, 0x00,
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.MinValueForDistribution.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	n2, err2 := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.InternalUptime, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.InternalUptime):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintParams(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x22
	if len(m.UnrestrictedCreatorWhitelist) > 0 {
		for iNdEx := len(m.UnrestrictedCreatorWhitelist) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.UnrestrictedCreatorWhitelist[iNdEx])
			copy(dAtA[i:], m.UnrestrictedCreatorWhitelist[iNdEx])
			i = encodeVarintParams(dAtA, i, uint64(len(m.UnrestrictedCreatorWhitelist[iNdEx])))
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.GroupCreationFee) > 0 {
		for iNdEx := len(m.GroupCreationFee) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.GroupCreationFee[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintParams(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.DistrEpochIdentifier) > 0 {
		i -= len(m.DistrEpochIdentifier)
		copy(dAtA[i:], m.DistrEpochIdentifier)
		i = encodeVarintParams(dAtA, i, uint64(len(m.DistrEpochIdentifier)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintParams(dAtA []byte, offset int, v uint64) int {
	offset -= sovParams(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.DistrEpochIdentifier)
	if l > 0 {
		n += 1 + l + sovParams(uint64(l))
	}
	if len(m.GroupCreationFee) > 0 {
		for _, e := range m.GroupCreationFee {
			l = e.Size()
			n += 1 + l + sovParams(uint64(l))
		}
	}
	if len(m.UnrestrictedCreatorWhitelist) > 0 {
		for _, s := range m.UnrestrictedCreatorWhitelist {
			l = len(s)
			n += 1 + l + sovParams(uint64(l))
		}
	}
	l = github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.InternalUptime)
	n += 1 + l + sovParams(uint64(l))
	l = m.MinValueForDistribution.Size()
	n += 1 + l + sovParams(uint64(l))
	return n
}

func sovParams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozParams(x uint64) (n int) {
	return sovParams(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowParams
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DistrEpochIdentifier", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DistrEpochIdentifier = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GroupCreationFee", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.GroupCreationFee = append(m.GroupCreationFee, types.Coin{})
			if err := m.GroupCreationFee[len(m.GroupCreationFee)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UnrestrictedCreatorWhitelist", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.UnrestrictedCreatorWhitelist = append(m.UnrestrictedCreatorWhitelist, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InternalUptime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdDurationUnmarshal(&m.InternalUptime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinValueForDistribution", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.MinValueForDistribution.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipParams(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthParams
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
func skipParams(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowParams
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
					return 0, ErrIntOverflowParams
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
					return 0, ErrIntOverflowParams
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
				return 0, ErrInvalidLengthParams
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupParams
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthParams
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthParams        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowParams          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupParams = fmt.Errorf("proto: unexpected end of group")
)
