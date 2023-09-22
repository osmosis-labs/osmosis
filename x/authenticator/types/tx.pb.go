// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/authenticator/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
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

// MsgAddAuthenticatorRequest defines the Msg/AddAuthenticator request type.
type MsgAddAuthenticator struct {
	Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Type   string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Data   []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *MsgAddAuthenticator) Reset()         { *m = MsgAddAuthenticator{} }
func (m *MsgAddAuthenticator) String() string { return proto.CompactTextString(m) }
func (*MsgAddAuthenticator) ProtoMessage()    {}
func (*MsgAddAuthenticator) Descriptor() ([]byte, []int) {
	return fileDescriptor_1aa1d7e4dc71ed44, []int{0}
}
func (m *MsgAddAuthenticator) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgAddAuthenticator) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgAddAuthenticator.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgAddAuthenticator) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgAddAuthenticator.Merge(m, src)
}
func (m *MsgAddAuthenticator) XXX_Size() int {
	return m.Size()
}
func (m *MsgAddAuthenticator) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgAddAuthenticator.DiscardUnknown(m)
}

var xxx_messageInfo_MsgAddAuthenticator proto.InternalMessageInfo

func (m *MsgAddAuthenticator) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

func (m *MsgAddAuthenticator) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *MsgAddAuthenticator) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

// MsgAddAuthenticatorResponse defines the Msg/AddAuthenticator response type.
type MsgAddAuthenticatorResponse struct {
	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (m *MsgAddAuthenticatorResponse) Reset()         { *m = MsgAddAuthenticatorResponse{} }
func (m *MsgAddAuthenticatorResponse) String() string { return proto.CompactTextString(m) }
func (*MsgAddAuthenticatorResponse) ProtoMessage()    {}
func (*MsgAddAuthenticatorResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1aa1d7e4dc71ed44, []int{1}
}
func (m *MsgAddAuthenticatorResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgAddAuthenticatorResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgAddAuthenticatorResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgAddAuthenticatorResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgAddAuthenticatorResponse.Merge(m, src)
}
func (m *MsgAddAuthenticatorResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgAddAuthenticatorResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgAddAuthenticatorResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgAddAuthenticatorResponse proto.InternalMessageInfo

func (m *MsgAddAuthenticatorResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

// MsgRemoveAuthenticatorRequest defines the Msg/RemoveAuthenticator request
// type.
type MsgRemoveAuthenticator struct {
	Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Id     uint64 `protobuf:"varint,2,opt,name=id,proto3" json:"id,omitempty"`
}

func (m *MsgRemoveAuthenticator) Reset()         { *m = MsgRemoveAuthenticator{} }
func (m *MsgRemoveAuthenticator) String() string { return proto.CompactTextString(m) }
func (*MsgRemoveAuthenticator) ProtoMessage()    {}
func (*MsgRemoveAuthenticator) Descriptor() ([]byte, []int) {
	return fileDescriptor_1aa1d7e4dc71ed44, []int{2}
}
func (m *MsgRemoveAuthenticator) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRemoveAuthenticator) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRemoveAuthenticator.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRemoveAuthenticator) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRemoveAuthenticator.Merge(m, src)
}
func (m *MsgRemoveAuthenticator) XXX_Size() int {
	return m.Size()
}
func (m *MsgRemoveAuthenticator) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRemoveAuthenticator.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRemoveAuthenticator proto.InternalMessageInfo

func (m *MsgRemoveAuthenticator) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

func (m *MsgRemoveAuthenticator) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

