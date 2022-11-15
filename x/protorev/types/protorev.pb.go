// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/protorev/v1beta1/protorev.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	types "github.com/cosmos/cosmos-sdk/types"
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

type NeedToArb struct {
	NeedToArb bool `protobuf:"varint,1,opt,name=need_to_arb,json=needToArb,proto3" json:"need_to_arb,omitempty"`
}

func (m *NeedToArb) Reset()         { *m = NeedToArb{} }
func (m *NeedToArb) String() string { return proto.CompactTextString(m) }
func (*NeedToArb) ProtoMessage()    {}
func (*NeedToArb) Descriptor() ([]byte, []int) {
	return fileDescriptor_1e9f2391fd9fec01, []int{0}
}
func (m *NeedToArb) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *NeedToArb) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_NeedToArb.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *NeedToArb) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NeedToArb.Merge(m, src)
}
func (m *NeedToArb) XXX_Size() int {
	return m.Size()
}
func (m *NeedToArb) XXX_DiscardUnknown() {
	xxx_messageInfo_NeedToArb.DiscardUnknown(m)
}

var xxx_messageInfo_NeedToArb proto.InternalMessageInfo

func (m *NeedToArb) GetNeedToArb() bool {
	if m != nil {
		return m.NeedToArb
	}
	return false
}

type SwapInfo struct {
	PooldId uint64      `protobuf:"varint,1,opt,name=poold_id,json=pooldId,proto3" json:"poold_id,omitempty"`
	Input   *types.Coin `protobuf:"bytes,2,opt,name=input,proto3" json:"input,omitempty"`
	Output  *types.Coin `protobuf:"bytes,3,opt,name=output,proto3" json:"output,omitempty"`
}

func (m *SwapInfo) Reset()         { *m = SwapInfo{} }
func (m *SwapInfo) String() string { return proto.CompactTextString(m) }
func (*SwapInfo) ProtoMessage()    {}
func (*SwapInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_1e9f2391fd9fec01, []int{1}
}
func (m *SwapInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SwapInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SwapInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SwapInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SwapInfo.Merge(m, src)
}
func (m *SwapInfo) XXX_Size() int {
	return m.Size()
}
func (m *SwapInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_SwapInfo.DiscardUnknown(m)
}

var xxx_messageInfo_SwapInfo proto.InternalMessageInfo

func (m *SwapInfo) GetPooldId() uint64 {
	if m != nil {
		return m.PooldId
	}
	return 0
}

func (m *SwapInfo) GetInput() *types.Coin {
	if m != nil {
		return m.Input
	}
	return nil
}

func (m *SwapInfo) GetOutput() *types.Coin {
	if m != nil {
		return m.Output
	}
	return nil
}

