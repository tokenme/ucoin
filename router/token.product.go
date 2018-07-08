package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/handler/token/product"
)

func tokenProductRouter(r *gin.Engine) {
	tokenProductGroup := r.Group("/token/product")
	tokenProductGroup.Use(AuthMiddleware.MiddlewareFunc())
	{
		tokenProductGroup.POST("/create", product.CreateHandler)
		tokenProductGroup.POST("/update", product.UpdateHandler)
	}
	r.GET("/token/product/list", AuthCheckerFunc(), product.ListHandler)
	r.GET("/token/product/info", AuthCheckerFunc(), product.InfoHandler)
}
