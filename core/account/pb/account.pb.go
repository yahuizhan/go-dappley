// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.13.0
// source: github.com/dappley/go-dappley/core/account/pb/account.proto

package accountpb

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

type Account struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	KeyPair    *KeyPair `protobuf:"bytes,1,opt,name=keyPair,proto3" json:"keyPair,omitempty"`
	Address    *Address `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	PubKeyHash []byte   `protobuf:"bytes,3,opt,name=pubKeyHash,proto3" json:"pubKeyHash,omitempty"`
}

func (x *Account) Reset() {
	*x = Account{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Account) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Account) ProtoMessage() {}

func (x *Account) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Account.ProtoReflect.Descriptor instead.
func (*Account) Descriptor() ([]byte, []int) {
	return file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescGZIP(), []int{0}
}

func (x *Account) GetKeyPair() *KeyPair {
	if x != nil {
		return x.KeyPair
	}
	return nil
}

func (x *Account) GetAddress() *Address {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *Account) GetPubKeyHash() []byte {
	if x != nil {
		return x.PubKeyHash
	}
	return nil
}

type TransactionAccount struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address    *Address `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	PubKeyHash []byte   `protobuf:"bytes,2,opt,name=pubKeyHash,proto3" json:"pubKeyHash,omitempty"`
}

func (x *TransactionAccount) Reset() {
	*x = TransactionAccount{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TransactionAccount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransactionAccount) ProtoMessage() {}

func (x *TransactionAccount) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransactionAccount.ProtoReflect.Descriptor instead.
func (*TransactionAccount) Descriptor() ([]byte, []int) {
	return file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescGZIP(), []int{1}
}

func (x *TransactionAccount) GetAddress() *Address {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *TransactionAccount) GetPubKeyHash() []byte {
	if x != nil {
		return x.PubKeyHash
	}
	return nil
}

type KeyPair struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PrivateKey []byte `protobuf:"bytes,1,opt,name=privateKey,proto3" json:"privateKey,omitempty"`
	PublicKey  []byte `protobuf:"bytes,2,opt,name=publicKey,proto3" json:"publicKey,omitempty"`
}

func (x *KeyPair) Reset() {
	*x = KeyPair{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeyPair) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeyPair) ProtoMessage() {}

func (x *KeyPair) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KeyPair.ProtoReflect.Descriptor instead.
func (*KeyPair) Descriptor() ([]byte, []int) {
	return file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescGZIP(), []int{2}
}

func (x *KeyPair) GetPrivateKey() []byte {
	if x != nil {
		return x.PrivateKey
	}
	return nil
}

func (x *KeyPair) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

type Address struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *Address) Reset() {
	*x = Address{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Address) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Address) ProtoMessage() {}

func (x *Address) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Address.ProtoReflect.Descriptor instead.
func (*Address) Descriptor() ([]byte, []int) {
	return file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescGZIP(), []int{3}
}

func (x *Address) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

type AccountConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FilePath string `protobuf:"bytes,1,opt,name=file_path,json=filePath,proto3" json:"file_path,omitempty"`
}

func (x *AccountConfig) Reset() {
	*x = AccountConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AccountConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccountConfig) ProtoMessage() {}

func (x *AccountConfig) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccountConfig.ProtoReflect.Descriptor instead.
func (*AccountConfig) Descriptor() ([]byte, []int) {
	return file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescGZIP(), []int{4}
}

func (x *AccountConfig) GetFilePath() string {
	if x != nil {
		return x.FilePath
	}
	return ""
}

var File_github_com_dappley_go_dappley_core_account_pb_account_proto protoreflect.FileDescriptor

var file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDesc = []byte{
	0x0a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x61, 0x70,
	0x70, 0x6c, 0x65, 0x79, 0x2f, 0x67, 0x6f, 0x2d, 0x64, 0x61, 0x70, 0x70, 0x6c, 0x65, 0x79, 0x2f,
	0x63, 0x6f, 0x72, 0x65, 0x2f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2f, 0x70, 0x62, 0x2f,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x61,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x70, 0x62, 0x22, 0x85, 0x01, 0x0a, 0x07, 0x41, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x12, 0x2c, 0x0a, 0x07, 0x6b, 0x65, 0x79, 0x50, 0x61, 0x69, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x70,
	0x62, 0x2e, 0x4b, 0x65, 0x79, 0x50, 0x61, 0x69, 0x72, 0x52, 0x07, 0x6b, 0x65, 0x79, 0x50, 0x61,
	0x69, 0x72, 0x12, 0x2c, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x70, 0x62, 0x2e,
	0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x48, 0x61, 0x73, 0x68, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x70, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x48, 0x61, 0x73, 0x68,
	0x22, 0x62, 0x0a, 0x12, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x41,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x2c, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x70, 0x62, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x07, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x48, 0x61,
	0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x70, 0x75, 0x62, 0x4b, 0x65, 0x79,
	0x48, 0x61, 0x73, 0x68, 0x22, 0x47, 0x0a, 0x07, 0x4b, 0x65, 0x79, 0x50, 0x61, 0x69, 0x72, 0x12,
	0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x0a, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x12,
	0x1c, 0x0a, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x22, 0x23, 0x0a,
	0x07, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x22, 0x2c, 0x0a, 0x0d, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x12, 0x1b, 0x0a, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x70, 0x61, 0x74, 0x68,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x50, 0x61, 0x74, 0x68,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescOnce sync.Once
	file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescData = file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDesc
)

func file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescGZIP() []byte {
	file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescOnce.Do(func() {
		file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescData)
	})
	return file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDescData
}

var file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_github_com_dappley_go_dappley_core_account_pb_account_proto_goTypes = []interface{}{
	(*Account)(nil),            // 0: accountpb.Account
	(*TransactionAccount)(nil), // 1: accountpb.TransactionAccount
	(*KeyPair)(nil),            // 2: accountpb.KeyPair
	(*Address)(nil),            // 3: accountpb.Address
	(*AccountConfig)(nil),      // 4: accountpb.AccountConfig
}
var file_github_com_dappley_go_dappley_core_account_pb_account_proto_depIdxs = []int32{
	2, // 0: accountpb.Account.keyPair:type_name -> accountpb.KeyPair
	3, // 1: accountpb.Account.address:type_name -> accountpb.Address
	3, // 2: accountpb.TransactionAccount.address:type_name -> accountpb.Address
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_github_com_dappley_go_dappley_core_account_pb_account_proto_init() }
func file_github_com_dappley_go_dappley_core_account_pb_account_proto_init() {
	if File_github_com_dappley_go_dappley_core_account_pb_account_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Account); i {
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
		file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TransactionAccount); i {
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
		file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KeyPair); i {
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
		file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Address); i {
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
		file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AccountConfig); i {
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
			RawDescriptor: file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_dappley_go_dappley_core_account_pb_account_proto_goTypes,
		DependencyIndexes: file_github_com_dappley_go_dappley_core_account_pb_account_proto_depIdxs,
		MessageInfos:      file_github_com_dappley_go_dappley_core_account_pb_account_proto_msgTypes,
	}.Build()
	File_github_com_dappley_go_dappley_core_account_pb_account_proto = out.File
	file_github_com_dappley_go_dappley_core_account_pb_account_proto_rawDesc = nil
	file_github_com_dappley_go_dappley_core_account_pb_account_proto_goTypes = nil
	file_github_com_dappley_go_dappley_core_account_pb_account_proto_depIdxs = nil
}
