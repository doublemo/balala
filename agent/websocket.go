// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"errors"
	"net/http"
	"time"

	"github.com/doublemo/balala/agent/session"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
	"github.com/gorilla/websocket"
)

func websocketInit(store *session.Store, r *gin.Engine, websocketOpts *WebSocketOptions, logger log.Logger) {
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	var webSocketUpgrader websocket.Upgrader
	{
		webSocketUpgrader = websocket.Upgrader{
			ReadBufferSize:  websocketOpts.ReadBufferSize,
			WriteBufferSize: websocketOpts.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	}

	// webscoket
	r.GET("/websocket", func(ctx *gin.Context) {
		if !ctx.IsWebsocket() {
			ctx.AbortWithError(http.StatusNotFound, errors.New("404 Not found."))
			return
		}

		webscoketHandler(ctx.Writer, ctx.Request, webSocketUpgrader, store, websocketOpts, logger)
	})
}

// webscoketHandler WebSocket 处理
func webscoketHandler(w http.ResponseWriter, req *http.Request, upgrader websocket.Upgrader, store *session.Store, websocketOpts *WebSocketOptions, logger log.Logger) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		kitlog.Error(logger).Log("error", err)
		return
	}

	exit := make(chan struct{})
	sess := store.NewClient(conn, "", time.Duration(websocketOpts.ReadDeadline)*time.Second, time.Duration(websocketOpts.WriteDeadline)*time.Second, websocketOpts.MaxMessageSize)
	defer func() {
		// 删除session 并关闭相服务
		store.RemoveAndExit(sess.Sid)
		close(exit)
		conn.Close()
	}()

	socketClient(sess, exit, logger)
}
