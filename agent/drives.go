// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"context"
	"net"
	"time"

	"github.com/doublemo/balala/agent/service"
	"github.com/doublemo/balala/cores/networks"
	"github.com/doublemo/balala/cores/process"
	"github.com/doublemo/balala/cores/utils"
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

func makeSocket(opts *Options, logger log.Logger) *process.RuntimeActor {
	socketOpts := opts.Socket
	if socketOpts == nil {
		return nil
	}

	var socket networks.Socket
	{
		socket.CallBack(func(conn net.Conn, exit chan struct{}) {
			socketClient()
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
