// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/downtime-detector/v1beta1/downtime_duration.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
	math "math"
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

type Downtime int32

const (
	Downtime_DURATION_30S  Downtime = 0
	Downtime_DURATION_1M   Downtime = 1
	Downtime_DURATION_2M   Downtime = 2
	Downtime_DURATION_3M   Downtime = 3
	Downtime_DURATION_4M   Downtime = 4
	Downtime_DURATION_5M   Downtime = 5
	Downtime_DURATION_10M  Downtime = 6
	Downtime_DURATION_20M  Downtime = 7
	Downtime_DURATION_30M  Downtime = 8
	Downtime_DURATION_40M  Downtime = 9
	Downtime_DURATION_50M  Downtime = 10
	Downtime_DURATION_1H   Downtime = 11
	Downtime_DURATION_1_5H Downtime = 12
	Downtime_DURATION_2H   Downtime = 13
	Downtime_DURATION_2_5H Downtime = 14
	Downtime_DURATION_3H   Downtime = 15
	Downtime_DURATION_4H   Downtime = 16
	Downtime_DURATION_5H   Downtime = 17
	Downtime_DURATION_6H   Downtime = 18
	Downtime_DURATION_9H   Downtime = 19
	Downtime_DURATION_12H  Downtime = 20
	Downtime_DURATION_18H  Downtime = 21
	Downtime_DURATION_24H  Downtime = 22
	Downtime_DURATION_36H  Downtime = 23
	Downtime_DURATION_48H  Downtime = 24
)

var Downtime_name = map[int32]string{
	0:  "DURATION_30S",
	1:  "DURATION_1M",
	2:  "DURATION_2M",
	3:  "DURATION_3M",
	4:  "DURATION_4M",
	5:  "DURATION_5M",
	6:  "DURATION_10M",
	7:  "DURATION_20M",
	8:  "DURATION_30M",
	9:  "DURATION_40M",
	10: "DURATION_50M",
	11: "DURATION_1H",
	12: "DURATION_1_5H",
	13: "DURATION_2H",
	14: "DURATION_2_5H",
	15: "DURATION_3H",
	16: "DURATION_4H",
	17: "DURATION_5H",
	18: "DURATION_6H",
	19: "DURATION_9H",
	20: "DURATION_12H",
	21: "DURATION_18H",
	22: "DURATION_24H",
	23: "DURATION_36H",
	24: "DURATION_48H",
}

var Downtime_value = map[string]int32{
	"DURATION_30S":  0,
	"DURATION_1M":   1,
	"DURATION_2M":   2,
	"DURATION_3M":   3,
	"DURATION_4M":   4,
	"DURATION_5M":   5,
	"DURATION_10M":  6,
	"DURATION_20M":  7,
	"DURATION_30M":  8,
	"DURATION_40M":  9,
	"DURATION_50M":  10,
	"DURATION_1H":   11,
	"DURATION_1_5H": 12,
	"DURATION_2H":   13,
	"DURATION_2_5H": 14,
	"DURATION_3H":   15,
	"DURATION_4H":   16,
	"DURATION_5H":   17,
	"DURATION_6H":   18,
	"DURATION_9H":   19,
	"DURATION_12H":  20,
	"DURATION_18H":  21,
	"DURATION_24H":  22,
	"DURATION_36H":  23,
	"DURATION_48H":  24,
}

func (x Downtime) String() string {
	return proto.EnumName(Downtime_name, int32(x))
}

func (Downtime) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_21a1969f22fb2a7e, []int{0}
}

func init() {
	proto.RegisterEnum("osmosis.downtimedetector.v1beta1.Downtime", Downtime_name, Downtime_value)
}

func init() {
	proto.RegisterFile("osmosis/downtime-detector/v1beta1/downtime_duration.proto", fileDescriptor_21a1969f22fb2a7e)
}

