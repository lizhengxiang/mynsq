package nsqlookupd

import (
	"log"
	"net"
	"os"
	"sync"

	"github.com/nsqio/nsq/internal/http_api"
	"github.com/nsqio/nsq/internal/lg"
	"github.com/nsqio/nsq/internal/protocol"
	"github.com/nsqio/nsq/internal/util"
	"github.com/nsqio/nsq/internal/version"
)

type NSQLookupd struct {
	//读写锁
	sync.RWMutex
	//在文件nsqlookupd\options.go中定义，记录NSQLookupd的配置信息
	opts         *Options
	//使用上面的tcpAddr建立的Listener
	tcpListener  net.Listener
	//使用上面的httpAddr建立的Listener
	httpListener net.Listener
	//在util\wait_group_wrapper.go文件中定义,与sync.WaitGroup相关，用于线程同步。
	waitGroup    util.WaitGroupWrapper
	//在nsqlookupd\registration_db.go文件中定义，看字面意思DB(database)就可知道这涉及到数据的存取
	DB           *RegistrationDB
}

//根据配置的nsqlookupdOptions创建一个NSQLookupd的实例
func New(opts *Options) *NSQLookupd {
	if opts.Logger == nil {
		opts.Logger = log.New(os.Stderr, opts.LogPrefix, log.Ldate|log.Ltime|log.Lmicroseconds)
	}
	n := &NSQLookupd{
		opts: opts,
		DB:   NewRegistrationDB(),
	}

	var err error
	opts.logLevel, err = lg.ParseLogLevel(opts.LogLevel, opts.Verbose)
	if err != nil {
		n.logf(LOG_FATAL, "%s", err)
		os.Exit(1)
	}

	n.logf(LOG_INFO, version.String("nsqlookupd"))
	return n
}

//
//Main函数，启动时首先执行本函数
//补注：阅读options.go时，发现nsqlookupd启动时，首先运行的并不是这个Main方法。而是apps\nsqlookupd\nsqlookupd.go里的main方法，这个下篇文章会提到。
//
func (l *NSQLookupd) Main() {
	//定义了Context的实例，Context在nsqlookupd\context.go文件中定义，其中只包含了一个nsqlookupd的指针,注意花括号里是字符L的小写，不是数字一.
	ctx := &Context{l}
	//监听TCP
	tcpListener, err := net.Listen("tcp", l.opts.TCPAddress)
	if err != nil {
		l.logf(LOG_FATAL, "listen (%s) failed - %s", l.opts.TCPAddress, err)
		os.Exit(1)
	}
	l.Lock()
	//把Listener存在NSQLookupd的struct里
	l.tcpListener = tcpListener
	l.Unlock()
	//创建tcpServer的实例，tcpServer在nsqlookupd\tcp.go文件中定义，用于处理TCP连接中接收到的数据。通过前面阅读知道，context里只是一个NSQLookupd类型的指针。
	tcpServer := &tcpServer{ctx: ctx}
	//调用util.TCPServer方法（在util\tcp_server.go中定义）开始接收监听并注册handler。
	// 传入的两个参数第一个是tcpListener
	//第二个tcpServer实现了util\tcp_server.go中定义的TCPHandler接口。
	//tcpServer接到TCP数据时，会调用其Handle方法(见nsqlookupd\tcp.go)来处理。
	//此处为何要用到waitGroup，目前还比较迷糊
	l.waitGroup.Wrap(func() {
		protocol.TCPServer(tcpListener, tcpServer, l.logf)
	})
	//监听HTTP
	httpListener, err := net.Listen("tcp", l.opts.HTTPAddress)
	if err != nil {
		l.logf(LOG_FATAL, "listen (%s) failed - %s", l.opts.HTTPAddress, err)
		os.Exit(1)
	}
	l.Lock()
	//把Listener存在NSQLookupd的struct里
	l.httpListener = httpListener
	l.Unlock()
	httpServer := newHTTPServer(ctx)
	l.waitGroup.Wrap(func() {
		http_api.Serve(httpListener, httpServer, "HTTP", l.logf)
	})

}

func (l *NSQLookupd) RealTCPAddr() *net.TCPAddr {
	l.RLock()
	defer l.RUnlock()
	return l.tcpListener.Addr().(*net.TCPAddr)
}

func (l *NSQLookupd) RealHTTPAddr() *net.TCPAddr {
	l.RLock()
	defer l.RUnlock()
	return l.httpListener.Addr().(*net.TCPAddr)
}
//退出 关闭两个Listener

func (l *NSQLookupd) Exit() {
	if l.tcpListener != nil {
		l.tcpListener.Close()
	}

	if l.httpListener != nil {
		l.httpListener.Close()
	}
	l.waitGroup.Wait()
}
