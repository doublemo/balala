// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/doublemo/balala/agent/session"
	"github.com/doublemo/balala/cores/networks"
	"github.com/doublemo/balala/cores/process"
	"github.com/doublemo/balala/cores/services"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
)

func makeSocketRuntimeActor(serviceOpts *services.Options, opts *Options, store *session.Store, logger log.Logger) *process.RuntimeActor {
	socketOpts := opts.Socket
	if socketOpts == nil {
		return nil
	}

	serviceOpts.Params["socket"] = socketOpts.Addr
	var socket networks.Socket
	{
		socket.CallBack(func(conn net.Conn, exit chan struct{}) {
			sess := store.NewClient(conn, "", time.Duration(socketOpts.ReadDeadline)*time.Second, time.Duration(socketOpts.WriteDeadline)*time.Second, 0)
			defer func() {
				store.RemoveAndExit(sess.ID())
			}()

			socketLoop(sess, exit, socketOpts.RPMLimit, logger)
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

func socketLoop(sess *session.Client, exit chan struct{}, rpmLimit int, logger log.Logger) {
	defer func() {
		if r := recover(); r != nil {
			kitlog.Error(logger).Log("panic", fmt.Sprint(r))
			i := 0
			funcName, file, line, ok := runtime.Caller(i)
			for ok {
				kitlog.Error(logger).Log("panic", fmt.Sprintf("frame %v:[func:%v,file:%v,line:%v]", i, runtime.FuncForPC(funcName).Name(), file, line))
				i++
				funcName, file, line, ok = runtime.Caller(i)
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
		kitlog.Info(logger).Log("offline", sess.ID())
	}()

	createAt := time.Now()
	packetCounter := 0
	rpm1Min := 0
	kitlog.Info(logger).Log("online", sess.ID())
	for {
		select {
		case frame, ok := <-sess.GetRecvChan():
			if !ok {
				return
			}

			packetCounter++
			kitlog.Info(logger).Log("frame", frame)

		case <-ticker.C:
			rpm1Min++
			flag := sess.Flag()
			if flag&session.FlagAuthorized != 0 {
				if rpm1Min >= 60 {
					// 如果在1分钟内超过了200个包,踢掉
					if packetCounter > rpmLimit {
						return
					}

					rpm1Min = 0
					packetCounter = 0
				}
			} else {
				date := time.Now().Sub(createAt)
				diff := date.Seconds()
				if diff > 5 && flag&session.FlagEncrypt == 0 {
					return
				} else if diff > 60 {
					return
				}
			}

		case <-sess.GetRecvExitChan():
			return

		case <-sess.GetSendExitChan():
			return

		case <-exit:
			return
		}
	}
}
