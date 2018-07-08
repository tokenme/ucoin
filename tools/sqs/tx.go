package sqs

import (
	"context"
	//"github.com/davecgh/go-spew/spew"
	"fmt"
	"github.com/go-msgqueue/msgqueue"
	"github.com/mkideal/log"
	"github.com/panjf2000/ants"
	"github.com/tokenme/ucoin/coins/eth/utils"
	"github.com/tokenme/ucoin/common"
	"sync"
	"time"
)

type TxMessage struct {
	Type MsgType `codec:"type"`
	Tx   string  `codec:"tx"`
}

func (this TxMessage) String() string {
	return fmt.Sprintf("%d-%s", this.Type, this.Tx)
}

type TxQueue struct {
	service   *common.Service
	config    common.Config
	queue     msgqueue.Queue
	processor *msgqueue.Processor
	reSendCh  chan string
	txCh      chan string
	exitCh    chan struct{}
	canStopCh chan struct{}
}

func NewTxQueue(m msgqueue.Manager, service *common.Service, config common.Config) *TxQueue {
	queue := &TxQueue{
		service:   service,
		config:    config,
		reSendCh:  make(chan string, 1000),
		txCh:      make(chan string, 1000),
		exitCh:    make(chan struct{}, 1),
		canStopCh: make(chan struct{}, 1),
	}
	opt := &msgqueue.Options{
		Name:    config.SQS.TxQueue,
		Handler: queue.Handler,
	}
	q := m.NewQueue(opt)
	queue.queue = q
	queue.processor = q.Processor()
	return queue
}

func (this *TxQueue) Start() {
	this.processor.Start()
	ctx := context.Background()
	var wg sync.WaitGroup
	pool, _ := ants.NewPoolWithFunc(10000, func(tx interface{}) error {
		defer wg.Done()
		txHex := tx.(string)
		log.Info("Checking New Tx Receipt:%s", txHex)
		receipt, err := utils.TransactionReceipt(this.service.Geth, ctx, txHex)
		if err != nil {
			log.Error(err.Error())
			this.reSendCh <- txHex
			return err
		}
		if receipt == nil {
			this.reSendCh <- txHex
			return nil
		}
		db := this.service.Db
		_, _, err = db.Query(`INSERT INTO ucoin.txs (tx, status) VALUES ('%s', %d)`, db.Escape(txHex), receipt.Status)
		if err != nil {
			log.Error(err.Error())
			this.reSendCh <- txHex
			return err
		}
		log.Info("Updated Tx Receipt:%s, Status:%d", txHex, receipt.Status)
		return nil
	})

	delayPool, _ := ants.NewPoolWithFunc(10000, func(tx interface{}) error {
		defer wg.Done()
		time.Sleep(10 * time.Second)
		return this.NewTx(tx.(string))
	})
	shouldStop := false
	for !shouldStop {
		select {
		case tx := <-this.txCh:
			wg.Add(1)
			pool.Serve(tx)
		case tx := <-this.reSendCh:
			wg.Add(1)
			delayPool.Serve(tx)
		case <-this.exitCh:
			shouldStop = true
			break
		}
	}
	wg.Wait()
	this.canStopCh <- struct{}{}
}

func (this *TxQueue) Stop() {
	this.processor.Stop()
	this.exitCh <- struct{}{}
	<-this.canStopCh
}

func (this *TxQueue) NewTx(tx string) error {
	return this.queue.Call(TxMessage{Type: TransactionMsg, Tx: tx})
}

func (this *TxQueue) Handler(msg TxMessage) error {
	switch msg.Type {
	case TransactionMsg:
		this.txCh <- msg.Tx
	}
	return nil
}
