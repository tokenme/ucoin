package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/handler/user"
)

func userRouter(r *gin.Engine) {
	userGroup := r.Group("/user")
	userGroup.Use(AuthMiddleware.MiddlewareFunc())
	{
		userGroup.GET("/info", user.InfoGetHandler)
		userGroup.POST("/update", user.UpdateHandler)
	}
	r.POST("/user/create", user.CreateHandler)
	r.POST("/user/reset-password", user.ResetPasswordHandler)
	r.GET("/user/avatar/:key", user.AvatarGetHandler)
}
