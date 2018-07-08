package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/handler/token"
)

func tokenRouter(r *gin.Engine) {
	tokenGroup := r.Group("/token")
	tokenGroup.Use(AuthMiddleware.MiddlewareFunc())
	{
		tokenGroup.POST("/create", token.CreateHandler)
		tokenGroup.POST("/update", token.UpdateHandler)
		tokenGroup.GET("/owned/list", token.OwnedListHandler)
	}

	r.GET("/token/info/:address", AuthCheckerFunc(), token.InfoHandler)
}
