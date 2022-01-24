package common

type CollectEntry struct {
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

// 将日志记录在etcd下面的节点下
// 这时候etcd下的节点和日志文件的路径是一一对应的
// 为此构建一个结构体

// 所以我以后可以以这个为思路 将有关联属性的两个值放入一个结构体里面
