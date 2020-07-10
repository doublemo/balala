// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package proto 协议处理
package proto

import "errors"

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

// 错误信息定义
var (

	// ErrInvalidCommand 非法的命令
	ErrInvalidCommand = errors.New("ErrInvalidCommand")

	// ErrInvalidMetadata 非法的metadata信息
	ErrInvalidMetadata = errors.New("ErrInvalidMetadata")
)
