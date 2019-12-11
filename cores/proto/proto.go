// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package proto 协议处理
package proto

// Types 定义协议类型
type Types int

const (
	// None 无协议
	None Types = iota

	// Socket tcp socket 协议
	Socket

	// Websocket web socket
	Websocket

	// GRPC grpc
	GRPC

	// HTTP http
	HTTP
)
