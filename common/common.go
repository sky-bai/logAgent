package common

type CollectEntry struct {
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

// 将日志记录在etcd下面的节点下
// 这时候etcd下的节点和日志文件的路径是一一对应的
// 为此构建一个结构体

// 所以我以后可以以这个为思路 将有关联属性的两个值放入一个结构体里面

type Config struct {
	KafkaConfig   `ini:"kafka"`
	CollectConfig `ini:"collect"`
	EtcdConfig    `ini:"etcd"`
	ESConf        `ini:"es"`
}

type KafkaConfig struct {
	Address  string `ini:"address"`
	Topic    string `ini:"topic"`
	ChanSize int64  `ini:"chan_size"`
}
type CollectConfig struct {
	LogFilePath string `ini:"logfile_path"`
}
type EtcdConfig struct {
	Address string `ini:"address"`
	// 下面这个管理了多个日志文件和对应的topic 在etcd一个节点下面储存日志文件地址和对应的topic
	CollectKey string `ini:"collect_key"`
}

type ESConf struct {
	Address      string `ini:"address"`
	Index        string `ini:"index"`
	MaxSize      int    `ini:"max_chan_size"`
	GoroutineNum int    `ini:"goroutine_num"`
}
