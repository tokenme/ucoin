package qrcode

import (
	//"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"net/http"
)

type ParseRequest struct {
	Uri string `form:"uri" json:"uri" binding:"required"`
}

func ParseHandler(c *gin.Context) {
	/*userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)*/
	var req ParseRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}
	qrcode, err := common.ParseQrcode([]byte(Config.LinkSalt), req.Uri)
	if CheckErr(err, c) {
		return
	}
	c.JSON(http.StatusOK, qrcode)
}
