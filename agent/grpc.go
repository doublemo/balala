// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/doublemo/balala/agent/endpoint"
	"github.com/doublemo/balala/agent/session"
	"github.com/doublemo/balala/agent/transport"
	"github.com/doublemo/balala/cores/process"
	"github.com/doublemo/balala/cores/proto/pb"
	"github.com/doublemo/balala/cores/utils"
	"github.com/go-kit/kit/auth/jwt"
	stdendpoint "github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
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

	return nil, nil
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

func makeGRPCRuntimeActor(opts *Options, store *session.Store, logger log.Logger) (*process.RuntimeActor, error) {
	grpcOpts := opts.GRPC
	if grpcOpts == nil {
		return nil, nil
	}

	var duration metrics.Histogram
	{
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: opts.ID,
			Subsystem: "agent",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}

	var counter metrics.Gauge
	{
		counter = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: opts.ID,
			Subsystem: "agent",
			Name:      "connect_to_counter",
			Help:      "Total count of client connect to the agent server.",
		}, []string{"method"})
	}

	var endpointExtends []stdendpoint.Middleware
	{
		endpointExtends = make([]stdendpoint.Middleware, 1)
		endpointExtends[0] = jwt.NewParser(makeKeyFuncByJWT(opts.ServiceSecurityKey), jwtgo.SigningMethodHS256, jwt.StandardClaimsFactory)
	}

	var (
		s          = newBaseGRPCServer(logger)
		endpoints  = endpoint.NewSet(s, logger, duration, counter, endpointExtends...)
		grpcServer = transport.NewGRPCServer(endpoints, logger)
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
			lis.Close()
		},
	}, nil
}

// makeKeyFuncByJWT GRPC
func makeKeyFuncByJWT(key string) func(token *jwtgo.Token) (interface{}, error) {
	return func(token *jwtgo.Token) (interface{}, error) {
		return key, nil
	}
}