package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/handler/token/task"
)

func tokenTaskRouter(r *gin.Engine) {
	tokenTaskGroup := r.Group("/token/task")
	tokenTaskGroup.Use(AuthMiddleware.MiddlewareFunc())
	{
		tokenTaskGroup.POST("/create", task.CreateHandler)
		tokenTaskGroup.POST("/update", task.UpdateHandler)
	}
	r.GET("/token/task/list", AuthCheckerFunc(), task.ListHandler)
	r.GET("/token/task/info", AuthCheckerFunc(), task.InfoHandler)
}
