package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// redis数据迁移工具

var (
	cmd_params_help    = flag.Bool("help", false, "show usage")
	cmd_params_version = flag.Bool("version", false, "show version and buildtime")
	cmd_params_conf    = flag.String("conf", "./config.json", "migrate config file")
	cmd_params_verbos  = flag.Bool("verbose", false, "display details")
	version            = "0.0.1"
	buildtime          = ""
)

func show_usage() {
	fmt.Println("usage for redisDataMigrate")
	fmt.Println("\tredisDataMigrate --version")
	fmt.Println("\tredisDataMigrate --help")
	fmt.Println("\tredisDataMigrate --conf=./config.json")
	return
}

func parse_params() {
	flag.Parse()

	if *cmd_params_help {
		show_usage()
		os.Exit(0)
	} else if *cmd_params_version {
		fmt.Printf("redisDataMigrate version[%s] buildtime[%s]\n", version, buildtime)
		os.Exit(0)
	}
}

func main() {
	// 解析参数
	parse_params()

	// 初始化日志
	InitLog("migrate.log")
	gLog.Debug("Migrate Starting")
	defer gLog.Close()

	// 读取配置
	config, err := ParseConfig(*cmd_params_conf)
	if err != nil {
		gLog.Error("Parse config file failed!")
	} else {
		// 初始化连接池
		err := InitRedisConnPools(config)
		if err != nil {
			gLog.Error("Init Redis Connect Pool failed!")
		} else {
			defer gRedisConnPoolMgr.Clean()
			DoMigrate(config)
		}
	}

	gLog.Debug("Migrate Exit")

	time.Sleep(time.Second)
}