// MsgRemoveAuthenticatorResponse defines the Msg/RemoveAuthenticator response
// type.
type MsgRemoveAuthenticatorResponse struct {
	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (m *MsgRemoveAuthenticatorResponse) Reset()         { *m = MsgRemoveAuthenticatorResponse{} }
func (m *MsgRemoveAuthenticatorResponse) String() string { return proto.CompactTextString(m) }
func (*MsgRemoveAuthenticatorResponse) ProtoMessage()    {}
func (*MsgRemoveAuthenticatorResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1aa1d7e4dc71ed44, []int{3}
}
func (m *MsgRemoveAuthenticatorResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRemoveAuthenticatorResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRemoveAuthenticatorResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRemoveAuthenticatorResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRemoveAuthenticatorResponse.Merge(m, src)
}
func (m *MsgRemoveAuthenticatorResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgRemoveAuthenticatorResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRemoveAuthenticatorResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRemoveAuthenticatorResponse proto.InternalMessageInfo

func (m *MsgRemoveAuthenticatorResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

// TxExtension allows for additional authenticator-specific data in
// transactions.
type TxExtension struct {
	// selected_authenticators holds indices for the chosen authenticators per
	// message.
	SelectedAuthenticators []int32 `protobuf:"varint,1,rep,packed,name=selected_authenticators,json=selectedAuthenticators,proto3" json:"selected_authenticators,omitempty"`
}

func (m *TxExtension) Reset()         { *m = TxExtension{} }
func (m *TxExtension) String() string { return proto.CompactTextString(m) }
func (*TxExtension) ProtoMessage()    {}
func (*TxExtension) Descriptor() ([]byte, []int) {
	return fileDescriptor_1aa1d7e4dc71ed44, []int{4}
}
func (m *TxExtension) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TxExtension) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TxExtension.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TxExtension) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TxExtension.Merge(m, src)
}
func (m *TxExtension) XXX_Size() int {
	return m.Size()
}
func (m *TxExtension) XXX_DiscardUnknown() {
	xxx_messageInfo_TxExtension.DiscardUnknown(m)
}

var xxx_messageInfo_TxExtension proto.InternalMessageInfo

func (m *TxExtension) GetSelectedAuthenticators() []int32 {
	if m != nil {
		return m.SelectedAuthenticators
	}
	return nil
}

func init() {
	proto.RegisterType((*MsgAddAuthenticator)(nil), "osmosis.authenticator.MsgAddAuthenticator")
	proto.RegisterType((*MsgAddAuthenticatorResponse)(nil), "osmosis.authenticator.MsgAddAuthenticatorResponse")
	proto.RegisterType((*MsgRemoveAuthenticator)(nil), "osmosis.authenticator.MsgRemoveAuthenticator")
	proto.RegisterType((*MsgRemoveAuthenticatorResponse)(nil), "osmosis.authenticator.MsgRemoveAuthenticatorResponse")
	proto.RegisterType((*TxExtension)(nil), "osmosis.authenticator.TxExtension")
}

func init() { proto.RegisterFile("osmosis/authenticator/tx.proto", fileDescriptor_1aa1d7e4dc71ed44) }

