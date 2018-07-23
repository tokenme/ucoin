package qiniu

import (
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/nu7hatch/gouuid"
	. "github.com/tokenme/ucoin/handler"
	"github.com/tokenme/ucoin/tools/qiniu"
	commonutils "github.com/tokenme/ucoin/utils"
	"net/http"
	"strconv"
	"time"
)

type TokenTaskEvidenceRequest struct {
	TaskId uint64 `form:"task_id" json:"task_id" binding:"required"`
	Amount uint   `form:"amount" json:"amount" binding:"required"`
}

func TokenTaskEvidenceHandler(c *gin.Context) {
	_, exists := c.Get("USER")
	if Check(!exists, "need login", c) {
		return
	}
	var req TokenTaskEvidenceRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}

	db := Service.Db
	rows, _, err := db.Query(`SELECT t.id FROM ucoin.tasks AS t WHERE t.id=%d LIMIT 1`, req.TaskId)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	if Check(len(rows) == 0, "not found", c) {
		return
	}

	var (
		i        uint
		upTokens []gin.H
	)
	imageId, err := uuid.NewV4()
	if CheckErr(err, c) {
		return
	}
	imageIdStr := commonutils.Sha1(imageId.String())
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	for i < req.Amount {
		upToken, key, link := qiniu.UpToken(Config.Qiniu, fmt.Sprintf("%s/%d/%s/%d", Config.Qiniu.TokenTaskEvidencePath, req.TaskId, imageIdStr, i), timestamp)
		upTokens = append(upTokens, gin.H{"uptoken": upToken, "key": key, "link": link, "index": i})
		i += 1
	}
	c.JSON(http.StatusOK, upTokens)
}
