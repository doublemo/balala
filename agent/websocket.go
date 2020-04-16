// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/doublemo/balala/agent/session"
	"github.com/doublemo/balala/cores/process"
	"github.com/doublemo/balala/cores/services"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
	"github.com/gorilla/websocket"
)

func makeWebsocketRuntimeActor(serviceOpts *services.Options, opts *Options, store *session.Store, logger log.Logger) *process.RuntimeActor {
	websocketOpts := opts.WebSocket
	if websocketOpts == nil {
		return nil
	}

	serviceOpts.Params["websocket"] = websocketOpts.Addr
	r := gin.New()
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
			ctx.AbortWithError(http.StatusNotFound, errors.New("404 Not found"))
			return
		}

		webscoketHandler(ctx.Writer, ctx.Request, webSocketUpgrader, store, websocketOpts, logger)
	})

	// http server
	s := &http.Server{
		Addr:           websocketOpts.Addr,
		Handler:        r,
		ReadTimeout:    time.Duration(websocketOpts.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(websocketOpts.WriteTimeout) * time.Second,
		MaxHeaderBytes: websocketOpts.MaxHeaderBytes,
	}

	return &process.RuntimeActor{
		Exec: func() error {
			logger.Log("transport", "websocket", "on", websocketOpts.Addr, "ssl", websocketOpts.SSL)
			if websocketOpts.SSL {
				return s.ListenAndServeTLS(websocketOpts.Cert, websocketOpts.Key)
			}

			return s.ListenAndServe()
		},
		Interrupt: func(err error) {
			if err != nil && err != http.ErrServerClosed {
				kitlog.Error(logger).Log("transport", "websocket", "error", err)
			}
		},

		Close: func() {
			logger.Log("transport", "websocket", "on", "shutdown")
			s.Shutdown(context.Background())
		},
	}
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
		store.RemoveAndExit(sess.ID())
		close(exit)
		conn.Close()
	}()

	socketLoop(sess, exit, websocketOpts.RPMLimit, logger)
}