var fileDescriptor_1aa1d7e4dc71ed44 = []byte{
	// 345 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0x41, 0x4b, 0xf3, 0x40,
	0x10, 0x6d, 0xda, 0x7e, 0xfd, 0xbe, 0x6f, 0x14, 0x91, 0x2d, 0xd6, 0x50, 0x61, 0x29, 0x39, 0x15,
	0xa1, 0x09, 0x56, 0xa4, 0xd4, 0x93, 0x15, 0xf4, 0xd6, 0xcb, 0xaa, 0x17, 0x2f, 0x92, 0x66, 0x87,
	0x34, 0xd0, 0x66, 0x43, 0x66, 0x5b, 0x52, 0xfc, 0x13, 0xfe, 0x2c, 0x8f, 0x3d, 0x7a, 0x94, 0xf6,
	0x3f, 0x78, 0x96, 0x86, 0x46, 0xac, 0x46, 0xa8, 0xb7, 0x99, 0x79, 0x2f, 0xef, 0x4d, 0xde, 0x0e,
	0x70, 0x45, 0x63, 0x45, 0x01, 0x39, 0xee, 0x44, 0x0f, 0x31, 0xd4, 0x81, 0xe7, 0x6a, 0x15, 0x3b,
	0x3a, 0xb1, 0xa3, 0x58, 0x69, 0xc5, 0x0e, 0xd6, 0xb8, 0xbd, 0x81, 0x5b, 0x77, 0x50, 0xed, 0x93,
	0xdf, 0x93, 0xb2, 0xf7, 0x79, 0xcc, 0x6a, 0x50, 0x21, 0x0c, 0x25, 0xc6, 0xa6, 0xd1, 0x30, 0x9a,
	0xff, 0xc5, 0xba, 0x63, 0x0c, 0xca, 0x7a, 0x16, 0xa1, 0x59, 0x4c, 0xa7, 0x69, 0xbd, 0x9a, 0x49,
	0x57, 0xbb, 0x66, 0xa9, 0x61, 0x34, 0x77, 0x45, 0x5a, 0x5b, 0x1d, 0x38, 0xca, 0x91, 0x15, 0x48,
	0x91, 0x0a, 0x09, 0x99, 0x09, 0x7f, 0x69, 0xe2, 0x79, 0x48, 0x94, 0xea, 0xff, 0x13, 0x59, 0x6b,
	0x5d, 0x40, 0xad, 0x4f, 0xbe, 0xc0, 0xb1, 0x9a, 0xe2, 0x76, 0x2b, 0xed, 0x41, 0x31, 0x90, 0xe9,
	0x42, 0x65, 0x51, 0x0c, 0xa4, 0x75, 0x0e, 0x3c, 0x5f, 0x61, 0x0b, 0xf7, 0x6b, 0xd8, 0xb9, 0x4d,
	0xae, 0x12, 0x8d, 0x21, 0x05, 0x2a, 0x64, 0x1d, 0x38, 0x24, 0x1c, 0xa1, 0xa7, 0x51, 0x3e, 0x6c,
	0xc4, 0xb6, 0xfa, 0xb0, 0xd4, 0xfc, 0x23, 0x6a, 0x19, 0xbc, 0x61, 0x44, 0xed, 0x37, 0x03, 0x4a,
	0x7d, 0xf2, 0x59, 0x0c, 0xfb, 0xdf, 0xa2, 0x3d, 0xb6, 0x73, 0x5f, 0xc2, 0xce, 0xc9, 0xab, 0xde,
	0xde, 0x9e, 0xfb, 0xf1, 0x77, 0x8f, 0x50, 0xcd, 0x8b, 0xaf, 0xf5, 0xb3, 0x54, 0x0e, 0xbd, 0x7e,
	0xf6, 0x2b, 0x7a, 0x66, 0x7e, 0x79, 0xf3, 0xbc, 0xe0, 0xc6, 0x7c, 0xc1, 0x8d, 0xd7, 0x05, 0x37,
	0x9e, 0x96, 0xbc, 0x30, 0x5f, 0xf2, 0xc2, 0xcb, 0x92, 0x17, 0xee, 0xbb, 0x7e, 0xa0, 0x87, 0x93,
	0x81, 0xed, 0xa9, 0xb1, 0xb3, 0x96, 0x6e, 0x8d, 0xdc, 0x01, 0x65, 0x8d, 0x33, 0x3d, 0xe9, 0x3a,
	0xc9, 0xd7, 0xeb, 0x9d, 0x45, 0x48, 0x83, 0x4a, 0x7a, 0xc1, 0xa7, 0xef, 0x01, 0x00, 0x00, 0xff,
	0xff, 0x09, 0x34, 0xf7, 0x66, 0xe3, 0x02, 0x00, 0x00,
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
	AddAuthenticator(ctx context.Context, in *MsgAddAuthenticator, opts ...grpc.CallOption) (*MsgAddAuthenticatorResponse, error)
	RemoveAuthenticator(ctx context.Context, in *MsgRemoveAuthenticator, opts ...grpc.CallOption) (*MsgRemoveAuthenticatorResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) AddAuthenticator(ctx context.Context, in *MsgAddAuthenticator, opts ...grpc.CallOption) (*MsgAddAuthenticatorResponse, error) {
	out := new(MsgAddAuthenticatorResponse)
	err := c.cc.Invoke(ctx, "/osmosis.authenticator.Msg/AddAuthenticator", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RemoveAuthenticator(ctx context.Context, in *MsgRemoveAuthenticator, opts ...grpc.CallOption) (*MsgRemoveAuthenticatorResponse, error) {
	out := new(MsgRemoveAuthenticatorResponse)
	err := c.cc.Invoke(ctx, "/osmosis.authenticator.Msg/RemoveAuthenticator", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	AddAuthenticator(context.Context, *MsgAddAuthenticator) (*MsgAddAuthenticatorResponse, error)
	RemoveAuthenticator(context.Context, *MsgRemoveAuthenticator) (*MsgRemoveAuthenticatorResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) AddAuthenticator(ctx context.Context, req *MsgAddAuthenticator) (*MsgAddAuthenticatorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddAuthenticator not implemented")
}
func (*UnimplementedMsgServer) RemoveAuthenticator(ctx context.Context, req *MsgRemoveAuthenticator) (*MsgRemoveAuthenticatorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveAuthenticator not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_AddAuthenticator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAddAuthenticator)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).AddAuthenticator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.authenticator.Msg/AddAuthenticator",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).AddAuthenticator(ctx, req.(*MsgAddAuthenticator))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RemoveAuthenticator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRemoveAuthenticator)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RemoveAuthenticator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.authenticator.Msg/RemoveAuthenticator",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RemoveAuthenticator(ctx, req.(*MsgRemoveAuthenticator))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "osmosis.authenticator.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddAuthenticator",
			Handler:    _Msg_AddAuthenticator_Handler,
		},
		{
			MethodName: "RemoveAuthenticator",
			Handler:    _Msg_RemoveAuthenticator_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "osmosis/authenticator/tx.proto",
}

func (m *MsgAddAuthenticator) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgAddAuthenticator) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgAddAuthenticator) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		i -= len(m.Data)
		copy(dAtA[i:], m.Data)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Data)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Type) > 0 {
		i -= len(m.Type)
		copy(dAtA[i:], m.Type)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Type)))
		i--
		dAtA[i] = 0x12
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

func (m *MsgAddAuthenticatorResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgAddAuthenticatorResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgAddAuthenticatorResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Success {
		i--
		if m.Success {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *MsgRemoveAuthenticator) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRemoveAuthenticator) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRemoveAuthenticator) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Id != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Id))
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

func (m *MsgRemoveAuthenticatorResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRemoveAuthenticatorResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRemoveAuthenticatorResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Success {
		i--
		if m.Success {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *TxExtension) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TxExtension) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TxExtension) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SelectedAuthenticators) > 0 {
		dAtA2 := make([]byte, len(m.SelectedAuthenticators)*10)
		var j1 int
		for _, num1 := range m.SelectedAuthenticators {
			num := uint64(num1)
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
		i = encodeVarintTx(dAtA, i, uint64(j1))
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
func (m *MsgAddAuthenticator) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Type)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgAddAuthenticatorResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Success {
		n += 2
	}
	return n
}

func (m *MsgRemoveAuthenticator) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Id != 0 {
		n += 1 + sovTx(uint64(m.Id))
	}
	return n
}

func (m *MsgRemoveAuthenticatorResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Success {
		n += 2
	}
	return n
}

func (m *TxExtension) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.SelectedAuthenticators) > 0 {
		l = 0
		for _, e := range m.SelectedAuthenticators {
			l += sovTx(uint64(e))
		}
		n += 1 + sovTx(uint64(l)) + l
	}
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgAddAuthenticator) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgAddAuthenticator: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgAddAuthenticator: illegal tag %d (wire type %d)", fieldNum, wire)
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
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
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
			m.Type = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
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
func (m *MsgAddAuthenticatorResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgAddAuthenticatorResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgAddAuthenticatorResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Success", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
			m.Success = bool(v != 0)
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
func (m *MsgRemoveAuthenticator) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgRemoveAuthenticator: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRemoveAuthenticator: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
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
func (m *MsgRemoveAuthenticatorResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgRemoveAuthenticatorResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRemoveAuthenticatorResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Success", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
			m.Success = bool(v != 0)
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
func (m *TxExtension) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: TxExtension: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TxExtension: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType == 0 {
				var v int32
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTx
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= int32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.SelectedAuthenticators = append(m.SelectedAuthenticators, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTx
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
					return ErrInvalidLengthTx
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthTx
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
				if elementCount != 0 && len(m.SelectedAuthenticators) == 0 {
					m.SelectedAuthenticators = make([]int32, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v int32
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowTx
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= int32(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.SelectedAuthenticators = append(m.SelectedAuthenticators, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field SelectedAuthenticators", wireType)
			}
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
