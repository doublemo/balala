// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package service

import (
	"context"

	"github.com/doublemo/balala/cores/proto/pb"
)

// GRPC 内部通信接口
type GRPC interface {
	// Call 远程调用
	Call(context.Context, *pb.Request) (*pb.Response, error)

	// Stream 流服务
	Stream(context.Context, pb.Internal_StreamServer) error
}
