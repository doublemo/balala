// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"github.com/doublemo/balala/agent/session"
	"github.com/doublemo/balala/cores/proto"
)

// route 路由
func route(s *session.Client, frame []byte) ([]byte, error) {
	switch s.Proto {
	case proto.Socket, proto.Websocket:
		return socketHandler(s, frame)

	}
	return nil, nil
}

// socketHandler 处理socket协议
func socketHandler(s *session.Client, b []byte) ([]byte, error) {
	return nil, nil
}
