package main

/*
 * kafka 消费测试代码
 *
 */

import (
	"github.com/cihub/seelog"
	"strings"
	"github.com/Shopify/sarama"
	"os"
	"log"
	"github.com/bsm/sarama-cluster"
	"os/signal"
	"sync"
)

// 消息处理协程启动参数
type ProcInfo struct {
	routineId int                          // 协程编号
	msgChan   chan *sarama.ConsumerMessage // kafka 消息队列
	exitChan  chan int32                   // 协程退出 channel
	wg        *sync.WaitGroup              //
}

// consumer 协程启动参数
type ConsumerInfo struct {
	ProcInfo
	group   string   // 消费组
	brokers []string //
	topics  []string
	config  *cluster.Config
}

// 处理消息
func ProcMessage(info ProcInfo) {
	info.wg.Add(1)
	defer info.wg.Done()

	for {
		select {
		case msg := <-info.msgChan:
			seelog.Debugf("routine[%d] recive msg, len:%d", info.routineId, len(msg.Value))
		case <-info.exitChan:
			seelog.Debugf("message proc routine[%d] exit", info.routineId)
			return
		}
	}
}

// consumer 协程
func ConsumerProcess(info ConsumerInfo) {
	info.wg.Add(1)
	defer info.wg.Done()

	consumer, err := cluster.NewConsumer(info.brokers, info.group, info.topics, info.config)
	if err != nil {
		seelog.Errorf("NewConsumer error:%s", err.Error())
		return
	}
	defer consumer.Close()

	// consume errors
	go func() {
		for err := range consumer.Errors() {
			seelog.Errorf("Error: %s", err.Error())
		}
	}()

	// consume notifications
	go func() {
		for ntf := range consumer.Notifications() {
			seelog.Errorf("Rebalanced: %+v", ntf)
		}
	}()

LABEL_CONSUMER_LOOP:
	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				info.msgChan <- msg
			}
		case <-info.exitChan:
			seelog.Debugf("consumer routine[%d] exit", info.routineId)
			break LABEL_CONSUMER_LOOP
		}
	}

}

func main() {

	LogInit()
	defer seelog.Flush()

	seelog.Debugf("LogInit Success!")

	// 读取配置文件
	consumerConf := KafkaConsumerConf{}
	err := LoadConfig("./app_conf.json", &consumerConf)
	if err != nil {
		seelog.Errorf("loadConfig error!")
		return
	}
	seelog.Debugf("AppConf[%+v]", consumerConf)

	var waitGroup sync.WaitGroup

	allMessage := make(chan *sarama.ConsumerMessage, consumerConf.NumConsumerRoutine*20)
	sigChannel := make(chan int32, consumerConf.NumProcessRoutine)

	// 启动消息处理协程
	for i := 0; i < consumerConf.NumProcessRoutine; i++ {
		procInfo := ProcInfo{routineId: i, msgChan: allMessage, exitChan: sigChannel, wg: &waitGroup}
		go ProcMessage(procInfo)
	}

	whole, err := os.Open(os.DevNull)
	if err != nil {
		seelog.Errorf("open file /dev/null error! exit now.")
		return
	}
	// sarama.Logger = log.New(os.Stderr, "[srama]", log.LstdFlags)
	sarama.Logger = log.New(whole, "[srama]", log.LstdFlags)

	// init (custom) config, enable errors and notifications
	config := cluster.NewConfig()

	config.ClientID = "goClient"
	config.ChannelBufferSize = 20480

	// config.Net.MaxOpenRequests = 2048
	//config.Net.DialTimeout = 100 * time.Millisecond
	//config.Net.ReadTimeout = 10 * time.Millisecond
	//config.Net.WriteTimeout = 10 * time.Millisecond
	//config.Net.KeepAlive = 10 * time.Millisecond // 5 * time.Second

	//config.Consumer.MaxWaitTime = 250 * time.Millisecond
	//config.Consumer.MaxProcessingTime = 100 * time.Millisecond
	//config.Consumer.Retry.Backoff = 10 * time.Millisecond
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// init consumer
	brokers := strings.Split(consumerConf.BrokersAddr, ",")
	topics := strings.Split(consumerConf.Topics, ",")

	// consumer 退出信号
	consumerExitSignal := make(chan int32, consumerConf.NumConsumerRoutine)
	// 启动 consumer 协程
	for i := 0; i < consumerConf.NumConsumerRoutine; i++ {
		consumerInfo := ConsumerInfo{ProcInfo: ProcInfo{routineId: i, msgChan: allMessage, exitChan: consumerExitSignal,
			wg: &waitGroup}, group: consumerConf.PartionGroup, brokers: brokers, topics: topics, config: config}
		go ConsumerProcess(consumerInfo)
		//go ProcMessage(procInfo)
	}

	// trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

LABEL_MAIN_LOOP:
	for {
		select {
		case <-signals:
			break LABEL_MAIN_LOOP
		}
	}

	// 向 consumer 协程发送退出信号
	for i := 0; i < consumerConf.NumConsumerRoutine; i++ {
		consumerExitSignal <- 1
	}

	// 向消息处理协程发送退出信号
	for i := 0; i < consumerConf.NumProcessRoutine; i++ {
		sigChannel <- 1
	}
	seelog.Debugf("wait all routine exit!")
	waitGroup.Wait()

	seelog.Debugf("proc end.")
}
