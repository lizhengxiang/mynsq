package nsqlookupd

import (
	"log"
	"os"
	"time"
	"github.com/nsqio/nsq/internal/lg"
)

type Options struct {
	LogLevel  string `flag:"log-level"`
	LogPrefix string `flag:"log-prefix"`
	//是否开启啰嗦模式，开启后，会打很多LOG，一般在调试或定位问题时使用。
	Verbose   bool   `flag:"verbose"` // for backwards compatibility
	Logger    Logger
	logLevel  lg.LogLevel // private, not really an option
	// TCP 监听地址
	TCPAddress       string `flag:"tcp-address"`
	//HTTP监听地址
	HTTPAddress      string `flag:"http-address"`
	//这个lookup节点的对外地址(BroadcastAddress广播地址)
	BroadcastAddress string `flag:"broadcast-address"`
	//producer的交互超时时间，默认是5分钟。就是说，如果5分钟内nsqlookupd没有收到producer的PING(类似心跳包),则会认为producer已掉线。
	InactiveProducerTimeout time.Duration `flag:"inactive-producer-timeout"`
	//字面直译是墓碑时间
	//在nsqadmin的http界面中访问/tombstone_topic_producer URL时，nsqlookupd会给producer TombstoneLifetime长度的时间来注销
	//默认为45秒，在这45秒内，producer不会再被任何consumer通过nsqadmin的/lookup操作找到，同时producer还会进行删除topic等操作。
	//45秒之后，producer就会与nsqlookupd断开连接，同时通过nsqlookupd TCP连接中的UNREGISTER操作在数据记录中把该producer删除。
	TombstoneLifetime       time.Duration `flag:"tombstone-lifetime"`
}

//
//新建nsqlookupdOptions类型的变量的指针
//

func NewOptions() *Options {
	//获取主机名
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	return &Options{
		LogPrefix:        "[nsqlookupd] ",
		LogLevel:         "info",
		TCPAddress:       "0.0.0.0:4160",
		HTTPAddress:      "0.0.0.0:4161",
		BroadcastAddress: hostname,
		//5分钟超时
		InactiveProducerTimeout: 300 * time.Second,
		//45秒
		TombstoneLifetime:       45 * time.Second,
	}
}
