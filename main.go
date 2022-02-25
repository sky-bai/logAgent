package main

import (
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"logAgent/common"
	"logAgent/es"
	"logAgent/etcd"
	"logAgent/kafka"
	"logAgent/tailfile"
	"net"
)

func run() (err error) {
	select {}
}

// GetOutBoundIp 获取本机IP
func GetOutBoundIp() (string, error) {
	conn, err := net.Dial("udp", "baidu.com:80")
	if err != nil {
		logrus.Errorf("get outbound ip failed,err:%v", err)
		return "", err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String(), nil
}

func main() {
	// 0.获取本机ip
	//ip, err := GetOutBoundIp()
	//if err != nil {
	//	logrus.Errorf("get outbound ip failed,err:%v", err)
	//	return
	//}

	// 1. 加载配置文件
	var configObj = new(common.Config)
	err := ini.MapTo(configObj, "./conf/config.ini")
	if err != nil {
		logrus.Error("read config file error:%v", err)
		return
	}

	logrus.Info("加载配置文件成功")

	// 2. 初始化kafka 然后后台起一个协程 去读取通道里面的数据 然后发向kafka里面
	err = kafka.Init([]string{configObj.KafkaConfig.Address}, configObj.KafkaConfig.ChanSize)
	if err != nil {
		logrus.Error("kafka init error:%v", err)
		return
	}
	logrus.Info("kafka 初始化成功")

	// 3. 初始化etcd 连接
	err = etcd.Init([]string{configObj.EtcdConfig.Address})
	if err != nil {
		logrus.Errorf("etcd init error:%v", err)
		return
	}
	logrus.Info("etcd 初始化成功")
	//collectKey := fmt.Sprintf(configObj.EtcdConfig.CollectKey, ip)

	// 4. 从etcd中获取要搜集日志的所有配置项 json格式的字符串
	allConf, err := etcd.GetConf(configObj.EtcdConfig.CollectKey)
	if err != nil {
		logrus.Errorf("etcd get conf error:%v", err)
		return
	}
	logrus.Info("初始化获取到的配置:", allConf)

	// 5.监听etcd中的配置项变化
	go etcd.WatchConf(configObj.EtcdConfig.CollectKey)

	// 5. 根据全部配置中的日志路径初始化tail tail只能获取一个日志文件地址然后创建一个tail对象
	err = tailfile.Init(allConf)
	if err != nil {
		logrus.Error("tailfile init error:%v", err)
	}

	logrus.Info("日志文件tailfile 初始化成功")

	// 6.连接ES
	err = es.Init(configObj.ESConf.Address, configObj.ESConf.Index, configObj.ESConf.GoroutineNum, configObj.ESConf.MaxSize)
	if err != nil {
		logrus.Errorf("connect es error: %v", err)
		return
	}
	logrus.Info("es 初始化成功")

	// 6.从kafka里面读取消息
	err = kafka.Consumer([]string{configObj.KafkaConfig.Address}, "configObj.KafkaConfig.Topic")
	if err != nil {
		logrus.Errorf("从kafka里面读取消息 error: %v", err)
		return
	}
	logrus.Info("从kafka里面读取消息成功")
	// 7.把数据发送给es
	run()
}

// 1.支持多个日志文件搜集
// 问题2 需要对多个日志文件进行收集
// 需要搜集多个业务线的数据
// --------------------

// 问题 1 需要动态的更新变化的日志文件
// 需要监听etcd中 日志节点下面的内容

// 问题 2 如何在etcd节点下储存多个日志文件的配置信息
// 将单个的日志文件的配置信息定义成一个结构体 多个结构体组成切片然后序列化成 json字符串

// 问题 3 在节点内容更新后，如何通知tailFile对象创建新的收集者去收集日志
// 应该是在监听etcd某一节点的时候，去更改tailFile对象的收集者

// 利用通道进行通信
// tail模块利用chan去获取数据 然后提供sendNewConf方法向外暴露一个可以写入chan的方法
// etcd模块监听节点变化那里，如果有新的配置项就向tail模块发送一个chan

// 问题 4
