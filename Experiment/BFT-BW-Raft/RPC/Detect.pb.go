// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.14.0
// source: Detect.proto

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

type DetectArgs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value int32 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *DetectArgs) Reset() {
	*x = DetectArgs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Detect_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DetectArgs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DetectArgs) ProtoMessage() {}

func (x *DetectArgs) ProtoReflect() protoreflect.Message {
	mi := &file_Detect_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DetectArgs.ProtoReflect.Descriptor instead.
func (*DetectArgs) Descriptor() ([]byte, []int) {
	return file_Detect_proto_rawDescGZIP(), []int{0}
}

func (x *DetectArgs) GetValue() int32 {
	if x != nil {
		return x.Value
	}
	return 0
}

type DetectReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Value   int32  `protobuf:"varint,2,opt,name=value,proto3" json:"value,omitempty"`
	Success bool   `protobuf:"varint,3,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *DetectReply) Reset() {
	*x = DetectReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Detect_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DetectReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DetectReply) ProtoMessage() {}

func (x *DetectReply) ProtoReflect() protoreflect.Message {
	mi := &file_Detect_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DetectReply.ProtoReflect.Descriptor instead.
func (*DetectReply) Descriptor() ([]byte, []int) {
	return file_Detect_proto_rawDescGZIP(), []int{1}
}

func (x *DetectReply) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *DetectReply) GetValue() int32 {
	if x != nil {
		return x.Value
	}
	return 0
}

func (x *DetectReply) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type BroadcastByzArgs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SendByzantine []byte `protobuf:"bytes,1,opt,name=sendByzantine,proto3" json:"sendByzantine,omitempty"`
	SendSuspicion []byte `protobuf:"bytes,2,opt,name=sendSuspicion,proto3" json:"sendSuspicion,omitempty"`
}

func (x *BroadcastByzArgs) Reset() {
	*x = BroadcastByzArgs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Detect_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BroadcastByzArgs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BroadcastByzArgs) ProtoMessage() {}

func (x *BroadcastByzArgs) ProtoReflect() protoreflect.Message {
	mi := &file_Detect_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BroadcastByzArgs.ProtoReflect.Descriptor instead.
func (*BroadcastByzArgs) Descriptor() ([]byte, []int) {
	return file_Detect_proto_rawDescGZIP(), []int{2}
}

func (x *BroadcastByzArgs) GetSendByzantine() []byte {
	if x != nil {
		return x.SendByzantine
	}
	return nil
}

func (x *BroadcastByzArgs) GetSendSuspicion() []byte {
	if x != nil {
		return x.SendSuspicion
	}
	return nil
}

type BroadcastByzRely struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReceiveByzantine []byte `protobuf:"bytes,1,opt,name=receiveByzantine,proto3" json:"receiveByzantine,omitempty"`
	ReceiveSuspicion []byte `protobuf:"bytes,2,opt,name=receiveSuspicion,proto3" json:"receiveSuspicion,omitempty"`
}

func (x *BroadcastByzRely) Reset() {
	*x = BroadcastByzRely{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Detect_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BroadcastByzRely) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BroadcastByzRely) ProtoMessage() {}

func (x *BroadcastByzRely) ProtoReflect() protoreflect.Message {
	mi := &file_Detect_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BroadcastByzRely.ProtoReflect.Descriptor instead.
func (*BroadcastByzRely) Descriptor() ([]byte, []int) {
	return file_Detect_proto_rawDescGZIP(), []int{3}
}

func (x *BroadcastByzRely) GetReceiveByzantine() []byte {
	if x != nil {
		return x.ReceiveByzantine
	}
	return nil
}

func (x *BroadcastByzRely) GetReceiveSuspicion() []byte {
	if x != nil {
		return x.ReceiveSuspicion
	}
	return nil
}

var File_Detect_proto protoreflect.FileDescriptor

var file_Detect_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03,
	0x52, 0x50, 0x43, 0x22, 0x22, 0x0a, 0x0a, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x41, 0x72, 0x67,
	0x73, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x57, 0x0a, 0x0b, 0x44, 0x65, 0x74, 0x65, 0x63,
	0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x22, 0x5e, 0x0a, 0x10, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x42, 0x79, 0x7a,
	0x41, 0x72, 0x67, 0x73, 0x12, 0x24, 0x0a, 0x0d, 0x73, 0x65, 0x6e, 0x64, 0x42, 0x79, 0x7a, 0x61,
	0x6e, 0x74, 0x69, 0x6e, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x73, 0x65, 0x6e,
	0x64, 0x42, 0x79, 0x7a, 0x61, 0x6e, 0x74, 0x69, 0x6e, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x73, 0x65,
	0x6e, 0x64, 0x53, 0x75, 0x73, 0x70, 0x69, 0x63, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x0d, 0x73, 0x65, 0x6e, 0x64, 0x53, 0x75, 0x73, 0x70, 0x69, 0x63, 0x69, 0x6f, 0x6e,
	0x22, 0x6a, 0x0a, 0x10, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x42, 0x79, 0x7a,
	0x52, 0x65, 0x6c, 0x79, 0x12, 0x2a, 0x0a, 0x10, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x42,
	0x79, 0x7a, 0x61, 0x6e, 0x74, 0x69, 0x6e, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x10,
	0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x42, 0x79, 0x7a, 0x61, 0x6e, 0x74, 0x69, 0x6e, 0x65,
	0x12, 0x2a, 0x0a, 0x10, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x53, 0x75, 0x73, 0x70, 0x69,
	0x63, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x10, 0x72, 0x65, 0x63, 0x65,
	0x69, 0x76, 0x65, 0x53, 0x75, 0x73, 0x70, 0x69, 0x63, 0x69, 0x6f, 0x6e, 0x32, 0x7f, 0x0a, 0x06,
	0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x12, 0x34, 0x0a, 0x0f, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74,
	0x42, 0x79, 0x7a, 0x61, 0x6e, 0x74, 0x69, 0x6e, 0x65, 0x12, 0x0f, 0x2e, 0x52, 0x50, 0x43, 0x2e,
	0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x41, 0x72, 0x67, 0x73, 0x1a, 0x10, 0x2e, 0x52, 0x50, 0x43,
	0x2e, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x3f, 0x0a, 0x0f,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x42, 0x79, 0x7a, 0x61, 0x6e, 0x74, 0x69, 0x6e, 0x65, 0x12,
	0x15, 0x2e, 0x52, 0x50, 0x43, 0x2e, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x42,
	0x79, 0x7a, 0x41, 0x72, 0x67, 0x73, 0x1a, 0x15, 0x2e, 0x52, 0x50, 0x43, 0x2e, 0x42, 0x72, 0x6f,
	0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x42, 0x79, 0x7a, 0x52, 0x65, 0x6c, 0x79, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_Detect_proto_rawDescOnce sync.Once
	file_Detect_proto_rawDescData = file_Detect_proto_rawDesc
)

func file_Detect_proto_rawDescGZIP() []byte {
	file_Detect_proto_rawDescOnce.Do(func() {
		file_Detect_proto_rawDescData = protoimpl.X.CompressGZIP(file_Detect_proto_rawDescData)
	})
	return file_Detect_proto_rawDescData
}

var file_Detect_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_Detect_proto_goTypes = []interface{}{
	(*DetectArgs)(nil),       // 0: RPC.DetectArgs
	(*DetectReply)(nil),      // 1: RPC.DetectReply
	(*BroadcastByzArgs)(nil), // 2: RPC.BroadcastByzArgs
	(*BroadcastByzRely)(nil), // 3: RPC.BroadcastByzRely
}
var file_Detect_proto_depIdxs = []int32{
	0, // 0: RPC.Detect.DetectByzantine:input_type -> RPC.DetectArgs
	2, // 1: RPC.Detect.UpdateByzantine:input_type -> RPC.BroadcastByzArgs
	1, // 2: RPC.Detect.DetectByzantine:output_type -> RPC.DetectReply
	3, // 3: RPC.Detect.UpdateByzantine:output_type -> RPC.BroadcastByzRely
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_Detect_proto_init() }
func file_Detect_proto_init() {
	if File_Detect_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_Detect_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DetectArgs); i {
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
		file_Detect_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DetectReply); i {
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
		file_Detect_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BroadcastByzArgs); i {
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
		file_Detect_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BroadcastByzRely); i {
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
			RawDescriptor: file_Detect_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_Detect_proto_goTypes,
		DependencyIndexes: file_Detect_proto_depIdxs,
		MessageInfos:      file_Detect_proto_msgTypes,
	}.Build()
	File_Detect_proto = out.File
	file_Detect_proto_rawDesc = nil
	file_Detect_proto_goTypes = nil
	file_Detect_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// DetectClient is the client API for Detect service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type DetectClient interface {
	DetectByzantine(ctx context.Context, in *DetectArgs, opts ...grpc.CallOption) (*DetectReply, error)
	UpdateByzantine(ctx context.Context, in *BroadcastByzArgs, opts ...grpc.CallOption) (*BroadcastByzRely, error)
}

type detectClient struct {
	cc grpc.ClientConnInterface
}

func NewDetectClient(cc grpc.ClientConnInterface) DetectClient {
	return &detectClient{cc}
}

func (c *detectClient) DetectByzantine(ctx context.Context, in *DetectArgs, opts ...grpc.CallOption) (*DetectReply, error) {
	out := new(DetectReply)
	err := c.cc.Invoke(ctx, "/RPC.Detect/DetectByzantine", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *detectClient) UpdateByzantine(ctx context.Context, in *BroadcastByzArgs, opts ...grpc.CallOption) (*BroadcastByzRely, error) {
	out := new(BroadcastByzRely)
	err := c.cc.Invoke(ctx, "/RPC.Detect/UpdateByzantine", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DetectServer is the server API for Detect service.
type DetectServer interface {
	DetectByzantine(context.Context, *DetectArgs) (*DetectReply, error)
	UpdateByzantine(context.Context, *BroadcastByzArgs) (*BroadcastByzRely, error)
}

// UnimplementedDetectServer can be embedded to have forward compatible implementations.
type UnimplementedDetectServer struct {
}

func (*UnimplementedDetectServer) DetectByzantine(context.Context, *DetectArgs) (*DetectReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DetectByzantine not implemented")
}
func (*UnimplementedDetectServer) UpdateByzantine(context.Context, *BroadcastByzArgs) (*BroadcastByzRely, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateByzantine not implemented")
}

func RegisterDetectServer(s *grpc.Server, srv DetectServer) {
	s.RegisterService(&_Detect_serviceDesc, srv)
}

func _Detect_DetectByzantine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DetectArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DetectServer).DetectByzantine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/RPC.Detect/DetectByzantine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DetectServer).DetectByzantine(ctx, req.(*DetectArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _Detect_UpdateByzantine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BroadcastByzArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DetectServer).UpdateByzantine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/RPC.Detect/UpdateByzantine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DetectServer).UpdateByzantine(ctx, req.(*BroadcastByzArgs))
	}
	return interceptor(ctx, in, info, handler)
}

var _Detect_serviceDesc = grpc.ServiceDesc{
	ServiceName: "RPC.Detect",
	HandlerType: (*DetectServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DetectByzantine",
			Handler:    _Detect_DetectByzantine_Handler,
		},
		{
			MethodName: "UpdateByzantine",
			Handler:    _Detect_UpdateByzantine_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "Detect.proto",
}
