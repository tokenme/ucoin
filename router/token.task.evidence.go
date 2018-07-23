package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/handler/token/task/evidence"
)

func tokenTaskEvidenceRouter(r *gin.Engine) {
	tokenTaskEvidenceGroup := r.Group("/token/task/evidence")
	tokenTaskEvidenceGroup.Use(AuthMiddleware.MiddlewareFunc())
	{
		tokenTaskEvidenceGroup.POST("/create", evidence.CreateHandler)
		tokenTaskEvidenceGroup.GET("/list", evidence.ListHandler)
		tokenTaskEvidenceGroup.POST("/approve", evidence.ApproveHandler)
	}
}
