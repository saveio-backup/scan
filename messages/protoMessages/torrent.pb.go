// Code generated by protoc-gen-go. DO NOT EDIT.
// source: torrent.proto

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

type Torrent struct {
	InfoHash             []byte   `protobuf:"bytes,1,opt,name=infoHash,proto3" json:"infoHash,omitempty"`
	Torrent              []byte   `protobuf:"bytes,2,opt,name=torrent,proto3" json:"torrent,omitempty"`
	Type                 int32    `protobuf:"varint,3,opt,name=type,proto3" json:"type,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Torrent) Reset()         { *m = Torrent{} }
func (m *Torrent) String() string { return proto.CompactTextString(m) }
func (*Torrent) ProtoMessage()    {}
func (*Torrent) Descriptor() ([]byte, []int) {
	return fileDescriptor_5390c0f93d3a953f, []int{0}
}

func (m *Torrent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Torrent.Unmarshal(m, b)
}
func (m *Torrent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Torrent.Marshal(b, m, deterministic)
}
func (m *Torrent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Torrent.Merge(m, src)
}
func (m *Torrent) XXX_Size() int {
	return xxx_messageInfo_Torrent.Size(m)
}
func (m *Torrent) XXX_DiscardUnknown() {
	xxx_messageInfo_Torrent.DiscardUnknown(m)
}

var xxx_messageInfo_Torrent proto.InternalMessageInfo

func (m *Torrent) GetInfoHash() []byte {
	if m != nil {
		return m.InfoHash
	}
	return nil
}

func (m *Torrent) GetTorrent() []byte {
	if m != nil {
		return m.Torrent
	}
	return nil
}

func (m *Torrent) GetType() int32 {
	if m != nil {
		return m.Type
	}
	return 0
}

type Bytes struct {
	Data                 []byte   `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Bytes) Reset()         { *m = Bytes{} }
func (m *Bytes) String() string { return proto.CompactTextString(m) }
func (*Bytes) ProtoMessage()    {}
func (*Bytes) Descriptor() ([]byte, []int) {
	return fileDescriptor_5390c0f93d3a953f, []int{1}
}

func (m *Bytes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Bytes.Unmarshal(m, b)
}
func (m *Bytes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Bytes.Marshal(b, m, deterministic)
}
func (m *Bytes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Bytes.Merge(m, src)
}
func (m *Bytes) XXX_Size() int {
	return xxx_messageInfo_Bytes.Size(m)
}
func (m *Bytes) XXX_DiscardUnknown() {
	xxx_messageInfo_Bytes.DiscardUnknown(m)
}

var xxx_messageInfo_Bytes proto.InternalMessageInfo

func (m *Bytes) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type Keepalive struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Keepalive) Reset()         { *m = Keepalive{} }
func (m *Keepalive) String() string { return proto.CompactTextString(m) }
func (*Keepalive) ProtoMessage()    {}
func (*Keepalive) Descriptor() ([]byte, []int) {
	return fileDescriptor_5390c0f93d3a953f, []int{2}
}

func (m *Keepalive) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Keepalive.Unmarshal(m, b)
}
func (m *Keepalive) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Keepalive.Marshal(b, m, deterministic)
}
func (m *Keepalive) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Keepalive.Merge(m, src)
}
func (m *Keepalive) XXX_Size() int {
	return xxx_messageInfo_Keepalive.Size(m)
}
func (m *Keepalive) XXX_DiscardUnknown() {
	xxx_messageInfo_Keepalive.DiscardUnknown(m)
}

var xxx_messageInfo_Keepalive proto.InternalMessageInfo

type KeepaliveResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KeepaliveResponse) Reset()         { *m = KeepaliveResponse{} }
func (m *KeepaliveResponse) String() string { return proto.CompactTextString(m) }
func (*KeepaliveResponse) ProtoMessage()    {}
func (*KeepaliveResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_5390c0f93d3a953f, []int{3}
}

