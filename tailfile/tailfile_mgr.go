package tailfile

import (
	"github.com/sirupsen/logrus"
	"logAgent/common"
)

type TailFileMgr struct {
	tailFileMap map[string]*tailTask       // 所有的tailTask任务
	confChan    chan []common.CollectEntry // 所有配置项
	// 日志配置项变化的时候会往etcd中写入配置项，我们需要监听etcd中的配置项变化，然后从中读取配置项写入confChan
	collectEntryList []common.CollectEntry // 等待新配置的通道
}

// Watch 既然已经抽象出一个结构体去管理多个日志文件，那么我们就去定义一些这个结构体的方法
func (t *TailFileMgr) Watch() {
	for {
		// 1.因为etcd模块与tail是分开的，所以我们需要监听etcd中的配置项变化，然后从中读取配置项写入confChan
		newConf := <-t.confChan
		logrus.Infof("新的配置信息 conf:%v", newConf)

		// 如果原来有的任务，我就不动了，如果新增我就增加
		for _, value := range newConf {
			// 1.如果原来有就不动了
			if t.confIsExist(value) {
				continue
			}
			// 2.如果没有就新增
			if t.addConf(value) {
				continue
			}
			// 3.创建一个配置项

		}

		// 原来有的现在没有需要停掉
		for key, value := range TtMgr.tailFileMap {
			logrus.Infof("key:%v,value:%v", key, value)
			var find bool
			for _, conf := range newConf {
				if key == conf.Path {
					find = true
					break
				}
			}
			logrus.Infof("find:%v", find)
			if !find {
				delete(TtMgr.tailFileMap, key)
				logrus.Infof("delete conf:%v", key)
				logrus.Infof("path %v,stop", key)
				value.cancel()
			}
		}
	}
}

func (t *TailFileMgr) confIsExist(conf common.CollectEntry) bool {
	_, ok := t.tailFileMap[conf.Path]
	return ok
}

func (t *TailFileMgr) addConf(conf common.CollectEntry) bool {
	task := newTailTask(conf.Path, conf.Topic)
	// 2.为这个文件创建一个tail对象
	err := task.Init()
	if err != nil {
		logrus.Errorf("创建tailObj失败 path：%s , err: %v", conf.Path, err)
		return false
	}
	TtMgr.tailFileMap[conf.Path] = task
	logrus.Infof("创建tailObj成功 path：%s , topic: %s", conf.Path, conf.Topic)
	// 3.搜集日志
	go task.run()

	return true
}

// 问题1 当配置项发生变化时
// 1.1 如果新增了配置项，我们就需要新增一个tailTask
// 1.2 如果删除了配置项，我们就需要删除一个tailTask
// 1.3 如果更新了配置项，我们就需要更新一个tailTask
