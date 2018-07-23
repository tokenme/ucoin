package evidence

import (
	//"github.com/davecgh/go-spew/spew"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	//"github.com/mkideal/log"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"net/http"
)

type ApproveRequest struct {
	EvidenceId    uint64 `form:"evidence_id" json:"evidence_id" binding:"required"`
	ApproveStatus int    `form:"approve_status" json:"approve_status" binding:"required"`
}

func ApproveHandler(c *gin.Context) {
	userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)
	var req ApproveRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}
	if Check(req.ApproveStatus != 1 && req.ApproveStatus != -1, "bad request", c) {
		return
	}
	db := Service.Db
	_, _, err := db.Query(`UPDATE ucoin.task_evidences AS te, ucoin.erc20 AS e SET te.approve_status=%d WHERE e.address=te.erc20 AND e.owner='%s' AND te.id=%d`, req.ApproveStatus, db.Escape(user.Wallet), req.EvidenceId)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	evidence := common.TokenTaskEvidence{
		Id:            req.EvidenceId,
		ApproveStatus: req.ApproveStatus,
	}

	c.JSON(http.StatusOK, evidence)
}
