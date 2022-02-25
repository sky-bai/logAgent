package es

import (
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

// 将日志数据写入es里

type ESClient struct {
	client      *elastic.Client
	index       string
	logDataChan chan interface{}
}

var esClient = &ESClient{}

func Init(addr, index string, goroutineNum, maxSize int) (err error) {
	//client, err := elastic.NewClient(elastic.SetURL("http://" + addr))
	logrus.Info("es addr:", "http://"+addr)
	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL("http://"+addr))
	if err != nil {
		logrus.Errorf("es init failed, err:%v", err)
		return err
	}
	esClient.client = client
	esClient.index = index
	esClient.logDataChan = make(chan interface{}, maxSize)
	logrus.Info("es init success")

	for i := 0; i < goroutineNum; i++ {
		go sendToES()
	}
	return
}

// PutLogDate 通过方法将外部数据放入内部结构体私有变量中
func PutLogDate(msg interface{}) {
	esClient.logDataChan <- msg

}

func sendToES() {
	for msg := range esClient.logDataChan {
		put1, err := esClient.client.Index().
			Index(esClient.index).
			BodyJson(msg).
			Do(context.Background())
		if err != nil {
			panic(err)
		}
		logrus.Infof("往es发送消息成功 Indexed user %s to index %s ,type %s\n", put1.Id, put1.Index, put1.Type)
	}
}
