// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package endpoint go-kit endpoint
package endpoint

import (
	"context"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/doublemo/balala/sss/proto/pb"
	"github.com/doublemo/balala/sss/service"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

// Set 定义方法
type Set struct {
	// SubscribeEndpoint 订阅
	SubscribeEndpoint endpoint.Endpoint

	// BroadcastEndpoint 广播
	BroadcastEndpoint endpoint.Endpoint

	// NewEndpoint 创建新的连接状态
	NewEndpoint endpoint.Endpoint

	// RemoveEndpoint 删除连接状态
	RemoveEndpoint endpoint.Endpoint

	// ParamsEndpoint 修改连接参数
	ParamsEndpoint endpoint.Endpoint
}

// Subscribe 订阅
func (s Set) Subscribe(ctx context.Context, stream pb.SessionStateServer_SubscribeServer) error {
	_, err := s.SubscribeEndpoint(ctx, stream)
	if err != nil {
		return err
	}
	return nil
}

// Broadcast 广播
func (s Set) Broadcast(ctx context.Context, in *pb.SessionStateServerAPI_BroadcastRequest) (*pb.SessionStateServerAPI_BroadcastResponse, error) {
	resp, err := s.BroadcastEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.SessionStateServerAPI_BroadcastResponse), nil
}

// New 创建新的连接状态
func (s Set) New(ctx context.Context, in *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error) {
	resp, err := s.NewEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.SessionStateServerAPI_Nil), nil
}

// Remove 删除连接状态
func (s Set) Remove(ctx context.Context, in *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error) {
	resp, err := s.RemoveEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.SessionStateServerAPI_Nil), nil
}

// Params 修改连接参数
func (s Set) Params(ctx context.Context, in *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error) {
	resp, err := s.ParamsEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.SessionStateServerAPI_Nil), nil
}

// MakeSubscribeEndpoint Subscribe
func MakeSubscribeEndpoint(s service.GRPC) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(pb.SessionStateServer_SubscribeServer)
		err = s.Subscribe(ctx, req)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

// MakeBroadcastEndpoint Broadcast
func MakeBroadcastEndpoint(s service.GRPC) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.SessionStateServerAPI_BroadcastRequest)
		return s.Broadcast(ctx, req)
	}
}

// MakeNewEndpoint New
func MakeNewEndpoint(s service.GRPC) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.SessionStateServerAPI_NewRequest)
		return s.New(ctx, req)
	}
}

// MakeRemoveEndpoint Remove
func MakeRemoveEndpoint(s service.GRPC) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.SessionStateServerAPI_NewRequest)
		return s.Remove(ctx, req)
	}
}

// MakeParamsEndpoint Params
func MakeParamsEndpoint(s service.GRPC) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.SessionStateServerAPI_NewRequest)
		return s.Params(ctx, req)
	}
}

// NewSet 创建内部通信节点
func NewSet(s service.GRPC, logger log.Logger, duration metrics.Histogram, counter metrics.Gauge, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, jwtToken jwtgo.Keyfunc) Set {

	jwtEndpoint := jwt.NewParser(jwtToken, jwtgo.SigningMethodHS256, jwt.StandardClaimsFactory)
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second*1/1024), 102400))

	var broadcastEndpoint endpoint.Endpoint
	{
		broadcastEndpoint = MakeBroadcastEndpoint(s)
		broadcastEndpoint = limiter(broadcastEndpoint)
		broadcastEndpoint = jwtEndpoint(broadcastEndpoint)
		broadcastEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(broadcastEndpoint)
		broadcastEndpoint = opentracing.TraceServer(otTracer, "Broadcast")(broadcastEndpoint)

		if zipkinTracer != nil {
			broadcastEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Broadcast")(broadcastEndpoint)
		}
		broadcastEndpoint = LoggingMiddleware(log.With(logger, "method", "Broadcast"))(broadcastEndpoint)
		broadcastEndpoint = InstrumentingMiddleware(duration.With("method", "Broadcast"))(broadcastEndpoint)

	}

	var newEndpoint endpoint.Endpoint
	{
		newEndpoint = MakeNewEndpoint(s)
		newEndpoint = limiter(newEndpoint)
		newEndpoint = jwtEndpoint(newEndpoint)
		newEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(newEndpoint)
		newEndpoint = opentracing.TraceServer(otTracer, "New")(newEndpoint)

		if zipkinTracer != nil {
			newEndpoint = zipkin.TraceEndpoint(zipkinTracer, "New")(newEndpoint)
		}
		newEndpoint = LoggingMiddleware(log.With(logger, "method", "New"))(newEndpoint)
		newEndpoint = InstrumentingMiddleware(duration.With("method", "New"))(newEndpoint)

	}

	var removeEndpoint endpoint.Endpoint
	{
		removeEndpoint = MakeRemoveEndpoint(s)
		removeEndpoint = limiter(removeEndpoint)
		removeEndpoint = jwtEndpoint(removeEndpoint)
		removeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(removeEndpoint)
		removeEndpoint = opentracing.TraceServer(otTracer, "Remove")(removeEndpoint)

		if zipkinTracer != nil {
			removeEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Remove")(removeEndpoint)
		}
		removeEndpoint = LoggingMiddleware(log.With(logger, "method", "Remove"))(removeEndpoint)
		removeEndpoint = InstrumentingMiddleware(duration.With("method", "Remove"))(removeEndpoint)
	}

	var paramsEndpoint endpoint.Endpoint
	{
		paramsEndpoint = MakeRemoveEndpoint(s)
		paramsEndpoint = limiter(paramsEndpoint)
		paramsEndpoint = jwtEndpoint(paramsEndpoint)
		paramsEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(paramsEndpoint)
		paramsEndpoint = opentracing.TraceServer(otTracer, "Params")(paramsEndpoint)

		if zipkinTracer != nil {
			paramsEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Params")(paramsEndpoint)
		}
		paramsEndpoint = LoggingMiddleware(log.With(logger, "method", "Params"))(paramsEndpoint)
		paramsEndpoint = InstrumentingMiddleware(duration.With("method", "Params"))(paramsEndpoint)
	}

	var subscribeEndpoint endpoint.Endpoint
	{
		subscribeEndpoint = MakeSubscribeEndpoint(s)
		subscribeEndpoint = CounterClientMiddleware(counter.With("method", "Subscribe"))(subscribeEndpoint)
		subscribeEndpoint = jwtEndpoint(subscribeEndpoint)
	}

	return Set{
		SubscribeEndpoint: subscribeEndpoint,
		BroadcastEndpoint: broadcastEndpoint,
		NewEndpoint:       newEndpoint,
		RemoveEndpoint:    removeEndpoint,
		ParamsEndpoint:    paramsEndpoint,
	}
}
