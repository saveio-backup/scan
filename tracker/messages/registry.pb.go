// Code generated by protoc-gen-go. DO NOT EDIT.
// source: registry.proto

package messages

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

type Registry struct {
	WalletAddr           string   `protobuf:"bytes,1,opt,name=walletAddr,proto3" json:"walletAddr,omitempty"`
	HostPort             string   `protobuf:"bytes,2,opt,name=hostPort,proto3" json:"hostPort,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Registry) Reset()         { *m = Registry{} }
func (m *Registry) String() string { return proto.CompactTextString(m) }
func (*Registry) ProtoMessage()    {}
func (*Registry) Descriptor() ([]byte, []int) {
	return fileDescriptor_41af05d40a615591, []int{0}
}

func (m *Registry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Registry.Unmarshal(m, b)
}
func (m *Registry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Registry.Marshal(b, m, deterministic)
}
func (m *Registry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Registry.Merge(m, src)
}
func (m *Registry) XXX_Size() int {
	return xxx_messageInfo_Registry.Size(m)
}
func (m *Registry) XXX_DiscardUnknown() {
	xxx_messageInfo_Registry.DiscardUnknown(m)
}

var xxx_messageInfo_Registry proto.InternalMessageInfo

func (m *Registry) GetWalletAddr() string {
	if m != nil {
		return m.WalletAddr
	}
	return ""
}

func (m *Registry) GetHostPort() string {
	if m != nil {
		return m.HostPort
	}
	return ""
}

func init() {
	proto.RegisterType((*Registry)(nil), "messages.Registry")
}

func init() { proto.RegisterFile("registry.proto", fileDescriptor_41af05d40a615591) }

var fileDescriptor_41af05d40a615591 = []byte{
	// 103 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2b, 0x4a, 0x4d, 0xcf,
	0x2c, 0x2e, 0x29, 0xaa, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0xc8, 0x4d, 0x2d, 0x2e,
	0x4e, 0x4c, 0x4f, 0x2d, 0x56, 0x72, 0xe3, 0xe2, 0x08, 0x82, 0xca, 0x09, 0xc9, 0x71, 0x71, 0x95,
	0x27, 0xe6, 0xe4, 0xa4, 0x96, 0x38, 0xa6, 0xa4, 0x14, 0x49, 0x30, 0x2a, 0x30, 0x6a, 0x70, 0x06,
	0x21, 0x89, 0x08, 0x49, 0x71, 0x71, 0x64, 0xe4, 0x17, 0x97, 0x04, 0xe4, 0x17, 0x95, 0x48, 0x30,
	0x81, 0x65, 0xe1, 0xfc, 0x24, 0x36, 0xb0, 0xc1, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x93,
	0x36, 0xd3, 0x9f, 0x6a, 0x00, 0x00, 0x00,
}
