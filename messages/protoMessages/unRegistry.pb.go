// Code generated by protoc-gen-go. DO NOT EDIT.
// source: unRegistry.proto

package protoMessages

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

type UnRegistry struct {
	WalletAddr           string   `protobuf:"bytes,1,opt,name=walletAddr,proto3" json:"walletAddr,omitempty"`
	Type                 int32    `protobuf:"varint,3,opt,name=type,proto3" json:"type,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UnRegistry) Reset()         { *m = UnRegistry{} }
func (m *UnRegistry) String() string { return proto.CompactTextString(m) }
func (*UnRegistry) ProtoMessage()    {}
func (*UnRegistry) Descriptor() ([]byte, []int) {
	return fileDescriptor_39a561863e9f22be, []int{0}
}

func (m *UnRegistry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UnRegistry.Unmarshal(m, b)
}
func (m *UnRegistry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UnRegistry.Marshal(b, m, deterministic)
}
func (m *UnRegistry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UnRegistry.Merge(m, src)
}
func (m *UnRegistry) XXX_Size() int {
	return xxx_messageInfo_UnRegistry.Size(m)
}
func (m *UnRegistry) XXX_DiscardUnknown() {
	xxx_messageInfo_UnRegistry.DiscardUnknown(m)
}

var xxx_messageInfo_UnRegistry proto.InternalMessageInfo

func (m *UnRegistry) GetWalletAddr() string {
	if m != nil {
		return m.WalletAddr
	}
	return ""
}

func (m *UnRegistry) GetType() int32 {
	if m != nil {
		return m.Type
	}
	return 0
}

func init() {
	proto.RegisterType((*UnRegistry)(nil), "protoMessages.UnRegistry")
}

func init() { proto.RegisterFile("unRegistry.proto", fileDescriptor_39a561863e9f22be) }

var fileDescriptor_39a561863e9f22be = []byte{
	// 106 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x28, 0xcd, 0x0b, 0x4a,
	0x4d, 0xcf, 0x2c, 0x2e, 0x29, 0xaa, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x05, 0x53,
	0xbe, 0xa9, 0xc5, 0xc5, 0x89, 0xe9, 0xa9, 0xc5, 0x4a, 0x0e, 0x5c, 0x5c, 0xa1, 0x70, 0x25, 0x42,
	0x72, 0x5c, 0x5c, 0xe5, 0x89, 0x39, 0x39, 0xa9, 0x25, 0x8e, 0x29, 0x29, 0x45, 0x12, 0x8c, 0x0a,
	0x8c, 0x1a, 0x9c, 0x41, 0x48, 0x22, 0x42, 0x42, 0x5c, 0x2c, 0x25, 0x95, 0x05, 0xa9, 0x12, 0xcc,
	0x0a, 0x8c, 0x1a, 0xac, 0x41, 0x60, 0x76, 0x12, 0x1b, 0xd8, 0x40, 0x63, 0x40, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xd5, 0xad, 0x4b, 0x50, 0x6b, 0x00, 0x00, 0x00,
}
