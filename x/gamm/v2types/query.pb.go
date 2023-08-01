// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/gamm/v2/query.proto

package v2types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/query"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/osmosis-labs/osmosis/v17/x/gamm/types"
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

// Deprecated: please use alternate in x/poolmanager
//
// Deprecated: Do not use.
type QuerySpotPriceRequest struct {
	PoolId          uint64 `protobuf:"varint,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty" yaml:"pool_id"`
	BaseAssetDenom  string `protobuf:"bytes,2,opt,name=base_asset_denom,json=baseAssetDenom,proto3" json:"base_asset_denom,omitempty" yaml:"base_asset_denom"`
	QuoteAssetDenom string `protobuf:"bytes,3,opt,name=quote_asset_denom,json=quoteAssetDenom,proto3" json:"quote_asset_denom,omitempty" yaml:"quote_asset_denom"`
}

func (m *QuerySpotPriceRequest) Reset()         { *m = QuerySpotPriceRequest{} }
func (m *QuerySpotPriceRequest) String() string { return proto.CompactTextString(m) }
func (*QuerySpotPriceRequest) ProtoMessage()    {}
func (*QuerySpotPriceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_49ff000e88fc374c, []int{0}
}
func (m *QuerySpotPriceRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QuerySpotPriceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QuerySpotPriceRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QuerySpotPriceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QuerySpotPriceRequest.Merge(m, src)
}
func (m *QuerySpotPriceRequest) XXX_Size() int {
	return m.Size()
}
func (m *QuerySpotPriceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QuerySpotPriceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QuerySpotPriceRequest proto.InternalMessageInfo

func (m *QuerySpotPriceRequest) GetPoolId() uint64 {
	if m != nil {
		return m.PoolId
	}
	return 0
}

func (m *QuerySpotPriceRequest) GetBaseAssetDenom() string {
	if m != nil {
		return m.BaseAssetDenom
	}
	return ""
}

func (m *QuerySpotPriceRequest) GetQuoteAssetDenom() string {
	if m != nil {
		return m.QuoteAssetDenom
	}
	return ""
}

// Depreacted: please use alternate in x/poolmanager
//
// Deprecated: Do not use.
type QuerySpotPriceResponse struct {
	// String of the Dec. Ex) 10.203uatom
	SpotPrice string `protobuf:"bytes,1,opt,name=spot_price,json=spotPrice,proto3" json:"spot_price,omitempty" yaml:"spot_price"`
}

func (m *QuerySpotPriceResponse) Reset()         { *m = QuerySpotPriceResponse{} }
func (m *QuerySpotPriceResponse) String() string { return proto.CompactTextString(m) }
func (*QuerySpotPriceResponse) ProtoMessage()    {}
func (*QuerySpotPriceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_49ff000e88fc374c, []int{1}
}
func (m *QuerySpotPriceResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QuerySpotPriceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QuerySpotPriceResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QuerySpotPriceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QuerySpotPriceResponse.Merge(m, src)
}
func (m *QuerySpotPriceResponse) XXX_Size() int {
	return m.Size()
}
func (m *QuerySpotPriceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QuerySpotPriceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QuerySpotPriceResponse proto.InternalMessageInfo

func (m *QuerySpotPriceResponse) GetSpotPrice() string {
	if m != nil {
		return m.SpotPrice
	}
	return ""
}

func init() {
	proto.RegisterType((*QuerySpotPriceRequest)(nil), "osmosis.gamm.v2.QuerySpotPriceRequest")
	proto.RegisterType((*QuerySpotPriceResponse)(nil), "osmosis.gamm.v2.QuerySpotPriceResponse")
}

func init() { proto.RegisterFile("osmosis/gamm/v2/query.proto", fileDescriptor_49ff000e88fc374c) }

var fileDescriptor_49ff000e88fc374c = []byte{
	// 469 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0x41, 0x6f, 0xd3, 0x30,
	0x14, 0xc7, 0xeb, 0x02, 0x43, 0xf5, 0x61, 0x63, 0x16, 0x83, 0xd2, 0x8d, 0x74, 0xca, 0x81, 0x0d,
	0x10, 0xb1, 0x1a, 0x90, 0x90, 0x76, 0xa3, 0x02, 0x09, 0x24, 0x0e, 0x10, 0x6e, 0x5c, 0x22, 0xa7,
	0x35, 0xc1, 0x52, 0x93, 0xe7, 0xd6, 0x4e, 0xb5, 0x0a, 0x71, 0xe1, 0xc4, 0x71, 0x12, 0x37, 0x3e,
	0x11, 0xc7, 0x49, 0x5c, 0xe0, 0x52, 0xa1, 0x96, 0x4f, 0xd0, 0x4f, 0x80, 0x6c, 0x27, 0x1d, 0x0b,
	0x48, 0xbb, 0xc5, 0xef, 0xf7, 0xf7, 0xff, 0xff, 0x9e, 0xf3, 0xf0, 0x2e, 0xa8, 0x0c, 0x94, 0x50,
	0x34, 0x65, 0x59, 0x46, 0xa7, 0x21, 0x1d, 0x17, 0x7c, 0x32, 0x0b, 0xe4, 0x04, 0x34, 0x90, 0xad,
	0x12, 0x06, 0x06, 0x06, 0xd3, 0xb0, 0x73, 0x3d, 0x85, 0x14, 0x2c, 0xa3, 0xe6, 0xcb, 0xc9, 0x3a,
	0xb7, 0xcf, 0x7b, 0xf4, 0x12, 0xae, 0x59, 0x8f, 0xea, 0xe3, 0x12, 0x7b, 0x03, 0xcb, 0x69, 0xc2,
	0x14, 0x5f, 0xd3, 0x01, 0x88, 0xbc, 0xe4, 0xf7, 0xfe, 0xe6, 0x36, 0x7e, 0xad, 0x92, 0x2c, 0x15,
	0x39, 0xd3, 0x02, 0x2a, 0xed, 0x5e, 0x0a, 0x90, 0x8e, 0x38, 0x65, 0x52, 0x50, 0x96, 0xe7, 0xa0,
	0x2d, 0x54, 0x25, 0xbd, 0x55, 0x52, 0x7b, 0x4a, 0x8a, 0x77, 0x94, 0xe5, 0xb3, 0x0a, 0xb9, 0x90,
	0xd8, 0x35, 0xef, 0x0e, 0x0e, 0xf9, 0x3f, 0x11, 0xde, 0x79, 0x6d, 0x62, 0xdf, 0x48, 0xd0, 0xaf,
	0x26, 0x62, 0xc0, 0x23, 0x3e, 0x2e, 0xb8, 0xd2, 0xe4, 0x3e, 0xbe, 0x2a, 0x01, 0x46, 0xb1, 0x18,
	0xb6, 0xd1, 0x3e, 0x3a, 0xbc, 0xdc, 0x27, 0xab, 0x79, 0x77, 0x73, 0xc6, 0xb2, 0xd1, 0x91, 0x5f,
	0x02, 0x3f, 0xda, 0x30, 0x5f, 0x2f, 0x86, 0xe4, 0x19, 0xbe, 0x66, 0x26, 0x88, 0x99, 0x52, 0x5c,
	0xc7, 0x43, 0x9e, 0x43, 0xd6, 0x6e, 0xee, 0xa3, 0xc3, 0x56, 0x7f, 0x77, 0x35, 0xef, 0xde, 0x74,
	0xb7, 0xea, 0x0a, 0x3f, 0xda, 0x34, 0xa5, 0x27, 0xa6, 0xf2, 0xd4, 0x14, 0xc8, 0x73, 0xbc, 0x3d,
	0x2e, 0x40, 0x9f, 0xf7, 0xb9, 0x64, 0x7d, 0xf6, 0x56, 0xf3, 0x6e, 0xdb, 0xf9, 0xfc, 0x23, 0xf1,
	0xa3, 0x2d, 0x5b, 0x3b, 0x73, 0x3a, 0x6a, 0xb6, 0x91, 0x1f, 0xe1, 0x1b, 0xf5, 0xd1, 0x94, 0x84,
	0x5c, 0x71, 0xf2, 0x08, 0x63, 0x25, 0x41, 0xc7, 0xd2, 0x54, 0xed, 0x78, 0xad, 0xfe, 0xce, 0x6a,
	0xde, 0xdd, 0x76, 0x01, 0x67, 0xcc, 0x8f, 0x5a, 0xaa, 0xba, 0x6d, 0x3c, 0xc3, 0xaf, 0x08, 0x5f,
	0xb1, 0xa6, 0xe4, 0x04, 0xe1, 0xd6, 0xda, 0x99, 0xdc, 0x09, 0x6a, 0xeb, 0x12, 0xfc, 0xf7, 0x55,
	0x3b, 0x07, 0x17, 0xea, 0x5c, 0x8b, 0x7e, 0xf8, 0xe9, 0xfb, 0xef, 0x2f, 0xcd, 0xbb, 0xe4, 0x80,
	0xd6, 0x97, 0xd4, 0x3c, 0xb9, 0xa2, 0x1f, 0xca, 0x7f, 0xf0, 0x91, 0xda, 0x46, 0xd5, 0xe7, 0x26,
	0xea, 0xbf, 0xfc, 0xb6, 0xf0, 0xd0, 0xe9, 0xc2, 0x43, 0xbf, 0x16, 0x1e, 0x3a, 0x59, 0x7a, 0x8d,
	0xd3, 0xa5, 0xd7, 0xf8, 0xb1, 0xf4, 0x1a, 0x6f, 0xc3, 0x54, 0xe8, 0xf7, 0x45, 0x12, 0x0c, 0x20,
	0xab, 0xfc, 0x1e, 0x8c, 0x58, 0xa2, 0xd6, 0xe6, 0xd3, 0xde, 0x63, 0x7a, 0x5c, 0x45, 0xe8, 0x99,
	0xe4, 0x2a, 0xd9, 0xb0, 0x1b, 0xf2, 0xf0, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x72, 0x56, 0x2d,
	0xc1, 0x26, 0x03, 0x00, 0x00,
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
	// Deprecated: please use alternate in x/poolmanager
	SpotPrice(ctx context.Context, in *QuerySpotPriceRequest, opts ...grpc.CallOption) (*QuerySpotPriceResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

// Deprecated: Do not use.
func (c *queryClient) SpotPrice(ctx context.Context, in *QuerySpotPriceRequest, opts ...grpc.CallOption) (*QuerySpotPriceResponse, error) {
	out := new(QuerySpotPriceResponse)
	err := c.cc.Invoke(ctx, "/osmosis.gamm.v2.Query/SpotPrice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// Deprecated: please use alternate in x/poolmanager
	SpotPrice(context.Context, *QuerySpotPriceRequest) (*QuerySpotPriceResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) SpotPrice(ctx context.Context, req *QuerySpotPriceRequest) (*QuerySpotPriceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SpotPrice not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_SpotPrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuerySpotPriceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).SpotPrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.gamm.v2.Query/SpotPrice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).SpotPrice(ctx, req.(*QuerySpotPriceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "osmosis.gamm.v2.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SpotPrice",
			Handler:    _Query_SpotPrice_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "osmosis/gamm/v2/query.proto",
}

func (m *QuerySpotPriceRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QuerySpotPriceRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QuerySpotPriceRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.QuoteAssetDenom) > 0 {
		i -= len(m.QuoteAssetDenom)
		copy(dAtA[i:], m.QuoteAssetDenom)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.QuoteAssetDenom)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.BaseAssetDenom) > 0 {
		i -= len(m.BaseAssetDenom)
		copy(dAtA[i:], m.BaseAssetDenom)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.BaseAssetDenom)))
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

func (m *QuerySpotPriceResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QuerySpotPriceResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QuerySpotPriceResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SpotPrice) > 0 {
		i -= len(m.SpotPrice)
		copy(dAtA[i:], m.SpotPrice)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.SpotPrice)))
		i--
		dAtA[i] = 0xa
	}
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
func (m *QuerySpotPriceRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PoolId != 0 {
		n += 1 + sovQuery(uint64(m.PoolId))
	}
	l = len(m.BaseAssetDenom)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = len(m.QuoteAssetDenom)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QuerySpotPriceResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.SpotPrice)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QuerySpotPriceRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QuerySpotPriceRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QuerySpotPriceRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field BaseAssetDenom", wireType)
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
			m.BaseAssetDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field QuoteAssetDenom", wireType)
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
			m.QuoteAssetDenom = string(dAtA[iNdEx:postIndex])
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
func (m *QuerySpotPriceResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QuerySpotPriceResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QuerySpotPriceResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpotPrice", wireType)
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
			m.SpotPrice = string(dAtA[iNdEx:postIndex])
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
