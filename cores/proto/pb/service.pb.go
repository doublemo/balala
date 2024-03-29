// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service.proto

package pb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type Header struct {
	SID                  string   `protobuf:"bytes,1,opt,name=SID,proto3" json:"SID,omitempty"`
	UserID               uint64   `protobuf:"varint,2,opt,name=UserID,proto3" json:"UserID,omitempty"`
	V                    int32    `protobuf:"varint,3,opt,name=V,proto3" json:"V,omitempty"`
	From                 int32    `protobuf:"varint,4,opt,name=From,proto3" json:"From,omitempty"`
	FromAddress          string   `protobuf:"bytes,5,opt,name=FromAddress,proto3" json:"FromAddress,omitempty"`
	Method               int32    `protobuf:"varint,6,opt,name=Method,proto3" json:"Method,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Header) Reset()         { *m = Header{} }
func (m *Header) String() string { return proto.CompactTextString(m) }
func (*Header) ProtoMessage()    {}
func (*Header) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{0}
}

func (m *Header) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Header.Unmarshal(m, b)
}
func (m *Header) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Header.Marshal(b, m, deterministic)
}
func (m *Header) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Header.Merge(m, src)
}
func (m *Header) XXX_Size() int {
	return xxx_messageInfo_Header.Size(m)
}
func (m *Header) XXX_DiscardUnknown() {
	xxx_messageInfo_Header.DiscardUnknown(m)
}

var xxx_messageInfo_Header proto.InternalMessageInfo

func (m *Header) GetSID() string {
	if m != nil {
		return m.SID
	}
	return ""
}

func (m *Header) GetUserID() uint64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

func (m *Header) GetV() int32 {
	if m != nil {
		return m.V
	}
	return 0
}

func (m *Header) GetFrom() int32 {
	if m != nil {
		return m.From
	}
	return 0
}

func (m *Header) GetFromAddress() string {
	if m != nil {
		return m.FromAddress
	}
	return ""
}

func (m *Header) GetMethod() int32 {
	if m != nil {
		return m.Method
	}
	return 0
}

type Request struct {
	Header               *Header  `protobuf:"bytes,1,opt,name=Header,proto3" json:"Header,omitempty"`
	Command              int32    `protobuf:"varint,2,opt,name=Command,proto3" json:"Command,omitempty"`
	Body                 []byte   `protobuf:"bytes,3,opt,name=Body,proto3" json:"Body,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{1}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetHeader() *Header {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *Request) GetCommand() int32 {
	if m != nil {
		return m.Command
	}
	return 0
}

func (m *Request) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

