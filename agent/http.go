// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"github.com/doublemo/balala/agent/session"
	"github.com/gin-gonic/gin"
)

func httpInit(store *session.Store, r *gin.Engine) {
	gin.DisableConsoleColor()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
}

// webAuthenticationMiddleware http 身份验证插件
func webAuthenticationMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//token := ctx.GetHeader("X-Session-Token")
		// if c := r.SessionStore.Get(token); c == nil {
		// 	if r.Configuration.Runmode != "dev" {
		// 		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": foxchat.ErrorInvalidToken, "message": ""})
		// 		return
		// 	}
		// }
	}
}
