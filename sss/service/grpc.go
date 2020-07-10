// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package service

import (
	"context"

	"github.com/doublemo/balala/sss/proto/pb"
)

// GRPC 内部通信接口
type GRPC interface {
	// Subscribe 订阅
	Subscribe(context.Context, pb.SessionStateServer_SubscribeServer) error

	// Broadcast 广播
	Broadcast(context.Context, *pb.SessionStateServerAPI_BroadcastRequest) (*pb.SessionStateServerAPI_BroadcastResponse, error)

	// New 新状态
	New(context.Context, *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error)

	// Remove 删除
	Remove(context.Context, *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error)

	// Params 参数修改
	Params(context.Context, *pb.SessionStateServerAPI_NewRequest) (*pb.SessionStateServerAPI_Nil, error)
}
