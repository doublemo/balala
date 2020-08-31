// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package transport

import (
	"context"
	"errors"
	"io"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/doublemo/balala/cores/services"
	sssendpoint "github.com/doublemo/balala/sss/endpoint"
	"github.com/doublemo/balala/sss/proto/pb"
	"github.com/doublemo/balala/sss/service"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/etcdv3"
	"github.com/go-kit/kit/sd/lb"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// GRPCServer 内部响应GRPC服务
type GRPCServer struct {
	subscribe grpctransport.Handler
	broadcast grpctransport.Handler
	new       grpctransport.Handler
	remove    grpctransport.Handler
	params    grpctransport.Handler
}

// Subscribe 订阅
func (s *GRPCServer) Subscribe(stream pb.SessionStateServer_SubscribeServer) error {
	s.subscribe.ServeGRPC(context.Background(), stream)
	return nil
}

// Broadcast 广播
func (s *GRPCServer) Broadcast(ctx context.Context, in *pb.SessionStateServerAPI_BroadcastRequest) (*pb.SessionStateServerAPI_BroadcastResponse, error) {
	_, rep, err := s.broadcast.ServeGRPC(ctx, in)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.SessionStateServerAPI_BroadcastResponse), nil
}

// New 新状态
func (s *GRPCServer) New(ctx context.Context, in *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error) {
	_, rep, err := s.new.ServeGRPC(ctx, in)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.SessionStateServerAPI_Nil), nil
}

// Remove 删除
func (s *GRPCServer) Remove(ctx context.Context, in *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error) {
	_, rep, err := s.remove.ServeGRPC(ctx, in)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.SessionStateServerAPI_Nil), nil
}

// Params 参数修改
func (s *GRPCServer) Params(ctx context.Context, in *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error) {
	_, rep, err := s.params.ServeGRPC(ctx, in)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.SessionStateServerAPI_Nil), nil
}

// NewGRPCServer 创建内部服务grpc server
func NewGRPCServer(endpoints sssendpoint.Set, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) pb.SessionStateServerServer {
	options := []grpctransport.ServerOption{grpctransport.ServerErrorLogger(logger)}

	if zipkinTracer != nil {
		// Zipkin GRPC Server Trace can either be instantiated per gRPC method with a
		// provided operation name or a global tracing service can be instantiated
		// without an operation name and fed to each Go kit gRPC server as a
		// ServerOption.
		// In the latter case, the operation name will be the endpoint's grpc method
		// path if used in combination with the Go kit gRPC Interceptor.
		//
		// In this example, we demonstrate a global Zipkin tracing service with
		// Go kit gRPC Interceptor.
		options = append(options, zipkin.GRPCServerTrace(zipkinTracer))
	}

	return &GRPCServer{
		subscribe: grpctransport.NewServer(
			endpoints.SubscribeEndpoint,
			decodeGRPCRequest,
			encodeGRPCResponse,
			append(options, grpctransport.ServerBefore(jwt.GRPCToContext(), opentracing.GRPCToContext(otTracer, "Subscribe", logger)))...,
		),

		broadcast: grpctransport.NewServer(
			endpoints.BroadcastEndpoint,
			decodeGRPCRequest,
			encodeGRPCResponse,
			append(options, grpctransport.ServerBefore(jwt.GRPCToContext(), opentracing.GRPCToContext(otTracer, "Broadcast", logger)))...,
		),

		new: grpctransport.NewServer(
			endpoints.NewEndpoint,
			decodeGRPCRequest,
			encodeGRPCResponse,
			append(options, grpctransport.ServerBefore(jwt.GRPCToContext(), opentracing.GRPCToContext(otTracer, "New", logger)))...,
		),

		remove: grpctransport.NewServer(
			endpoints.RemoveEndpoint,
			decodeGRPCRequest,
			encodeGRPCResponse,
			append(options, grpctransport.ServerBefore(jwt.GRPCToContext(), opentracing.GRPCToContext(otTracer, "Remove", logger)))...,
		),

		params: grpctransport.NewServer(
			endpoints.ParamsEndpoint,
			decodeGRPCRequest,
			encodeGRPCResponse,
			append(options, grpctransport.ServerBefore(jwt.GRPCToContext(), opentracing.GRPCToContext(otTracer, "Params", logger)))...,
		),
	}
}

