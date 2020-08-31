// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package sss

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/doublemo/balala/cores/process"
	"github.com/doublemo/balala/cores/services"
	"github.com/doublemo/balala/cores/utils"
	"github.com/doublemo/balala/sss/endpoint"
	"github.com/doublemo/balala/sss/proto/pb"
	"github.com/doublemo/balala/sss/session"
	"github.com/doublemo/balala/sss/transport"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	stdzipkin "github.com/openzipkin/zipkin-go"
	zipkinreporter "github.com/openzipkin/zipkin-go/reporter"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// baseGRPCServer 服务于内部通信的grpc
type baseGRPCServer struct {
	// subscribes 订阅信息
	subscribes *session.SubscribeStore

	// 客户端信息存储
	sessionStore *session.Store

	// 集群中其它sss服务连接信息

	// logger 日志
	logger log.Logger
}

// Broadcast 广播
func (s *baseGRPCServer) Broadcast(ctx context.Context, in *pb.SessionStateServerAPI_BroadcastRequest) (*pb.SessionStateServerAPI_BroadcastResponse, error) {
	defer utils.RecoverStackPanic(s.logger, in)

	return &pb.SessionStateServerAPI_BroadcastResponse{}, nil
}

// New 新状态
func (s *baseGRPCServer) New(ctx context.Context, in *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error) {
	defer utils.RecoverStackPanic(s.logger, in)
	if in.GetClientID() == "" {
		return nil, nil
	}

	client := s.sessionStore.NewClient(in.GetClientID())
	storeParams := make([]*session.Param, len(in.Params))
	for i, v := range in.Params {
		storeParams[i] = &session.Param{Key: v.GetKey(), Value: v.GetValue()}
	}

	client.SetParam(storeParams...)

	// 通知集群其它服务
	return &pb.SessionStateServerAPI_Nil{}, nil
}

// Remove 删除
func (s *baseGRPCServer) Remove(ctx context.Context, in *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error) {
	defer utils.RecoverStackPanic(s.logger, in)

	if in.GetClientID() == "" {
		return nil, nil
	}

	s.sessionStore.Remove(in.GetClientID())
	return &pb.SessionStateServerAPI_Nil{}, nil
}

// Params 参数修改
func (s *baseGRPCServer) Params(ctx context.Context, in *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error) {
	defer utils.RecoverStackPanic(s.logger, in)

	if in.GetClientID() == "" {
		return nil, nil
	}

	client := s.sessionStore.Get(in.GetClientID())
	if client == nil {
		return nil, nil
	}

	return &pb.SessionStateServerAPI_Nil{}, nil
}

func (s *baseGRPCServer) Subscribe(_ context.Context, stream pb.SessionStateServer_SubscribeServer) error {
	defer utils.RecoverStackPanic(s.logger)
	metadata, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("Invalid metadata")
	}

	id, ok := metadata["id"]
	if !ok || len(id) < 1 {
		return errors.New("Invalid id")
	}

	serviceID, ok := metadata["serviceID"]
	if !ok || len(serviceID) < 1 {
		return errors.New("Invalid serviceID")
	}

	sid, err := strconv.Atoi(serviceID[0])
	if err != nil {
		return err
	}

	events := make([]int32, 0)
	if m, ok := metadata["events"]; ok {
		for _, eid := range m {
			evid, err := strconv.Atoi(eid)
			if err != nil {
				return err
			}

			events = append(events, int32(evid))
		}
	}

	if len(events) < 1 {
		return errors.New("Invalid events")
	}
	subscriber, err := s.subscribes.NewSubscriber(id[0], int32(sid), events)
	if err != nil {
		return err
	}

	defer func() {
		s.subscribes.RemoveByServiceIDAndID(subscriber.GetID(), subscriber.GetServiceID())
	}()

	ticker := time.NewTicker(time.Second)
	recvChan := make(chan *pb.SessionStateServerAPI_Nil, 4096)
	recvErr := make(chan error)
	go s.recv(stream, recvChan, recvErr)
	for {
		select {
		case frame, ok := <-recvChan:
			if !ok {
				return nil
			}
			s.logger.Log("frame", frame)

		case frame, ok := <-subscriber.GetRecv():
			if !ok {
				return nil
			}

			stream.Send(frame)

		case err, ok := <-recvErr:
			if !ok {
				return nil
			}

			s.logger.Log("error", err)
			return nil

		case <-ticker.C:
		}
	}
}

