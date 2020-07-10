package robot

import (
	coreproto "github.com/doublemo/balala/cores/proto"
	"github.com/doublemo/balala/robot/proto/pb"
)

var apiHandlers = make(map[int32]map[coreproto.Command]pb.HandleFunc)

// makeRoutes se
func makeRoutes() {
	// 接口版本号1的接口
	apiHandlers[1] = make(map[coreproto.Command]pb.HandleFunc)
	apiHandlers[1][pb.CommandCreate] = createRobotV1
}
