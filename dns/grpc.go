// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package dns

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/doublemo/balala/cores/process"
	"github.com/doublemo/balala/cores/proto/pb"
	"github.com/doublemo/balala/cores/services"
	"github.com/doublemo/balala/cores/utils"
	"github.com/doublemo/balala/dns/endpoint"
	"github.com/doublemo/balala/dns/session"
	"github.com/doublemo/balala/dns/transport"
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

	// logger 日志
	logger log.Logger
}

func (s *baseGRPCServer) Call(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	defer utils.RecoverStackPanic(s.logger, in)

	return &pb.Response{Command: 1}, nil
}

func (s *baseGRPCServer) Stream(_ context.Context, stream pb.Internal_StreamServer) error {
	defer utils.RecoverStackPanic(s.logger)
	_, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("Invalid metadata")
	}

	ticker := time.NewTicker(time.Second)
	recvChan := make(chan *pb.Request, 4096)
	recvErr := make(chan error)
	go s.recv(stream, recvChan, recvErr)
	for {
		select {
		case frame, ok := <-recvChan:
			if !ok {
				return nil
			}

			s.logger.Log("frame", frame)

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

func (s *baseGRPCServer) recv(stream pb.Internal_StreamServer, recvChan chan *pb.Request, recvErr chan error) {
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
		logger: logger,
	}
}

func makeGRPCRuntimeActor(serviceOpts *services.Options, opts *Options, store *session.Store, logger log.Logger) (*process.RuntimeActor, error) {
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
			Subsystem: "dns",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}

	var counter metrics.Gauge
	{
		counter = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: opts.ID,
			Subsystem: "dns",
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
		zEP, _ := stdzipkin.NewEndpoint("dns", port)
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
			pb.RegisterInternalServer(baseServer, grpcServer)
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
