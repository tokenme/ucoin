package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/handler/qiniu"
)

func qiniuRouter(r *gin.Engine) {
	qiniuGroup := r.Group("/qiniu")
	qiniuGroup.Use(AuthMiddleware.MiddlewareFunc())
	{
		qiniuGroup.POST("/token/product", qiniu.TokenProductHandler)
		qiniuGroup.POST("/token/task", qiniu.TokenTaskHandler)
		qiniuGroup.POST("/token/logo", qiniu.TokenLogoHandler)
	}
}
