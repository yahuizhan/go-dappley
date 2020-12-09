// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.13.0
// source: github.com/yahuizhan/go-dappley/tool/performance_testing/pb/config.proto

package performance_configpb

import (
	proto "github.com/golang/protobuf/proto"
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

type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GoCount          int32   `protobuf:"varint,1,opt,name=goCount,proto3" json:"goCount,omitempty"`
	Tps              float32 `protobuf:"fixed32,2,opt,name=tps,proto3" json:"tps,omitempty"`
	MinerPrivKey     string  `protobuf:"bytes,3,opt,name=minerPrivKey,proto3" json:"minerPrivKey,omitempty"`
	AmountFromMinner uint64  `protobuf:"varint,4,opt,name=amountFromMinner,proto3" json:"amountFromMinner,omitempty"`
	AmountPerTx      uint64  `protobuf:"varint,5,opt,name=amountPerTx,proto3" json:"amountPerTx,omitempty"`
	Ip               string  `protobuf:"bytes,6,opt,name=ip,proto3" json:"ip,omitempty"`
	Port             string  `protobuf:"bytes,7,opt,name=port,proto3" json:"port,omitempty"`
	LogOpen          bool    `protobuf:"varint,8,opt,name=logOpen,proto3" json:"logOpen,omitempty"`
	LogName          string  `protobuf:"bytes,9,opt,name=logName,proto3" json:"logName,omitempty"`
	LogLevel         string  `protobuf:"bytes,10,opt,name=logLevel,proto3" json:"logLevel,omitempty"`
	LogCount         int32   `protobuf:"varint,11,opt,name=logCount,proto3" json:"logCount,omitempty"`
	LogRotateTime    int32   `protobuf:"varint,12,opt,name=logRotateTime,proto3" json:"logRotateTime,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetGoCount() int32 {
	if x != nil {
		return x.GoCount
	}
	return 0
}

func (x *Config) GetTps() float32 {
	if x != nil {
		return x.Tps
	}
	return 0
}

func (x *Config) GetMinerPrivKey() string {
	if x != nil {
		return x.MinerPrivKey
	}
	return ""
}

func (x *Config) GetAmountFromMinner() uint64 {
	if x != nil {
		return x.AmountFromMinner
	}
	return 0
}

func (x *Config) GetAmountPerTx() uint64 {
	if x != nil {
		return x.AmountPerTx
	}
	return 0
}

func (x *Config) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *Config) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

func (x *Config) GetLogOpen() bool {
	if x != nil {
		return x.LogOpen
	}
	return false
}

func (x *Config) GetLogName() string {
	if x != nil {
		return x.LogName
	}
	return ""
}

func (x *Config) GetLogLevel() string {
	if x != nil {
		return x.LogLevel
	}
	return ""
}

func (x *Config) GetLogCount() int32 {
	if x != nil {
		return x.LogCount
	}
	return 0
}

func (x *Config) GetLogRotateTime() int32 {
	if x != nil {
		return x.LogRotateTime
	}
	return 0
}

var File_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto protoreflect.FileDescriptor

var file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDesc = []byte{
	0x0a, 0x48, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x79, 0x61, 0x68,
	0x75, 0x69, 0x7a, 0x68, 0x61, 0x6e, 0x2f, 0x67, 0x6f, 0x2d, 0x64, 0x61, 0x70, 0x70, 0x6c, 0x65,
	0x79, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x2f, 0x70, 0x65, 0x72, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e,
	0x63, 0x65, 0x5f, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x70, 0x65, 0x72, 0x66,
	0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x70, 0x62,
	0x22, 0xdc, 0x02, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x18, 0x0a, 0x07, 0x67,
	0x6f, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x67, 0x6f,
	0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x70, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x02, 0x52, 0x03, 0x74, 0x70, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x6d, 0x69, 0x6e, 0x65, 0x72,
	0x50, 0x72, 0x69, 0x76, 0x4b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6d,
	0x69, 0x6e, 0x65, 0x72, 0x50, 0x72, 0x69, 0x76, 0x4b, 0x65, 0x79, 0x12, 0x2a, 0x0a, 0x10, 0x61,
	0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x46, 0x72, 0x6f, 0x6d, 0x4d, 0x69, 0x6e, 0x6e, 0x65, 0x72, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x10, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x46, 0x72, 0x6f,
	0x6d, 0x4d, 0x69, 0x6e, 0x6e, 0x65, 0x72, 0x12, 0x20, 0x0a, 0x0b, 0x61, 0x6d, 0x6f, 0x75, 0x6e,
	0x74, 0x50, 0x65, 0x72, 0x54, 0x78, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x61, 0x6d,
	0x6f, 0x75, 0x6e, 0x74, 0x50, 0x65, 0x72, 0x54, 0x78, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72,
	0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x18, 0x0a,
	0x07, 0x6c, 0x6f, 0x67, 0x4f, 0x70, 0x65, 0x6e, 0x18, 0x08, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07,
	0x6c, 0x6f, 0x67, 0x4f, 0x70, 0x65, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x6c, 0x6f, 0x67, 0x4e, 0x61,
	0x6d, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6c, 0x6f, 0x67, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x0a, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x1a, 0x0a,
	0x08, 0x6c, 0x6f, 0x67, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x08, 0x6c, 0x6f, 0x67, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x24, 0x0a, 0x0d, 0x6c, 0x6f, 0x67,
	0x52, 0x6f, 0x74, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0d, 0x6c, 0x6f, 0x67, 0x52, 0x6f, 0x74, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x42,
	0x52, 0x5a, 0x50, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x79, 0x61,
	0x68, 0x75, 0x69, 0x7a, 0x68, 0x61, 0x6e, 0x2f, 0x67, 0x6f, 0x2d, 0x64, 0x61, 0x70, 0x70, 0x6c,
	0x65, 0x79, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x2f, 0x70, 0x65, 0x72, 0x66, 0x6f, 0x72, 0x6d, 0x61,
	0x6e, 0x63, 0x65, 0x5f, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x62, 0x3b, 0x70,
	0x65, 0x72, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDescOnce sync.Once
	file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDescData = file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDesc
)

func file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDescGZIP() []byte {
	file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDescOnce.Do(func() {
		file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDescData)
	})
	return file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDescData
}

var file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_goTypes = []interface{}{
	(*Config)(nil), // 0: performance_configpb.Config
}
var file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_init() }
func file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_init() {
	if File_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
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
			RawDescriptor: file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_goTypes,
		DependencyIndexes: file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_depIdxs,
		MessageInfos:      file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_msgTypes,
	}.Build()
	File_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto = out.File
	file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_rawDesc = nil
	file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_goTypes = nil
	file_github_com_yahuizhan_go_dappley_tool_performance_testing_pb_config_proto_depIdxs = nil
}