// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pagination/internal/model/model.proto

package libs_pagination_internal_proto

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Status struct {
	ParamsChecksum       string   `protobuf:"bytes,1,opt,name=params_checksum,json=paramsChecksum,proto3" json:"params_checksum,omitempty"`
	Cursor               int64    `protobuf:"varint,3,opt,name=cursor,proto3" json:"cursor,omitempty"`
	End                  bool     `protobuf:"varint,4,opt,name=end,proto3" json:"end,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Status) Reset()         { *m = Status{} }
func (m *Status) String() string { return proto.CompactTextString(m) }
func (*Status) ProtoMessage()    {}
func (*Status) Descriptor() ([]byte, []int) {
	return fileDescriptor_211e01d7a5d40e9f, []int{0}
}

func (m *Status) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Status.Unmarshal(m, b)
}
func (m *Status) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Status.Marshal(b, m, deterministic)
}
func (m *Status) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Status.Merge(m, src)
}
func (m *Status) XXX_Size() int {
	return xxx_messageInfo_Status.Size(m)
}
func (m *Status) XXX_DiscardUnknown() {
	xxx_messageInfo_Status.DiscardUnknown(m)
}

var xxx_messageInfo_Status proto.InternalMessageInfo

func (m *Status) GetParamsChecksum() string {
	if m != nil {
		return m.ParamsChecksum
	}
	return ""
}

func (m *Status) GetCursor() int64 {
	if m != nil {
		return m.Cursor
	}
	return 0
}

func (m *Status) GetEnd() bool {
	if m != nil {
		return m.End
	}
	return false
}

func init() {
	proto.RegisterType((*Status)(nil), "libs.pagination.internal.proto.Status")
}

func init() {
	proto.RegisterFile("pagination/internal/model/model.proto", fileDescriptor_211e01d7a5d40e9f)
}

var fileDescriptor_211e01d7a5d40e9f = []byte{
	// 153 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x2d, 0x48, 0x4c, 0xcf,
	0xcc, 0x4b, 0x2c, 0xc9, 0xcc, 0xcf, 0xd3, 0xcf, 0xcc, 0x2b, 0x49, 0x2d, 0xca, 0x4b, 0xcc, 0xd1,
	0xcf, 0xcd, 0x4f, 0x49, 0x85, 0x92, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42, 0x72, 0x39, 0x99,
	0x49, 0xc5, 0x7a, 0x08, 0xb5, 0x7a, 0x30, 0xb5, 0x10, 0x79, 0xa5, 0x68, 0x2e, 0xb6, 0xe0, 0x92,
	0xc4, 0x92, 0xd2, 0x62, 0x21, 0x75, 0x2e, 0xfe, 0x82, 0xc4, 0xa2, 0xc4, 0xdc, 0xe2, 0xf8, 0xe4,
	0x8c, 0xd4, 0xe4, 0xec, 0xe2, 0xd2, 0x5c, 0x09, 0x46, 0x05, 0x46, 0x0d, 0xce, 0x20, 0x3e, 0x88,
	0xb0, 0x33, 0x54, 0x54, 0x48, 0x8c, 0x8b, 0x2d, 0xb9, 0xb4, 0xa8, 0x38, 0xbf, 0x48, 0x82, 0x59,
	0x81, 0x51, 0x83, 0x39, 0x08, 0xca, 0x13, 0x12, 0xe0, 0x62, 0x4e, 0xcd, 0x4b, 0x91, 0x60, 0x51,
	0x60, 0xd4, 0xe0, 0x08, 0x02, 0x31, 0x93, 0xd8, 0xc0, 0x76, 0x18, 0x03, 0x02, 0x00, 0x00, 0xff,
	0xff, 0x0f, 0x0f, 0x7a, 0xc3, 0xac, 0x00, 0x00, 0x00,
}
