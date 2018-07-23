package handler

import (
	"errors"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/mkideal/log"
	"github.com/tokenme/ucoin/common"
	"github.com/tokenme/ucoin/tools/sqs"
	"net"
	"net/http"
	"strconv"
	"strings"
)

var (
	Service *common.Service
	Config  common.Config
	Queues  map[string]sqs.Queue
)

func InitHandler(s *common.Service, c common.Config, queues map[string]sqs.Queue) {
	Service = s
	Config = c
	Queues = queues
	raven.SetDSN(Config.SentryDSN)
}

type APIResponse struct {
	Msg string `json:"message,omitempty"`
}

type ErrorCode uint

const (
	BADREQUEST_ERROR                 ErrorCode = 400
	INTERNAL_ERROR                   ErrorCode = 500
	NOTFOUND_ERROR                   ErrorCode = 404
	UNAUTHORIZED_ERROR               ErrorCode = 401
	INVALID_PASSWD_ERROR             ErrorCode = 409
	DUPLICATE_USER_ERROR             ErrorCode = 202
	UNACTIVATED_USER_ERROR           ErrorCode = 502
	NOT_ENOUGH_TOKEN_ERROR           ErrorCode = 600
	NOT_ENOUGH_TOKEN_PRODUCT_ERROR   ErrorCode = 601
	NOT_ENOUGH_TOKEN_TASK_ERROR      ErrorCode = 700
	DUPLICATE_EVIDENCE_ERROR         ErrorCode = 701
	TOKEN_UNDER_CONSTRUCTION_ERROR   ErrorCode = 800
	PRODUCT_UNDER_CONSTRUCTION_ERROR ErrorCode = 801
)

type APIError struct {
	Code ErrorCode `json:"code,omitempty"`
	Msg  string    `json:"message,omitempty"`
}

func (this APIError) Error() string {
	return fmt.Sprintf("CODE:%d, MSG:%s", this.Code, this.Msg)
}

func Check(flag bool, err string, c *gin.Context) (ret bool) {
	ret = flag
	if ret {
		log.Error(err)
		c.JSON(http.StatusOK, APIError{Code: BADREQUEST_ERROR, Msg: err})
	}
	return
}

func CheckErr(err error, c *gin.Context) (ret bool) {
	ret = err != nil
	if ret {
		log.Error(err.Error())
		c.JSON(http.StatusOK, APIError{Code: BADREQUEST_ERROR, Msg: err.Error()})
	}
	return
}

func CheckWithCode(flag bool, code ErrorCode, err string, c *gin.Context) (ret bool) {
	ret = flag
	if ret {
		log.Error(err)
		c.JSON(http.StatusOK, APIError{Code: code, Msg: err})
	}
	return
}

func Uint64Value(val string, defaultVal uint64) (uint64, error) {
	if val == "" {
		return defaultVal, nil
	}

	i, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func Uint64NonZero(val string, err string) (uint64, error) {
	if val == "" {
		return 0, errors.New(err)
	}

	i, e := strconv.ParseUint(val, 10, 64)
	if e != nil {
		return 0, e
	}

	return i, nil
}

func ClientIP(c *gin.Context) string {
	if values, _ := c.Request.Header["X-Forwarded-For"]; len(values) > 0 {
		clientIP := values[0]
		if index := strings.IndexByte(clientIP, ','); index >= 0 {
			clientIP = clientIP[0:index]
		}
		clientIP = strings.TrimSpace(clientIP)
		if len(clientIP) > 0 {
			return clientIP
		}
	}
	if values, _ := c.Request.Header["X-Real-Ip"]; len(values) > 0 {
		clientIP := strings.TrimSpace(values[0])
		if len(clientIP) > 0 {
			return clientIP
		}
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

func IP2Long(IpStr string) (int64, error) {
	bits := strings.Split(IpStr, ".")
	if len(bits) != 4 {
		return 0, errors.New("ip format error")
	}

	var sum int64
	for i, n := range bits {
		bit, _ := strconv.ParseInt(n, 10, 64)
		sum += bit << uint(24-8*i)
	}

	return sum, nil
}
