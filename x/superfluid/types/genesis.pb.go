// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/superfluid/genesis.proto

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

// GenesisState defines the module's genesis state.
type GenesisState struct {
	Params                    Params                           `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	SuperfluidAssets          []SuperfluidAsset                `protobuf:"bytes,2,rep,name=superfluid_assets,json=superfluidAssets,proto3" json:"superfluid_assets"`
	OsmoEquivalentMultipliers []OsmoEquivalentMultiplierRecord `protobuf:"bytes,3,rep,name=osmo_equivalent_multipliers,json=osmoEquivalentMultipliers,proto3" json:"osmo_equivalent_multipliers"`
	IntermediaryAccounts      []SuperfluidIntermediaryAccount  `protobuf:"bytes,4,rep,name=intermediary_accounts,json=intermediaryAccounts,proto3" json:"intermediary_accounts"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_d5256ebb7c83fff3, []int{0}
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

func (m *GenesisState) GetSuperfluidAssets() []SuperfluidAsset {
	if m != nil {
		return m.SuperfluidAssets
	}
	return nil
}

func (m *GenesisState) GetOsmoEquivalentMultipliers() []OsmoEquivalentMultiplierRecord {
	if m != nil {
		return m.OsmoEquivalentMultipliers
	}
	return nil
}

func (m *GenesisState) GetIntermediaryAccounts() []SuperfluidIntermediaryAccount {
	if m != nil {
		return m.IntermediaryAccounts
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "osmosis.superfluid.GenesisState")
}

func init() { proto.RegisterFile("osmosis/superfluid/genesis.proto", fileDescriptor_d5256ebb7c83fff3) }

var fileDescriptor_d5256ebb7c83fff3 = []byte{
	// 342 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0x31, 0x4f, 0xc2, 0x40,
	0x14, 0xc7, 0x5b, 0x21, 0x0c, 0xc5, 0x41, 0x1b, 0x4c, 0x2a, 0x26, 0x85, 0xc8, 0xc2, 0x62, 0x1b,
	0x31, 0x46, 0x57, 0x48, 0x8c, 0x71, 0x30, 0x18, 0x48, 0x1c, 0x5c, 0x9a, 0xa3, 0x3c, 0xeb, 0x25,
	0x6d, 0xaf, 0xde, 0xbb, 0x23, 0xf0, 0x0d, 0x1c, 0xfd, 0x58, 0x8c, 0x8c, 0x4e, 0xc6, 0xc0, 0x17,
	0x31, 0xb4, 0x27, 0xc5, 0x50, 0xdd, 0x5e, 0xfb, 0x7e, 0xff, 0xf7, 0x7b, 0x97, 0x67, 0x34, 0x19,
	0x46, 0x0c, 0x29, 0xba, 0x28, 0x13, 0xe0, 0xcf, 0xa1, 0xa4, 0x63, 0x37, 0x80, 0x18, 0x90, 0xa2,
	0x93, 0x70, 0x26, 0x98, 0x69, 0x2a, 0xc2, 0xc9, 0x89, 0x7a, 0x2d, 0x60, 0x01, 0x4b, 0xdb, 0xee,
	0xba, 0xca, 0xc8, 0x7a, 0xab, 0x60, 0x56, 0x5e, 0x2a, 0xa8, 0x51, 0x00, 0x25, 0x84, 0x93, 0x48,
	0xf9, 0x4e, 0xdf, 0x4a, 0xc6, 0xfe, 0x6d, 0xb6, 0xc1, 0x50, 0x10, 0x01, 0xe6, 0xb5, 0x51, 0xc9,
	0x00, 0x4b, 0x6f, 0xea, 0xed, 0x6a, 0xa7, 0xee, 0xec, 0x6e, 0xe4, 0x3c, 0xa4, 0x44, 0xaf, 0x3c,
	0xff, 0x6c, 0x68, 0x03, 0xc5, 0x9b, 0x8f, 0xc6, 0x61, 0x8e, 0x78, 0x04, 0x11, 0x04, 0x5a, 0x7b,
	0xcd, 0x52, 0xbb, 0xda, 0x69, 0x15, 0x0d, 0x19, 0x6e, 0xca, 0xee, 0x9a, 0x55, 0xd3, 0x0e, 0xf0,
	0xf7, 0x6f, 0x34, 0xa7, 0xc6, 0xc9, 0x3a, 0xed, 0xc1, 0xab, 0xa4, 0x13, 0x12, 0x42, 0x2c, 0xbc,
	0x48, 0x86, 0x82, 0x26, 0x21, 0x05, 0x8e, 0x56, 0x29, 0x35, 0x74, 0x8a, 0x0c, 0x7d, 0x8c, 0xd8,
	0xcd, 0x26, 0x75, 0xbf, 0x09, 0x0d, 0xc0, 0x67, 0x7c, 0xac, 0x84, 0xc7, 0xec, 0x0f, 0x0a, 0xcd,
	0xd0, 0x38, 0xa2, 0xb1, 0x00, 0x1e, 0xc1, 0x98, 0x12, 0x3e, 0xf3, 0x88, 0xef, 0x33, 0x19, 0x0b,
	0xb4, 0xca, 0xa9, 0xf3, 0xfc, 0xff, 0x57, 0xdd, 0x6d, 0x45, 0xbb, 0x59, 0x52, 0x29, 0x6b, 0x74,
	0xb7, 0x85, 0xbd, 0xfe, 0x7c, 0x69, 0xeb, 0x8b, 0xa5, 0xad, 0x7f, 0x2d, 0x6d, 0xfd, 0x7d, 0x65,
	0x6b, 0x8b, 0x95, 0xad, 0x7d, 0xac, 0x6c, 0xed, 0xe9, 0x32, 0xa0, 0xe2, 0x45, 0x8e, 0x1c, 0x9f,
	0x45, 0xae, 0x52, 0x9e, 0x85, 0x64, 0x84, 0x3f, 0x1f, 0xee, 0xe4, 0xca, 0x9d, 0x6e, 0x9f, 0x58,
	0xcc, 0x12, 0xc0, 0x51, 0x25, 0x3d, 0xf1, 0xc5, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0xa5, 0x7a,
	0xbf, 0x48, 0x76, 0x02, 0x00, 0x00,
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
	if len(m.IntermediaryAccounts) > 0 {
		for iNdEx := len(m.IntermediaryAccounts) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.IntermediaryAccounts[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.OsmoEquivalentMultipliers) > 0 {
		for iNdEx := len(m.OsmoEquivalentMultipliers) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.OsmoEquivalentMultipliers[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.SuperfluidAssets) > 0 {
		for iNdEx := len(m.SuperfluidAssets) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.SuperfluidAssets[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
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
	if len(m.SuperfluidAssets) > 0 {
		for _, e := range m.SuperfluidAssets {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.OsmoEquivalentMultipliers) > 0 {
		for _, e := range m.OsmoEquivalentMultipliers {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.IntermediaryAccounts) > 0 {
		for _, e := range m.IntermediaryAccounts {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
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
				return fmt.Errorf("proto: wrong wireType = %d for field SuperfluidAssets", wireType)
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
			m.SuperfluidAssets = append(m.SuperfluidAssets, SuperfluidAsset{})
			if err := m.SuperfluidAssets[len(m.SuperfluidAssets)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OsmoEquivalentMultipliers", wireType)
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
			m.OsmoEquivalentMultipliers = append(m.OsmoEquivalentMultipliers, OsmoEquivalentMultiplierRecord{})
			if err := m.OsmoEquivalentMultipliers[len(m.OsmoEquivalentMultipliers)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IntermediaryAccounts", wireType)
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
			m.IntermediaryAccounts = append(m.IntermediaryAccounts, SuperfluidIntermediaryAccount{})
			if err := m.IntermediaryAccounts[len(m.IntermediaryAccounts)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
