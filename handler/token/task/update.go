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
	"github.com/tokenme/ucoin/coins/eth"
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

type UpdateRequest struct {
	Id           uint64 `form:"id" json:"id" binding:"required"`
	Title        string `form:"title" json:"title"`
	Bonus        uint64 `form:"bonus" json:"bonus"`
	Amount       uint   `form:"amount" json:"amount"`
	StartDate    string `form:"start_date" json:"start_date"`
	EndDate      string `form:"end_date" json:"end_date"`
	Description  string `form:"desc" json:"desc"`
	Tags         string `form:"tags" json:"tags"`
	Images       string `form:"images" json:"images"`
	OnlineStatus int    `form:"online_status" json:"online_status"`
	NeedEvidence int    `form:"need_evidence" json:"need_evidence"`
}

func UpdateHandler(c *gin.Context) {
	userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)
	var req UpdateRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}
	db := Service.Db

	rows, _, err := db.Query(`SELECT wallet, wallet_salt FROM ucoin.users WHERE id=%d LIMIT 1`, user.Id)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	if CheckWithCode(len(rows) == 0, UNAUTHORIZED_ERROR, "user not exists", c) {
		return
	}
	row := rows[0]
	walletEncrypted := row.Str(0)
	walletSalt := row.Str(1)

	privKey, err := commonutils.AddressDecrypt(walletEncrypted, walletSalt, Config.TokenSalt)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	pubKey, err := eth.AddressFromHexPrivateKey(privKey)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	rows, _, err = db.Query(`SELECT p.id, p.name, p.bonus, p.amount, p.start_date, p.end_date, p.online_status, t.address, t.name, t.symbol, t.decimals, t.logo FROM ucoin.tasks AS p INNER JOIN ucoin.erc20 AS t ON (t.address=p.erc20) WHERE p.id=%d AND t.owner='%s' LIMIT 1`, req.Id, db.Escape(pubKey))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	if CheckWithCode(len(rows) == 0, NOTFOUND_ERROR, "not found", c) {
		return
	}
	row = rows[0]
	erc20Token := &common.Token{
		Address:  row.Str(7),
		Name:     row.Str(8),
		Symbol:   row.Str(9),
		Decimals: row.Uint(10),
		Logo:     row.Str(11),
	}

	tokenTask := common.TokenTask{
		Id:           row.Uint64(0),
		Owner:        pubKey,
		Title:        row.Str(1),
		Token:        erc20Token,
		Bonus:        new(big.Int).SetUint64(row.Uint64(2)),
		Amount:       row.Uint(3),
		StartDate:    row.ForceLocaltime(4).Format(time.RFC3339),
		EndDate:      row.ForceLocaltime(5).Format(time.RFC3339),
		OnlineStatus: row.Int(6),
	}

	timeFormatter := "2006-01-02T15:04:05-0700"
	var updateSets []string
	if req.StartDate != "" {
		startTime, err := time.Parse(timeFormatter, req.StartDate)
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		startTimeStr := startTime.UTC().Format("2006-01-02 15:04:05")
		updateSets = append(updateSets, fmt.Sprintf("start_date='%s'", startTimeStr))
		tokenTask.StartDate = req.StartDate
	}

	if req.EndDate != "" {
		endTime, err := time.Parse(timeFormatter, req.EndDate)
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		endTimeStr := endTime.UTC().Format("2006-01-02 15:04:05")
		updateSets = append(updateSets, fmt.Sprintf("end_date='%s'", endTimeStr))
		tokenTask.EndDate = req.EndDate
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
				imageURL, _, err := qiniu.Upload(c, Config.Qiniu, fmt.Sprintf("%s/%s/%s/%d", Config.Qiniu.TokenTaskImagePath, tokenTask.Token.Address, imageIdStr, i), timestamp, buf.Bytes())
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
	if len(imageURLs) > 0 {
		updateSets = append(updateSets, fmt.Sprintf("images='%s'", db.Escape(strings.Join(imageURLs, ","))))
		tokenTask.Images = imageURLs
	}
	var (
		tags    []string
		tagsArr = strings.Split(req.Tags, " ")
	)
	for _, tag := range tagsArr {
		t := strings.TrimSpace(tag)
		if t == "" {
			continue
		}
		tags = append(tags, strings.ToLower(t))
	}
	if len(tags) > 0 {
		updateSets = append(updateSets, fmt.Sprintf("tags='%s'", db.Escape(strings.Join(tags, " "))))
		tokenTask.Tags = tags
	}
	if req.Bonus > 0 {
		updateSets = append(updateSets, fmt.Sprintf("bonus=%d", req.Bonus))
		tokenTask.Bonus = new(big.Int).SetUint64(req.Bonus)
	}
	if req.Amount > 0 {
		updateSets = append(updateSets, fmt.Sprintf("amount=%d", req.Amount))
		tokenTask.Amount = req.Amount
	}
	if req.Description != "" {
		updateSets = append(updateSets, fmt.Sprintf("summary='%s'", strings.TrimSpace(req.Description)))
		tokenTask.Description = req.Description
	}
	if req.Title != "" {
		updateSets = append(updateSets, fmt.Sprintf("name='%s'", strings.TrimSpace(req.Title)))
		tokenTask.Title = req.Title
	}
	if req.OnlineStatus == 1 || req.OnlineStatus == -1 {
		updateSets = append(updateSets, fmt.Sprintf("online_status=%d", req.OnlineStatus))
		tokenTask.OnlineStatus = req.OnlineStatus
	}
	if req.NeedEvidence == 1 || req.NeedEvidence == -1 {
		updateSets = append(updateSets, fmt.Sprintf("need_evidence=%d", req.NeedEvidence))
		tokenTask.NeedEvidence = req.NeedEvidence
	}
	_, _, err = db.Query(`UPDATE ucoin.tasks SET %s WHERE id='%d'`, strings.Join(updateSets, ","), req.Id)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	c.JSON(http.StatusOK, tokenTask)
}
