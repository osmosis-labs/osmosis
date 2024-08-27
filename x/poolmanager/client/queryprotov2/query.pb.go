// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/poolmanager/v2/query.proto

package queryprotov2

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_osmosis_labs_osmosis_osmomath "github.com/osmosis-labs/osmosis/osmomath"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	_ "google.golang.org/protobuf/types/known/timestamppb"
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

// SpotPriceRequest defines the gRPC request structure for a SpotPrice
// query.
type SpotPriceRequest struct {
	PoolId          uint64 `protobuf:"varint,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty" yaml:"pool_id"`
	BaseAssetDenom  string `protobuf:"bytes,2,opt,name=base_asset_denom,json=baseAssetDenom,proto3" json:"base_asset_denom,omitempty" yaml:"base_asset_denom"`
	QuoteAssetDenom string `protobuf:"bytes,3,opt,name=quote_asset_denom,json=quoteAssetDenom,proto3" json:"quote_asset_denom,omitempty" yaml:"quote_asset_denom"`
}

func (m *SpotPriceRequest) Reset()         { *m = SpotPriceRequest{} }
func (m *SpotPriceRequest) String() string { return proto.CompactTextString(m) }
func (*SpotPriceRequest) ProtoMessage()    {}
func (*SpotPriceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb2850debe8fb398, []int{0}
}
func (m *SpotPriceRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SpotPriceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SpotPriceRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SpotPriceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpotPriceRequest.Merge(m, src)
}
func (m *SpotPriceRequest) XXX_Size() int {
	return m.Size()
}
func (m *SpotPriceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SpotPriceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SpotPriceRequest proto.InternalMessageInfo

func (m *SpotPriceRequest) GetPoolId() uint64 {
	if m != nil {
		return m.PoolId
	}
	return 0
}

func (m *SpotPriceRequest) GetBaseAssetDenom() string {
	if m != nil {
		return m.BaseAssetDenom
	}
	return ""
}

func (m *SpotPriceRequest) GetQuoteAssetDenom() string {
	if m != nil {
		return m.QuoteAssetDenom
	}
	return ""
}

// SpotPriceResponse defines the gRPC response structure for a SpotPrice
// query.
type SpotPriceResponse struct {
	// String of the BigDec. Ex) 10.203uatom
	SpotPrice github_com_osmosis_labs_osmosis_osmomath.BigDec `protobuf:"bytes,1,opt,name=spot_price,json=spotPrice,proto3,customtype=github.com/osmosis-labs/osmosis/osmomath.BigDec" json:"spot_price" yaml:"spot_price"`
}

func (m *SpotPriceResponse) Reset()         { *m = SpotPriceResponse{} }
func (m *SpotPriceResponse) String() string { return proto.CompactTextString(m) }
func (*SpotPriceResponse) ProtoMessage()    {}
func (*SpotPriceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb2850debe8fb398, []int{1}
}
func (m *SpotPriceResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SpotPriceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SpotPriceResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SpotPriceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpotPriceResponse.Merge(m, src)
}
func (m *SpotPriceResponse) XXX_Size() int {
	return m.Size()
}
func (m *SpotPriceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SpotPriceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SpotPriceResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*SpotPriceRequest)(nil), "osmosis.poolmanager.v2.SpotPriceRequest")
	proto.RegisterType((*SpotPriceResponse)(nil), "osmosis.poolmanager.v2.SpotPriceResponse")
}

func init() {
	proto.RegisterFile("osmosis/poolmanager/v2/query.proto", fileDescriptor_eb2850debe8fb398)
}

