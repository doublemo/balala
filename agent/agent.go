// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package agent 代理服务器
package agent

import (
	"fmt"
	"os"

	"github.com/doublemo/balala/cores/process"
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

	// logger
	logger log.Logger
}

// Start 启动服务
func (s *Agent) Start() {
	defer func() {
		close(s.exitChan)
		kitlog.Debug(s.logger).Log("Agent", "started")
	}()

	kitlog.Debug(s.logger).Log("Agent", "start")
	// init etcd
	makeEtcdv3Client(s)

	// 开始注册服务
	// 注意服务注册顺序就是服务的启动顺序
	// 关闭服务时会反顺关闭
	// socket
	s.process.Add(makeSocket(s.configureOptions.Read(), s.logger), true)

	// 创建服务
	s.process.Add(makeServices(s), true)
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

// ServiceName 返回服务名称
func (s *Agent) ServiceName() string {
	return "agent"
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
		logger:           logger,
	}
}
