package order

import (
	//"github.com/davecgh/go-spew/spew"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	//"github.com/mkideal/log"
	"github.com/tokenme/ucoin/coins/eth"
	"github.com/tokenme/ucoin/coins/eth/utils"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"github.com/tokenme/ucoin/tools/sqs"
	commonutils "github.com/tokenme/ucoin/utils"
	"math"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type CreateRequest struct {
	Address string `form:"address" json:"address" binding:"required"`
}

func CreateHandler(c *gin.Context) {
	userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)
	var req CreateRequest
	if CheckErr(c.BindJSON(&req), c) {
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

	rows, _, err = db.Query(`SELECT
	erc20.address, 
	erc20.owner, 
	erc20.name, 
	erc20.symbol, 
	erc20.decimals,
	erc20.total_supply AS initial_supply,
	erc20.logo,
	erc20.summary,
	t.address, 
	t.owner, 
	t.name, 
	t.price, 
	t.amount,
	t.start_date,
	t.end_date,
	t.images,
	t.tags,
	t.summary,
	t.online_status,
	t.tx_status,
	u.wallet,
	u.wallet_salt
FROM ucoin.erc721 AS t 
INNER JOIN ucoin.users AS u ON (u.wallet_addr = t.owner)
INNER JOIN ucoin.erc20 AS erc20 ON (erc20.address = t.erc20)
WHERE t.address = '%s' AND t.owner != '%s' LIMIT 1`, db.Escape(req.Address), db.Escape(pubKey))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	if CheckWithCode(len(rows) == 0, NOTFOUND_ERROR, "not found", c) {
		return
	}
	row = rows[0]
	var (
		decimals      = row.Uint(4)
		decimalsPow   = new(big.Int).SetUint64(uint64(math.Pow10(int(decimals))))
		initialSupply = new(big.Int).Mul(new(big.Int).SetUint64(row.Uint64(5)), decimalsPow)
	)
	token := &common.Token{
		Address:       row.Str(0),
		Owner:         row.Str(1),
		Name:          row.Str(2),
		Symbol:        row.Str(3),
		Decimals:      decimals,
		InitialSupply: initialSupply,
		Logo:          row.Str(6),
		Description:   row.Str(7),
	}
	erc721 := common.TokenProduct{
		Address:      row.Str(8),
		Owner:        row.Str(9),
		Title:        row.Str(10),
		Price:        new(big.Int).SetUint64(row.Uint64(11)),
		Amount:       row.Uint(12),
		StartDate:    row.ForceLocaltime(13).Format(time.RFC3339),
		EndDate:      row.ForceLocaltime(14).Format(time.RFC3339),
		Description:  row.Str(17),
		Token:        token,
		OnlineStatus: row.Int(18),
		TxStatus:     row.Int(19),
	}
	if CheckWithCode(erc721.TxStatus != 1, PRODUCT_UNDER_CONSTRUCTION_ERROR, "product under construct", c) {
		return
	}
	imgArr := strings.Split(row.Str(15), ",")
	for _, img := range imgArr {
		if img != "" {
			erc721.Images = append(erc721.Images, img)
		}
	}
	tagArr := strings.Split(row.Str(16), " ")
	for _, tag := range tagArr {
		if tag != "" {
			erc721.Tags = append(erc721.Tags, tag)
		}
	}

	walletEncrypted = row.Str(20)
	walletSalt = row.Str(21)

	erc721PrivKey, err := commonutils.AddressDecrypt(walletEncrypted, walletSalt, Config.TokenSalt)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	erc20Token, err := utils.NewToken(token.Address, Service.Geth)
	if CheckErr(err, c) {
		return
	}
	erc20TokenBalance, err := utils.TokenBalanceOf(erc20Token, pubKey)
	if CheckErr(err, c) {
		return
	}
	if CheckWithCode(erc20TokenBalance.Cmp(erc721.Price) == -1, NOT_ENOUGH_TOKEN_ERROR, "not enough token", c) {
		return
	}
	erc721Token, err := utils.NewNFToken(erc721.Address, Service.Geth)
	if CheckErr(err, c) {
		return
	}
	if erc721.Amount > 0 {
		erc721TotalSupply, err := erc721Token.TotalSupply(nil)
		if CheckErr(err, c) {
			return
		}
		if CheckWithCode(erc721TotalSupply.Cmp(big.NewInt(int64(erc721.Amount))) >= 0, NOT_ENOUGH_TOKEN_PRODUCT_ERROR, "sold out", c) {
			return
		}
	}
	_, ret, err := db.Query(`INSERT INTO ucoin.erc721_orders (buyer, seller, erc721, erc20, price) VALUES ('%s', '%s', '%s', '%s', %d)`, db.Escape(pubKey), db.Escape(erc721.Owner), db.Escape(erc721.Address), db.Escape(token.Address), erc721.Price.Uint64())
	if CheckErr(err, c) {
		return
	}
	orderId := ret.InsertId()
	erc20Transactor := eth.TransactorAccount(privKey)
	nonce, err := eth.Nonce(c, Service.Geth, Service.Redis.Master, pubKey, eth.UC_CHAIN)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	erc20TransactorOpts := eth.TransactorOptions{
		Nonce:    nonce,
		GasLimit: 2100000,
	}
	eth.TransactorUpdate(erc20Transactor, erc20TransactorOpts, c)
	tx, err := utils.Transfer(erc20Token, erc20Transactor, erc721.Owner, erc721.Price)
	if CheckErr(err, c) {
		return
	}
	eth.NonceIncr(c, Service.Geth, Service.Redis.Master, pubKey, eth.UC_CHAIN)
	erc20Tx := tx.Hash().Hex()
	erc721Transactor := eth.TransactorAccount(erc721PrivKey)
	nonce, err = eth.Nonce(c, Service.Geth, Service.Redis.Master, erc721.Owner, eth.UC_CHAIN)
	erc721TransactorOpts := eth.TransactorOptions{
		Nonce:    nonce,
		GasLimit: 2100000,
	}
	if privKey == erc721PrivKey {
		nonce, err := eth.PendingNonce(Service.Geth, c, erc721.Owner)
		if CheckErr(err, c) {
			return
		}
		erc721TransactorOpts.Nonce = nonce + 1
	}
	eth.TransactorUpdate(erc721Transactor, erc721TransactorOpts, c)
	tx, err = utils.NFTMint(erc721Token, erc721Transactor, pubKey, new(big.Int).SetUint64(orderId), fmt.Sprintf("%s/token/product/info?address=%s", Config.BaseUrl, erc721.Address))
	if CheckErr(err, c) {
		return
	}
	eth.NonceIncr(c, Service.Geth, Service.Redis.Master, erc721.Owner, eth.UC_CHAIN)
	erc721Tx := tx.Hash().Hex()
	_, _, err = db.Query(`UPDATE ucoin.erc721_orders SET token_id=%d, erc721_tx='%s', erc20_tx='%s' WHERE id=%d`, orderId, db.Escape(erc721Tx), db.Escape(erc20Tx), orderId)
	if CheckErr(err, c) {
		return
	}
	err = Queues[Config.SQS.TxQueue].(*sqs.TxQueue).NewTx(erc20Tx, "erc721_orders", "erc20_tx", "erc20_tx_status")
	if err != nil {
		raven.CaptureError(err, nil)
		return
	}
	err = Queues[Config.SQS.TxQueue].(*sqs.TxQueue).NewTx(erc721Tx, "erc721_orders", "erc721_tx", "erc721_tx_status")
	if err != nil {
		raven.CaptureError(err, nil)
		return
	}
	order := common.Order{
		TokenId:    orderId,
		Buyer:      user,
		Seller:     common.User{Wallet: erc721.Owner},
		Product:    erc721,
		Price:      erc721.Price,
		InsertedAt: time.Now().Format(time.RFC3339),
		UpdatedAt:  time.Now().Format(time.RFC3339),
		Tx:         erc721Tx,
	}

	balance, err := eth.BalanceOf(Service.Geth, c, order.Buyer.Wallet)
	if err != nil {
		raven.CaptureError(err, nil)
	} else if balance.Cmp(eth.MinGas) == -1 {
		err = Queues[Config.SQS.GasQueue].(*sqs.GasQueue).Deposit(order.Buyer.Wallet, eth.MinGas.Uint64())
		if err != nil {
			raven.CaptureError(err, nil)
		}
	}

	balance, err = eth.BalanceOf(Service.Geth, c, order.Seller.Wallet)
	if err != nil {
		raven.CaptureError(err, nil)
	} else if balance.Cmp(eth.MinGas) == -1 {
		err = Queues[Config.SQS.GasQueue].(*sqs.GasQueue).Deposit(order.Seller.Wallet, eth.MinGas.Uint64())
		if err != nil {
			raven.CaptureError(err, nil)
		}
	}
	c.JSON(http.StatusOK, order)
}
