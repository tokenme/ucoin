package token

import (
	"bufio"
	"bytes"
	"fmt"
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
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CreateRequest struct {
	Name        string `form:"name" json:"name" binding:"required"`
	Symbol      string `form:"symbol" json:"symbol" binding:"required"`
	TotalSupply uint64 `form:"total_supply" json:"total_supply" binding:"required"`
	Decimals    uint   `form:"decimals" json:"decimals" binding:"required"`
	Description string `form:"desc" json:"desc"`
	Logo        string `form:"logo" json:"logo"`
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
	contractAddress, tx, _, err := utils.DeployToken(transactor, Service.Geth, req.Name, req.Symbol, req.Decimals, req.TotalSupply)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	eth.NonceIncr(c, Service.Geth, Service.Redis.Master, pubKey, eth.UC_CHAIN)
	tokenAddress := strings.ToLower(contractAddress.Hex())
	txHash := tx.Hash()
	err = Queues[Config.SQS.TxQueue].(*sqs.TxQueue).NewTx(txHash.Hex(), "erc20", "tx", "tx_status")
	if err != nil {
		raven.CaptureError(err, nil)
		return
	}

	var logoURL string
	if req.Logo == "" {
		file, _, err := c.Request.FormFile("logo")
		if err == nil {
			buf := new(bytes.Buffer)
			var w = bufio.NewWriter(buf)
			w.ReadFrom(file)
			w.Flush()
			timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
			logoURL, _, err = qiniu.Upload(c, Config.Qiniu, fmt.Sprintf("%s/%s", Config.Qiniu.LogoPath, tokenAddress), timestamp, buf.Bytes())
			if CheckErr(err, c) {
				raven.CaptureError(err, nil)
				return
			}
		}
	} else {
		logoURL = req.Logo
	}

	var (
		desc = "NULL"
		logo = "NULL"
	)

	if req.Description != "" {
		desc = fmt.Sprintf("'%s'", db.Escape(req.Description))
	}

	if logoURL != "" {
		logo = fmt.Sprintf("'%s'", db.Escape(logoURL))
	}

	_, _, err = db.Query(`INSERT INTO ucoin.erc20 (address, owner, name, symbol, total_supply, decimals, summary, logo, tx) VALUES ('%s', '%s', '%s', '%s', %d, %d, %s, %s, '%s')`, db.Escape(tokenAddress), db.Escape(pubKey), db.Escape(req.Name), db.Escape(req.Symbol), req.TotalSupply, req.Decimals, desc, logo, db.Escape(txHash.Hex()))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	var (
		decimalsPow   = new(big.Int).SetUint64(uint64(math.Pow10(int(req.Decimals))))
		initialSupply = new(big.Int).Mul(new(big.Int).SetUint64(req.TotalSupply), decimalsPow)
	)

	token := common.Token{
		Address:       tokenAddress,
		Owner:         pubKey,
		Name:          req.Name,
		Symbol:        req.Symbol,
		InitialSupply: initialSupply,
		TotalSupply:   initialSupply,
		Decimals:      req.Decimals,
		Description:   req.Description,
	}

	if logoURL != "" {
		token.Logo = logoURL
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

	c.JSON(http.StatusOK, token)
}
