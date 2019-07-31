// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/dappley/go-dappley/logic/pb/account_manager.proto

package logicpb

import (
	fmt "fmt"
	math "math"

	pb "github.com/dappley/go-dappley/core/client/pb"
	proto "github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type AccountManager struct {
	Accounts             []*pb.Account `protobuf:"bytes,1,rep,name=accounts,proto3" json:"accounts,omitempty"`
	PassPhrase           []byte        `protobuf:"bytes,2,opt,name=passPhrase,proto3" json:"passPhrase,omitempty"`
	Locked               bool          `protobuf:"varint,3,opt,name=locked,proto3" json:"locked,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *AccountManager) Reset()         { *m = AccountManager{} }
func (m *AccountManager) String() string { return proto.CompactTextString(m) }
func (*AccountManager) ProtoMessage()    {}
func (*AccountManager) Descriptor() ([]byte, []int) {
	return fileDescriptor_414cff4bce54c36c, []int{0}
}

func (m *AccountManager) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccountManager.Unmarshal(m, b)
}
func (m *AccountManager) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccountManager.Marshal(b, m, deterministic)
}
func (m *AccountManager) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountManager.Merge(m, src)
}
func (m *AccountManager) XXX_Size() int {
	return xxx_messageInfo_AccountManager.Size(m)
}
func (m *AccountManager) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountManager.DiscardUnknown(m)
}

var xxx_messageInfo_AccountManager proto.InternalMessageInfo

func (m *AccountManager) GetAccounts() []*pb.Account {
	if m != nil {
		return m.Accounts
	}
	return nil
}

func (m *AccountManager) GetPassPhrase() []byte {
	if m != nil {
		return m.PassPhrase
	}
	return nil
}

func (m *AccountManager) GetLocked() bool {
	if m != nil {
		return m.Locked
	}
	return false
}

func init() {
	proto.RegisterType((*AccountManager)(nil), "logicpb.AccountManager")
}

func init() {
	proto.RegisterFile("github.com/dappley/go-dappley/logic/pb/account_manager.proto", fileDescriptor_414cff4bce54c36c)
}

var fileDescriptor_414cff4bce54c36c = []byte{
	// 183 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xb2, 0x49, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0x49, 0x2c, 0x28, 0xc8, 0x49, 0xad, 0xd4, 0x4f,
	0xcf, 0xd7, 0x85, 0x31, 0x73, 0xf2, 0xd3, 0x33, 0x93, 0xf5, 0x0b, 0x92, 0xf4, 0x13, 0x93, 0x93,
	0xf3, 0x4b, 0xf3, 0x4a, 0xe2, 0x73, 0x13, 0xf3, 0x12, 0xd3, 0x53, 0x8b, 0xf4, 0x0a, 0x8a, 0xf2,
	0x4b, 0xf2, 0x85, 0xd8, 0xc1, 0xf2, 0x05, 0x49, 0x52, 0xa6, 0xf8, 0x8d, 0x49, 0xce, 0xc9, 0x4c,
	0xcd, 0x2b, 0x41, 0x32, 0x07, 0xa2, 0x5f, 0xa9, 0x82, 0x8b, 0xcf, 0x11, 0x22, 0xe0, 0x0b, 0x31,
	0x57, 0x48, 0x8f, 0x8b, 0x03, 0xaa, 0xa4, 0x58, 0x82, 0x51, 0x81, 0x59, 0x83, 0xdb, 0x48, 0x48,
	0x0f, 0x2a, 0x50, 0x90, 0xa4, 0x07, 0x55, 0x1c, 0x04, 0x57, 0x23, 0x24, 0xc7, 0xc5, 0x55, 0x90,
	0x58, 0x5c, 0x1c, 0x90, 0x51, 0x94, 0x58, 0x9c, 0x2a, 0xc1, 0xa4, 0xc0, 0xa8, 0xc1, 0x13, 0x84,
	0x24, 0x22, 0x24, 0xc6, 0xc5, 0x96, 0x93, 0x9f, 0x9c, 0x9d, 0x9a, 0x22, 0xc1, 0xac, 0xc0, 0xa8,
	0xc1, 0x11, 0x04, 0xe5, 0x25, 0xb1, 0x81, 0x1d, 0x60, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x56,
	0x8a, 0xdd, 0x3d, 0x00, 0x01, 0x00, 0x00,
}
