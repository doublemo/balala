// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package dns

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/doublemo/balala/cores/process"
	"github.com/doublemo/balala/cores/services"
	"github.com/doublemo/balala/dns/session"
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

	// 代理
	r.GET("/px", ReverseProxy())
	r.POST("/px", ReverseProxy())

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

// ReverseProxy 实现反向代理
func ReverseProxy() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		target, err := url.Parse("http://localhost:9090/metrics")
		if err != nil {
			return
		}

		targetQuery := target.RawQuery
		director := func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
			req.URL.Path = "/metrics"
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		}

		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