func (m *KeepaliveResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeepaliveResponse.Unmarshal(m, b)
}
func (m *KeepaliveResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeepaliveResponse.Marshal(b, m, deterministic)
}
func (m *KeepaliveResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeepaliveResponse.Merge(m, src)
}
func (m *KeepaliveResponse) XXX_Size() int {
	return xxx_messageInfo_KeepaliveResponse.Size(m)
}
func (m *KeepaliveResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_KeepaliveResponse.DiscardUnknown(m)
}

var xxx_messageInfo_KeepaliveResponse proto.InternalMessageInfo

type Disconnect struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Disconnect) Reset()         { *m = Disconnect{} }
func (m *Disconnect) String() string { return proto.CompactTextString(m) }
func (*Disconnect) ProtoMessage()    {}
func (*Disconnect) Descriptor() ([]byte, []int) {
	return fileDescriptor_5390c0f93d3a953f, []int{4}
}

func (m *Disconnect) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Disconnect.Unmarshal(m, b)
}
func (m *Disconnect) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Disconnect.Marshal(b, m, deterministic)
}
func (m *Disconnect) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Disconnect.Merge(m, src)
}
func (m *Disconnect) XXX_Size() int {
	return xxx_messageInfo_Disconnect.Size(m)
}
func (m *Disconnect) XXX_DiscardUnknown() {
	xxx_messageInfo_Disconnect.DiscardUnknown(m)
}

var xxx_messageInfo_Disconnect proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Torrent)(nil), "protoMessages.Torrent")
	proto.RegisterType((*Bytes)(nil), "protoMessages.Bytes")
	proto.RegisterType((*Keepalive)(nil), "protoMessages.Keepalive")
	proto.RegisterType((*KeepaliveResponse)(nil), "protoMessages.KeepaliveResponse")
	proto.RegisterType((*Disconnect)(nil), "protoMessages.Disconnect")
}

func init() { proto.RegisterFile("torrent.proto", fileDescriptor_5390c0f93d3a953f) }

var fileDescriptor_5390c0f93d3a953f = []byte{
	// 171 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x8e, 0x31, 0xca, 0xc2, 0x40,
	0x10, 0x46, 0xc9, 0xff, 0x1b, 0xa3, 0x63, 0x52, 0xb8, 0x36, 0x8b, 0x36, 0x61, 0xab, 0x54, 0x36,
	0xde, 0x40, 0x2c, 0x04, 0xb1, 0x89, 0x5e, 0x60, 0x8d, 0xa3, 0x06, 0x64, 0x67, 0xd9, 0x19, 0x84,
	0xdc, 0x5e, 0x5c, 0x63, 0xaa, 0xf9, 0x1e, 0x0f, 0x1e, 0x03, 0x85, 0x50, 0x08, 0xe8, 0x64, 0xed,
	0x03, 0x09, 0xa9, 0x22, 0x9e, 0x23, 0x32, 0xdb, 0x3b, 0xb2, 0x39, 0x41, 0x76, 0xfe, 0x7a, 0xb5,
	0x84, 0x49, 0xeb, 0x6e, 0xb4, 0xb7, 0xfc, 0xd0, 0x49, 0x99, 0x54, 0x79, 0x3d, 0xb0, 0xd2, 0x90,
	0xf5, 0x19, 0xfd, 0x17, 0xd5, 0x0f, 0x95, 0x82, 0x91, 0x74, 0x1e, 0xf5, 0x7f, 0x99, 0x54, 0x69,
	0x1d, 0xb7, 0x59, 0x41, 0xba, 0xed, 0x04, 0xf9, 0x23, 0xaf, 0x56, 0x6c, 0x9f, 0x8b, 0xdb, 0xcc,
	0x60, 0x7a, 0x40, 0xf4, 0xf6, 0xd9, 0xbe, 0xd0, 0x2c, 0x60, 0x3e, 0x40, 0x8d, 0xec, 0xc9, 0x31,
	0x9a, 0x1c, 0x60, 0xd7, 0x72, 0x43, 0xce, 0x61, 0x23, 0x97, 0x71, 0x7c, 0x78, 0xf3, 0x0e, 0x00,
	0x00, 0xff, 0xff, 0x2a, 0x85, 0xbc, 0x3f, 0xc8, 0x00, 0x00, 0x00,
}
