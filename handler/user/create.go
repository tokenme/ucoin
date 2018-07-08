package user

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/params"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/o1egl/govatar"
	"image/png"
	//"github.com/mkideal/log"
	"github.com/nu7hatch/gouuid"
	"github.com/tokenme/ucoin/coins/eth"
	. "github.com/tokenme/ucoin/handler"
	"github.com/tokenme/ucoin/tools/qiniu"
	"github.com/tokenme/ucoin/utils"
	tokenUtils "github.com/tokenme/ucoin/utils/token"
	"github.com/tokenme/ucoin/utils/twilio"
	"github.com/ziutek/mymysql/mysql"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CreateRequest struct {
	Mobile      string         `form:"mobile" json:"mobile"`
	CountryCode uint           `form:"country_code" json:"country_code"`
	VerifyCode  string         `form:"verify_code" json:"verify_code"`
	Password    string         `form:"passwd" json:"passwd"`
	RePassword  string         `form:"repasswd" json:"repasswd"`
	Wechat      *WechatRequest `form:"wechat" json:"wechat"`
}

type WechatRequest struct {
	AccessKey     string `form:"accessKey", json:"accessKey"`
	UnionId       string `form:"unionId", json:"unionId"`
	Nick          string `form:"nickName", json:"nickName"`
	Gender        uint   `form:"gender", json:"gender"`
	City          string `form:"city" json:"city"`
	Province      string `form:"province" json:"province"`
	Country       string `form:"country" json:"country"`
	Avatar        string `form:"avatarUrl" json:"avatarUrl"`
	Language      string `form:"language" json:"language"`
	EncryptedData string `from:"encryptedData" json:"encryptedData" binding:"required"`
	Iv            string `from:"iv" json:"iv" binding:"required"`
}

func CreateHandler(c *gin.Context) {
	var req CreateRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}
	if req.Wechat != nil {
		createByWechat(c, req)
		return
	}
	createByMobile(c, req)
}

func createByWechat(c *gin.Context, req CreateRequest) {
	if Check(req.Wechat.AccessKey == "", "missing wx_session_key", c) {
		return
	}
	db := Service.Db
	rows, _, err := db.Query(`SELECT union_id, session_key FROM ucoin.wx_oauth WHERE k='%s' LIMIT 1`, db.Escape(req.Wechat.AccessKey))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	if Check(len(rows) == 0, "unauthorized", c) {
		return
	}
	unionId := rows[0].Str(0)
	privateKey, _, err := eth.GenerateAccount()
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	walletSalt, wallet, err := utils.AddressEncrypt(privateKey, Config.TokenSalt)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	pubKey, err := eth.AddressFromHexPrivateKey(privateKey)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	_, _, err = db.Query(`INSERT IGNORE INTO ucoin.users (wallet, wallet_salt, wallet_addr, wx_unionid, wx_nick, wx_avatar, wx_gender, wx_city, wx_province, wx_country, wx_language) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', %d, '%s', '%s', '%s', '%s')`, db.Escape(wallet), db.Escape(walletSalt), db.Escape(pubKey), db.Escape(unionId), db.Escape(req.Wechat.Nick), db.Escape(req.Wechat.Avatar), req.Wechat.Gender, db.Escape(req.Wechat.City), db.Escape(req.Wechat.Province), db.Escape(req.Wechat.Country), db.Escape(req.Wechat.Language))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	outputWalletPrivateKey, err := utils.AddressDecrypt(Config.OutputWallet.Data, Config.OutputWallet.Salt, Config.TokenSalt)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	transactor := eth.TransactorAccount(outputWalletPrivateKey)
	transactor.Value = new(big.Int).Mul(big.NewInt(2), big.NewInt(params.Ether))
	_, err = eth.Transfer(transactor, Service.Geth, c, pubKey)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	c.JSON(http.StatusOK, APIResponse{Msg: "ok"})
}

func createByMobile(c *gin.Context, req CreateRequest) {
	if Check(req.Mobile == "" || req.CountryCode == 0 || req.VerifyCode == "" || req.Password == "" || req.RePassword == "", "missing params", c) {
		return
	}
	if Check(req.Password != req.RePassword, "repassword!=password", c) {
		return
	}
	passwdLength := len(req.Password)
	if Check(passwdLength < 8 || passwdLength > 64, "password length must between 8-32", c) {
		return
	}
	token, err := uuid.NewV4()
	if CheckErr(err, c) {
		return
	}
	salt := utils.Sha1(token.String())
	passwd := utils.Sha1(fmt.Sprintf("%s%s%s", salt, req.Password, salt))
	mobile := strings.Replace(req.Mobile, " ", "", 0)

	ret, err := twilio.AuthVerification(Config.TwilioToken, mobile, req.CountryCode, req.VerifyCode)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	if Check(!ret.Success, ret.Message, c) {
		return
	}

	privateKey, _, err := eth.GenerateAccount()
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	walletSalt, wallet, err := utils.AddressEncrypt(privateKey, Config.TokenSalt)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	pubKey, err := eth.AddressFromHexPrivateKey(privateKey)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	gender := govatar.MALE
	maleOrFemale := utils.RangeRandUint64(0, 1)
	if maleOrFemale == 1 {
		gender = govatar.FEMALE
	}
	avatarImg, err := govatar.GenerateFromUsername(gender, pubKey)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	avatarBuf := new(bytes.Buffer)
	err = png.Encode(avatarBuf, avatarImg)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	avatar, _, err := qiniu.Upload(c, Config.Qiniu, fmt.Sprintf("%s/%s", Config.Qiniu.AvatarPath, pubKey), timestamp, avatarBuf.Bytes())
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	db := Service.Db
	_, userRet, err := db.Query(`INSERT INTO ucoin.users (country_code, mobile, passwd, avatar, salt, active, wallet, wallet_salt, wallet_addr) VALUES (%d, '%s', '%s', '%s', '%s', 1, '%s', '%s', '%s')`, req.CountryCode, db.Escape(mobile), db.Escape(passwd), db.Escape(avatar), db.Escape(salt), db.Escape(wallet), db.Escape(walletSalt), db.Escape(pubKey))
	if err != nil && err.(*mysql.Error).Code == mysql.ER_DUP_ENTRY {
		c.JSON(http.StatusOK, APIResponse{Msg: "account already exists"})
		return
	}
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	userId := userRet.InsertId()

	for {
		nickname := tokenUtils.New()
		rows, _, err := db.Query(`UPDATE ucoin.users SET nickname='%s' WHERE id=%d LIMIT 1`, db.Escape(nickname.Encode()), userId)
		if err != nil {
			continue
		}
		if len(rows) == 0 {
			break
		}
	}

	outputWalletPrivateKey, err := utils.AddressDecrypt(Config.OutputWallet.Data, Config.OutputWallet.Salt, Config.TokenSalt)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	transactor := eth.TransactorAccount(outputWalletPrivateKey)
	transactor.Value = eth.InitialGas
	_, err = eth.Transfer(transactor, Service.Geth, c, pubKey)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	c.JSON(http.StatusOK, APIResponse{Msg: "ok"})
}
