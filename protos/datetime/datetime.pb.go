// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.11.4
// source: protos/datetime/datetime.proto

package datetime

import (
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Date struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Day   int32 `protobuf:"varint,1,opt,name=day,proto3" json:"day,omitempty"`
	Month int32 `protobuf:"varint,2,opt,name=month,proto3" json:"month,omitempty"`
	Year  int32 `protobuf:"varint,3,opt,name=year,proto3" json:"year,omitempty"`
}

func (x *Date) Reset() {
	*x = Date{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_datetime_datetime_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Date) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Date) ProtoMessage() {}

func (x *Date) ProtoReflect() protoreflect.Message {
	mi := &file_protos_datetime_datetime_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Date.ProtoReflect.Descriptor instead.
func (*Date) Descriptor() ([]byte, []int) {
	return file_protos_datetime_datetime_proto_rawDescGZIP(), []int{0}
}

func (x *Date) GetDay() int32 {
	if x != nil {
		return x.Day
	}
	return 0
}

func (x *Date) GetMonth() int32 {
	if x != nil {
		return x.Month
	}
	return 0
}

func (x *Date) GetYear() int32 {
	if x != nil {
		return x.Year
	}
	return 0
}

type DateRange struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start *Date `protobuf:"bytes,1,opt,name=start,proto3" json:"start,omitempty"`
	End   *Date `protobuf:"bytes,2,opt,name=end,proto3" json:"end,omitempty"`
}

func (x *DateRange) Reset() {
	*x = DateRange{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_datetime_datetime_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DateRange) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DateRange) ProtoMessage() {}

func (x *DateRange) ProtoReflect() protoreflect.Message {
	mi := &file_protos_datetime_datetime_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DateRange.ProtoReflect.Descriptor instead.
func (*DateRange) Descriptor() ([]byte, []int) {
	return file_protos_datetime_datetime_proto_rawDescGZIP(), []int{1}
}

func (x *DateRange) GetStart() *Date {
	if x != nil {
		return x.Start
	}
	return nil
}

func (x *DateRange) GetEnd() *Date {
	if x != nil {
		return x.End
	}
	return nil
}

type DateRanges struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ranges []*DateRange `protobuf:"bytes,1,rep,name=ranges,proto3" json:"ranges,omitempty"`
}

func (x *DateRanges) Reset() {
	*x = DateRanges{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_datetime_datetime_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DateRanges) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DateRanges) ProtoMessage() {}

func (x *DateRanges) ProtoReflect() protoreflect.Message {
	mi := &file_protos_datetime_datetime_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DateRanges.ProtoReflect.Descriptor instead.
func (*DateRanges) Descriptor() ([]byte, []int) {
	return file_protos_datetime_datetime_proto_rawDescGZIP(), []int{2}
}

func (x *DateRanges) GetRanges() []*DateRange {
	if x != nil {
		return x.Ranges
	}
	return nil
}

var File_protos_datetime_datetime_proto protoreflect.FileDescriptor

var file_protos_datetime_datetime_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x64, 0x61, 0x74, 0x65, 0x74, 0x69, 0x6d,
	0x65, 0x2f, 0x64, 0x61, 0x74, 0x65, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x0d, 0x6c, 0x69, 0x62, 0x73, 0x2e, 0x64, 0x61, 0x74, 0x65, 0x74, 0x69, 0x6d, 0x65, 0x22,
	0x42, 0x0a, 0x04, 0x44, 0x61, 0x74, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x64, 0x61, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x64, 0x61, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x6f, 0x6e,
	0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6d, 0x6f, 0x6e, 0x74, 0x68, 0x12,
	0x12, 0x0a, 0x04, 0x79, 0x65, 0x61, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x79,
	0x65, 0x61, 0x72, 0x22, 0x5d, 0x0a, 0x09, 0x44, 0x61, 0x74, 0x65, 0x52, 0x61, 0x6e, 0x67, 0x65,
	0x12, 0x29, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x6c, 0x69, 0x62, 0x73, 0x2e, 0x64, 0x61, 0x74, 0x65, 0x74, 0x69, 0x6d, 0x65, 0x2e,
	0x44, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x25, 0x0a, 0x03, 0x65,
	0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x6c, 0x69, 0x62, 0x73, 0x2e,
	0x64, 0x61, 0x74, 0x65, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x44, 0x61, 0x74, 0x65, 0x52, 0x03, 0x65,
	0x6e, 0x64, 0x22, 0x3e, 0x0a, 0x0a, 0x44, 0x61, 0x74, 0x65, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x73,
	0x12, 0x30, 0x0a, 0x06, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x18, 0x2e, 0x6c, 0x69, 0x62, 0x73, 0x2e, 0x64, 0x61, 0x74, 0x65, 0x74, 0x69, 0x6d, 0x65,
	0x2e, 0x44, 0x61, 0x74, 0x65, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x06, 0x72, 0x61, 0x6e, 0x67,
	0x65, 0x73, 0x42, 0x29, 0x5a, 0x27, 0x6c, 0x69, 0x62, 0x73, 0x2e, 0x61, 0x6c, 0x74, 0x69, 0x70,
	0x6c, 0x61, 0x2e, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6c, 0x74, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x64, 0x61, 0x74, 0x65, 0x74, 0x69, 0x6d, 0x65, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protos_datetime_datetime_proto_rawDescOnce sync.Once
	file_protos_datetime_datetime_proto_rawDescData = file_protos_datetime_datetime_proto_rawDesc
)

func file_protos_datetime_datetime_proto_rawDescGZIP() []byte {
	file_protos_datetime_datetime_proto_rawDescOnce.Do(func() {
		file_protos_datetime_datetime_proto_rawDescData = protoimpl.X.CompressGZIP(file_protos_datetime_datetime_proto_rawDescData)
	})
	return file_protos_datetime_datetime_proto_rawDescData
}

var file_protos_datetime_datetime_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_protos_datetime_datetime_proto_goTypes = []interface{}{
	(*Date)(nil),       // 0: libs.datetime.Date
	(*DateRange)(nil),  // 1: libs.datetime.DateRange
	(*DateRanges)(nil), // 2: libs.datetime.DateRanges
}
var file_protos_datetime_datetime_proto_depIdxs = []int32{
	0, // 0: libs.datetime.DateRange.start:type_name -> libs.datetime.Date
	0, // 1: libs.datetime.DateRange.end:type_name -> libs.datetime.Date
	1, // 2: libs.datetime.DateRanges.ranges:type_name -> libs.datetime.DateRange
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_protos_datetime_datetime_proto_init() }
func file_protos_datetime_datetime_proto_init() {
	if File_protos_datetime_datetime_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protos_datetime_datetime_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Date); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protos_datetime_datetime_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DateRange); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protos_datetime_datetime_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DateRanges); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protos_datetime_datetime_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_datetime_datetime_proto_goTypes,
		DependencyIndexes: file_protos_datetime_datetime_proto_depIdxs,
		MessageInfos:      file_protos_datetime_datetime_proto_msgTypes,
	}.Build()
	File_protos_datetime_datetime_proto = out.File
	file_protos_datetime_datetime_proto_rawDesc = nil
	file_protos_datetime_datetime_proto_goTypes = nil
	file_protos_datetime_datetime_proto_depIdxs = nil
}
