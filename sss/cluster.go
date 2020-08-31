// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package sss

import (
	"github.com/doublemo/balala/cores/services"
	"github.com/go-kit/kit/endpoint"
	"google.golang.org/grpc"
)

// cluster 集群处理
type cluster struct {
	endpoints     map[string]endpoint.Endpoint
	endpointChan  chan *services.Options
	broadcastChan chan string
}

func (cluster *cluster) serve() {
	for {
		select {
		case o, ok := <-cluster.endpointChan:
			if !ok {
				return
			}

			cluster.newEndpoint(o)

		case <-cluster.broadcastChan:
		}
	}
}

func (cluster *cluster) newEndpoint(o *services.Options) error {
	conn, err := grpc.Dial(instance, grpc.WithInsecure())
	return nil
}
