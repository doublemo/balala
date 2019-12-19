// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package transport

import (
	"context"
	"errors"
	"io"
	"time"

	agentendpoint "github.com/doublemo/balala/agent/endpoint"
	"github.com/doublemo/balala/agent/service"
	"github.com/doublemo/balala/cores/proto/pb"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/etcdv3"
	"github.com/go-kit/kit/sd/lb"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// GRPCServer 内部响应GRPC服务
type GRPCServer struct {
	call   grpctransport.Handler
	stream grpctransport.Handler
}

// Call 远程调用
func (s *GRPCServer) Call(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	_, rep, err := s.call.ServeGRPC(ctx, in)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.Response), nil
}

// Stream 流支持
func (s *GRPCServer) Stream(stream pb.Internal_StreamServer) error {
	s.stream.ServeGRPC(context.Background(), stream)
	return nil
}

// NewGRPCServer 创建内部服务grpc server
func NewGRPCServer(endpoints agentendpoint.Set, logger log.Logger) pb.InternalServer {
	options := []grpctransport.ServerOption{grpctransport.ServerErrorLogger(logger)}

	return &GRPCServer{
		call: grpctransport.NewServer(
			endpoints.CallEndpoint,
			decodeGRPCRequest,
			encodeGRPCResponse,
			append(options, grpctransport.ServerBefore(jwt.GRPCToContext()))...,
		),

		stream: grpctransport.NewServer(
			endpoints.StreamEndpoint,
			decodeGRPCRequest,
			encodeGRPCResponse,
			append(options, grpctransport.ServerBefore(jwt.GRPCToContext()))...,
		),
	}
}

// NewGRPCClient 创建内部服务grpc client
func NewGRPCClient(conn *grpc.ClientConn, logger log.Logger, extends ...endpoint.Middleware) service.GRPC {
	var callEndpoint endpoint.Endpoint
	{
		callEndpoint = grpctransport.NewClient(conn,
			"pb.InternalServer",
			"Call",
			encodeGRPRequest,
			decodeGRPCResponse,
			pb.Response{},
			grpctransport.ClientBefore(jwt.ContextToGRPC()),
		).Endpoint()
		for _, e := range extends {
			callEndpoint = e(callEndpoint)
		}
	}

	return agentendpoint.Set{
		CallEndpoint: callEndpoint,
	}
}

// MakeInstancer 创建服务实例
func MakeInstancer(client etcdv3.Client, firefix string, logger log.Logger) (sd.Instancer, error) {
	return etcdv3.NewInstancer(client, service.MakeKey(firefix, ""), logger)
}

// MakeRetry 创建Subscribe方法的客户调用
// 支持RoundRobin 多节点分布式调用
func MakeRetry(instancer sd.Instancer, factory sd.Factory, retryMax int, retryTimeout time.Duration, logger log.Logger) endpoint.Endpoint {
	endpointer := sd.NewEndpointer(instancer, factory, logger)
	balancer := lb.NewRoundRobin(endpointer)
	retry := lb.Retry(retryMax, retryTimeout, balancer)
	return retry
}

// MakeRetryStream 流服务支持
func MakeRetryStream(instancer sd.Instancer, logger log.Logger) endpoint.Endpoint {
	endpointer := sd.NewEndpointer(instancer, MakeFactoryStream(), logger)
	balancer := lb.NewRoundRobin(endpointer)
	return func(ctx context.Context, req interface{}) (response interface{}, err error) {
		fn, err := balancer.Endpoint()
		if err != nil {
			return nil, err
		}

		connect, err := fn(nil, req)
		if err != nil {
			return nil, err
		}
		return connect, nil
	}
}

// MakeFactoryStream 创建流服务支持
func MakeFactoryStream() sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		return makeStreamEndpoint(conn), conn, nil
	}
}

func makeStreamEndpoint(conn *grpc.ClientConn) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (response interface{}, err error) {
		param, ok := req.(map[string]string)
		if !ok {
			return nil, errors.New("Invalid params")
		}

		cli := pb.NewInternalClient(conn)
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := cli.Stream(metadata.NewOutgoingContext(ctx, metadata.New(param)))
		if err != nil {
			cancel()
			return nil, err
		}

		return &DefaultGRPCStream{stream: stream, cancel: cancel}, nil
	}
}

// MakeFactoryCall Call
func MakeFactoryCall(logger log.Logger, extends ...endpoint.Middleware) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}

		s := NewGRPCClient(conn, logger, extends...)
		doEndpoint := agentendpoint.MakeCallEndpoint(s)
		return doEndpoint, conn, nil
	}
}

func encodeGRPRequest(_ context.Context, request interface{}) (interface{}, error) {
	return request, nil
}

func decodeGRPCRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	return grpcReq, nil
}

func decodeGRPCResponse(_ context.Context, response interface{}) (interface{}, error) {
	return response, nil
}

func encodeGRPCResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	return grpcReply, nil
}
