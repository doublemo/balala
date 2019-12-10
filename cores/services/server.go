// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package services

// Command 定义服务命令
type Command string

const (
	// CommandStop 停止服务
	CommandStop Command = "stop"

	// CommandQuit 退出服务
	CommandQuit Command = "quit"

	// CommandReload 重载服务
	CommandReload Command = "reload"

	// CommandUSR1 自定义命令1
	CommandUSR1 Command = "usr1"

	// CommandUSR2 自定义命令2
	CommandUSR2 Command = "usr2"
)

// Server 服务接口
type Server interface {
	// Start 启动服务
	Start()

	// Readyed 服务是否已经准备就绪
	Readyed() bool

	// Shutdown 关闭服务
	Shutdown()

	// Reload 重新加载服务
	Reload()

	// ServiceName 返回服务名称
	ServiceName() string

	// OtherCommand 响应其他自定义命令
	OtherCommand(int)

	// QuitCh 退出信息号
	QuitCh() <-chan struct{}

	// Fatalf Fatal信息处理
	Fatalf(string, ...interface{})

	// Errorf Error信息处理
	Errorf(string, ...interface{})

	// Warnf Warn信息处理
	Warnf(string, ...interface{})

	// Debugf Debug信息处理
	Debugf(string, ...interface{})

	// Tracef Trace信息处理
	Tracef(string, ...interface{})

	// Tracef Print信息处理
	Printf(string, ...interface{})
}
