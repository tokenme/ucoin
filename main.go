package main

import (
	"flag"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/fvbock/endless"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/configor"
	"github.com/mkideal/log"
	"github.com/tokenme/ucoin/common"
	"github.com/tokenme/ucoin/handler"
	"github.com/tokenme/ucoin/router"
	"github.com/tokenme/ucoin/tools/gc"
	"github.com/tokenme/ucoin/tools/sqs"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var (
		config     common.Config
		configFlag common.Config
		configPath string
	)

	os.Setenv("CONFIGOR_ENV_PREFIX", "-")

	flag.StringVar(&configPath, "config", "config.toml", "configuration file")
	flag.IntVar(&configFlag.Port, "port", 0, "set port")
	flag.StringVar(&configFlag.StaticPath, "static", "", "set static file path without filename")
	flag.StringVar(&configFlag.LogPath, "log", "", "set log file path without filename")
	flag.BoolVar(&configFlag.Debug, "debug", false, "set debug mode")
	flag.BoolVar(&configFlag.EnableWeb, "web", false, "enable http web server")
	flag.BoolVar(&configFlag.EnableGC, "gc", false, "enable gc")
	flag.BoolVar(&configFlag.EnableTx, "tx", false, "enable tx queue handler")
	flag.Parse()

	configor.New(&configor.Config{Verbose: configFlag.Debug, ErrorOnUnmatchedKeys: true, Environment: "production"}).Load(&config, configPath)

	if configFlag.Port > 0 {
		config.Port = configFlag.Port
	}
	if configFlag.StaticPath != "" {
		config.StaticPath = configFlag.StaticPath
	}
	if configFlag.LogPath != "" {
		config.LogPath = configFlag.LogPath
	}
	if configFlag.EnableWeb {
		config.EnableWeb = configFlag.EnableWeb
	}

	if configFlag.EnableGC {
		config.EnableGC = configFlag.EnableGC
	}

	if configFlag.EnableTx {
		config.EnableTx = configFlag.EnableTx
	}

	if configFlag.Debug {
		config.Debug = configFlag.Debug
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Error(err.Error())
		return
	}
	var logPath string
	if path.IsAbs(config.LogPath) {
		logPath = config.LogPath
	} else {
		logPath = path.Join(wd, config.LogPath)
	}
	defer log.Uninit(log.InitMultiFileAndConsole(logPath, "ucoin.log", log.LvERROR))

	raven.SetDSN(config.SentryDSN)
	service := common.NewService(config)
	defer service.Close()
	service.Db.Reconnect()

	gcHandler := gc.New(service, config)
	if config.EnableGC {
		go gcHandler.Start()
	}

	queueManager := sqs.NewManager(config.SQS)
	queues := make(map[string]sqs.Queue)
	queues = map[string]sqs.Queue{
		config.SQS.TxQueue:  sqs.NewTxQueue(queueManager, service, config),
		config.SQS.GasQueue: sqs.NewGasQueue(queueManager, service, config),
	}
	if config.EnableTx {
		for _, queue := range queues {
			go queue.Start()
		}
	}

	if config.EnableWeb {
		handler.InitHandler(service, config, queues)
		if config.Debug {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
		//gin.DisableBindValidation()
		var staticPath string
		if path.IsAbs(config.StaticPath) {
			staticPath = config.StaticPath
		} else {
			staticPath = path.Join(wd, config.StaticPath)
		}
		log.Info("Static StaticPath path: %s", staticPath)
		r := router.NewRouter(staticPath)
		log.Info("%s started at:0.0.0.0:%d", config.AppName, config.Port)
		defer log.Info("%s exit from:0.0.0.0:%d", config.AppName, config.Port)
		endless.ListenAndServe(fmt.Sprintf(":%d", config.Port), r)
	} else {
		exitChan := make(chan struct{}, 1)
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGSTOP, syscall.SIGTERM)
			<-ch
			exitChan <- struct{}{}
			close(ch)
		}()
		<-exitChan
	}
	if config.EnableTx {
		for _, queue := range queues {
			queue.Stop()
		}
	}
	gcHandler.Stop()
}
