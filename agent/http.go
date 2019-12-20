// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"context"
	"net/http"
	"time"

	"github.com/doublemo/balala/agent/session"
	"github.com/doublemo/balala/cores/process"
	"github.com/doublemo/balala/cores/services"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func makeHTTPRuntimeActor(serviceOpts *services.Options, opts *Options, store *session.Store, logger log.Logger) *process.RuntimeActor {
	httpOpts := opts.HTTP
	if httpOpts == nil {
		return nil
	}

	serviceOpts.Params["http"] = httpOpts.Addr
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 定义命令路由
	if opts.Runmode == "dev" {
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// 启动http服务
	s := &http.Server{
		Addr:           httpOpts.Addr,
		Handler:        r,
		ReadTimeout:    time.Duration(httpOpts.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(httpOpts.WriteTimeout) * time.Second,
		MaxHeaderBytes: httpOpts.MaxHeaderBytes,
	}

	return &process.RuntimeActor{
		Exec: func() error {
			logger.Log("transport", "http", "on", httpOpts.Addr, "ssl", httpOpts.SSL)
			if httpOpts.SSL {
				return s.ListenAndServeTLS(httpOpts.Cert, httpOpts.Key)
			}

			return s.ListenAndServe()
		},
		Interrupt: func(err error) {
			if err != nil && err != http.ErrServerClosed {
				kitlog.Error(logger).Log("transport", "http", "error", err)
			}
		},

		Close: func() {
			logger.Log("transport", "http", "on", "shutdown")
			s.Shutdown(context.Background())
		},
	}
}
