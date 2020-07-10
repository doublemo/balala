// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package service 服务
package service

import (
	"github.com/doublemo/balala/cores/services"
)

// 定义软件信息
const (
	// 软件运行时显示的服务名称
	Name string = "robot"
)

var Caches = services.NewCaches()
