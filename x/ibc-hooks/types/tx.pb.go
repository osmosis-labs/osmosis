// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/ibchooks/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type MsgEmitIBCAck struct {
	Sender         string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty" yaml:"sender"`
	PacketSequence uint64 `protobuf:"varint,2,opt,name=packet_sequence,json=packetSequence,proto3" json:"packet_sequence,omitempty" yaml:"packet_sequence"`
	Channel        string `protobuf:"bytes,3,opt,name=channel,proto3" json:"channel,omitempty" yaml:"channel"`
}

func (m *MsgEmitIBCAck) Reset()         { *m = MsgEmitIBCAck{} }
func (m *MsgEmitIBCAck) String() string { return proto.CompactTextString(m) }
func (*MsgEmitIBCAck) ProtoMessage()    {}
func (*MsgEmitIBCAck) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb5a795bb7f479a3, []int{0}
}
func (m *MsgEmitIBCAck) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgEmitIBCAck) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgEmitIBCAck.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgEmitIBCAck) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgEmitIBCAck.Merge(m, src)
}
func (m *MsgEmitIBCAck) XXX_Size() int {
	return m.Size()
}
func (m *MsgEmitIBCAck) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgEmitIBCAck.DiscardUnknown(m)
}

var xxx_messageInfo_MsgEmitIBCAck proto.InternalMessageInfo

func (m *MsgEmitIBCAck) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

func (m *MsgEmitIBCAck) GetPacketSequence() uint64 {
	if m != nil {
		return m.PacketSequence
	}
	return 0
}

func (m *MsgEmitIBCAck) GetChannel() string {
	if m != nil {
		return m.Channel
	}
	return ""
}

type MsgEmitIBCAckResponse struct {
	ContractResult string `protobuf:"bytes,1,opt,name=contract_result,json=contractResult,proto3" json:"contract_result,omitempty" yaml:"contract_result"`
	IbcAck         string `protobuf:"bytes,2,opt,name=ibc_ack,json=ibcAck,proto3" json:"ibc_ack,omitempty" yaml:"ibc_ack"`
}

