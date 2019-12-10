// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// +build !windows

package services

// runService 启动服务
// 在除windows系统外的系统都采用这个方法启动
func runService(s Server) error {
	s.Start()
	return nil
}

// IsWindowsService 确认系统是否运行在windows服务系统上
func IsWindowsService() bool {
	return false
}
