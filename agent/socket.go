// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"fmt"
	"runtime"
	"time"

	"github.com/doublemo/balala/agent/session"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
)

func socketClient(sess *session.Client, exit chan struct{}, logger log.Logger) {
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
		kitlog.Debug(logger).Log("offline", sess.Sid)
	}()

	kitlog.Debug(logger).Log("online", sess.Sid)
	for {
		select {
		case frame, ok := <-sess.GetRecvChan():
			if !ok {
				return
			}

			sess.PacketCounter++
			resp, err := sess.Call(frame, func(s *session.Client, frame []byte) ([]byte, error) {
				return route(s, frame)
			})

			if err != nil {
				kitlog.Error(logger).Log("error", err, "frame", fmt.Sprintf("%v", frame))
				return
			}

			sess.Send(resp)

		case <-ticker.C:
			// 如果客户端连接上来,在出现以下情况将直接踢掉客户端
			// 5秒内没有和服务端握手
			// 1分钟之内没有进行登录
			flag := sess.Flag()
			if flag&session.FlagAuthorized != 0 {
				ticker.Stop()
				continue
			}

			date := time.Now().Sub(sess.CreateAt)
			diff := date.Seconds()
			if diff > 5 && flag&session.FlagEncrypt == 0 {
				//return
			} else if diff > 60 {
				//return
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
