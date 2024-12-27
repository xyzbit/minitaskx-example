/*
 * @Author: xiaoyan 1425895909@qq.com
 * @Date: 2024-12-09 22:41:50
 * @LastEditors: xiaoyan 1425895909@qq.com
 * @LastEditTime: 2024-12-09 23:07:16
 * @FilePath: /minitaskx/example/worker/main.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xyzbit/minitaskx-contrib/discover/nacos"
	"github.com/xyzbit/minitaskx-contrib/taskrepo/mysql"
	example "github.com/xyzbit/minitaskx-example/pkg"
	"github.com/xyzbit/minitaskx/core/components/log"
	"github.com/xyzbit/minitaskx/core/model"
	"github.com/xyzbit/minitaskx/core/worker"
	"github.com/xyzbit/minitaskx/core/worker/executor"
	"github.com/xyzbit/minitaskx/core/worker/executor/goroutine"
	"github.com/xyzbit/minitaskx/core/worker/infomer"
	"github.com/xyzbit/minitaskx/pkg/util"
	"go.uber.org/zap/zapcore"
)

var (
	port int
	id   string
)

func init() {
	executor.RegisterExecutor("goroutine", goroutine.NewExecutor(new(bizLogic).Do))

	flag.StringVar(&id, "id", "", "worker id, if empty, will be auto set to discover instance id")
	flag.IntVar(&port, "port", 0, "worker port")
	flag.Parse()
}

type bizLogic struct {
	index int
}

func (b *bizLogic) Do(task *model.Task) (bool, error) {
	b.index++
	fmt.Printf("[%s] logic run step(%d) \n", task.TaskKey, b.index)
	time.Sleep(2 * time.Second)
	if b.index >= 15 {
		b.index = 0
		return true, nil
	}
	return false, nil
}

func main() {
	// init dependecy
	ip, err := util.GlobalUnicastIPString()
	if err != nil {
		panic(err)
	}
	nacosDiscover, err := nacos.NewNacosDiscover(nacos.NacosConfig{
		IpAddr:      "localhost",
		Port:        8848,
		ServiceName: "example-workers",
		GroupName:   "default",
		ClusterName: "default",
		LogLevel:    "debug",
	})
	if err != nil {
		log.Panic("创建 Nacos 客户端失败: %v", err)
	}
	taskrepo := mysql.NewTaskRepo(example.NewGormDB())
	logger := newLogger(ip)

	// new worker
	indexer := infomer.NewIndexer(&executor.Global{}, 5*time.Minute)
	worker := worker.NewWorker(
		id, ip, port,
		nacosDiscover,
		infomer.New(indexer, taskrepo, logger),
		worker.WithLogger(logger),
		worker.WithTriggerResync(1*time.Second),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-quit
		cancel()
	}()

	if err := worker.Run(ctx); err != nil {
		log.Panic("启动 Worker 失败: %v", err)
	}
}

func newLogger(ip string) log.Logger {
	var field zapcore.Field
	if id == "" {
		field = zapcore.Field{Key: "worker_id", String: fmt.Sprintf("%s:%d", ip, port), Type: zapcore.StringType}
	} else {
		field = zapcore.Field{Key: "worker_id", String: id, Type: zapcore.StringType}
	}
	return example.NewLogger(field)
}
