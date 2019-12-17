// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/doublemo/balala/agent/service"
	"github.com/doublemo/balala/agent/session"
	"github.com/doublemo/balala/cores/networks"
	"github.com/doublemo/balala/cores/process"
	"github.com/doublemo/balala/cores/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/sd/etcdv3"
)

func makeEtcdv3Client(s *Agent) {
	c := s.configureOptions.Read()
	if c.ETCD == nil {
		panic("ectd config is nil")
	}

	etcd := c.ETCD
	client, err := etcdv3.NewClient(context.Background(), etcd.Address, etcdv3.ClientOptions{
		CACert:        etcd.CACert,
		Cert:          etcd.Cert,
		Key:           etcd.Key,
		Username:      etcd.Username,
		Password:      etcd.Password,
		DialTimeout:   time.Duration(etcd.DialTimeout) * time.Second,
		DialKeepAlive: time.Duration(etcd.DialKeepAlive) * time.Second,
	})

	utils.Assert(err)
	s.etcdV3Client = client
}

func makeServices(s *Agent) *process.RuntimeActor {
	opts := s.configureOptions.Read()
	if opts.GRPC == nil {
		opts.GRPC = &GRPCOptions{Addr: ":9092"}
	}

	addr, port, err := net.SplitHostPort(opts.GRPC.Addr)
	utils.Assert(err)

	if addr == "" {
		addr = opts.LocalIP
		if ip := net.ParseIP(addr); ip == nil {
			panic("local IP is nil")
		}
	}

	addr = net.JoinHostPort(addr, port)
	// 注册服务
	registrar := etcdv3.NewRegistrar(s.etcdV3Client, etcdv3.Service{
		Key:   service.MakeKey(s.ServiceName(), opts.ETCD.Frefix, addr),
		Value: addr,
	}, s.logger)

	serviceChan := make(chan struct{})
	return &process.RuntimeActor{
		Exec: func() error {
			registrar.Register()
			close(s.readyedChan)
			for {
				select {
				case <-s.exitChan:
					return nil

				case <-serviceChan:
					return nil
				}
			}
		},
		Interrupt: func(err error) {},

		Close: func() {
			kitlog.Debug(s.logger).Log("Deregister", "Deregister")
			registrar.Deregister()
			close(serviceChan)
		},
	}
}

func makeSocket(opts *Options, store *session.Store, logger log.Logger) *process.RuntimeActor {
	socketOpts := opts.Socket
	if socketOpts == nil {
		return nil
	}

	var socket networks.Socket
	{
		socket.CallBack(func(conn net.Conn, exit chan struct{}) {
			sess := store.NewClient(conn, "", time.Duration(socketOpts.ReadDeadline)*time.Second, time.Duration(socketOpts.WriteDeadline)*time.Second, 0)
			defer func() {
				store.RemoveAndExit(sess.Sid)
			}()
			socketClient(sess, exit, logger)
		})
	}

	return &process.RuntimeActor{
		Exec: func() error {
			logger.Log("transport", "socket", "on", socketOpts.Addr)
			return socket.Serve(socketOpts.Addr, socketOpts.ReadBufferSize, socketOpts.WriteBufferSize)
		},
		Interrupt: func(err error) {
			if err != nil {
				kitlog.Error(logger).Log("transport", "socket", "error", err)
			}
		},

		Close: func() {
			logger.Log("transport", "socket", "on", "shutdown")
			socket.Shutdown()
		},
	}
}

func makeHTTP(opts *Options, store *session.Store, logger log.Logger) *process.RuntimeActor {
	gin.SetMode(gin.ReleaseMode)
	httpOpts := opts.HTTP
	if httpOpts == nil {
		return nil
	}

	if opts.Runmode == "dev" {
		gin.SetMode(gin.DebugMode)
	}

	handler := gin.New()
	httpInit(store, handler)
	s := &http.Server{
		Addr:           httpOpts.Addr,
		Handler:        handler,
		ReadTimeout:    time.Duration(httpOpts.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(httpOpts.WriteTimeout) * time.Second,
		MaxHeaderBytes: httpOpts.MaxHeaderBytes,
	}

	return &process.RuntimeActor{
		Exec: func() error {
			logger.Log("transport", "http", "on", httpOpts.Addr, "ssl", httpOpts.SSL)
			if httpOpts.SSL {
				return s.ListenAndServeTLS(httpOpts.Cert, httpOpts.Key)
			}

			return s.ListenAndServe()
		},
		Interrupt: func(err error) {
			if err != nil && err != http.ErrServerClosed {
				kitlog.Error(logger).Log("transport", "http", "error", err)
			}
		},

		Close: func() {
			logger.Log("transport", "http", "on", "shutdown")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.Shutdown(ctx)

			select {
			case <-ctx.Done():
				logger.Log("http", httpOpts.Addr, "error", "timeout of 5 seconds.")
			}
		},
	}
}

func makeWebsocket(opts *Options, store *session.Store, logger log.Logger) *process.RuntimeActor {
	gin.SetMode(gin.ReleaseMode)
	websocketOpts := opts.WebSocket
	if websocketOpts == nil {
		return nil
	}

	if opts.Runmode == "dev" {
		gin.SetMode(gin.DebugMode)
	}

	handler := gin.New()
	websocketInit(store, handler, websocketOpts, logger)
	s := &http.Server{
		Addr:           websocketOpts.Addr,
		Handler:        handler,
		ReadTimeout:    time.Duration(websocketOpts.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(websocketOpts.WriteTimeout) * time.Second,
		MaxHeaderBytes: websocketOpts.MaxHeaderBytes,
	}

	return &process.RuntimeActor{
		Exec: func() error {
			logger.Log("transport", "websocket", "on", websocketOpts.Addr, "ssl", websocketOpts.SSL)
			if websocketOpts.SSL {
				return s.ListenAndServeTLS(websocketOpts.Cert, websocketOpts.Key)
			}

			return s.ListenAndServe()
		},
		Interrupt: func(err error) {
			if err != nil && err != http.ErrServerClosed {
				kitlog.Error(logger).Log("transport", "websocket", "error", err)
			}
		},

		Close: func() {
			logger.Log("transport", "websocket", "on", "shutdown")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.Shutdown(ctx)

			select {
			case <-ctx.Done():
				logger.Log("websocket", websocketOpts.Addr, "error", "timeout of 5 seconds.")
			}
		},
	}
}
