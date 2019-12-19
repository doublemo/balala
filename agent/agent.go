// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package agent 代理服务器
package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/doublemo/balala/agent/service"
	"github.com/doublemo/balala/agent/session"
	"github.com/doublemo/balala/cores/process"
	"github.com/doublemo/balala/cores/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/sd/etcdv3"
)

// Agent 代理服务器
// 代理服务器支持通过服务接口
type Agent struct {
	// exitChan 退出信息
	exitChan chan struct{}

	// readyedChan 准备就绪信号
	readyedChan chan struct{}

	// configureOptions 配置文件
	configureOptions *ConfigureOptions

	// process 服务进程管理
	process *process.RuntimeContainer

	// etcdV3Client 连接实例,用于服务发现
	etcdV3Client etcdv3.Client

	// sessionStore session存储
	sessionStore *session.Store

	// servicesCaches 集群服务信息缓存
	servicesCaches map[int32]string

	// logger
	logger log.Logger
}

// Start 启动服务
func (s *Agent) Start() {
	defer func() {
		close(s.exitChan)
	}()

	// 读取一个配置文件副本
	opts := s.configureOptions.Read()

	// gin web framework
	gin.SetMode(gin.ReleaseMode)
	if opts.Runmode == "dev" {
		gin.SetMode(gin.DebugMode)
	}

	// Disable Console Color
	gin.DisableConsoleColor()

	// init etcd
	utils.Assert(s.makeEtcdv3Client())

	// 开始注册服务
	// 注意服务注册顺序就是服务的启动顺序
	// 关闭服务时会反顺关闭
	// internal grpc
	s.process.Add(s.mustRuntimeActor(makeGRPCRuntimeActor(s.configureOptions.Read(), s.sessionStore, s.logger)), true)

	// socket
	s.process.Add(makeSocketRuntimeActor(s.configureOptions.Read(), s.sessionStore, s.logger), true)

	// http
	s.process.Add(makeHTTPRuntimeActor(s.configureOptions.Read(), s.sessionStore, s.logger), true)

	// websocket
	s.process.Add(makeWebsocketRuntimeActor(s.configureOptions.Read(), s.sessionStore, s.logger), true)

	// 创建服务
	s.process.Add(s.mustRuntimeActor(s.makeServices()), true)
	s.process.Run()
}

// Readyed 返回服务准备就绪信号
func (s *Agent) Readyed() bool {
	<-s.readyedChan
	return true
}

// Shutdown 关闭服务
func (s *Agent) Shutdown() {
	kitlog.Debug(s.logger).Log("Agent", "shutdown")
	s.process.Stop()
}

// Reload 重新加载服务
func (s *Agent) Reload() {}

// ServiceName 返回唯一服务名称
func (s *Agent) ServiceName() string {
	return service.Name
}

// ServiceID 返回服务唯一服务编号
func (s *Agent) ServiceID() int32 {
	return service.ID
}

// OtherCommand 响应其他自定义命令
func (s *Agent) OtherCommand(cmd int) {}

// QuitCh 退出信息号
func (s *Agent) QuitCh() <-chan struct{} {
	return s.exitChan
}

// Fatalf Fatal信息处理
func (s *Agent) Fatalf(format string, args ...interface{}) {
	kitlog.Error(s.logger).Log("fatal", fmt.Sprintf(format, args...))
}

// Errorf Error信息处理
func (s *Agent) Errorf(format string, args ...interface{}) {
	kitlog.Error(s.logger).Log("error", fmt.Sprintf(format, args...))
}

// Warnf Warn信息处理
func (s *Agent) Warnf(format string, args ...interface{}) {
	kitlog.Warn(s.logger).Log("wran", fmt.Sprintf(format, args...))
}

// Debugf Debug信息处理
func (s *Agent) Debugf(format string, args ...interface{}) {
	kitlog.Debug(s.logger).Log("debug", fmt.Sprintf(format, args...))
}

// Tracef Trace信息处理
func (s *Agent) Tracef(format string, args ...interface{}) {
	kitlog.Info(s.logger).Log("trace", fmt.Sprintf(format, args...))
}

// Printf Print信息处理
func (s *Agent) Printf(format string, args ...interface{}) {
	kitlog.Info(s.logger).Log("info", fmt.Sprintf(format, args...))
}

func (s *Agent) makeEtcdv3Client() error {
	opts := s.configureOptions.Read()
	if opts.ETCD == nil {
		return errors.New("ETCD options is nil")
	}

	etcd := opts.ETCD
	client, err := etcdv3.NewClient(context.Background(), etcd.Address, etcdv3.ClientOptions{
		CACert:        etcd.CACert,
		Cert:          etcd.Cert,
		Key:           etcd.Key,
		Username:      etcd.Username,
		Password:      etcd.Password,
		DialTimeout:   time.Duration(etcd.DialTimeout) * time.Second,
		DialKeepAlive: time.Duration(etcd.DialKeepAlive) * time.Second,
	})

	if err != nil {
		return err
	}

	s.etcdV3Client = client
	return nil
}

func (s *Agent) makeServices() (*process.RuntimeActor, error) {
	opts := s.configureOptions.Read()
	if opts.ETCD == nil {
		return nil, errors.New("ETCD options is nil")
	}

	var value service.Value
	{
		value.ID = s.ServiceID()
		value.Name = s.ServiceName()
		value.LocalID = opts.LocalIP
		value.MachineID = opts.ID
		value.Frefix = opts.ETCD.Frefix
	}

	if opts.GRPC != nil {
		value.GRPCAddr = opts.GRPC.Addr
	}

	if opts.HTTP != nil {
		value.HTTPAddr = opts.HTTP.Addr
	}

	if opts.Socket != nil {
		value.SocketAddr = opts.Socket.Addr
	}

	if opts.WebSocket != nil {
		value.WebsocketAddr = opts.WebSocket.Addr
	}

	registrar := etcdv3.NewRegistrar(s.etcdV3Client, etcdv3.Service{
		Key:   value.Key(),
		Value: value.String(),
	}, s.logger)
	serviceChan := make(chan struct{})
	return &process.RuntimeActor{
		Exec: func() error {
			registrar.Register()
			close(s.readyedChan)
			ch := make(chan struct{})
			go s.etcdV3Client.WatchPrefix(opts.ETCD.Frefix, ch)
			for {
				select {
				case <-ch:
					instances, err := s.etcdV3Client.GetEntries(opts.ETCD.Frefix)
					fmt.Println("----------ss---", instances, err)
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
	}, nil
}

func (s *Agent) mustRuntimeActor(actor *process.RuntimeActor, err error) *process.RuntimeActor {
	if err != nil {
		kitlog.Error(s.logger).Log("error", err)
		panic(err)
	}

	return actor
}

// New 创建网关服务
func New(opts *ConfigureOptions) *Agent {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.WithPrefix(logger, "o", "[Balala]Agent")
	if opts.Read().Runmode == "dev" {
		logger = kitlog.NewFilter(logger, kitlog.AllowAll())
	} else {
		logger = kitlog.NewFilter(logger, kitlog.AllowError(), kitlog.AllowWarn())
	}

	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	return &Agent{
		exitChan:         make(chan struct{}),
		readyedChan:      make(chan struct{}),
		configureOptions: opts,
		process:          process.NewRuntimeContainer(),
		sessionStore:     session.NewStore(logger),
		logger:           logger,
	}
}
