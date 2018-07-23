package evidence

import (
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	//"github.com/mkideal/log"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"math/big"
	"net/http"
	"strings"
	"time"
)

const DefaultPageSize uint = 10

type ListRequest struct {
	TaskId        uint64 `form:"task_id" json:"task_id" binding:"required"`
	ApproveStatus int    `form:"approve_status" json:"approve_status"`
	Page          uint   `form:"page" json:"page"`
	PageSize      uint   `form:"page_size" json:"page_size"`
}

func ListHandler(c *gin.Context) {
	userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)
	var req ListRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}
	var page uint
	if req.Page > 0 {
		page = req.Page
	}
	var pageSize = DefaultPageSize
	if req.PageSize > 0 && req.PageSize < DefaultPageSize {
		pageSize = req.PageSize
	}

	db := Service.Db
	var (
		where  string
		wheres []string
	)
	wheres = append(wheres, fmt.Sprintf("t.owner='%s'", db.Escape(user.Wallet)))
	wheres = append(wheres, fmt.Sprintf("te.task_id=%d", req.TaskId))
	wheres = append(wheres, fmt.Sprintf("te.approve_status=%d", req.ApproveStatus))

	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	rows, _, err := db.Query(`SELECT
    erc20.address, 
    erc20.decimals,
    t.id, 
    u.wallet_addr,
    u.nickname,
    u.avatar,
    te.id,
    te.bonus,
    te.summary,
    te.images,
    te.inserted_at,
    te.updated_at,
    te.tx,
    IFNULL(txs.status, 0),
    te.approve_status
FROM ucoin.task_evidences AS te
INNER JOIN ucoin.tasks AS t ON (t.id = te.id)
INNER JOIN ucoin.users AS u ON (u.wallet_addr = te.submitter)
INNER JOIN ucoin.erc20 AS erc20 ON (erc20.address = te.erc20)
LEFT JOIN ucoin.txs AS txs ON (txs.tx = te.tx)
%s
ORDER BY te.id DESC LIMIT %d, %d`, where, page*pageSize, pageSize)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	var evidences []common.TokenTaskEvidence
	for _, row := range rows {
		token := &common.Token{
			Address:  row.Str(0),
			Decimals: row.Uint(1),
		}
		task := common.TokenTask{
			Id:    row.Uint64(2),
			Token: token,
		}
		u := common.User{
			Wallet: row.Str(3),
			Nick:   row.Str(4),
			Avatar: row.Str(5),
		}
		u.ShowName = u.GetShowName()
		evidence := common.TokenTaskEvidence{
			Id:            row.Uint64(6),
			Bonus:         new(big.Int).SetUint64(row.Uint64(7)),
			Description:   row.Str(8),
			CreateTime:    row.ForceLocaltime(10).Format(time.RFC3339),
			UpdateTime:    row.ForceLocaltime(11).Format(time.RFC3339),
			Tx:            row.Str(12),
			TxStatus:      row.Uint(13),
			ApproveStatus: row.Int(14),
			Task:          task,
			User:          u,
		}
		imgArr := strings.Split(row.Str(9), ",")
		for _, img := range imgArr {
			if img != "" {
				evidence.Images = append(evidence.Images, img)
			}
		}
		evidences = append(evidences, evidence)
	}

	c.JSON(http.StatusOK, evidences)
}
