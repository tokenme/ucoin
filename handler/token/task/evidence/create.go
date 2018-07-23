package evidence

import (
	"bufio"
	"bytes"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	//"github.com/mkideal/log"
	"github.com/nu7hatch/gouuid"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"github.com/tokenme/ucoin/tools/qiniu"
	commonutils "github.com/tokenme/ucoin/utils"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CreateRequest struct {
	TaskId      uint64 `form:"task_id" json:"task_id" binding:"required"`
	Description string `form:"desc" json:"desc" binding:"required"`
	Images      string `form:"images" json:"images"`
}

func CreateHandler(c *gin.Context) {
	userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)
	var req CreateRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}

	db := Service.Db
	rows, _, err := db.Query(`SELECT e.address, e.name, e.symbol, e.decimals, e.logo, t.bonus, t.amount, IFNULL(te.id, 0), IFNULL(te.approve_status, 0) FROM ucoin.tasks AS t INNER JOIN ucoin.erc20 AS e ON (e.address=t.erc20) LEFT JOIN ucoin.task_evidences AS te ON (te.task_id=t.id AND te.submitter='%s') WHERE t.id=%d LIMIT 1`, db.Escape(user.Wallet), req.TaskId)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	if CheckWithCode(len(rows) == 0, NOTFOUND_ERROR, "not found", c) {
		return
	}
	row := rows[0]
	erc20Token := &common.Token{
		Address:  row.Str(0),
		Name:     row.Str(1),
		Symbol:   row.Str(2),
		Decimals: row.Uint(3),
		Logo:     row.Str(4),
	}

	bonus := row.Uint64(5)
	taskAmount := row.Uint(6)
	evidenceId := row.Uint64(7)
	approveStatus := row.Uint(8)

	if taskAmount > 0 {
		rows, _, err := db.Query(`SELECT COUNT(1) FROM ucoin.task_evidences WHERE task_id=%d`, req.TaskId)
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		if len(rows) > 0 {
			paticipants := rows[0].Uint(0)
			if CheckWithCode(paticipants >= taskAmount, NOT_ENOUGH_TOKEN_TASK_ERROR, "exceeded maximum submitters", c) {
				return
			}
		}
	}
	var imageURLs []string
	if req.Images == "" {
		formData := c.Request.MultipartForm
		timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
		if formData != nil {
			imageId, err := uuid.NewV4()
			if CheckErr(err, c) {
				return
			}
			imageIdStr := commonutils.Sha1(imageId.String())
			files := formData.File["images"]
			for i, _ := range files { // loop through the files one by one
				file, err := files[i].Open()
				defer file.Close()
				if CheckErr(err, c) {
					return
				}
				buf := new(bytes.Buffer)
				var w = bufio.NewWriter(buf)
				w.ReadFrom(file)
				w.Flush()
				imageURL, _, err := qiniu.Upload(c, Config.Qiniu, fmt.Sprintf("%s/%s/%s/%d", Config.Qiniu.TokenTaskEvidencePath, req.TaskId, imageIdStr, i), timestamp, buf.Bytes())
				if CheckErr(err, c) {
					raven.CaptureError(err, nil)
					return
				}
				imageURLs = append(imageURLs, imageURL)
			}
		}
	} else {
		imagesArr := strings.Split(req.Images, ",")
		for _, img := range imagesArr {
			t := strings.TrimSpace(img)
			if t == "" {
				continue
			}
			imageURLs = append(imageURLs, img)
		}
	}

	var imagesStr = "NULL"

	if len(imageURLs) > 0 {
		imagesStr = fmt.Sprintf("'%s'", db.Escape(strings.Join(imageURLs, ",")))
	}
	if evidenceId > 0 {
		if CheckWithCode(approveStatus != 0, DUPLICATE_EVIDENCE_ERROR, "duplicate evidence", c) {
			return
		}
		_, _, err := db.Query(`UPDATE ucoin.task_evidences SET summary='%s', images=%s WHERE id=%d`, db.Escape(req.Description), imagesStr, evidenceId)
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
	} else {
		_, ret, err := db.Query(`INSERT INTO ucoin.task_evidences (submitter, task_id, erc20, bonus, summary, images) VALUES ('%s', %d, '%s', %d, '%s', %s)`, db.Escape(user.Wallet), req.TaskId, db.Escape(erc20Token.Address), bonus, db.Escape(req.Description), imagesStr)
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		evidenceId = ret.InsertId()
	}

	evidence := common.TokenTaskEvidence{
		Id:   evidenceId,
		User: user,
		Task: common.TokenTask{
			Id:    req.TaskId,
			Token: erc20Token,
		},
		Bonus:       new(big.Int).SetUint64(bonus),
		Description: req.Description,
		Images:      imageURLs,
		CreateTime:  time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, evidence)
}