func (m *MsgEmitIBCAckResponse) Reset()         { *m = MsgEmitIBCAckResponse{} }
func (m *MsgEmitIBCAckResponse) String() string { return proto.CompactTextString(m) }
func (*MsgEmitIBCAckResponse) ProtoMessage()    {}
func (*MsgEmitIBCAckResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb5a795bb7f479a3, []int{1}
}
func (m *MsgEmitIBCAckResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgEmitIBCAckResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgEmitIBCAckResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgEmitIBCAckResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgEmitIBCAckResponse.Merge(m, src)
}
func (m *MsgEmitIBCAckResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgEmitIBCAckResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgEmitIBCAckResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgEmitIBCAckResponse proto.InternalMessageInfo

func (m *MsgEmitIBCAckResponse) GetContractResult() string {
	if m != nil {
		return m.ContractResult
	}
	return ""
}

func (m *MsgEmitIBCAckResponse) GetIbcAck() string {
	if m != nil {
		return m.IbcAck
	}
	return ""
}

func init() {
	proto.RegisterType((*MsgEmitIBCAck)(nil), "osmosis.ibchooks.MsgEmitIBCAck")
	proto.RegisterType((*MsgEmitIBCAckResponse)(nil), "osmosis.ibchooks.MsgEmitIBCAckResponse")
}

func init() { proto.RegisterFile("osmosis/ibchooks/tx.proto", fileDescriptor_eb5a795bb7f479a3) }

var fileDescriptor_eb5a795bb7f479a3 = []byte{
	// 364 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0x4f, 0x4f, 0xc2, 0x30,
	0x18, 0xc6, 0xa9, 0x18, 0x88, 0x4d, 0x40, 0x6d, 0xd4, 0x20, 0x87, 0x8d, 0xf4, 0x22, 0x46, 0xd9,
	0x0c, 0xc6, 0x8b, 0x37, 0x46, 0x3c, 0x78, 0x20, 0x26, 0x35, 0xf1, 0x60, 0x62, 0xc8, 0x56, 0x9b,
	0xb1, 0xec, 0x4f, 0xe7, 0x5a, 0x0c, 0x7c, 0x04, 0x6f, 0x7e, 0x11, 0xbf, 0x87, 0x47, 0x8e, 0x9e,
	0x88, 0x81, 0x6f, 0xc0, 0x27, 0x30, 0x6c, 0x5d, 0x02, 0xbb, 0x78, 0xdb, 0x9e, 0xe7, 0xf7, 0xb6,
	0xcf, 0xfb, 0xf6, 0x85, 0xa7, 0x5c, 0x84, 0x5c, 0x78, 0xc2, 0xf4, 0x1c, 0x3a, 0xe2, 0xdc, 0x17,
	0xa6, 0x9c, 0x18, 0x71, 0xc2, 0x25, 0x47, 0x07, 0xca, 0x32, 0x72, 0xab, 0x79, 0xe4, 0x72, 0x97,
	0xa7, 0xa6, 0xb9, 0xfe, 0xca, 0x38, 0xfc, 0x05, 0x60, 0x6d, 0x20, 0xdc, 0xbb, 0xd0, 0x93, 0xf7,
	0x56, 0xbf, 0x47, 0x7d, 0x74, 0x0e, 0x2b, 0x82, 0x45, 0xaf, 0x2c, 0x69, 0x80, 0x16, 0x68, 0xef,
	0x59, 0x87, 0xab, 0xb9, 0x5e, 0x9b, 0xda, 0x61, 0x70, 0x8b, 0x33, 0x1d, 0x13, 0x05, 0xa0, 0x3e,
	0xdc, 0x8f, 0x6d, 0xea, 0x33, 0x39, 0x14, 0xec, 0x6d, 0xcc, 0x22, 0xca, 0x1a, 0x3b, 0x2d, 0xd0,
	0xde, 0xb5, 0x9a, 0xab, 0xb9, 0x7e, 0x92, 0xd5, 0x14, 0x00, 0x4c, 0xea, 0x99, 0xf2, 0xa8, 0x04,
	0x74, 0x09, 0xab, 0x74, 0x64, 0x47, 0x11, 0x0b, 0x1a, 0xe5, 0xf4, 0x42, 0xb4, 0x9a, 0xeb, 0xf5,
	0xac, 0x58, 0x19, 0x98, 0xe4, 0x08, 0xfe, 0x00, 0xf0, 0x78, 0x2b, 0x2f, 0x61, 0x22, 0xe6, 0x91,
	0x60, 0xeb, 0x30, 0x94, 0x47, 0x32, 0xb1, 0xa9, 0x1c, 0x26, 0x4c, 0x8c, 0x03, 0xa9, 0x1a, 0xd8,
	0x08, 0x53, 0x00, 0x30, 0xa9, 0xe7, 0x0a, 0x49, 0x05, 0x74, 0x01, 0xab, 0x9e, 0x43, 0x87, 0x36,
	0xf5, 0xd3, 0x4e, 0xb6, 0xc2, 0x28, 0x03, 0x93, 0x8a, 0xe7, 0xd0, 0x1e, 0xf5, 0xbb, 0x2f, 0xb0,
	0x3c, 0x10, 0x2e, 0x7a, 0x82, 0x70, 0x63, 0x7c, 0xba, 0x51, 0x9c, 0xbc, 0xb1, 0x95, 0xb7, 0x79,
	0xf6, 0x0f, 0x90, 0x37, 0x64, 0x3d, 0x7c, 0x2f, 0x34, 0x30, 0x5b, 0x68, 0xe0, 0x77, 0xa1, 0x81,
	0xcf, 0xa5, 0x56, 0x9a, 0x2d, 0xb5, 0xd2, 0xcf, 0x52, 0x2b, 0x3d, 0xdf, 0xb8, 0x9e, 0x1c, 0x8d,
	0x1d, 0x83, 0xf2, 0xd0, 0x54, 0x87, 0x75, 0x02, 0xdb, 0x11, 0xf9, 0x8f, 0xf9, 0xde, 0xbd, 0x32,
	0x27, 0xeb, 0xad, 0xe8, 0xa8, 0xb5, 0x98, 0xc6, 0x4c, 0x38, 0x95, 0xf4, 0xc9, 0xaf, 0xff, 0x02,
	0x00, 0x00, 0xff, 0xff, 0xdc, 0x85, 0xb6, 0x32, 0x37, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// EmitIBCAck checks the sender can emit the ack and writes the IBC
	// acknowledgement
	EmitIBCAck(ctx context.Context, in *MsgEmitIBCAck, opts ...grpc.CallOption) (*MsgEmitIBCAckResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) EmitIBCAck(ctx context.Context, in *MsgEmitIBCAck, opts ...grpc.CallOption) (*MsgEmitIBCAckResponse, error) {
	out := new(MsgEmitIBCAckResponse)
	err := c.cc.Invoke(ctx, "/osmosis.ibchooks.Msg/EmitIBCAck", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// EmitIBCAck checks the sender can emit the ack and writes the IBC
	// acknowledgement
	EmitIBCAck(context.Context, *MsgEmitIBCAck) (*MsgEmitIBCAckResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) EmitIBCAck(ctx context.Context, req *MsgEmitIBCAck) (*MsgEmitIBCAckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EmitIBCAck not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_EmitIBCAck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgEmitIBCAck)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).EmitIBCAck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.ibchooks.Msg/EmitIBCAck",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).EmitIBCAck(ctx, req.(*MsgEmitIBCAck))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "osmosis.ibchooks.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "EmitIBCAck",
			Handler:    _Msg_EmitIBCAck_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "osmosis/ibchooks/tx.proto",
}

func (m *MsgEmitIBCAck) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgEmitIBCAck) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgEmitIBCAck) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Channel) > 0 {
		i -= len(m.Channel)
		copy(dAtA[i:], m.Channel)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Channel)))
		i--
		dAtA[i] = 0x1a
	}
	if m.PacketSequence != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.PacketSequence))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgEmitIBCAckResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgEmitIBCAckResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgEmitIBCAckResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.IbcAck) > 0 {
		i -= len(m.IbcAck)
		copy(dAtA[i:], m.IbcAck)
		i = encodeVarintTx(dAtA, i, uint64(len(m.IbcAck)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ContractResult) > 0 {
		i -= len(m.ContractResult)
		copy(dAtA[i:], m.ContractResult)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ContractResult)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgEmitIBCAck) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.PacketSequence != 0 {
		n += 1 + sovTx(uint64(m.PacketSequence))
	}
	l = len(m.Channel)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgEmitIBCAckResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ContractResult)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.IbcAck)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgEmitIBCAck) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgEmitIBCAck: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgEmitIBCAck: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PacketSequence", wireType)
			}
			m.PacketSequence = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PacketSequence |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Channel", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Channel = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgEmitIBCAckResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgEmitIBCAckResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgEmitIBCAckResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractResult", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContractResult = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IbcAck", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IbcAck = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
