// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"net"
	"time"

	"github.com/doublemo/balala/agent/session"
	"github.com/doublemo/balala/cores/networks"
	"github.com/doublemo/balala/cores/process"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
)

func makeSocketRuntimeActor(opts *Options, store *session.Store, logger log.Logger) *process.RuntimeActor {
	socketOpts := opts.Socket
	if socketOpts == nil {
		return nil
	}

	var socket networks.Socket
	{
		socket.CallBack(func(conn net.Conn, exit chan struct{}) {
			sess := store.NewClient(conn, "", time.Duration(socketOpts.ReadDeadline)*time.Second, time.Duration(socketOpts.WriteDeadline)*time.Second, 0)
			defer func() {
				store.RemoveAndExit(sess.Sid)
			}()
			//socketClient(sess, exit, logger)
		})
	}

	return &process.RuntimeActor{
		Exec: func() error {
			logger.Log("transport", "socket", "on", socketOpts.Addr)
			return socket.Serve(socketOpts.Addr, socketOpts.ReadBufferSize, socketOpts.WriteBufferSize)
		},
		Interrupt: func(err error) {
			if err != nil {
				kitlog.Error(logger).Log("transport", "socket", "error", err)
			}
		},

		Close: func() {
			logger.Log("transport", "socket", "on", "shutdown")
			socket.Shutdown()
		},
	}
}
