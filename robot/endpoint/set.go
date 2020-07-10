// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package endpoint go-kit endpoint
package endpoint

import (
	"context"
	"fmt"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/doublemo/balala/cores/proto/pb"
	"github.com/doublemo/balala/robot/service"
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
	// CallEndpoint 远程调用
	CallEndpoint endpoint.Endpoint

	// StreamEndpoint 流服务支持
	StreamEndpoint endpoint.Endpoint
}

// Call 内部远程调用
func (s Set) Call(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	resp, err := s.CallEndpoint(ctx, in)
	fmt.Println("000000----", err)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.Response), nil
}

// Stream 流服务支持
func (s Set) Stream(ctx context.Context, stream pb.Internal_StreamServer) error {
	_, err := s.StreamEndpoint(ctx, stream)
	if err != nil {
		return err
	}
	return nil
}

// MakeCallEndpoint 创建call
func MakeCallEndpoint(s service.GRPC) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.Request)
		return s.Call(ctx, req)
	}
}

// MakeStreamEndpoint 创建stream
func MakeStreamEndpoint(s service.GRPC) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(pb.Internal_StreamServer)
		err = s.Stream(ctx, req)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

// NewSet 创建内部通信节点
func NewSet(s service.GRPC, logger log.Logger, duration metrics.Histogram, counter metrics.Gauge, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, jwtToken jwtgo.Keyfunc) Set {

	jwtEndpoint := jwt.NewParser(jwtToken, jwtgo.SigningMethodHS256, jwt.StandardClaimsFactory)
	var callEndpoint endpoint.Endpoint
	{
		callEndpoint = MakeCallEndpoint(s)
		callEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second*1/1024), 102400))(callEndpoint)
		callEndpoint = jwtEndpoint(callEndpoint)
		callEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(callEndpoint)
		callEndpoint = opentracing.TraceServer(otTracer, "Call")(callEndpoint)

		if zipkinTracer != nil {
			callEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Call")(callEndpoint)
		}
		callEndpoint = LoggingMiddleware(log.With(logger, "method", "Call"))(callEndpoint)
		callEndpoint = InstrumentingMiddleware(duration.With("method", "Call"))(callEndpoint)

	}

	var streamEndpoint endpoint.Endpoint
	{
		streamEndpoint = MakeStreamEndpoint(s)
		streamEndpoint = CounterClientMiddleware(counter.With("method", "Stream"))(streamEndpoint)
		streamEndpoint = jwtEndpoint(streamEndpoint)
	}

	return Set{
		CallEndpoint:   callEndpoint,
		StreamEndpoint: streamEndpoint,
	}
}
