// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package proto 协议处理
package proto

// Request 定义请求协议
// ----------------------------------------------------------------
// |  SID   | PageCount |  V   | Command | SubCommand |  Payload  |
// | uint32 |   int8    | int8 |  int16  |    int16   |   bytes   |
// ----------------------------------------------------------------
// 当PageCount大于1,说明需要进行分包处理
// 分包协议中将带page
// -----------------------------------------------------------------------
// |  SID   | PageCount | Page |  V   | Command | SubCommand |  Payload  |
// | uint32 |   int8    | int8 | int8 |  int16  |    int16   |   bytes   |
// -----------------------------------------------------------------------
// 分页开始后第一个包头是完整的
// 分页包
// ------------------------------------------
// |  SID   | PageCount | Page  |  Payload  |
// | uint32 |   int8    | int8  |   bytes   |
// ------------------------------------------
type Request interface {
	// V 版本号
	V() int8

	// SID 协议编号
	SID() uint32

	// Command 主命令
	Command() Command

	// SubCommand 子命令
	SubCommand() Command

	// Page 分包
	Page() int

	// PageCount 分包总数
	PageCount() int

	// Body 内容
	Body() []byte

	// Marshal 组包
	Marshal() ([]byte, error)

	// Unmarshal 解包
	Unmarshal([]byte) error
}
