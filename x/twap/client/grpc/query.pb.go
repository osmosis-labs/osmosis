// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/twap/v1beta1/query.proto

package grpc

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/cosmos/cosmos-sdk/types"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/query"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/osmosis-labs/osmosis/v10/x/twap/types"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type GetArithmeticTwapRequest struct {
	PoolId     uint64 `protobuf:"varint,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	BaseAsset  string `protobuf:"bytes,2,opt,name=base_asset,json=baseAsset,proto3" json:"base_asset,omitempty"`
	QuoteAsset string `protobuf:"bytes,3,opt,name=quote_asset,json=quoteAsset,proto3" json:"quote_asset,omitempty"`
	StartTime  string `protobuf:"bytes,4,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	EndTime    string `protobuf:"bytes,5,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`
}

func (m *GetArithmeticTwapRequest) Reset()         { *m = GetArithmeticTwapRequest{} }
func (m *GetArithmeticTwapRequest) String() string { return proto.CompactTextString(m) }
func (*GetArithmeticTwapRequest) ProtoMessage()    {}
func (*GetArithmeticTwapRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_141a22dba58615af, []int{0}
}
func (m *GetArithmeticTwapRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GetArithmeticTwapRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GetArithmeticTwapRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GetArithmeticTwapRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetArithmeticTwapRequest.Merge(m, src)
}
func (m *GetArithmeticTwapRequest) XXX_Size() int {
	return m.Size()
}
func (m *GetArithmeticTwapRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetArithmeticTwapRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetArithmeticTwapRequest proto.InternalMessageInfo

func (m *GetArithmeticTwapRequest) GetPoolId() uint64 {
	if m != nil {
		return m.PoolId
	}
	return 0
}

func (m *GetArithmeticTwapRequest) GetBaseAsset() string {
	if m != nil {
		return m.BaseAsset
	}
	return ""
}

func (m *GetArithmeticTwapRequest) GetQuoteAsset() string {
	if m != nil {
		return m.QuoteAsset
	}
	return ""
}

func (m *GetArithmeticTwapRequest) GetStartTime() string {
	if m != nil {
		return m.StartTime
	}
	return ""
}

func (m *GetArithmeticTwapRequest) GetEndTime() string {
	if m != nil {
		return m.EndTime
	}
	return ""
}

type GetArithmeticTwapResponse struct {
	ArithmeticTwap github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,1,opt,name=arithmetic_twap,json=arithmeticTwap,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"arithmetic_twap" yaml:"arithmetic_twap"`
}

func (m *GetArithmeticTwapResponse) Reset()         { *m = GetArithmeticTwapResponse{} }
func (m *GetArithmeticTwapResponse) String() string { return proto.CompactTextString(m) }
func (*GetArithmeticTwapResponse) ProtoMessage()    {}
func (*GetArithmeticTwapResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_141a22dba58615af, []int{1}
}
func (m *GetArithmeticTwapResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GetArithmeticTwapResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GetArithmeticTwapResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GetArithmeticTwapResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetArithmeticTwapResponse.Merge(m, src)
}
func (m *GetArithmeticTwapResponse) XXX_Size() int {
	return m.Size()
}
func (m *GetArithmeticTwapResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetArithmeticTwapResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetArithmeticTwapResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*GetArithmeticTwapRequest)(nil), "osmosis.gamm.twap.v1beta1.GetArithmeticTwapRequest")
	proto.RegisterType((*GetArithmeticTwapResponse)(nil), "osmosis.gamm.twap.v1beta1.GetArithmeticTwapResponse")
}

func init() { proto.RegisterFile("osmosis/twap/v1beta1/query.proto", fileDescriptor_141a22dba58615af) }

