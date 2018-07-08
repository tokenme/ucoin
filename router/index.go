package router

import (
	"github.com/danielkov/gin-helmet"
	"github.com/dvwright/xss-mw"
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/router/static"
)

func NewRouter(staticPath string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(helmet.Default())
	xssMdlwr := &xss.XssMw{
		FieldsToSkip: []string{"password", "start_date", "end_date", "token"},
		BmPolicy:     "UGCPolicy",
	}
	r.Use(xssMdlwr.RemoveXss())
	r.Use(static.Serve("/", static.LocalFile(staticPath, 0, true)))
	r.LoadHTMLGlob("templates/contracts/erc20/*")
	authRouter(r)
	userRouter(r)
	tokenRouter(r)
	tokenProductRouter(r)
	tokenTaskRouter(r)
	orderRouter(r)
	qiniuRouter(r)
	return r
}
