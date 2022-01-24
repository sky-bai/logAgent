package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

func Consumer(addr []string, topic string) (err error) {

	// 1.创建消费者
	consumer, err := sarama.NewConsumer(addr, nil)
	if err != nil {
		logrus.Debugf("创建消费者失败 error:%v", err)
		return err
	}
	// 2.根据topic获取到该topic下的所有分区
	partitionList, err := consumer.Partitions(topic)
	if err != nil {
		logrus.Debugf("获取分区列表失败 error:%v", err)
		return err
	}
	// 3.根据分区创建分区消费者
	for partition := range partitionList {
		pc, err := consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			logrus.Debugf("创建分区消费者失败 error:%v", err)
			return err
		}
		// defer pc.AsyncClose() return 会调用 defer
		// 4.消费者开始消费
		go func(sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				logrus.Debugf("Partition:%d, Offset:%d, Key:%s, Value:%s", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				fmt.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				var m1 map[string]interface{}
				err := json.Unmarshal(msg.Value, &m1)
				if err != nil {
					logrus.Debugf("json解析失败 error:%v,要解析的消息为%s", err, string(msg.Value))
					continue
				}
				// 这里将map的指针传递给es  pass the pointer of map to es to store
				//es.PutLogDate(m1)
			}
		}(pc)
	}
	return
}
