package tailfile

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
	"logAgent/common"
	"logAgent/kafka"
	"strings"
	"time"
)

// confChan 保存每一个配置更新后的配置信息
var confChan chan []common.CollectEntry

type tailTask struct {
	path  string
	topic string
	// TailObj 根据这个路径 去管理这个日志文件 它里面管理着这个日志文件的内容 日志文件实例化
	tailObj *tail.Tail
	// 每次会启一个协程去读取日志文件的内容 但是有时候需要kill掉这个协程
	ctx    context.Context
	cancel context.CancelFunc
}

// newTailTask 创建一个带有path 和 topic 的对象
func newTailTask(path, topic string) *tailTask {
	// 每次创建一个新的tailTask 对象，它都会以协程的方式运行。
	// 建立一个可以取消的上下文 协程执行完毕后会自动取消
	ctx, cancel := context.WithCancel(context.Background())
	return &tailTask{
		path:   path,
		topic:  topic,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Init 为tailTask 创建一个tail对象
func (t *tailTask) Init() (err error) {
	config := tail.Config{
		ReOpen:    true, // 日志分割后重新打开
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	}
	file, err := tail.TailFile(t.path, config)
	t.tailObj = file
	if err != nil {
		logrus.Errorf("tail file failed, err:%v", err)
		return err
	}

	return
}

// run 读取tailObj里面的日志文件的内容 然后向kafka里面发送数据
func (t *tailTask) run() {
	logrus.Infof("start tail file:%s", t.path)
	// 读取对象里面的文件然后发给kafka

	select {
	case <-t.ctx.Done():
		t.tailObj.Cleanup()
		logrus.Infof("tail file:%s goroutine exit", t.path)
		return
	case line, ok := <-t.tailObj.Lines:
		// 循环读数据

		if !ok {
			logrus.Error("tail file close reopen, filename:%s\n", t.path)
			time.Sleep(time.Second)
		}
		//strings.Trim(line.Text, "\r")
		if len(strings.Trim(line.Text, "\r")) == 0 {
			logrus.Info("出现空格 跳过")
		}

		fmt.Println("这一行的内容", line.Text)
		msg := &sarama.ProducerMessage{}
		msg.Topic = "test"
		msg.Value = sarama.StringEncoder(line.Text)
		// 将数据放入 chan
		kafka.GetMsgChan() <- msg
	}

}

// TtMgr 定一个管理所有tail任务的结构体
var TtMgr *TailFileMgr

// Init 根据文件path初始化tail allConf里面有多个文件的路径 和对应的topic 现在需要针对一个文件创建一个对应的tailObj 就是可以读这个文件的一个对象
func Init(allConf []common.CollectEntry) (err error) {
	//
	TtMgr = &TailFileMgr{
		tailFileMap:      make(map[string]*tailTask),
		collectEntryList: allConf,
		confChan:         make(chan []common.CollectEntry),
	}

	for _, eachLogConf := range allConf {
		// 1.创建一个配置项
		task := newTailTask(eachLogConf.Path, eachLogConf.Topic)
		// 2.为这个文件创建一个tail对象
		err = task.Init()
		if err != nil {
			logrus.Errorf("创建tailObj失败 path：%s , err: %v", eachLogConf.Path, err)
			return
		}
		TtMgr.tailFileMap[eachLogConf.Path] = task
		logrus.Infof("创建tailObj成功 path：%s , topic: %s", eachLogConf.Path, eachLogConf.Topic)
		// 3.搜集日志
		go task.run()
	}

	// 从全局chan里面获取新的配置
	//confChan = make(chan []common.CollectEntry)
	newConf := <-TtMgr.confChan
	logrus.Infof("收到新的配置: %v", newConf)

	go TtMgr.Watch()

	return
}

// SendConfChan 将读取到的信息放入全局保存配置信息的confChan中
func SendConfChan(newConf []common.CollectEntry) {
	TtMgr.confChan <- newConf
}

// 将本模块的全局chan通过方法暴露给其他模块，其他模块可以通过该方法向chan里面发送信息。

// 问题1 tail从chan里面获取到新的配置后如何管理tailTask

// 将需要管理的数据抽象出一个大的结构体来管理
