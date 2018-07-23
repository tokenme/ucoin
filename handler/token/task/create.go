package task

import (
	"bufio"
	"bytes"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/nu7hatch/gouuid"
	//"github.com/mkideal/log"
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
	Title        string `form:"title" json:"title" binding:"required"`
	Token        string `form:"token" json:"token" binding:"required"`
	Bonus        uint64 `form:"bonus" json:"bonus" binding:"required"`
	Amount       uint   `form:"amount" json:"amount"`
	StartDate    string `form:"start_date" json:"start_date" binding:"required"`
	EndDate      string `form:"end_date" json:"end_date" binding:"required"`
	Description  string `form:"desc" json:"desc" binding:"required"`
	Tags         string `form:"tags" json:"tags"`
	Images       string `form:"images" json:"images"`
	NeedEvidence int    `form:"need_evidence" json:"need_evidence"`
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
	timeFormatter := "2006-01-02T15:04:05-0700"
	startTime, err := time.Parse(timeFormatter, req.StartDate)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	endTime, err := time.Parse(timeFormatter, req.EndDate)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	if Check(endTime.Before(startTime), "start date should before end date", c) {
		return
	}

	db := Service.Db

	rows, _, err := db.Query(`SELECT address, name, symbol, decimals, logo, tx_status FROM ucoin.erc20 WHERE address="%s" AND owner="%s" LIMIT 1`, db.Escape(req.Token), db.Escape(user.Wallet))
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
		TxStatus: row.Int(5),
	}

	if CheckWithCode(erc20Token.TxStatus != 1, TOKEN_UNDER_CONSTRUCTION_ERROR, "token under construct", c) {
		return
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
				imageURL, _, err := qiniu.Upload(c, Config.Qiniu, fmt.Sprintf("%s/%s/%s/%d", Config.Qiniu.TokenTaskImagePath, req.Token, imageIdStr, i), timestamp, buf.Bytes())
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

	startTimeStr := startTime.UTC().Format("2006-01-02 15:04:05")
	endTimeStr := endTime.UTC().Format("2006-01-02 15:04:05")

	var (
		imagesStr    = "NULL"
		tagsStr      = "NULL"
		tagsArr      = strings.Split(req.Tags, " ")
		tags         []string
		needEvidence = -1
	)

	if len(imageURLs) > 0 {
		imagesStr = fmt.Sprintf("'%s'", db.Escape(strings.Join(imageURLs, ",")))
	}
	for _, tag := range tagsArr {
		t := strings.TrimSpace(tag)
		if t == "" {
			continue
		}
		tags = append(tags, strings.ToLower(t))
	}

	if len(tags) > 0 {
		tagsStr = fmt.Sprintf("'%s'", db.Escape(strings.Join(tags, " ")))
	}
	if req.NeedEvidence == 1 {
		needEvidence = 1
	}
	_, ret, err := db.Query(`INSERT INTO ucoin.tasks (owner, name, summary, erc20, bonus, amount, start_date, end_date, images, tags, need_evidence) VALUES ('%s', '%s', '%s', '%s', %d, %d, '%s', '%s', %s, %s, %d)`, db.Escape(user.Wallet), db.Escape(req.Title), db.Escape(req.Description), db.Escape(erc20Token.Address), req.Bonus, req.Amount, db.Escape(startTimeStr), db.Escape(endTimeStr), imagesStr, tagsStr, needEvidence)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	taskId := ret.InsertId()

	tokenTask := common.TokenTask{
		Id:           taskId,
		Owner:        user.Wallet,
		Title:        req.Title,
		Token:        erc20Token,
		Bonus:        new(big.Int).SetUint64(req.Bonus),
		Amount:       req.Amount,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Description:  req.Description,
		Images:       imageURLs,
		Tags:         tags,
		NeedEvidence: needEvidence,
		OnlineStatus: 1,
	}

	c.JSON(http.StatusOK, tokenTask)
}
