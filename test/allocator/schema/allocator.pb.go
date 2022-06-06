// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: allocator.proto

package schema

import (
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
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

// MakeAllocationRequest - запрос на аллокацию.
type MakeAllocationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// size - размер аллокации
	Size uint64 `protobuf:"varint,1,opt,name=size,proto3" json:"size,omitempty"`
	// duration - продолжительность времени, на которое надо заблокировать запрос после аллокации
	Duration *durationpb.Duration `protobuf:"bytes,2,opt,name=duration,proto3" json:"duration,omitempty"`
}

func (x *MakeAllocationRequest) Reset() {
	*x = MakeAllocationRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_allocator_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MakeAllocationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MakeAllocationRequest) ProtoMessage() {}

func (x *MakeAllocationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_allocator_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MakeAllocationRequest.ProtoReflect.Descriptor instead.
func (*MakeAllocationRequest) Descriptor() ([]byte, []int) {
	return file_allocator_proto_rawDescGZIP(), []int{0}
}

func (x *MakeAllocationRequest) GetSize() uint64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *MakeAllocationRequest) GetDuration() *durationpb.Duration {
	if x != nil {
		return x.Duration
	}
	return nil
}

// MakeAllocationResponse - ответ на запрос на аллокацию.
type MakeAllocationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// value - просто некоторое значение
	Value uint64 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *MakeAllocationResponse) Reset() {
	*x = MakeAllocationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_allocator_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MakeAllocationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MakeAllocationResponse) ProtoMessage() {}

func (x *MakeAllocationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_allocator_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MakeAllocationResponse.ProtoReflect.Descriptor instead.
func (*MakeAllocationResponse) Descriptor() ([]byte, []int) {
	return file_allocator_proto_rawDescGZIP(), []int{1}
}

func (x *MakeAllocationResponse) GetValue() uint64 {
	if x != nil {
		return x.Value
	}
	return 0
}

var File_allocator_proto protoreflect.FileDescriptor

var file_allocator_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x62, 0x0a, 0x15, 0x4d, 0x61, 0x6b,
	0x65, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x12, 0x35, 0x0a, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x2e, 0x0a,
	0x16, 0x4d, 0x61, 0x6b, 0x65, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x32, 0x5e, 0x0a,
	0x09, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x51, 0x0a, 0x0e, 0x4d, 0x61,
	0x6b, 0x65, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1d, 0x2e, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x4d, 0x61, 0x6b, 0x65, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x73, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x2e, 0x4d, 0x61, 0x6b, 0x65, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x43, 0x5a,
	0x41, 0x67, 0x69, 0x74, 0x6c, 0x61, 0x62, 0x2e, 0x73, 0x74, 0x61, 0x67, 0x65, 0x6f, 0x66, 0x66,
	0x69, 0x63, 0x65, 0x2e, 0x72, 0x75, 0x2f, 0x55, 0x43, 0x53, 0x2d, 0x43, 0x4f, 0x4d, 0x4d, 0x4f,
	0x4e, 0x2f, 0x6d, 0x65, 0x6d, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x65, 0x72, 0x2f, 0x74, 0x65, 0x73,
	0x74, 0x2f, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x73, 0x63, 0x68, 0x65,
	0x6d, 0x61, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_allocator_proto_rawDescOnce sync.Once
	file_allocator_proto_rawDescData = file_allocator_proto_rawDesc
)

func file_allocator_proto_rawDescGZIP() []byte {
	file_allocator_proto_rawDescOnce.Do(func() {
		file_allocator_proto_rawDescData = protoimpl.X.CompressGZIP(file_allocator_proto_rawDescData)
	})
	return file_allocator_proto_rawDescData
}

var file_allocator_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_allocator_proto_goTypes = []interface{}{
	(*MakeAllocationRequest)(nil),  // 0: schema.MakeAllocationRequest
	(*MakeAllocationResponse)(nil), // 1: schema.MakeAllocationResponse
	(*durationpb.Duration)(nil),    // 2: google.protobuf.Duration
}
var file_allocator_proto_depIdxs = []int32{
	2, // 0: schema.MakeAllocationRequest.duration:type_name -> google.protobuf.Duration
	0, // 1: schema.Allocator.MakeAllocation:input_type -> schema.MakeAllocationRequest
	1, // 2: schema.Allocator.MakeAllocation:output_type -> schema.MakeAllocationResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_allocator_proto_init() }
func file_allocator_proto_init() {
	if File_allocator_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_allocator_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MakeAllocationRequest); i {
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
		file_allocator_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MakeAllocationResponse); i {
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
			RawDescriptor: file_allocator_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_allocator_proto_goTypes,
		DependencyIndexes: file_allocator_proto_depIdxs,
		MessageInfos:      file_allocator_proto_msgTypes,
	}.Build()
	File_allocator_proto = out.File
	file_allocator_proto_rawDesc = nil
	file_allocator_proto_goTypes = nil
	file_allocator_proto_depIdxs = nil
}