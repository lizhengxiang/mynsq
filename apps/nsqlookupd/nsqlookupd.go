package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/judwhite/go-svc/svc"
	"github.com/mreiferson/go-options"
	"github.com/nsqio/nsq/internal/version"
	"github.com/nsqio/nsq/nsqlookupd"
)

func nsqlookupdFlagSet(opts *nsqlookupd.Options) *flag.FlagSet {
	//NewFlagSet创建一个新的、名为name，采用errorHandling为错误处理策略的FlagSet。
	flagSet := flag.NewFlagSet("nsqlookupd", flag.ExitOnError)
	//String用指定的名称、默认值、使用信息注册一个string类型flag。返回一个保存了该flag的值的指针。
	flagSet.String("config", "", "path to config file")
	//Bool用指定的名称、默认值、使用信息注册一个bool类型flag。返回一个保存了该flag的值的指针。
	flagSet.Bool("version", false, "print version string")

	flagSet.String("log-level", "info", "set log verbosity: debug, info, warn, error, or fatal")
	flagSet.String("log-prefix", "[nsqlookupd] ", "log message prefix")
	flagSet.Bool("verbose", false, "deprecated in favor of log-level")

	flagSet.String("tcp-address", opts.TCPAddress, "<addr>:<port> to listen on for TCP clients")
	flagSet.String("http-address", opts.HTTPAddress, "<addr>:<port> to listen on for HTTP clients")
	flagSet.String("broadcast-address", opts.BroadcastAddress, "address of this lookupd node, (default to the OS hostname)")
	//Duration用指定的名称、默认值、使用信息注册一个time.Duration类型flag。返回一个保存了该flag的值的指针。
	flagSet.Duration("inactive-producer-timeout", opts.InactiveProducerTimeout, "duration of time a producer will remain in the active list since its last ping")
	flagSet.Duration("tombstone-lifetime", opts.TombstoneLifetime, "duration of time a producer will remain tombstoned if registration remains")

	return flagSet
}

type program struct {
	nsqlookupd *nsqlookupd.NSQLookupd
}

func main() {
	prg := &program{}
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}

func (p *program) Init(env svc.Environment) error {
	if env.IsWindowsService() {
		//Dir返回路径除去最后一个路径元素的部分，即该路径最后一个元素所在的目录。在使用Split去掉最后一个元素后，
		// 会简化路径并去掉末尾的斜杠。如果路径是空字符串，会返回"."；如果路径由1到多个斜杠后跟0到多个非斜杠字符组成，
		// 会返回"/"；其他任何情况下都不会返回以斜杠结尾的路径。
		dir := filepath.Dir(os.Args[0])
		//Chdir将当前工作目录修改为dir指定的目录。如果出错，会返回*PathError底层类型的错误。
		return os.Chdir(dir)
	}
	return nil
}

func (p *program) Start() error {
	opts := nsqlookupd.NewOptions()
	flagSet := nsqlookupdFlagSet(opts)
	//从os.Args[1:]中解析注册的flag。必须在所有flag都注册好而未访问其值时执行。未注册却使用flag -help时，会返回ErrHelp。
	flagSet.Parse(os.Args[1:])
	//返回已经f中已注册flag的Flag结构体指针；如果flag不存在的话，返回nil。+
	if flagSet.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("nsqlookupd"))
		os.Exit(0)
	}

	var cfg map[string]interface{}
	configFile := flagSet.Lookup("config").Value.String()
	if configFile != "" {
		_, err := toml.DecodeFile(configFile, &cfg)
		if err != nil {
			log.Fatalf("ERROR: failed to load config file %s - %s", configFile, err.Error())
		}
	}

	options.Resolve(opts, flagSet, cfg)
	daemon := nsqlookupd.New(opts)

	daemon.Main()
	p.nsqlookupd = daemon
	return nil
}

func (p *program) Stop() error {
	if p.nsqlookupd != nil {
		p.nsqlookupd.Exit()
	}
	return nil
}
