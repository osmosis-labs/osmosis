// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/concentrated-liquidity/params.proto

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

type Params struct {
	// authorized_tick_spacing is an array of uint64s that represents the tick
	// spacing values concentrated-liquidity pools can be created with. For
	// example, an authorized_tick_spacing of [1, 10, 30] allows for pools
	// to be created with tick spacing of 1, 10, or 30.
	AuthorizedTickSpacing   []uint64                                 `protobuf:"varint,1,rep,packed,name=authorized_tick_spacing,json=authorizedTickSpacing,proto3" json:"authorized_tick_spacing,omitempty" yaml:"authorized_tick_spacing"`
	AuthorizedSpreadFactors []github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,2,rep,name=authorized_spread_factors,json=authorizedSpreadFactors,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"authorized_spread_factors" yaml:"authorized_spread_factors"`
	// balancer_shares_reward_discount is the rate by which incentives flowing
	// from CL to Balancer pools will be discounted to encourage LPs to migrate.
	// e.g. a rate of 0.05 means Balancer LPs get 5% less incentives than full
	// range CL LPs.
	// This field can range from (0,1]. If set to 1, it indicates that all
	// incentives stay at cl pool.
	BalancerSharesRewardDiscount github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,3,opt,name=balancer_shares_reward_discount,json=balancerSharesRewardDiscount,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"balancer_shares_reward_discount" yaml:"balancer_shares_reward_discount"`
	// authorized_quote_denoms is a list of quote denoms that can be used as
	// token1 when creating a pool. We limit the quote assets to a small set for
	// the purposes of having convinient price increments stemming from tick to
	// price conversion. These increments are in a human readable magnitude only
	// for token1 as a quote. For limit orders in the future, this will be a
	// desirable property in terms of UX as to allow users to set limit orders at
	// prices in terms of token1 (quote asset) that are easy to reason about.
	AuthorizedQuoteDenoms []string        `protobuf:"bytes,4,rep,name=authorized_quote_denoms,json=authorizedQuoteDenoms,proto3" json:"authorized_quote_denoms,omitempty" yaml:"authorized_quote_denoms"`
	AuthorizedUptimes     []time.Duration `protobuf:"bytes,5,rep,name=authorized_uptimes,json=authorizedUptimes,proto3,stdduration" json:"duration,omitempty" yaml:"authorized_uptimes"`
	// is_permissionless_pool_creation_enabled is a boolean that determines if
	// concentrated liquidity pools can be created via message. At launch,
	// we consider allowing only governance to create pools, and then later
	// allowing permissionless pool creation by switching this flag to true
	// with a governance proposal.
	IsPermissionlessPoolCreationEnabled bool `protobuf:"varint,6,opt,name=is_permissionless_pool_creation_enabled,json=isPermissionlessPoolCreationEnabled,proto3" json:"is_permissionless_pool_creation_enabled,omitempty" yaml:"is_permissionless_pool_creation_enabled"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd3784445b6f6ba7, []int{0}
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

func (m *Params) GetAuthorizedTickSpacing() []uint64 {
	if m != nil {
		return m.AuthorizedTickSpacing
	}
	return nil
}

func (m *Params) GetAuthorizedQuoteDenoms() []string {
	if m != nil {
		return m.AuthorizedQuoteDenoms
	}
	return nil
}

func (m *Params) GetAuthorizedUptimes() []time.Duration {
	if m != nil {
		return m.AuthorizedUptimes
	}
	return nil
}

func (m *Params) GetIsPermissionlessPoolCreationEnabled() bool {
	if m != nil {
		return m.IsPermissionlessPoolCreationEnabled
	}
	return false
}

func init() {
	proto.RegisterType((*Params)(nil), "osmosis.concentratedliquidity.Params")
}

func init() {
	proto.RegisterFile("osmosis/concentrated-liquidity/params.proto", fileDescriptor_cd3784445b6f6ba7)
}

var fileDescriptor_cd3784445b6f6ba7 = []byte{
	// 546 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0x31, 0x6f, 0xd3, 0x40,
	0x14, 0xc7, 0x63, 0x52, 0x22, 0x6a, 0x26, 0x2c, 0x10, 0x4e, 0x05, 0xb6, 0x65, 0xa4, 0x12, 0x09,
	0x6a, 0x8b, 0x32, 0x01, 0x9b, 0x09, 0xcc, 0xc5, 0x01, 0x09, 0x55, 0x48, 0xa7, 0xf3, 0xf9, 0xea,
	0x9c, 0x62, 0xfb, 0xb9, 0x77, 0x67, 0x20, 0xec, 0x20, 0x26, 0xc4, 0xc8, 0xc6, 0xd7, 0xe9, 0xd8,
	0x11, 0x31, 0x18, 0x94, 0x6c, 0x8c, 0xf9, 0x04, 0x28, 0x67, 0x87, 0x3a, 0x94, 0x0a, 0x98, 0xec,
	0xf7, 0xfe, 0xbf, 0xf7, 0xee, 0xdd, 0x7b, 0xef, 0xf4, 0x5b, 0x20, 0x32, 0x10, 0x4c, 0xf8, 0x04,
	0x72, 0x42, 0x73, 0xc9, 0xb1, 0xa4, 0xf1, 0x4e, 0xca, 0x0e, 0x4b, 0x16, 0x33, 0x39, 0xf5, 0x0b,
	0xcc, 0x71, 0x26, 0xbc, 0x82, 0x83, 0x04, 0xe3, 0x7a, 0x03, 0x7b, 0x6d, 0xf8, 0x17, 0xbb, 0x75,
	0x39, 0x81, 0x04, 0x14, 0xe9, 0x2f, 0xff, 0xea, 0xa0, 0xad, 0x3e, 0x51, 0x51, 0xa8, 0x16, 0x6a,
	0xa3, 0x91, 0xac, 0x04, 0x20, 0x49, 0xa9, 0xaf, 0xac, 0xa8, 0x3c, 0xf0, 0xe3, 0x92, 0x63, 0xc9,
	0x20, 0xaf, 0x75, 0xf7, 0x5d, 0x4f, 0xef, 0xed, 0xa9, 0x02, 0x8c, 0x7d, 0xfd, 0x2a, 0x2e, 0xe5,
	0x18, 0x38, 0x7b, 0x43, 0x63, 0x24, 0x19, 0x99, 0x20, 0x51, 0x60, 0xc2, 0xf2, 0xc4, 0xd4, 0x9c,
	0xee, 0x60, 0x23, 0x70, 0x17, 0x95, 0x6d, 0x4d, 0x71, 0x96, 0xde, 0x77, 0xcf, 0x00, 0xdd, 0xf0,
	0xca, 0x89, 0xf2, 0x94, 0x91, 0xc9, 0xa8, 0xf6, 0x1b, 0x1f, 0x34, 0xbd, 0xdf, 0x8a, 0x11, 0x05,
	0xa7, 0x38, 0x46, 0x07, 0x98, 0x48, 0xe0, 0xc2, 0x3c, 0xe7, 0x74, 0x07, 0x9b, 0x41, 0x78, 0x54,
	0xd9, 0x9d, 0xaf, 0x95, 0xbd, 0x9d, 0x30, 0x39, 0x2e, 0x23, 0x8f, 0x40, 0xd6, 0xdc, 0xa5, 0xf9,
	0xec, 0x88, 0x78, 0xe2, 0xcb, 0x69, 0x41, 0x85, 0x37, 0xa4, 0x64, 0x51, 0xd9, 0xce, 0xa9, 0x62,
	0xd6, 0x13, 0xbb, 0x61, 0xeb, 0x46, 0x23, 0x25, 0x3d, 0xae, 0x15, 0xe3, 0xb3, 0xa6, 0xdb, 0x11,
	0x4e, 0x71, 0x4e, 0x28, 0x47, 0x62, 0x8c, 0x39, 0x15, 0x88, 0xd3, 0x57, 0x98, 0xc7, 0x28, 0x66,
	0x82, 0x40, 0x99, 0x4b, 0xb3, 0xeb, 0x68, 0x83, 0xcd, 0xe0, 0xf9, 0x7f, 0x97, 0xb5, 0x5d, 0x97,
	0xf5, 0x97, 0xf4, 0x6e, 0x78, 0x6d, 0x45, 0x8c, 0x14, 0x10, 0x2a, 0x7d, 0xd8, 0xc8, 0xbf, 0x8d,
	0xe3, 0xb0, 0x04, 0x49, 0x51, 0x4c, 0x73, 0xc8, 0x84, 0xb9, 0xa1, 0xfa, 0xf5, 0xe7, 0x71, 0xb4,
	0xc1, 0xb5, 0x71, 0x3c, 0x59, 0x0a, 0x43, 0xe5, 0x37, 0xde, 0x6a, 0xba, 0xd1, 0x8a, 0x29, 0x0b,
	0xc9, 0x32, 0x2a, 0xcc, 0xf3, 0x4e, 0x77, 0x70, 0x71, 0xb7, 0xef, 0xd5, 0x3b, 0xe3, 0xad, 0x76,
	0xc6, 0x1b, 0x36, 0x3b, 0x13, 0x3c, 0x58, 0xf6, 0xe2, 0x47, 0x65, 0x1b, 0xab, 0x2d, 0xba, 0x0d,
	0x19, 0x93, 0x34, 0x2b, 0xe4, 0x74, 0x51, 0xd9, 0xfd, 0x53, 0xc5, 0x34, 0x89, 0xdd, 0x4f, 0xdf,
	0x6c, 0x2d, 0xbc, 0x74, 0x22, 0x3c, 0xab, 0xfd, 0xc6, 0x7b, 0x4d, 0xbf, 0xc9, 0x04, 0x2a, 0x28,
	0xcf, 0x98, 0x10, 0x0c, 0xf2, 0x94, 0x0a, 0x81, 0x0a, 0x80, 0x14, 0x11, 0x4e, 0xd5, 0x09, 0x88,
	0xe6, 0x38, 0x4a, 0x69, 0x6c, 0xf6, 0x1c, 0x6d, 0x70, 0x21, 0xd8, 0x5d, 0x54, 0xb6, 0x57, 0x9f,
	0xf3, 0x8f, 0x81, 0x6e, 0x78, 0x83, 0x89, 0xbd, 0x35, 0x70, 0x0f, 0x20, 0x7d, 0xd8, 0x60, 0x8f,
	0x6a, 0x2a, 0x78, 0x71, 0x34, 0xb3, 0xb4, 0xe3, 0x99, 0xa5, 0x7d, 0x9f, 0x59, 0xda, 0xc7, 0xb9,
	0xd5, 0x39, 0x9e, 0x5b, 0x9d, 0x2f, 0x73, 0xab, 0xb3, 0x1f, 0xb4, 0x06, 0xdf, 0xbc, 0xce, 0x9d,
	0x14, 0x47, 0x62, 0x65, 0xf8, 0x2f, 0xef, 0xdc, 0xf3, 0x5f, 0x9f, 0xf5, 0xba, 0xd5, 0x62, 0x44,
	0x3d, 0xd5, 0xcb, 0xbb, 0x3f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x0a, 0xfa, 0x5f, 0x9a, 0x0c, 0x04,
	0x00, 0x00,
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
	if m.IsPermissionlessPoolCreationEnabled {
		i--
		if m.IsPermissionlessPoolCreationEnabled {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x30
	}
	if len(m.AuthorizedUptimes) > 0 {
		for iNdEx := len(m.AuthorizedUptimes) - 1; iNdEx >= 0; iNdEx-- {
			n, err := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.AuthorizedUptimes[iNdEx], dAtA[i-github_com_gogo_protobuf_types.SizeOfStdDuration(m.AuthorizedUptimes[iNdEx]):])
			if err != nil {
				return 0, err
			}
			i -= n
			i = encodeVarintParams(dAtA, i, uint64(n))
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.AuthorizedQuoteDenoms) > 0 {
		for iNdEx := len(m.AuthorizedQuoteDenoms) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.AuthorizedQuoteDenoms[iNdEx])
			copy(dAtA[i:], m.AuthorizedQuoteDenoms[iNdEx])
			i = encodeVarintParams(dAtA, i, uint64(len(m.AuthorizedQuoteDenoms[iNdEx])))
			i--
			dAtA[i] = 0x22
		}
	}
	{
		size := m.BalancerSharesRewardDiscount.Size()
		i -= size
		if _, err := m.BalancerSharesRewardDiscount.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.AuthorizedSpreadFactors) > 0 {
		for iNdEx := len(m.AuthorizedSpreadFactors) - 1; iNdEx >= 0; iNdEx-- {
			{
				size := m.AuthorizedSpreadFactors[iNdEx].Size()
				i -= size
				if _, err := m.AuthorizedSpreadFactors[iNdEx].MarshalTo(dAtA[i:]); err != nil {
					return 0, err
				}
				i = encodeVarintParams(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.AuthorizedTickSpacing) > 0 {
		dAtA2 := make([]byte, len(m.AuthorizedTickSpacing)*10)
		var j1 int
		for _, num := range m.AuthorizedTickSpacing {
			for num >= 1<<7 {
				dAtA2[j1] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j1++
			}
			dAtA2[j1] = uint8(num)
			j1++
		}
		i -= j1
		copy(dAtA[i:], dAtA2[:j1])
		i = encodeVarintParams(dAtA, i, uint64(j1))
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
	if len(m.AuthorizedTickSpacing) > 0 {
		l = 0
		for _, e := range m.AuthorizedTickSpacing {
			l += sovParams(uint64(e))
		}
		n += 1 + sovParams(uint64(l)) + l
	}
	if len(m.AuthorizedSpreadFactors) > 0 {
		for _, e := range m.AuthorizedSpreadFactors {
			l = e.Size()
			n += 1 + l + sovParams(uint64(l))
		}
	}
	l = m.BalancerSharesRewardDiscount.Size()
	n += 1 + l + sovParams(uint64(l))
	if len(m.AuthorizedQuoteDenoms) > 0 {
		for _, s := range m.AuthorizedQuoteDenoms {
			l = len(s)
			n += 1 + l + sovParams(uint64(l))
		}
	}
	if len(m.AuthorizedUptimes) > 0 {
		for _, e := range m.AuthorizedUptimes {
			l = github_com_gogo_protobuf_types.SizeOfStdDuration(e)
			n += 1 + l + sovParams(uint64(l))
		}
	}
	if m.IsPermissionlessPoolCreationEnabled {
		n += 2
	}
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
			if wireType == 0 {
				var v uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowParams
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.AuthorizedTickSpacing = append(m.AuthorizedTickSpacing, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowParams
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthParams
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthParams
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA[iNdEx:postIndex] {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.AuthorizedTickSpacing) == 0 {
					m.AuthorizedTickSpacing = make([]uint64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowParams
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.AuthorizedTickSpacing = append(m.AuthorizedTickSpacing, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field AuthorizedTickSpacing", wireType)
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AuthorizedSpreadFactors", wireType)
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
			var v github_com_cosmos_cosmos_sdk_types.Dec
			m.AuthorizedSpreadFactors = append(m.AuthorizedSpreadFactors, v)
			if err := m.AuthorizedSpreadFactors[len(m.AuthorizedSpreadFactors)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BalancerSharesRewardDiscount", wireType)
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
			if err := m.BalancerSharesRewardDiscount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AuthorizedQuoteDenoms", wireType)
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
			m.AuthorizedQuoteDenoms = append(m.AuthorizedQuoteDenoms, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AuthorizedUptimes", wireType)
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
			m.AuthorizedUptimes = append(m.AuthorizedUptimes, time.Duration(0))
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&(m.AuthorizedUptimes[len(m.AuthorizedUptimes)-1]), dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IsPermissionlessPoolCreationEnabled", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.IsPermissionlessPoolCreationEnabled = bool(v != 0)
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
