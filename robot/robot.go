// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// robot 机器人系统
package robot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/doublemo/balala/cores/process"
	"github.com/doublemo/balala/cores/services"
	"github.com/doublemo/balala/cores/utils"
	"github.com/doublemo/balala/dns/service"
	"github.com/doublemo/balala/dns/session"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/sd/etcdv3"
)

// Robot 机器人服务
type Robot struct {
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

	// ServiceOpts 系统服务参数
	serviceOpts *services.Options

	// logger
	logger log.Logger
}

// Start 启动服务
func (s *Robot) Start() {
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

	// 创建服务
	s.process.Add(s.mustRuntimeActor(s.makeServices()), true)
	s.process.Run()
}

// Readyed 返回服务准备就绪信号
func (s *Robot) Readyed() bool {
	<-s.readyedChan
	return true
}

// Shutdown 关闭服务
func (s *Robot) Shutdown() {
	kitlog.Debug(s.logger).Log("Agent", "shutdown")
	s.process.Stop()
}

// Reload 重新加载服务
func (s *Robot) Reload() {}

// ServiceName 返回唯一服务名称
func (s *Robot) ServiceName() string {
	return service.Name
}

// ServiceID 返回服务唯一服务编号
func (s *Robot) ServiceID() int32 {
	return service.ID
}

// OtherCommand 响应其他自定义命令
func (s *Robot) OtherCommand(cmd int) {}

// QuitCh 退出信息号
func (s *Robot) QuitCh() <-chan struct{} {
	return s.exitChan
}

// Fatalf Fatal信息处理
func (s *Robot) Fatalf(format string, args ...interface{}) {
	kitlog.Error(s.logger).Log("fatal", fmt.Sprintf(format, args...))
}

// Errorf Error信息处理
func (s *Robot) Errorf(format string, args ...interface{}) {
	kitlog.Error(s.logger).Log("error", fmt.Sprintf(format, args...))
}

// Warnf Warn信息处理
func (s *Robot) Warnf(format string, args ...interface{}) {
	kitlog.Warn(s.logger).Log("wran", fmt.Sprintf(format, args...))
}

// Debugf Debug信息处理
func (s *Robot) Debugf(format string, args ...interface{}) {
	kitlog.Debug(s.logger).Log("debug", fmt.Sprintf(format, args...))
}

// Tracef Trace信息处理
func (s *Robot) Tracef(format string, args ...interface{}) {
	kitlog.Info(s.logger).Log("trace", fmt.Sprintf(format, args...))
}

// Printf Print信息处理
func (s *Robot) Printf(format string, args ...interface{}) {
	kitlog.Info(s.logger).Log("info", fmt.Sprintf(format, args...))
}

func (s *Robot) makeEtcdv3Client() error {
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

func (s *Robot) makeServices() (*process.RuntimeActor, error) {
	opts := s.configureOptions.Read()
	if opts.ETCD == nil {
		return nil, errors.New("ETCD options is nil")
	}

	registrar := etcdv3.NewRegistrar(s.etcdV3Client, etcdv3.Service{
		Key:   services.RegKey(opts.ETCD.Frefix, s.serviceOpts),
		Value: services.RegValue(s.serviceOpts),
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
					if err != nil {
						continue
					}

					service.Caches.Reset()
					for _, s := range instances {
						service.Caches.StoreFromString(s)
					}

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

func (s *Robot) mustRuntimeActor(actor *process.RuntimeActor, err error) *process.RuntimeActor {
	if err != nil {
		kitlog.Error(s.logger).Log("error", err)
		panic(err)
	}

	return actor
}

// New 创建网关服务
func New(serviceOpts *services.Options, opts *ConfigureOptions) *Robot {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.WithPrefix(logger, "o", "[BRob]")
	if opts.Read().Runmode == "dev" {
		logger = kitlog.NewFilter(logger, kitlog.AllowAll())
	} else {
		logger = kitlog.NewFilter(logger, kitlog.AllowError(), kitlog.AllowWarn())
	}

	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	return &Robot{
		exitChan:         make(chan struct{}),
		readyedChan:      make(chan struct{}),
		configureOptions: opts,
		process:          process.NewRuntimeContainer(),
		sessionStore:     session.NewStore(logger),
		logger:           logger,
		serviceOpts:      serviceOpts,
	}
}
