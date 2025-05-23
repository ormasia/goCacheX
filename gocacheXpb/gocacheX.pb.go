// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v4.24.2
// source: gocacheX.proto

package gocacheXpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Request struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Group         string                 `protobuf:"bytes,1,opt,name=group,proto3" json:"group,omitempty"`
	Key           string                 `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Request) Reset() {
	*x = Request{}
	mi := &file_gocacheX_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_gocacheX_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_gocacheX_proto_rawDescGZIP(), []int{0}
}

func (x *Request) GetGroup() string {
	if x != nil {
		return x.Group
	}
	return ""
}

func (x *Request) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

type Response struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Value         []byte                 `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Response) Reset() {
	*x = Response{}
	mi := &file_gocacheX_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_gocacheX_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_gocacheX_proto_rawDescGZIP(), []int{1}
}

func (x *Response) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

var File_gocacheX_proto protoreflect.FileDescriptor

const file_gocacheX_proto_rawDesc = "" +
	"\n" +
	"\x0egocacheX.proto\x12\n" +
	"gocacheXpb\"1\n" +
	"\aRequest\x12\x14\n" +
	"\x05group\x18\x01 \x01(\tR\x05group\x12\x10\n" +
	"\x03key\x18\x02 \x01(\tR\x03key\" \n" +
	"\bResponse\x12\x14\n" +
	"\x05value\x18\x01 \x01(\fR\x05value2>\n" +
	"\n" +
	"GroupCache\x120\n" +
	"\x03Get\x12\x13.gocacheXpb.Request\x1a\x14.gocacheXpb.ResponseB\x15Z\x13goCacheX/gocacheXpbb\x06proto3"

var (
	file_gocacheX_proto_rawDescOnce sync.Once
	file_gocacheX_proto_rawDescData []byte
)

func file_gocacheX_proto_rawDescGZIP() []byte {
	file_gocacheX_proto_rawDescOnce.Do(func() {
		file_gocacheX_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_gocacheX_proto_rawDesc), len(file_gocacheX_proto_rawDesc)))
	})
	return file_gocacheX_proto_rawDescData
}

var file_gocacheX_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_gocacheX_proto_goTypes = []any{
	(*Request)(nil),  // 0: gocacheXpb.Request
	(*Response)(nil), // 1: gocacheXpb.Response
}
var file_gocacheX_proto_depIdxs = []int32{
	0, // 0: gocacheXpb.GroupCache.Get:input_type -> gocacheXpb.Request
	1, // 1: gocacheXpb.GroupCache.Get:output_type -> gocacheXpb.Response
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_gocacheX_proto_init() }
func file_gocacheX_proto_init() {
	if File_gocacheX_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_gocacheX_proto_rawDesc), len(file_gocacheX_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_gocacheX_proto_goTypes,
		DependencyIndexes: file_gocacheX_proto_depIdxs,
		MessageInfos:      file_gocacheX_proto_msgTypes,
	}.Build()
	File_gocacheX_proto = out.File
	file_gocacheX_proto_goTypes = nil
	file_gocacheX_proto_depIdxs = nil
}