var fileDescriptor_21a1969f22fb2a7e = []byte{
	// 386 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0xbb, 0x6e, 0xe2, 0x50,
	0x10, 0x86, 0xed, 0x65, 0x97, 0x65, 0x0d, 0x2c, 0x83, 0x97, 0xbd, 0x51, 0x78, 0xb7, 0x8e, 0x84,
	0x8f, 0x2f, 0x80, 0xa0, 0x48, 0x91, 0x88, 0xe2, 0xa4, 0x38, 0x89, 0x94, 0x8b, 0x22, 0xa5, 0xb1,
	0x6c, 0x70, 0x1c, 0x4b, 0x98, 0x83, 0xf0, 0x81, 0x84, 0xb7, 0xc8, 0x33, 0xa5, 0x4a, 0x49, 0x99,
	0x32, 0x82, 0x17, 0x89, 0xf0, 0x85, 0x68, 0x50, 0x3a, 0xcf, 0x37, 0xf3, 0xfb, 0xfc, 0xff, 0x68,
	0x94, 0x3e, 0x8f, 0x23, 0x1e, 0x87, 0x31, 0x19, 0xf1, 0xfb, 0x89, 0x08, 0x23, 0xbf, 0x35, 0xf2,
	0x85, 0x3f, 0x14, 0x7c, 0x46, 0x16, 0xa6, 0xe7, 0x0b, 0xd7, 0xdc, 0x75, 0x9c, 0xd1, 0x7c, 0xe6,
	0x8a, 0x90, 0x4f, 0xf4, 0xe9, 0x8c, 0x0b, 0xae, 0xfe, 0xcf, 0xa4, 0x7a, 0x3e, 0x90, 0x2b, 0xf5,
	0x4c, 0xd9, 0x6c, 0x04, 0x3c, 0xe0, 0xc9, 0x30, 0xd9, 0x7e, 0xa5, 0xba, 0xe6, 0xdf, 0x80, 0xf3,
	0x60, 0xec, 0x93, 0xa4, 0xf2, 0xe6, 0xb7, 0xc4, 0x9d, 0x2c, 0xf3, 0xd6, 0x30, 0xf9, 0xa7, 0x93,
	0x6a, 0xd2, 0x22, 0x6b, 0x69, 0xfb, 0x2a, 0xec, 0xa6, 0xf9, 0x6f, 0xbf, 0xbf, 0x75, 0x14, 0x0b,
	0x37, 0x9a, 0xa6, 0x03, 0x07, 0x4f, 0x05, 0xa5, 0x34, 0xc8, 0x9c, 0xaa, 0xa0, 0x54, 0x06, 0x57,
	0xe7, 0x47, 0x97, 0x27, 0x67, 0xa7, 0x8e, 0x6d, 0x5c, 0x80, 0xa4, 0xd6, 0x94, 0xf2, 0x8e, 0x98,
	0x0c, 0x64, 0x04, 0x2c, 0x06, 0x9f, 0x10, 0xb0, 0x19, 0x14, 0x10, 0x68, 0x33, 0xf8, 0x8c, 0x40,
	0x87, 0xc1, 0x17, 0xf4, 0x8c, 0x69, 0x30, 0x28, 0x22, 0x62, 0x19, 0x0c, 0xbe, 0xee, 0x59, 0x61,
	0x50, 0x42, 0xa4, 0x6d, 0x30, 0xf8, 0x86, 0x48, 0xc7, 0x60, 0xa0, 0x60, 0xbb, 0x14, 0xca, 0x6a,
	0x5d, 0xa9, 0xbe, 0x03, 0xa7, 0x43, 0xa1, 0x82, 0x13, 0x50, 0xa8, 0xa2, 0x19, 0x6b, 0x3b, 0xf3,
	0x1d, 0x87, 0xa2, 0x50, 0xc3, 0xa1, 0x28, 0x00, 0x0e, 0x45, 0xa1, 0x8e, 0x40, 0x97, 0x82, 0x8a,
	0x40, 0x9f, 0xc2, 0x0f, 0x1c, 0xdb, 0xa2, 0xd0, 0xc0, 0xa4, 0x47, 0xe1, 0x27, 0x5e, 0x44, 0x9b,
	0xc2, 0x2f, 0xbc, 0x88, 0x2e, 0x85, 0xdf, 0x78, 0x11, 0x3d, 0x0a, 0x7f, 0x8e, 0xaf, 0x9f, 0xd7,
	0x9a, 0xbc, 0x5a, 0x6b, 0xf2, 0xeb, 0x5a, 0x93, 0x1f, 0x37, 0x9a, 0xb4, 0xda, 0x68, 0xd2, 0xcb,
	0x46, 0x93, 0x6e, 0x0e, 0x83, 0x50, 0xdc, 0xcd, 0x3d, 0x7d, 0xc8, 0x23, 0x92, 0x1d, 0x66, 0x6b,
	0xec, 0x7a, 0x71, 0x5e, 0x90, 0x85, 0x69, 0x93, 0x87, 0x0f, 0xce, 0x5c, 0x2c, 0xa7, 0x7e, 0xec,
	0x15, 0x93, 0x23, 0xb1, 0xdf, 0x02, 0x00, 0x00, 0xff, 0xff, 0x48, 0x2b, 0x3d, 0xaf, 0x10, 0x03,
	0x00, 0x00,
}
