// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// +build !windows

package main

// installService 安装服务
func installService(serviceName, dname, description, args string) {
	panic("This method is not supported on this platform")
}

// uninstallService 卸载服务
func uninstallService(serviceName string) {
	panic("This method is not supported on this platform")
}
