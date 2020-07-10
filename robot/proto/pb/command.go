package pb

import (
	coreproto "github.com/doublemo/balala/cores/proto"
	corepb "github.com/doublemo/balala/cores/proto/pb"
)

// HandleFunc 接口处理
type HandleFunc func(*corepb.Request) (*corepb.Response, error)

const (
	// CommandCreate 创建机器人
	CommandCreate coreproto.Command = 30000

	// CommandRun 运行指定机器人
	CommandRun coreproto.Command = 30001
)
