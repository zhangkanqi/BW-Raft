// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.14.0
// source: GetValue.proto

package RPC

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
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

type GetValueArgs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *GetValueArgs) Reset() {
	*x = GetValueArgs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_GetValue_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetValueArgs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetValueArgs) ProtoMessage() {}

func (x *GetValueArgs) ProtoReflect() protoreflect.Message {
	mi := &file_GetValue_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetValueArgs.ProtoReflect.Descriptor instead.
func (*GetValueArgs) Descriptor() ([]byte, []int) {
	return file_GetValue_proto_rawDescGZIP(), []int{0}
}

func (x *GetValueArgs) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

type GetValueReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Value   string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *GetValueReply) Reset() {
	*x = GetValueReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_GetValue_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetValueReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetValueReply) ProtoMessage() {}

func (x *GetValueReply) ProtoReflect() protoreflect.Message {
	mi := &file_GetValue_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetValueReply.ProtoReflect.Descriptor instead.
func (*GetValueReply) Descriptor() ([]byte, []int) {
	return file_GetValue_proto_rawDescGZIP(), []int{1}
}

func (x *GetValueReply) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *GetValueReply) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_GetValue_proto protoreflect.FileDescriptor

var file_GetValue_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x47, 0x65, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x20, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x41, 0x72, 0x67, 0x73,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x22, 0x3f, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x14, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x32, 0x35, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12,
	0x29, 0x0a, 0x08, 0x67, 0x65, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x0d, 0x2e, 0x47, 0x65,
	0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x41, 0x72, 0x67, 0x73, 0x1a, 0x0e, 0x2e, 0x47, 0x65, 0x74,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_GetValue_proto_rawDescOnce sync.Once
	file_GetValue_proto_rawDescData = file_GetValue_proto_rawDesc
)

func file_GetValue_proto_rawDescGZIP() []byte {
	file_GetValue_proto_rawDescOnce.Do(func() {
		file_GetValue_proto_rawDescData = protoimpl.X.CompressGZIP(file_GetValue_proto_rawDescData)
	})
	return file_GetValue_proto_rawDescData
}

var file_GetValue_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_GetValue_proto_goTypes = []interface{}{
	(*GetValueArgs)(nil),  // 0: GetValueArgs
	(*GetValueReply)(nil), // 1: GetValueReply
}
var file_GetValue_proto_depIdxs = []int32{
	0, // 0: GetValue.getValue:input_type -> GetValueArgs
	1, // 1: GetValue.getValue:output_type -> GetValueReply
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_GetValue_proto_init() }
func file_GetValue_proto_init() {
	if File_GetValue_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_GetValue_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetValueArgs); i {
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
		file_GetValue_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetValueReply); i {
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
			RawDescriptor: file_GetValue_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_GetValue_proto_goTypes,
		DependencyIndexes: file_GetValue_proto_depIdxs,
		MessageInfos:      file_GetValue_proto_msgTypes,
	}.Build()
	File_GetValue_proto = out.File
	file_GetValue_proto_rawDesc = nil
	file_GetValue_proto_goTypes = nil
	file_GetValue_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// GetValueClient is the client API for GetValue service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GetValueClient interface {
	GetValue(ctx context.Context, in *GetValueArgs, opts ...grpc.CallOption) (*GetValueReply, error)
}

type getValueClient struct {
	cc grpc.ClientConnInterface
}

func NewGetValueClient(cc grpc.ClientConnInterface) GetValueClient {
	return &getValueClient{cc}
}

func (c *getValueClient) GetValue(ctx context.Context, in *GetValueArgs, opts ...grpc.CallOption) (*GetValueReply, error) {
	out := new(GetValueReply)
	err := c.cc.Invoke(ctx, "/GetValue/getValue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetValueServer is the server API for GetValue service.
type GetValueServer interface {
	GetValue(context.Context, *GetValueArgs) (*GetValueReply, error)
}

// UnimplementedGetValueServer can be embedded to have forward compatible implementations.
type UnimplementedGetValueServer struct {
}

func (*UnimplementedGetValueServer) GetValue(context.Context, *GetValueArgs) (*GetValueReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetValue not implemented")
}

func RegisterGetValueServer(s *grpc.Server, srv GetValueServer) {
	s.RegisterService(&_GetValue_serviceDesc, srv)
}

func _GetValue_GetValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetValueArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GetValueServer).GetValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/GetValue/GetValue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GetValueServer).GetValue(ctx, req.(*GetValueArgs))
	}
	return interceptor(ctx, in, info, handler)
}

var _GetValue_serviceDesc = grpc.ServiceDesc{
	ServiceName: "GetValue",
	HandlerType: (*GetValueServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "getValue",
			Handler:    _GetValue_GetValue_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "GetValue.proto",
}