type Response struct {
	Command              int32    `protobuf:"varint,1,opt,name=Command,proto3" json:"Command,omitempty"`
	Body                 []byte   `protobuf:"bytes,2,opt,name=Body,proto3" json:"Body,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{2}
}

func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (m *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(m, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

func (m *Response) GetCommand() int32 {
	if m != nil {
		return m.Command
	}
	return 0
}

func (m *Response) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

func init() {
	proto.RegisterType((*Header)(nil), "pb.Header")
	proto.RegisterType((*Request)(nil), "pb.Request")
	proto.RegisterType((*Response)(nil), "pb.Response")
}

func init() { proto.RegisterFile("service.proto", fileDescriptor_a0b84a42fa06f626) }

var fileDescriptor_a0b84a42fa06f626 = []byte{
	// 272 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0x31, 0x4b, 0xfc, 0x40,
	0x10, 0xc5, 0xff, 0x93, 0x4b, 0x72, 0xf7, 0x9f, 0x44, 0x90, 0x29, 0x64, 0xb1, 0x0a, 0xb1, 0x89,
	0x4d, 0x90, 0xb3, 0xb1, 0xd5, 0x3b, 0xc4, 0x14, 0x36, 0x7b, 0x78, 0x85, 0x56, 0x89, 0x3b, 0xa0,
	0x90, 0x64, 0xe3, 0x6e, 0x14, 0xfc, 0x12, 0x7e, 0x66, 0xc9, 0x26, 0xc1, 0x2b, 0xc4, 0x6a, 0xdf,
	0x7b, 0xec, 0x7b, 0xfc, 0x60, 0xf0, 0xc8, 0xb2, 0xf9, 0x78, 0x7d, 0xe6, 0xbc, 0x33, 0xba, 0xd7,
	0xe4, 0x75, 0x55, 0xfa, 0x05, 0x18, 0xde, 0x71, 0xa9, 0xd8, 0xd0, 0x31, 0x2e, 0x76, 0xc5, 0x56,
	0x40, 0x02, 0xd9, 0x7f, 0x39, 0x48, 0x3a, 0xc1, 0xf0, 0xc1, 0xb2, 0x29, 0xb6, 0xc2, 0x4b, 0x20,
	0xf3, 0xe5, 0xe4, 0x28, 0x46, 0xd8, 0x8b, 0x45, 0x02, 0x59, 0x20, 0x61, 0x4f, 0x84, 0xfe, 0xad,
	0xd1, 0x8d, 0xf0, 0x5d, 0xe0, 0x34, 0x25, 0x18, 0x0d, 0xef, 0xb5, 0x52, 0x86, 0xad, 0x15, 0x81,
	0xdb, 0x3c, 0x8c, 0x86, 0xed, 0x7b, 0xee, 0x5f, 0xb4, 0x12, 0xa1, 0xeb, 0x4d, 0x2e, 0x7d, 0xc2,
	0xa5, 0xe4, 0xb7, 0x77, 0xb6, 0x3d, 0xa5, 0x33, 0x9a, 0x63, 0x8a, 0xd6, 0x98, 0x77, 0x55, 0x3e,
	0x26, 0x72, 0x86, 0x16, 0xb8, 0xdc, 0xe8, 0xa6, 0x29, 0x5b, 0xe5, 0x18, 0x03, 0x39, 0xdb, 0x01,
	0xeb, 0x46, 0xab, 0x4f, 0xc7, 0x19, 0x4b, 0xa7, 0xd3, 0x2b, 0x5c, 0x49, 0xb6, 0x9d, 0x6e, 0x2d,
	0x1f, 0x36, 0xe1, 0xf7, 0xa6, 0xf7, 0xd3, 0x5c, 0x3f, 0xe2, 0xaa, 0x68, 0x7b, 0x36, 0x6d, 0x59,
	0xd3, 0x19, 0xfa, 0x9b, 0xb2, 0xae, 0x29, 0x1a, 0x78, 0x26, 0xd8, 0xd3, 0x78, 0x34, 0xe3, 0x78,
	0xfa, 0x8f, 0xce, 0x31, 0xdc, 0xf5, 0x86, 0xcb, 0xe6, 0xcf, 0x6f, 0x19, 0x5c, 0x40, 0x15, 0xba,
	0x73, 0x5c, 0x7e, 0x07, 0x00, 0x00, 0xff, 0xff, 0x74, 0xde, 0x40, 0xb4, 0x9f, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// InternalClient is the client API for Internal service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type InternalClient interface {
	Call(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	Stream(ctx context.Context, opts ...grpc.CallOption) (Internal_StreamClient, error)
}

type internalClient struct {
	cc *grpc.ClientConn
}

func NewInternalClient(cc *grpc.ClientConn) InternalClient {
	return &internalClient{cc}
}

func (c *internalClient) Call(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/pb.Internal/Call", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *internalClient) Stream(ctx context.Context, opts ...grpc.CallOption) (Internal_StreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Internal_serviceDesc.Streams[0], "/pb.Internal/Stream", opts...)
	if err != nil {
		return nil, err
	}
	x := &internalStreamClient{stream}
	return x, nil
}

type Internal_StreamClient interface {
	Send(*Request) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type internalStreamClient struct {
	grpc.ClientStream
}

func (x *internalStreamClient) Send(m *Request) error {
	return x.ClientStream.SendMsg(m)
}

func (x *internalStreamClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// InternalServer is the server API for Internal service.
type InternalServer interface {
	Call(context.Context, *Request) (*Response, error)
	Stream(Internal_StreamServer) error
}

// UnimplementedInternalServer can be embedded to have forward compatible implementations.
type UnimplementedInternalServer struct {
}

func (*UnimplementedInternalServer) Call(ctx context.Context, req *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}
func (*UnimplementedInternalServer) Stream(srv Internal_StreamServer) error {
	return status.Errorf(codes.Unimplemented, "method Stream not implemented")
}

func RegisterInternalServer(s *grpc.Server, srv InternalServer) {
	s.RegisterService(&_Internal_serviceDesc, srv)
}

func _Internal_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InternalServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Internal/Call",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InternalServer).Call(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Internal_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(InternalServer).Stream(&internalStreamServer{stream})
}

type Internal_StreamServer interface {
	Send(*Response) error
	Recv() (*Request, error)
	grpc.ServerStream
}

type internalStreamServer struct {
	grpc.ServerStream
}

func (x *internalStreamServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *internalStreamServer) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Internal_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Internal",
	HandlerType: (*InternalServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Call",
			Handler:    _Internal_Call_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _Internal_Stream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "service.proto",
}
