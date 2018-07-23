package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/handler/qrcode"
)

func qrcodeRouter(r *gin.Engine) {
	qrcodeGroup := r.Group("/qrcode")
	qrcodeGroup.Use(AuthMiddleware.MiddlewareFunc())
	{
		qrcodeGroup.GET("/collect", qrcode.CollectHandler)
	}

	r.GET("/qrcode/parse", AuthCheckerFunc(), qrcode.ParseHandler)
}
