package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/handler/order"
)

func orderRouter(r *gin.Engine) {
	orderGroup := r.Group("/order")
	orderGroup.Use(AuthMiddleware.MiddlewareFunc())
	{
		orderGroup.POST("/create", order.CreateHandler)
		orderGroup.GET("/info", order.InfoHandler)
	}
	r.GET("/order/list", AuthCheckerFunc(), order.ListHandler)
}
