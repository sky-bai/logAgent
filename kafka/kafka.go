package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

var (
	client  sarama.SyncProducer
	msgChan chan *sarama.ProducerMessage
)

func Init(address []string, chanSize int64) (err error) {
	// 1.生产者配置
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出一个partition
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回

	// 2.连接kafka
	client, err = sarama.NewSyncProducer(address, config)
	if err != nil {
		logrus.Error("init kafka producer failed, err:", err)
		return
	}
	msgChan = make(chan *sarama.ProducerMessage, chanSize)
	// 启动一个goroutine发送消息
	go sendMsg()
	return
}

func sendMsg() {
	fmt.Println("---sendMsg---")
	for {
		select {
		case msg := <-msgChan:
			fmt.Println("读数据")
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				logrus.Warning("send message failed,err:%v", err)
				return
			}
			logrus.Info("发送消息成功 success,pid:%d,offset:%d", pid, offset)
		}
	}
}

// GetMsgChan 向外暴露只写的msgChan
func GetMsgChan() chan<- *sarama.ProducerMessage {
	return msgChan
}
