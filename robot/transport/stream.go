// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package transport

import (
	"github.com/doublemo/balala/cores/proto/pb"
	"golang.org/x/net/context"
)

// GRPCStramCallback 回调
type GRPCStramCallback func(*pb.Response) error

// GRPCStream grpc流连接
type GRPCStream interface {
	// Recv 接收
	Recv(GRPCStramCallback) error

	// Send 发送
	Send(*pb.Request) error

	// Close 关闭流连接
	Close()
}

type DefaultGRPCStream struct {
	stream pb.Internal_StreamClient
	cancel context.CancelFunc
}

func (g *DefaultGRPCStream) Recv(callback GRPCStramCallback) error {
	for {
		frame, err := g.stream.Recv()
		if err != nil {
			return err
		}

		if err := callback(frame); err != nil {
			return err
		}
	}
}

func (g *DefaultGRPCStream) Send(msg *pb.Request) error {
	return g.stream.Send(msg)
}

func (g *DefaultGRPCStream) Close() {
	if g.cancel != nil {
		g.cancel()
	}
}
