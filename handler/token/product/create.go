package product

import (
	"bufio"
	"bytes"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	//"github.com/mkideal/log"
	"github.com/tokenme/ucoin/coins/eth"
	"github.com/tokenme/ucoin/coins/eth/utils"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"github.com/tokenme/ucoin/tools/qiniu"
	"github.com/tokenme/ucoin/tools/sqs"
	commonutils "github.com/tokenme/ucoin/utils"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CreateRequest struct {
	Title       string `form:"title" json:"title" binding:"required"`
	Token       string `form:"token" json:"token" binding:"required"`
	Price       uint64 `form:"price" json:"price" binding:"required"`
	Amount      uint   `form:"amount" json:"amount"`
	StartDate   string `form:"start_date" json:"start_date" binding:"required"`
	EndDate     string `form:"end_date" json:"end_date" binding:"required"`
	Description string `form:"desc" json:"desc" binding:"required"`
	Tags        string `form:"tags" json:"tags"`
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

	rows, _, err = db.Query(`SELECT address, name, symbol, decimals, logo, tx_status FROM ucoin.erc20 WHERE address="%s" AND owner="%s" LIMIT 1`, db.Escape(req.Token), db.Escape(pubKey))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	if CheckWithCode(len(rows) == 0, NOTFOUND_ERROR, "not found", c) {
		return
	}
	row = rows[0]
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

	transactor := eth.TransactorAccount(privKey)
	nonce, err := eth.Nonce(c, Service.Geth, Service.Redis.Master, pubKey, eth.UC_CHAIN)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	transactorOpts := eth.TransactorOptions{
		Nonce:    nonce,
		GasLimit: 2100000,
	}
	eth.TransactorUpdate(transactor, transactorOpts, c)
	contractAddress, tx, _, err := utils.DeployNFToken(transactor, Service.Geth, req.Title, erc20Token.Symbol, Config.ERC721Template)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	eth.NonceIncr(c, Service.Geth, Service.Redis.Master, pubKey, eth.UC_CHAIN)
	tokenAddress := strings.ToLower(contractAddress.Hex())
	txHash := tx.Hash()
	err = Queues[Config.SQS.TxQueue].(*sqs.TxQueue).NewTx(txHash.Hex(), "erc721", "tx", "tx_status")
	if err != nil {
		raven.CaptureError(err, nil)
		return
	}
	var imageURLs []string
	if req.Images == "" {
		formData := c.Request.MultipartForm
		timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
		if formData != nil {
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
				imageURL, _, err := qiniu.Upload(c, Config.Qiniu, fmt.Sprintf("%s/%s/%d", Config.Qiniu.TokenProductImagePath, req.Token, i), timestamp, buf.Bytes())
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
		imagesStr = "NULL"
		tagsStr   = "NULL"
		tagsArr   = strings.Split(req.Tags, " ")
		tags      []string
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
	_, _, err = db.Query(`INSERT INTO ucoin.erc721 (address, owner, name, symbol, summary, erc20, price, amount, start_date, end_date, images, tags, tx) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', %d, %d, '%s', '%s', %s, %s, '%s')`, db.Escape(tokenAddress), db.Escape(pubKey), db.Escape(req.Title), db.Escape(Config.ERC721Symbol), db.Escape(req.Description), db.Escape(erc20Token.Address), req.Price, req.Amount, db.Escape(startTimeStr), db.Escape(endTimeStr), imagesStr, tagsStr, db.Escape(txHash.Hex()))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	tokenProduct := common.TokenProduct{
		Address:     tokenAddress,
		Owner:       pubKey,
		Title:       req.Title,
		Token:       erc20Token,
		Price:       new(big.Int).SetUint64(req.Price),
		Amount:      req.Amount,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Description: req.Description,
		Images:      imageURLs,
		Tags:        tags,
	}

	balance, err := eth.BalanceOf(Service.Geth, c, pubKey)
	if err != nil {
		raven.CaptureError(err, nil)
	} else if balance.Cmp(eth.MinGas) == -1 {
		err = Queues[Config.SQS.GasQueue].(*sqs.GasQueue).Deposit(pubKey, eth.MinGas.Uint64())
		if err != nil {
			raven.CaptureError(err, nil)
		}
	}

	c.JSON(http.StatusOK, tokenProduct)
}