type ArbDetails struct {
	Sender string      `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Swaps  []*SwapInfo `protobuf:"bytes,2,rep,name=swaps,proto3" json:"swaps,omitempty"`
}

func (m *ArbDetails) Reset()         { *m = ArbDetails{} }
func (m *ArbDetails) String() string { return proto.CompactTextString(m) }
func (*ArbDetails) ProtoMessage()    {}
func (*ArbDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_1e9f2391fd9fec01, []int{2}
}
func (m *ArbDetails) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ArbDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ArbDetails.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ArbDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ArbDetails.Merge(m, src)
}
func (m *ArbDetails) XXX_Size() int {
	return m.Size()
}
func (m *ArbDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_ArbDetails.DiscardUnknown(m)
}

var xxx_messageInfo_ArbDetails proto.InternalMessageInfo

func (m *ArbDetails) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

func (m *ArbDetails) GetSwaps() []*SwapInfo {
	if m != nil {
		return m.Swaps
	}
	return nil
}

type ConnectedTokens struct {
	Tokens []string `protobuf:"bytes,1,rep,name=tokens,proto3" json:"tokens,omitempty"`
}

func (m *ConnectedTokens) Reset()         { *m = ConnectedTokens{} }
func (m *ConnectedTokens) String() string { return proto.CompactTextString(m) }
func (*ConnectedTokens) ProtoMessage()    {}
func (*ConnectedTokens) Descriptor() ([]byte, []int) {
	return fileDescriptor_1e9f2391fd9fec01, []int{3}
}
func (m *ConnectedTokens) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ConnectedTokens) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ConnectedTokens.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ConnectedTokens) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConnectedTokens.Merge(m, src)
}
func (m *ConnectedTokens) XXX_Size() int {
	return m.Size()
}
func (m *ConnectedTokens) XXX_DiscardUnknown() {
	xxx_messageInfo_ConnectedTokens.DiscardUnknown(m)
}

var xxx_messageInfo_ConnectedTokens proto.InternalMessageInfo

func (m *ConnectedTokens) GetTokens() []string {
	if m != nil {
		return m.Tokens
	}
	return nil
}

type PairsToPoolIDs struct {
	PoolIds []uint64 `protobuf:"varint,1,rep,packed,name=pool_ids,json=poolIds,proto3" json:"pool_ids,omitempty"`
}

func (m *PairsToPoolIDs) Reset()         { *m = PairsToPoolIDs{} }
func (m *PairsToPoolIDs) String() string { return proto.CompactTextString(m) }
func (*PairsToPoolIDs) ProtoMessage()    {}
func (*PairsToPoolIDs) Descriptor() ([]byte, []int) {
	return fileDescriptor_1e9f2391fd9fec01, []int{4}
}
func (m *PairsToPoolIDs) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PairsToPoolIDs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PairsToPoolIDs.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PairsToPoolIDs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PairsToPoolIDs.Merge(m, src)
}
func (m *PairsToPoolIDs) XXX_Size() int {
	return m.Size()
}
func (m *PairsToPoolIDs) XXX_DiscardUnknown() {
	xxx_messageInfo_PairsToPoolIDs.DiscardUnknown(m)
}

var xxx_messageInfo_PairsToPoolIDs proto.InternalMessageInfo

func (m *PairsToPoolIDs) GetPoolIds() []uint64 {
	if m != nil {
		return m.PoolIds
	}
	return nil
}

type CyclicRoute struct {
	Id []uint64 `protobuf:"varint,1,rep,packed,name=id,proto3" json:"id,omitempty"`
}

func (m *CyclicRoute) Reset()         { *m = CyclicRoute{} }
func (m *CyclicRoute) String() string { return proto.CompactTextString(m) }
func (*CyclicRoute) ProtoMessage()    {}
func (*CyclicRoute) Descriptor() ([]byte, []int) {
	return fileDescriptor_1e9f2391fd9fec01, []int{5}
}
func (m *CyclicRoute) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CyclicRoute) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CyclicRoute.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CyclicRoute) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CyclicRoute.Merge(m, src)
}
func (m *CyclicRoute) XXX_Size() int {
	return m.Size()
}
func (m *CyclicRoute) XXX_DiscardUnknown() {
	xxx_messageInfo_CyclicRoute.DiscardUnknown(m)
}

var xxx_messageInfo_CyclicRoute proto.InternalMessageInfo

func (m *CyclicRoute) GetId() []uint64 {
	if m != nil {
		return m.Id
	}
	return nil
}

type ListOfCyclicRoutes struct {
	CyclicRoute []*CyclicRoute `protobuf:"bytes,1,rep,name=cyclic_route,json=cyclicRoute,proto3" json:"cyclic_route,omitempty"`
}

func (m *ListOfCyclicRoutes) Reset()         { *m = ListOfCyclicRoutes{} }
func (m *ListOfCyclicRoutes) String() string { return proto.CompactTextString(m) }
func (*ListOfCyclicRoutes) ProtoMessage()    {}
func (*ListOfCyclicRoutes) Descriptor() ([]byte, []int) {
	return fileDescriptor_1e9f2391fd9fec01, []int{6}
}
func (m *ListOfCyclicRoutes) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ListOfCyclicRoutes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ListOfCyclicRoutes.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ListOfCyclicRoutes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListOfCyclicRoutes.Merge(m, src)
}
func (m *ListOfCyclicRoutes) XXX_Size() int {
	return m.Size()
}
func (m *ListOfCyclicRoutes) XXX_DiscardUnknown() {
	xxx_messageInfo_ListOfCyclicRoutes.DiscardUnknown(m)
}

var xxx_messageInfo_ListOfCyclicRoutes proto.InternalMessageInfo

func (m *ListOfCyclicRoutes) GetCyclicRoute() []*CyclicRoute {
	if m != nil {
		return m.CyclicRoute
	}
	return nil
}

func init() {
	proto.RegisterType((*NeedToArb)(nil), "osmosis.protorev.v1beta1.NeedToArb")
	proto.RegisterType((*SwapInfo)(nil), "osmosis.protorev.v1beta1.SwapInfo")
	proto.RegisterType((*ArbDetails)(nil), "osmosis.protorev.v1beta1.ArbDetails")
	proto.RegisterType((*ConnectedTokens)(nil), "osmosis.protorev.v1beta1.ConnectedTokens")
	proto.RegisterType((*PairsToPoolIDs)(nil), "osmosis.protorev.v1beta1.PairsToPoolIDs")
	proto.RegisterType((*CyclicRoute)(nil), "osmosis.protorev.v1beta1.CyclicRoute")
	proto.RegisterType((*ListOfCyclicRoutes)(nil), "osmosis.protorev.v1beta1.ListOfCyclicRoutes")
}

func init() {
	proto.RegisterFile("osmosis/protorev/v1beta1/protorev.proto", fileDescriptor_1e9f2391fd9fec01)
}

var fileDescriptor_1e9f2391fd9fec01 = []byte{
	// 444 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0xcf, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0xe3, 0x84, 0x86, 0x64, 0x8c, 0x8a, 0xb4, 0x42, 0xc8, 0xa9, 0x84, 0x15, 0x59, 0x42,
	0x04, 0x55, 0xd8, 0x4a, 0xe0, 0xc0, 0xb5, 0xa4, 0x07, 0x22, 0xa1, 0x52, 0x99, 0x9e, 0x38, 0xd4,
	0xb2, 0xbd, 0xdb, 0xb0, 0xc2, 0xdd, 0xb1, 0x76, 0x37, 0x29, 0x7d, 0x04, 0x6e, 0x3c, 0x16, 0xc7,
	0x1e, 0x39, 0xa2, 0xe4, 0x45, 0x90, 0x77, 0xd7, 0x69, 0x2f, 0x11, 0xb7, 0xf9, 0xbe, 0xf9, 0xcd,
	0xce, 0x1f, 0x2d, 0xbc, 0x42, 0x75, 0x8d, 0x8a, 0xab, 0xa4, 0x96, 0xa8, 0x51, 0xb2, 0x75, 0xb2,
	0x9e, 0x16, 0x4c, 0xe7, 0xd3, 0x9d, 0x11, 0x9b, 0x80, 0x04, 0x0e, 0x8c, 0x77, 0xbe, 0x03, 0x8f,
	0x46, 0xa5, 0x49, 0x65, 0x26, 0x91, 0x58, 0x61, 0xa9, 0xa3, 0x67, 0x4b, 0x5c, 0xa2, 0xf5, 0x9b,
	0xc8, 0xb9, 0xa1, 0x65, 0x92, 0x22, 0x57, 0x6c, 0xd7, 0xae, 0x44, 0x2e, 0x6c, 0x3e, 0x3a, 0x86,
	0xe1, 0x19, 0x63, 0xf4, 0x02, 0x4f, 0x64, 0x41, 0x42, 0xf0, 0x05, 0x63, 0x34, 0xd3, 0x98, 0xe5,
	0xb2, 0x08, 0xbc, 0xb1, 0x37, 0x19, 0xa4, 0x43, 0xd1, 0xe6, 0xa3, 0x9f, 0x1e, 0x0c, 0xbe, 0xdc,
	0xe4, 0xf5, 0x42, 0x5c, 0x21, 0x19, 0xc1, 0xa0, 0x46, 0xac, 0x68, 0xc6, 0xa9, 0x21, 0x1f, 0xa5,
	0x8f, 0x8d, 0x5e, 0x50, 0x92, 0xc0, 0x01, 0x17, 0xf5, 0x4a, 0x07, 0xdd, 0xb1, 0x37, 0xf1, 0x67,
	0xa3, 0xd8, 0x0d, 0xda, 0x0c, 0xd1, 0xae, 0x12, 0xcf, 0x91, 0x8b, 0xd4, 0x72, 0x64, 0x0a, 0x7d,
	0x5c, 0xe9, 0xa6, 0xa2, 0xf7, 0xbf, 0x0a, 0x07, 0x46, 0x97, 0x00, 0x27, 0xb2, 0x38, 0x65, 0x3a,
	0xe7, 0x95, 0x22, 0xcf, 0xa1, 0xaf, 0x98, 0xa0, 0x4c, 0x9a, 0x51, 0x86, 0xa9, 0x53, 0xe4, 0x3d,
	0x1c, 0xa8, 0x9b, 0xbc, 0x56, 0x41, 0x77, 0xdc, 0x9b, 0xf8, 0xb3, 0x28, 0xde, 0x77, 0xd9, 0xb8,
	0xdd, 0x2b, 0xb5, 0x05, 0xd1, 0x6b, 0x78, 0x3a, 0x47, 0x21, 0x58, 0xa9, 0x9b, 0xed, 0xbf, 0x33,
	0x61, 0x9a, 0x68, 0x13, 0x05, 0xde, 0xb8, 0xd7, 0x34, 0xb1, 0x2a, 0x3a, 0x86, 0xc3, 0xf3, 0x9c,
	0x4b, 0x75, 0x81, 0xe7, 0x88, 0xd5, 0xe2, 0x54, 0xb5, 0xb7, 0xc9, 0x38, 0xb5, 0xac, 0xbb, 0xcd,
	0x82, 0xaa, 0xe8, 0x05, 0xf8, 0xf3, 0xdb, 0xb2, 0xe2, 0x65, 0x8a, 0x2b, 0xcd, 0xc8, 0x21, 0x74,
	0xcd, 0xfd, 0x1a, 0xa6, 0xcb, 0x69, 0x74, 0x09, 0xe4, 0x13, 0x57, 0xfa, 0xf3, 0xd5, 0x03, 0x48,
	0x91, 0x8f, 0xf0, 0xa4, 0x34, 0x3a, 0x93, 0x8d, 0x61, 0x78, 0x7f, 0xf6, 0x72, 0xff, 0x36, 0x0f,
	0xaa, 0x53, 0xbf, 0xbc, 0x17, 0x1f, 0xce, 0x7e, 0x6f, 0x42, 0xef, 0x6e, 0x13, 0x7a, 0x7f, 0x37,
	0xa1, 0xf7, 0x6b, 0x1b, 0x76, 0xee, 0xb6, 0x61, 0xe7, 0xcf, 0x36, 0xec, 0x7c, 0x7d, 0xb7, 0xe4,
	0xfa, 0xdb, 0xaa, 0x88, 0x4b, 0xbc, 0x4e, 0xdc, 0xbb, 0x6f, 0xaa, 0xbc, 0x50, 0xad, 0x48, 0xd6,
	0xd3, 0x59, 0xf2, 0xe3, 0xfe, 0xef, 0xea, 0xdb, 0x9a, 0xa9, 0xa2, 0x6f, 0xf4, 0xdb, 0x7f, 0x01,
	0x00, 0x00, 0xff, 0xff, 0xe2, 0xb3, 0x3d, 0x1b, 0xdc, 0x02, 0x00, 0x00,
}

func (m *NeedToArb) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NeedToArb) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *NeedToArb) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.NeedToArb {
		i--
		if m.NeedToArb {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *SwapInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SwapInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SwapInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Output != nil {
		{
			size, err := m.Output.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintProtorev(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if m.Input != nil {
		{
			size, err := m.Input.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintProtorev(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if m.PooldId != 0 {
		i = encodeVarintProtorev(dAtA, i, uint64(m.PooldId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *ArbDetails) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ArbDetails) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ArbDetails) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Swaps) > 0 {
		for iNdEx := len(m.Swaps) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Swaps[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintProtorev(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintProtorev(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *ConnectedTokens) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ConnectedTokens) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ConnectedTokens) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Tokens) > 0 {
		for iNdEx := len(m.Tokens) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Tokens[iNdEx])
			copy(dAtA[i:], m.Tokens[iNdEx])
			i = encodeVarintProtorev(dAtA, i, uint64(len(m.Tokens[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *PairsToPoolIDs) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PairsToPoolIDs) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PairsToPoolIDs) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.PoolIds) > 0 {
		dAtA4 := make([]byte, len(m.PoolIds)*10)
		var j3 int
		for _, num := range m.PoolIds {
			for num >= 1<<7 {
				dAtA4[j3] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j3++
			}
			dAtA4[j3] = uint8(num)
			j3++
		}
		i -= j3
		copy(dAtA[i:], dAtA4[:j3])
		i = encodeVarintProtorev(dAtA, i, uint64(j3))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *CyclicRoute) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CyclicRoute) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CyclicRoute) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Id) > 0 {
		dAtA6 := make([]byte, len(m.Id)*10)
		var j5 int
		for _, num := range m.Id {
			for num >= 1<<7 {
				dAtA6[j5] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j5++
			}
			dAtA6[j5] = uint8(num)
			j5++
		}
		i -= j5
		copy(dAtA[i:], dAtA6[:j5])
		i = encodeVarintProtorev(dAtA, i, uint64(j5))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *ListOfCyclicRoutes) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ListOfCyclicRoutes) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ListOfCyclicRoutes) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.CyclicRoute) > 0 {
		for iNdEx := len(m.CyclicRoute) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.CyclicRoute[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintProtorev(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintProtorev(dAtA []byte, offset int, v uint64) int {
	offset -= sovProtorev(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *NeedToArb) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.NeedToArb {
		n += 2
	}
	return n
}

func (m *SwapInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PooldId != 0 {
		n += 1 + sovProtorev(uint64(m.PooldId))
	}
	if m.Input != nil {
		l = m.Input.Size()
		n += 1 + l + sovProtorev(uint64(l))
	}
	if m.Output != nil {
		l = m.Output.Size()
		n += 1 + l + sovProtorev(uint64(l))
	}
	return n
}

func (m *ArbDetails) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovProtorev(uint64(l))
	}
	if len(m.Swaps) > 0 {
		for _, e := range m.Swaps {
			l = e.Size()
			n += 1 + l + sovProtorev(uint64(l))
		}
	}
	return n
}

func (m *ConnectedTokens) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Tokens) > 0 {
		for _, s := range m.Tokens {
			l = len(s)
			n += 1 + l + sovProtorev(uint64(l))
		}
	}
	return n
}

func (m *PairsToPoolIDs) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.PoolIds) > 0 {
		l = 0
		for _, e := range m.PoolIds {
			l += sovProtorev(uint64(e))
		}
		n += 1 + sovProtorev(uint64(l)) + l
	}
	return n
}

func (m *CyclicRoute) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Id) > 0 {
		l = 0
		for _, e := range m.Id {
			l += sovProtorev(uint64(e))
		}
		n += 1 + sovProtorev(uint64(l)) + l
	}
	return n
}

func (m *ListOfCyclicRoutes) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.CyclicRoute) > 0 {
		for _, e := range m.CyclicRoute {
			l = e.Size()
			n += 1 + l + sovProtorev(uint64(l))
		}
	}
	return n
}

func sovProtorev(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozProtorev(x uint64) (n int) {
	return sovProtorev(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *NeedToArb) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtorev
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
			return fmt.Errorf("proto: NeedToArb: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NeedToArb: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NeedToArb", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtorev
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
			m.NeedToArb = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipProtorev(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProtorev
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
func (m *SwapInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtorev
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
			return fmt.Errorf("proto: SwapInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SwapInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PooldId", wireType)
			}
			m.PooldId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtorev
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PooldId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Input", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtorev
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
				return ErrInvalidLengthProtorev
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthProtorev
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Input == nil {
				m.Input = &types.Coin{}
			}
			if err := m.Input.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Output", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtorev
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
				return ErrInvalidLengthProtorev
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthProtorev
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Output == nil {
				m.Output = &types.Coin{}
			}
			if err := m.Output.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtorev(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProtorev
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
func (m *ArbDetails) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtorev
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
			return fmt.Errorf("proto: ArbDetails: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ArbDetails: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtorev
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
				return ErrInvalidLengthProtorev
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProtorev
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Swaps", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtorev
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
				return ErrInvalidLengthProtorev
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthProtorev
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Swaps = append(m.Swaps, &SwapInfo{})
			if err := m.Swaps[len(m.Swaps)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtorev(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProtorev
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
func (m *ConnectedTokens) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtorev
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
			return fmt.Errorf("proto: ConnectedTokens: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ConnectedTokens: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tokens", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtorev
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
				return ErrInvalidLengthProtorev
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProtorev
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Tokens = append(m.Tokens, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtorev(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProtorev
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
func (m *PairsToPoolIDs) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtorev
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
			return fmt.Errorf("proto: PairsToPoolIDs: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PairsToPoolIDs: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType == 0 {
				var v uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowProtorev
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
				m.PoolIds = append(m.PoolIds, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowProtorev
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
					return ErrInvalidLengthProtorev
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthProtorev
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
				if elementCount != 0 && len(m.PoolIds) == 0 {
					m.PoolIds = make([]uint64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowProtorev
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
					m.PoolIds = append(m.PoolIds, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolIds", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipProtorev(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProtorev
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
func (m *CyclicRoute) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtorev
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
			return fmt.Errorf("proto: CyclicRoute: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CyclicRoute: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType == 0 {
				var v uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowProtorev
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
				m.Id = append(m.Id, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowProtorev
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
					return ErrInvalidLengthProtorev
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthProtorev
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
				if elementCount != 0 && len(m.Id) == 0 {
					m.Id = make([]uint64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowProtorev
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
					m.Id = append(m.Id, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipProtorev(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProtorev
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
func (m *ListOfCyclicRoutes) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtorev
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
			return fmt.Errorf("proto: ListOfCyclicRoutes: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ListOfCyclicRoutes: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CyclicRoute", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtorev
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
				return ErrInvalidLengthProtorev
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthProtorev
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CyclicRoute = append(m.CyclicRoute, &CyclicRoute{})
			if err := m.CyclicRoute[len(m.CyclicRoute)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtorev(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProtorev
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
func skipProtorev(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProtorev
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
					return 0, ErrIntOverflowProtorev
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
					return 0, ErrIntOverflowProtorev
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
				return 0, ErrInvalidLengthProtorev
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupProtorev
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthProtorev
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthProtorev        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProtorev          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupProtorev = fmt.Errorf("proto: unexpected end of group")
)