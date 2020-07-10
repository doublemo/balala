package robot

import (
	corepb "github.com/doublemo/balala/cores/proto/pb"
	"github.com/doublemo/balala/robot/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

// createRobotV1 创建机器人
func createRobotV1(req *corepb.Request) (*corepb.Response, error) {
	var reqBody pb.ApiV1_CreateRequest
	if err := grpcproto.Unmarshal(req.GetBody(), &reqBody); err != nil {
		return nil, err
	}

	switch reqBody.GetRobotID() {
	case "A":
	case "B":
	}

	return nil, nil
}
