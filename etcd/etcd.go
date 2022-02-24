package etcd

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/client/v3"
	"logAgent/common"
	"logAgent/tailfile"
	"time"
)

var client *clientv3.Client

type collectEntry struct {
	// 每一个日志搜集项 不同的日志搜集项 有不同的分类topic
	path  string
	topic string
}

// Init 根据配置文件初始化etcd
func Init(address []string) (err error) {
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logrus.Errorf("connect to etcd failed, err:%v\n", err)
		return
	}
	return
}

// GetConf 根据key获取etcd节点下的所有的日志配置项
func GetConf(key string) (collectEntryList []common.CollectEntry, err error) {
	// 1.根据key获取到etcd中的所有的日志配置项
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.Get(ctx, key)
	if err != nil {
		logrus.Errorf("get from etcd failed, err:%v\n", err)
		return
	}
	if len(resp.Kvs) == 0 {
		logrus.Errorf("get len = 0 from etcd by key:%s\n", key)
		return
	}
	ret := resp.Kvs[0].Value
	logrus.Info("从etcd获取到的日志项配置为", string(ret))
	err = json.Unmarshal(ret, &collectEntryList)
	if err != nil {
		logrus.Errorf("unmarshal failed, err:%v\n", err)
		return nil, err
	}

	//  2.监听该节点变化
	go WatchConf(key)
	return collectEntryList, nil
}

// WatchConf 监听etcd节点下的所有的日志配置项 问题1
func WatchConf(key string) {
	for {
		watchChan := client.Watch(context.Background(), key)
		var newConf []common.CollectEntry // 监听获取到信息到新的
		// watchChan records all response for key
		// response records events
		for value := range watchChan {
			for i, i2 := range value.Events {
				logrus.Infof("%d, 变更该节点的操作为%s操作, %s, %s\n", i, i2.Type, i2.Kv.Key, i2.Kv.Value)
				// 1.获取到新的配置项
				err := json.Unmarshal(i2.Kv.Value, &newConf)
				if err != nil {
					logrus.Errorf("unmarshal failed, err:%v\n", err)
					return
				}
				// 2.告诉tailFile对象去管理新的配置项
				tailfile.SendConfChan(newConf) // tail模块如果没有处理chan里面的内容,这里会一直阻塞

			}
		}

	}
}

// etcd 只需要一个在一个节点下存放日志配置项信息，也只需要监听这一个节点的变化
