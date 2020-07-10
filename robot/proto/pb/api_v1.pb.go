// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api_v1.proto

package pb

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

type ApiV1 struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ApiV1) Reset()         { *m = ApiV1{} }
func (m *ApiV1) String() string { return proto.CompactTextString(m) }
func (*ApiV1) ProtoMessage()    {}
func (*ApiV1) Descriptor() ([]byte, []int) {
	return fileDescriptor_42af7352bbfa1c23, []int{0}
}

func (m *ApiV1) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ApiV1.Unmarshal(m, b)
}
func (m *ApiV1) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ApiV1.Marshal(b, m, deterministic)
}
func (m *ApiV1) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ApiV1.Merge(m, src)
}
func (m *ApiV1) XXX_Size() int {
	return xxx_messageInfo_ApiV1.Size(m)
}
func (m *ApiV1) XXX_DiscardUnknown() {
	xxx_messageInfo_ApiV1.DiscardUnknown(m)
}

var xxx_messageInfo_ApiV1 proto.InternalMessageInfo

// 创建机器人
type ApiV1_CreateRequest struct {
	RobotID              string   `protobuf:"bytes,1,opt,name=RobotID,proto3" json:"RobotID,omitempty"`
	Nickname             string   `protobuf:"bytes,2,opt,name=Nickname,proto3" json:"Nickname,omitempty"`
	Username             string   `protobuf:"bytes,3,opt,name=Username,proto3" json:"Username,omitempty"`
	Password             string   `protobuf:"bytes,4,opt,name=Password,proto3" json:"Password,omitempty"`
	Body                 []byte   `protobuf:"bytes,5,opt,name=Body,proto3" json:"Body,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ApiV1_CreateRequest) Reset()         { *m = ApiV1_CreateRequest{} }
func (m *ApiV1_CreateRequest) String() string { return proto.CompactTextString(m) }
func (*ApiV1_CreateRequest) ProtoMessage()    {}
func (*ApiV1_CreateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_42af7352bbfa1c23, []int{0, 0}
}

func (m *ApiV1_CreateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ApiV1_CreateRequest.Unmarshal(m, b)
}
func (m *ApiV1_CreateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ApiV1_CreateRequest.Marshal(b, m, deterministic)
}
func (m *ApiV1_CreateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ApiV1_CreateRequest.Merge(m, src)
}
func (m *ApiV1_CreateRequest) XXX_Size() int {
	return xxx_messageInfo_ApiV1_CreateRequest.Size(m)
}
func (m *ApiV1_CreateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ApiV1_CreateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ApiV1_CreateRequest proto.InternalMessageInfo

func (m *ApiV1_CreateRequest) GetRobotID() string {
	if m != nil {
		return m.RobotID
	}
	return ""
}

func (m *ApiV1_CreateRequest) GetNickname() string {
	if m != nil {
		return m.Nickname
	}
	return ""
}

func (m *ApiV1_CreateRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *ApiV1_CreateRequest) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *ApiV1_CreateRequest) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

func init() {
	proto.RegisterType((*ApiV1)(nil), "pb.ApiV1")
	proto.RegisterType((*ApiV1_CreateRequest)(nil), "pb.ApiV1.CreateRequest")
}

func init() { proto.RegisterFile("api_v1.proto", fileDescriptor_42af7352bbfa1c23) }

var fileDescriptor_42af7352bbfa1c23 = []byte{
	// 158 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x49, 0x2c, 0xc8, 0x8c,
	0x2f, 0x33, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2a, 0x48, 0x52, 0x9a, 0xcd, 0xc8,
	0xc5, 0xea, 0x58, 0x90, 0x19, 0x66, 0x28, 0x35, 0x91, 0x91, 0x8b, 0xd7, 0xb9, 0x28, 0x35, 0xb1,
	0x24, 0x35, 0x28, 0xb5, 0xb0, 0x34, 0xb5, 0xb8, 0x44, 0x48, 0x82, 0x8b, 0x3d, 0x28, 0x3f, 0x29,
	0xbf, 0xc4, 0xd3, 0x45, 0x82, 0x51, 0x81, 0x51, 0x83, 0x33, 0x08, 0xc6, 0x15, 0x92, 0xe2, 0xe2,
	0xf0, 0xcb, 0x4c, 0xce, 0xce, 0x4b, 0xcc, 0x4d, 0x95, 0x60, 0x02, 0x4b, 0xc1, 0xf9, 0x20, 0xb9,
	0xd0, 0xe2, 0xd4, 0x22, 0xb0, 0x1c, 0x33, 0x44, 0x0e, 0xc6, 0x07, 0xc9, 0x05, 0x24, 0x16, 0x17,
	0x97, 0xe7, 0x17, 0xa5, 0x48, 0xb0, 0x40, 0xe4, 0x60, 0x7c, 0x21, 0x21, 0x2e, 0x16, 0xa7, 0xfc,
	0x94, 0x4a, 0x09, 0x56, 0x05, 0x46, 0x0d, 0x9e, 0x20, 0x30, 0x3b, 0x89, 0x0d, 0xec, 0x50, 0x63,
	0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0xb1, 0x6a, 0xd5, 0xe9, 0xb8, 0x00, 0x00, 0x00,
}