var fileDescriptor_eb2850debe8fb398 = []byte{
	// 472 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x52, 0xc1, 0x8a, 0xd4, 0x4c,
	0x10, 0x9e, 0xde, 0xff, 0x77, 0x65, 0x5a, 0x58, 0x77, 0x82, 0xe8, 0x38, 0x2e, 0xc9, 0xd2, 0xa7,
	0x11, 0x31, 0x2d, 0x11, 0xf6, 0xe0, 0xcd, 0xb0, 0x0b, 0x0a, 0x1e, 0x34, 0x82, 0x07, 0x2f, 0xa1,
	0x93, 0x69, 0xb3, 0x8d, 0x49, 0x2a, 0x33, 0xdd, 0x19, 0x1c, 0x44, 0x10, 0x9f, 0x40, 0xf0, 0xe6,
	0xd9, 0x87, 0xd9, 0xe3, 0x80, 0x17, 0xf1, 0x10, 0x64, 0xc6, 0x27, 0x98, 0x27, 0x90, 0xee, 0xce,
	0xb8, 0xbb, 0xa3, 0xa2, 0xa7, 0x54, 0xd5, 0xf7, 0xf5, 0x57, 0x55, 0xa9, 0x0f, 0x13, 0x90, 0x05,
	0x48, 0x21, 0x69, 0x05, 0x90, 0x17, 0xac, 0x64, 0x19, 0x9f, 0xd0, 0x69, 0x40, 0xc7, 0x35, 0x9f,
	0xcc, 0xfc, 0x6a, 0x02, 0x0a, 0x9c, 0xab, 0x2d, 0xc7, 0x3f, 0xc3, 0xf1, 0xa7, 0xc1, 0xe0, 0x4a,
	0x06, 0x19, 0x18, 0x0a, 0xd5, 0x91, 0x65, 0x0f, 0xf6, 0x32, 0x80, 0x2c, 0xe7, 0x94, 0x55, 0x82,
	0xb2, 0xb2, 0x04, 0xc5, 0x94, 0x80, 0x52, 0xb6, 0xe8, 0xf5, 0x16, 0x35, 0x59, 0x52, 0xbf, 0xa0,
	0xac, 0x9c, 0xad, 0xa1, 0xd4, 0xf4, 0x89, 0xad, 0xa2, 0x4d, 0x5a, 0xc8, 0xdb, 0x7c, 0xa5, 0x44,
	0xc1, 0xa5, 0x62, 0x45, 0x65, 0x09, 0x64, 0x8e, 0xf0, 0xee, 0xd3, 0x0a, 0xd4, 0xe3, 0x89, 0x48,
	0x79, 0xc4, 0xc7, 0x35, 0x97, 0xca, 0xb9, 0x85, 0x2f, 0xea, 0x89, 0x63, 0x31, 0xea, 0xa3, 0x7d,
	0x34, 0xfc, 0x3f, 0x74, 0x56, 0x8d, 0xb7, 0x33, 0x63, 0x45, 0x7e, 0x8f, 0xb4, 0x00, 0x89, 0xb6,
	0x75, 0xf4, 0x70, 0xe4, 0x1c, 0xe1, 0xdd, 0x84, 0x49, 0x1e, 0x33, 0x29, 0xb9, 0x8a, 0x47, 0xbc,
	0x84, 0xa2, 0xbf, 0xb5, 0x8f, 0x86, 0xdd, 0xf0, 0xc6, 0xaa, 0xf1, 0xae, 0xd9, 0x57, 0x9b, 0x0c,
	0x12, 0xed, 0xe8, 0xd2, 0x7d, 0x5d, 0x39, 0xd4, 0x05, 0xe7, 0x01, 0xee, 0x8d, 0x6b, 0x50, 0xe7,
	0x75, 0xfe, 0x33, 0x3a, 0x7b, 0xab, 0xc6, 0xeb, 0x5b, 0x9d, 0x5f, 0x28, 0x24, 0xba, 0x6c, 0x6a,
	0xa7, 0x4a, 0xe4, 0x2d, 0xc2, 0xbd, 0x33, 0x2b, 0xc9, 0x0a, 0x4a, 0xc9, 0x9d, 0x97, 0x18, 0xcb,
	0x0a, 0x54, 0x5c, 0xe9, 0xaa, 0x59, 0xab, 0x1b, 0x3e, 0x3a, 0x69, 0xbc, 0xce, 0xd7, 0xc6, 0xa3,
	0x99, 0x50, 0xc7, 0x75, 0xe2, 0xa7, 0x50, 0xd0, 0xf6, 0x64, 0xb7, 0x73, 0x96, 0xc8, 0x75, 0x62,
	0xbe, 0x05, 0x53, 0xc7, 0x7e, 0x28, 0xb2, 0x43, 0x9e, 0xae, 0x1a, 0xaf, 0x67, 0xe7, 0x39, 0x95,
	0x24, 0x51, 0x57, 0xae, 0x9b, 0x06, 0x9f, 0x10, 0xbe, 0xf0, 0x44, 0x1b, 0xc1, 0xf9, 0x88, 0xf0,
	0xa5, 0x9f, 0xc3, 0x3c, 0x0b, 0x9c, 0xa1, 0xff, 0x7b, 0x4f, 0xf8, 0x9b, 0x47, 0x18, 0xdc, 0xfc,
	0x07, 0xa6, 0xdd, 0x8d, 0x1c, 0xbc, 0xfb, 0xfc, 0xfd, 0xc3, 0xd6, 0x1d, 0xc7, 0xa7, 0x7f, 0x30,
	0xa5, 0x4e, 0x25, 0x7d, 0xdd, 0xde, 0xee, 0x0d, 0x35, 0x13, 0xcb, 0x30, 0x3e, 0x59, 0xb8, 0x68,
	0xbe, 0x70, 0xd1, 0xb7, 0x85, 0x8b, 0xde, 0x2f, 0xdd, 0xce, 0x7c, 0xe9, 0x76, 0xbe, 0x2c, 0xdd,
	0xce, 0xf3, 0xa3, 0xbf, 0xfd, 0x91, 0x69, 0x70, 0x40, 0x5f, 0x9d, 0x6b, 0x93, 0xe6, 0x82, 0x97,
	0xca, 0xfa, 0xdf, 0x78, 0x6b, 0x1a, 0x24, 0xdb, 0x26, 0xb8, 0xfb, 0x23, 0x00, 0x00, 0xff, 0xff,
	0x55, 0xcb, 0xa6, 0xac, 0x2d, 0x03, 0x00, 0x00,
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
	// SpotPriceV2 defines a gRPC query handler that returns the spot price given
	// a base denomination and a quote denomination.
	// The returned spot price has 36 decimal places. However, some of
	// modules perform sig fig rounding so most of the rightmost decimals can be
	// zeroes.
	SpotPriceV2(ctx context.Context, in *SpotPriceRequest, opts ...grpc.CallOption) (*SpotPriceResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) SpotPriceV2(ctx context.Context, in *SpotPriceRequest, opts ...grpc.CallOption) (*SpotPriceResponse, error) {
	out := new(SpotPriceResponse)
	err := c.cc.Invoke(ctx, "/osmosis.poolmanager.v2.Query/SpotPriceV2", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// SpotPriceV2 defines a gRPC query handler that returns the spot price given
	// a base denomination and a quote denomination.
	// The returned spot price has 36 decimal places. However, some of
	// modules perform sig fig rounding so most of the rightmost decimals can be
	// zeroes.
	SpotPriceV2(context.Context, *SpotPriceRequest) (*SpotPriceResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) SpotPriceV2(ctx context.Context, req *SpotPriceRequest) (*SpotPriceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SpotPriceV2 not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_SpotPriceV2_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SpotPriceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).SpotPriceV2(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.poolmanager.v2.Query/SpotPriceV2",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).SpotPriceV2(ctx, req.(*SpotPriceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "osmosis.poolmanager.v2.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SpotPriceV2",
			Handler:    _Query_SpotPriceV2_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "osmosis/poolmanager/v2/query.proto",
}

func (m *SpotPriceRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SpotPriceRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SpotPriceRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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

func (m *SpotPriceResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SpotPriceResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SpotPriceResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.SpotPrice.Size()
		i -= size
		if _, err := m.SpotPrice.MarshalTo(dAtA[i:]); err != nil {
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
func (m *SpotPriceRequest) Size() (n int) {
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

func (m *SpotPriceResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.SpotPrice.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *SpotPriceRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: SpotPriceRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SpotPriceRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *SpotPriceResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: SpotPriceResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SpotPriceResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
			if err := m.SpotPrice.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