var fileDescriptor_141a22dba58615af = []byte{
	// 496 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0xc7, 0xb3, 0xa5, 0x1f, 0x64, 0x91, 0x40, 0x58, 0x08, 0x92, 0x40, 0x9c, 0x28, 0x48, 0x55,
	0x85, 0x54, 0x2f, 0xa1, 0x88, 0x03, 0xb7, 0x56, 0x48, 0xc0, 0x0d, 0xa2, 0x9e, 0xb8, 0x58, 0x6b,
	0x7b, 0x70, 0x57, 0xd8, 0x1e, 0xc7, 0xbb, 0x6e, 0xc9, 0x95, 0x07, 0x40, 0x48, 0x3c, 0x02, 0xe7,
	0xbe, 0x47, 0x8f, 0x95, 0xb8, 0x20, 0x0e, 0x11, 0x4a, 0x78, 0x02, 0x9e, 0x00, 0xed, 0x47, 0x42,
	0x81, 0xf6, 0xd0, 0x93, 0x3d, 0xf3, 0xfb, 0xcf, 0xec, 0xec, 0xec, 0x9f, 0xf6, 0x51, 0xe6, 0x28,
	0x85, 0x64, 0xea, 0x88, 0x97, 0xec, 0x70, 0x18, 0x81, 0xe2, 0x43, 0x36, 0xae, 0xa1, 0x9a, 0x04,
	0x65, 0x85, 0x0a, 0xbd, 0xb6, 0x53, 0x04, 0x29, 0xcf, 0xf3, 0x40, 0xcb, 0x02, 0x27, 0xeb, 0xdc,
	0x4a, 0x31, 0x45, 0xa3, 0x62, 0xfa, 0xcf, 0x16, 0x74, 0x36, 0xcf, 0x6d, 0xa9, 0x83, 0xb0, 0x82,
	0x18, 0xab, 0xc4, 0xe9, 0xfc, 0xd8, 0x08, 0x59, 0xc4, 0x25, 0x2c, 0x65, 0x31, 0x8a, 0xc2, 0xf1,
	0x07, 0x67, 0xb9, 0x99, 0x68, 0xa9, 0x2a, 0x79, 0x2a, 0x0a, 0xae, 0x04, 0x2e, 0xb4, 0xf7, 0x52,
	0xc4, 0x34, 0x03, 0xc6, 0x4b, 0xc1, 0x78, 0x51, 0xa0, 0x32, 0x50, 0x3a, 0xda, 0x76, 0xd4, 0x44,
	0x51, 0xfd, 0x96, 0xf1, 0x62, 0xb2, 0x40, 0xf6, 0x90, 0xd0, 0xde, 0xc2, 0x06, 0x16, 0x0d, 0x8e,
	0x09, 0x6d, 0x3d, 0x07, 0xb5, 0x5b, 0x09, 0x75, 0x90, 0x83, 0x12, 0xf1, 0xfe, 0x11, 0x2f, 0x47,
	0x30, 0xae, 0x41, 0x2a, 0xef, 0x0e, 0xdd, 0x28, 0x11, 0xb3, 0x50, 0x24, 0x2d, 0xd2, 0x27, 0x5b,
	0xab, 0xa3, 0x75, 0x1d, 0xbe, 0x4c, 0xbc, 0x2e, 0xa5, 0x7a, 0xe0, 0x90, 0x4b, 0x09, 0xaa, 0xb5,
	0xd2, 0x27, 0x5b, 0xcd, 0x51, 0x53, 0x67, 0x76, 0x75, 0xc2, 0xeb, 0xd1, 0x6b, 0xe3, 0x1a, 0xd5,
	0x82, 0x5f, 0x31, 0x9c, 0x9a, 0x94, 0x15, 0x74, 0x29, 0x95, 0x8a, 0x57, 0x2a, 0x54, 0x22, 0x87,
	0xd6, 0xaa, 0xad, 0x37, 0x99, 0x7d, 0x91, 0x83, 0xd7, 0xa6, 0x57, 0xa1, 0x48, 0x2c, 0x5c, 0x33,
	0x70, 0x03, 0x8a, 0x44, 0xa3, 0xc1, 0x47, 0x42, 0xdb, 0xe7, 0xcc, 0x2b, 0x4b, 0x2c, 0x24, 0x78,
	0x63, 0x7a, 0x83, 0x2f, 0x49, 0xa8, 0x5f, 0xc3, 0x0c, 0xde, 0xdc, 0x7b, 0x71, 0x32, 0xed, 0x35,
	0xbe, 0x4f, 0x7b, 0x9b, 0xa9, 0x50, 0x07, 0x75, 0x14, 0xc4, 0x98, 0xbb, 0x3d, 0xb8, 0xcf, 0xb6,
	0x4c, 0xde, 0x31, 0x35, 0x29, 0x41, 0x06, 0xcf, 0x20, 0xfe, 0x35, 0xed, 0xdd, 0x9e, 0xf0, 0x3c,
	0x7b, 0x3a, 0xf8, 0xa7, 0xdd, 0x60, 0x74, 0x9d, 0xff, 0x75, 0xf4, 0xa3, 0x63, 0x42, 0xd7, 0x5e,
	0xeb, 0x77, 0xf3, 0xbe, 0x10, 0x7a, 0xf3, 0xbf, 0xd1, 0xbc, 0x9d, 0xe0, 0x42, 0x6b, 0x05, 0x17,
	0x2d, 0xbe, 0xf3, 0xf8, 0x72, 0x45, 0xf6, 0xf6, 0x83, 0xfb, 0x1f, 0xbe, 0xfe, 0xfc, 0xbc, 0xd2,
	0xf5, 0xee, 0xb2, 0x85, 0x39, 0x75, 0xf5, 0x1f, 0x3f, 0x21, 0x66, 0x72, 0xef, 0xd5, 0xc9, 0xcc,
	0x27, 0xa7, 0x33, 0x9f, 0xfc, 0x98, 0xf9, 0xe4, 0xd3, 0xdc, 0x6f, 0x9c, 0xce, 0xfd, 0xc6, 0xb7,
	0xb9, 0xdf, 0x78, 0xf3, 0xe4, 0xcc, 0x6e, 0x5c, 0x83, 0xed, 0x8c, 0x47, 0x72, 0xd9, 0xed, 0x70,
	0xf8, 0x90, 0xbd, 0xb7, 0x86, 0x8f, 0x33, 0x01, 0x85, 0x62, 0x69, 0x55, 0xc6, 0xd1, 0xba, 0x71,
	0xd2, 0xce, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0xce, 0x64, 0xe3, 0xa0, 0x66, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	GetArithmeticTwap(ctx context.Context, in *GetArithmeticTwapRequest, opts ...grpc.CallOption) (*GetArithmeticTwapResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) GetArithmeticTwap(ctx context.Context, in *GetArithmeticTwapRequest, opts ...grpc.CallOption) (*GetArithmeticTwapResponse, error) {
	out := new(GetArithmeticTwapResponse)
	err := c.cc.Invoke(ctx, "/osmosis.gamm.twap.v1beta1.Query/GetArithmeticTwap", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	GetArithmeticTwap(context.Context, *GetArithmeticTwapRequest) (*GetArithmeticTwapResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) GetArithmeticTwap(ctx context.Context, req *GetArithmeticTwapRequest) (*GetArithmeticTwapResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetArithmeticTwap not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_GetArithmeticTwap_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetArithmeticTwapRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GetArithmeticTwap(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.gamm.twap.v1beta1.Query/GetArithmeticTwap",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GetArithmeticTwap(ctx, req.(*GetArithmeticTwapRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "osmosis.gamm.twap.v1beta1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetArithmeticTwap",
			Handler:    _Query_GetArithmeticTwap_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "osmosis/twap/v1beta1/query.proto",
}

func (m *GetArithmeticTwapRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetArithmeticTwapRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GetArithmeticTwapRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.EndTime) > 0 {
		i -= len(m.EndTime)
		copy(dAtA[i:], m.EndTime)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.EndTime)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.StartTime) > 0 {
		i -= len(m.StartTime)
		copy(dAtA[i:], m.StartTime)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.StartTime)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.QuoteAsset) > 0 {
		i -= len(m.QuoteAsset)
		copy(dAtA[i:], m.QuoteAsset)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.QuoteAsset)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.BaseAsset) > 0 {
		i -= len(m.BaseAsset)
		copy(dAtA[i:], m.BaseAsset)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.BaseAsset)))
		i--
		dAtA[i] = 0x12
	}
	if m.PoolId != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.PoolId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *GetArithmeticTwapResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetArithmeticTwapResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GetArithmeticTwapResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.ArithmeticTwap.Size()
		i -= size
		if _, err := m.ArithmeticTwap.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GetArithmeticTwapRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PoolId != 0 {
		n += 1 + sovQuery(uint64(m.PoolId))
	}
	l = len(m.BaseAsset)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = len(m.QuoteAsset)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = len(m.StartTime)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = len(m.EndTime)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *GetArithmeticTwapResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.ArithmeticTwap.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GetArithmeticTwapRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: GetArithmeticTwapRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetArithmeticTwapRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolId", wireType)
			}
			m.PoolId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return fmt.Errorf("proto: wrong wireType = %d for field BaseAsset", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BaseAsset = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field QuoteAsset", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.QuoteAsset = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StartTime", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.StartTime = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EndTime", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.EndTime = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *GetArithmeticTwapResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: GetArithmeticTwapResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetArithmeticTwapResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ArithmeticTwap", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ArithmeticTwap.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)
