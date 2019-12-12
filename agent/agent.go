// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package agent 代理服务器
package agent

import "log"

// Agent 代理服务器
// 代理服务器支持通过服务接口
type Agent struct {
	// configureOptions 配置文件
	configureOptions *ConfigureOptions
}

// Start 启动服务
func (s *Agent) Start() {
	log.Println("START Agent")
}

// Readyed 返回服务准备就绪信号
func (s *Agent) Readyed() bool {
	return true
}

// Shutdown 关闭服务
func (s *Agent) Shutdown() {
	log.Println("Shutdown Agent")
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
	return nil
}

// Fatalf Fatal信息处理
func (s *Agent) Fatalf(string, ...interface{}) {}

// Errorf Error信息处理
func (s *Agent) Errorf(string, ...interface{}) {}

// Warnf Warn信息处理
func (s *Agent) Warnf(string, ...interface{}) {}

// Debugf Debug信息处理
func (s *Agent) Debugf(string, ...interface{}) {}

// Tracef Trace信息处理
func (s *Agent) Tracef(string, ...interface{}) {}

// Printf Print信息处理
func (s *Agent) Printf(string, ...interface{}) {}

// New 创建网关服务
func New(opts *ConfigureOptions) *Agent {
	return &Agent{configureOptions: opts}
}
