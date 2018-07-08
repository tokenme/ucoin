package token

import (
	//"github.com/davecgh/go-spew/spew"
	"bufio"
	"bytes"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"github.com/tokenme/ucoin/tools/qiniu"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type UpdateRequest struct {
	Address     string `form:"address" json:"address" binding:"required"`
	Description string `form:"description" json:"description"`
}

func UpdateHandler(c *gin.Context) {
	var req UpdateRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}
	userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)
	db := Service.Db

	token := common.Token{
		Address: req.Address,
		Owner:   user.Wallet,
	}

	var updateFields []string
	if req.Description != "" {
		description := strings.TrimSpace(req.Description)
		updateFields = append(updateFields, fmt.Sprintf("summary='%s'", db.Escape(description)))
		token.Description = description
	}

	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	file, _, err := c.Request.FormFile("logo")
	if err == nil {
		buf := new(bytes.Buffer)
		var w = bufio.NewWriter(buf)
		w.ReadFrom(file)
		w.Flush()
		logoURL, _, err := qiniu.Upload(c, Config.Qiniu, fmt.Sprintf("%s/%s", Config.Qiniu.LogoPath, req.Address), timestamp, buf.Bytes())
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		updateFields = append(updateFields, fmt.Sprintf("logo='%s'", db.Escape(logoURL)))
		token.Logo = logoURL
	}

	file, _, err = c.Request.FormFile("cover")
	if err == nil {
		buf := new(bytes.Buffer)
		var w = bufio.NewWriter(buf)
		w.ReadFrom(file)
		w.Flush()
		coverURL, _, err := qiniu.Upload(c, Config.Qiniu, fmt.Sprintf("%s/%s", Config.Qiniu.TokenCover, req.Address), timestamp, buf.Bytes())
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		updateFields = append(updateFields, fmt.Sprintf("cover='%s'", db.Escape(coverURL)))
		token.Cover = coverURL
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusOK, token)
		return
	}

	_, _, err = db.Query(`UPDATE ucoin.erc20 SET %s WHERE address='%s' AND owner='%s' LIMIT 1`, strings.Join(updateFields, ","), db.Escape(req.Address), db.Escape(user.Wallet))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	c.JSON(http.StatusOK, token)
}
