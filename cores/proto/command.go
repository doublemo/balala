// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package proto 协议处理
package proto

// Command  定义命令类型
type Command int16

// Int16 转换
func (c Command) Int16() int16 {
	return int16(c)
}

// 内部命令定义
const (
	InternalBad Command = 110
)