func (s *baseGRPCServer) recv(stream pb.SessionStateServer_SubscribeServer, recvChan chan *pb.SessionStateServerAPI_Nil, recvErr chan error) {
	defer func() {
		close(recvChan)
		close(recvErr)
	}()

	for {
		frame, err := stream.Recv()
		if err == io.EOF {
			continue
		}

		if err != nil {
			recvErr <- err
			return
		}

		recvChan <- frame
	}
}

func newBaseGRPCServer(logger log.Logger) *baseGRPCServer {
	return &baseGRPCServer{
		subscribes:   session.NewSubscribeStore(),
		sessionStore: session.NewStore(),
		logger:       logger,
	}
}

func makeGRPCRuntimeActor(serviceOpts *services.Options, opts *Options, logger log.Logger) (*process.RuntimeActor, error) {
	grpcOpts := opts.GRPC
	if grpcOpts == nil {
		return nil, nil
	}

	_, port, err := net.SplitHostPort(grpcOpts.Addr)
	if err != nil {
		return nil, err
	}

	serviceOpts.Port = port
	var duration metrics.Histogram
	{
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: opts.ID,
			Subsystem: "sss",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}

	var counter metrics.Gauge
	{
		counter = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: opts.ID,
			Subsystem: "sss",
			Name:      "connect_to_counter",
			Help:      "Total count of client connect to the dns server.",
		}, []string{"method"})
	}

	tracer := stdopentracing.GlobalTracer() // no-op

	var reporter zipkinreporter.Reporter
	zipkinTracer, err := stdzipkin.NewTracer(nil, stdzipkin.WithNoopTracer(true))
	if opts.Tracer != nil && opts.Tracer.ReporterURL != "" {

		// some http://192.168.31.20:9411/api/v2/spans
		reporter = zipkinhttp.NewReporter(opts.Tracer.ReporterURL)
		zEP, _ := stdzipkin.NewEndpoint("sss", port)
		zipkinTracer, err = stdzipkin.NewTracer(reporter, stdzipkin.WithLocalEndpoint(zEP))
	}

	if err != nil {
		return nil, err
	}

	tracer = zipkinot.Wrap(zipkinTracer)
	var (
		s          = newBaseGRPCServer(logger)
		endpoints  = endpoint.NewSet(s, logger, duration, counter, tracer, zipkinTracer, makeKeyFuncByJWT(opts.ServiceSecurityKey))
		grpcServer = transport.NewGRPCServer(endpoints, tracer, zipkinTracer, logger)
	)

	lis, err := net.Listen("tcp", grpcOpts.Addr)
	if err != nil {
		return nil, err
	}

	return &process.RuntimeActor{
		Exec: func() error {
			logger.Log("transport", "grpc", "on", grpcOpts.Addr)
			baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor), grpc.MaxConcurrentStreams(65535))
			pb.RegisterSessionStateServerServer(baseServer, grpcServer)
			return baseServer.Serve(lis)
		},
		Interrupt: func(err error) {
			if err != nil && err != http.ErrServerClosed {
				kitlog.Error(logger).Log("transport", "grpc", "error", err)
			}
		},

		Close: func() {
			logger.Log("transport", "grpc", "on", "shutdown")
			if reporter != nil {
				reporter.Close()
			}

			lis.Close()
		},
	}, nil
}

// makeKeyFuncByJWT GRPC
func makeKeyFuncByJWT(key string) func(token *jwtgo.Token) (interface{}, error) {
	return func(token *jwtgo.Token) (interface{}, error) {
		return []byte(key), nil
	}
}