// NewGRPCClient 创建内部服务grpc client
func NewGRPCClient(conn *grpc.ClientConn, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, jwtToken []byte, logger log.Logger) service.GRPC {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second*1/1024), 102400))
	jwtEndpoint := jwt.NewSigner("kid-gate", jwtToken, jwtgo.SigningMethodHS256, jwt.StandardClaimsFactory())
	// global client middlewares
	var options []grpctransport.ClientOption
	if zipkinTracer != nil {
		options = append(options, zipkin.GRPCClientTrace(zipkinTracer))
	}

	var broadcastEndpoint endpoint.Endpoint
	{
		broadcastEndpoint = grpctransport.NewClient(conn,
			"pb.SessionStateServer",
			"Broadcast",
			encodeGRPRequest,
			decodeGRPCResponse,
			pb.SessionStateServerAPI_BroadcastResponse{},
			append(options, grpctransport.ClientBefore(jwt.ContextToGRPC()))...,
		).Endpoint()

		broadcastEndpoint = jwtEndpoint(broadcastEndpoint)
		broadcastEndpoint = opentracing.TraceClient(otTracer, "Broadcast")(broadcastEndpoint)
		broadcastEndpoint = limiter(broadcastEndpoint)
		broadcastEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Broadcast",
			Timeout: 30 * time.Second,
		}))(broadcastEndpoint)
	}

	var newEndpoint endpoint.Endpoint
	{
		broadcastEndpoint = grpctransport.NewClient(conn,
			"pb.SessionStateServer",
			"New",
			encodeGRPRequest,
			decodeGRPCResponse,
			pb.SessionStateServerAPI_Nil{},
			append(options, grpctransport.ClientBefore(jwt.ContextToGRPC()))...,
		).Endpoint()

		newEndpoint = jwtEndpoint(newEndpoint)
		newEndpoint = opentracing.TraceClient(otTracer, "New")(newEndpoint)
		newEndpoint = limiter(newEndpoint)
		newEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "New",
			Timeout: 30 * time.Second,
		}))(newEndpoint)
	}

	var removeEndpoint endpoint.Endpoint
	{
		removeEndpoint = grpctransport.NewClient(conn,
			"pb.SessionStateServer",
			"Remove",
			encodeGRPRequest,
			decodeGRPCResponse,
			pb.SessionStateServerAPI_Nil{},
			append(options, grpctransport.ClientBefore(jwt.ContextToGRPC()))...,
		).Endpoint()

		removeEndpoint = jwtEndpoint(removeEndpoint)
		removeEndpoint = opentracing.TraceClient(otTracer, "Remove")(removeEndpoint)
		removeEndpoint = limiter(removeEndpoint)
		removeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Remove",
			Timeout: 30 * time.Second,
		}))(removeEndpoint)
	}

	var paramsEndpoint endpoint.Endpoint
	{
		paramsEndpoint = grpctransport.NewClient(conn,
			"pb.SessionStateServer",
			"Params",
			encodeGRPRequest,
			decodeGRPCResponse,
			pb.SessionStateServerAPI_Nil{},
			append(options, grpctransport.ClientBefore(jwt.ContextToGRPC()))...,
		).Endpoint()

		paramsEndpoint = jwtEndpoint(paramsEndpoint)
		paramsEndpoint = opentracing.TraceClient(otTracer, "Params")(paramsEndpoint)
		paramsEndpoint = limiter(paramsEndpoint)
		paramsEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Params",
			Timeout: 30 * time.Second,
		}))(paramsEndpoint)
	}

	return sssendpoint.Set{
		BroadcastEndpoint: broadcastEndpoint,
		NewEndpoint:       newEndpoint,
		RemoveEndpoint:    removeEndpoint,
		ParamsEndpoint:    paramsEndpoint,
	}
}

// MakeInstancer 创建服务实例
func MakeInstancer(client etcdv3.Client, key string, logger log.Logger) (sd.Instancer, error) {
	return etcdv3.NewInstancer(client, key, logger)
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
	endpointer := sd.NewEndpointer(instancer, MakeFactorySubscribe(), logger)
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

// MakeFactorySubscribe Subscribe
func MakeFactorySubscribe() sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		return makeSubscribeEndpoint(conn), conn, nil
	}
}

func makeSubscribeEndpoint(conn *grpc.ClientConn) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (response interface{}, err error) {
		param, ok := req.(map[string]string)
		if !ok {
			return nil, errors.New("Invalid params")
		}

		cli := pb.NewSessionStateServerClient(conn)
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := cli.Subscribe(metadata.NewOutgoingContext(ctx, metadata.New(param)))
		if err != nil {
			cancel()
			return nil, err
		}

		return &DefaultGRPCStream{stream: stream, cancel: cancel}, nil
	}
}

// MakeFactoryBroadcast Broadcast
func MakeFactoryBroadcast(logger log.Logger, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, jwtToken []byte) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		value, err := services.RegValueFromString(instance)
		if err != nil {
			return nil, nil, err
		}

		conn, err := grpc.Dial(value.IP+":"+value.Port, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
		if err != nil {
			return nil, nil, err
		}

		s := NewGRPCClient(conn, otTracer, zipkinTracer, jwtToken, logger)
		doEndpoint := sssendpoint.MakeBroadcastEndpoint(s)
		return doEndpoint, conn, nil
	}
}

// MakeFactoryNew New
func MakeFactoryNew(logger log.Logger, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, jwtToken []byte) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		value, err := services.RegValueFromString(instance)
		if err != nil {
			return nil, nil, err
		}

		conn, err := grpc.Dial(value.IP+":"+value.Port, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
		if err != nil {
			return nil, nil, err
		}

		s := NewGRPCClient(conn, otTracer, zipkinTracer, jwtToken, logger)
		doEndpoint := sssendpoint.MakeNewEndpoint(s)
		return doEndpoint, conn, nil
	}
}

// MakeFactoryRemove Remove
func MakeFactoryRemove(logger log.Logger, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, jwtToken []byte) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		value, err := services.RegValueFromString(instance)
		if err != nil {
			return nil, nil, err
		}

		conn, err := grpc.Dial(value.IP+":"+value.Port, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
		if err != nil {
			return nil, nil, err
		}

		s := NewGRPCClient(conn, otTracer, zipkinTracer, jwtToken, logger)
		doEndpoint := sssendpoint.MakeRemoveEndpoint(s)
		return doEndpoint, conn, nil
	}
}

// MakeFactoryParams Params
func MakeFactoryParams(logger log.Logger, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, jwtToken []byte) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		value, err := services.RegValueFromString(instance)
		if err != nil {
			return nil, nil, err
		}

		conn, err := grpc.Dial(value.IP+":"+value.Port, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
		if err != nil {
			return nil, nil, err
		}

		s := NewGRPCClient(conn, otTracer, zipkinTracer, jwtToken, logger)
		doEndpoint := sssendpoint.MakeParamsEndpoint(s)
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
