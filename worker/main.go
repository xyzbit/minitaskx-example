package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xyzbit/minitaskx-contrib/discover/nacos"
	"github.com/xyzbit/minitaskx-contrib/taskrepo/mysql"
	"github.com/xyzbit/minitaskx/core/components/log"
	"github.com/xyzbit/minitaskx/core/model"
	"github.com/xyzbit/minitaskx/core/worker"
	"github.com/xyzbit/minitaskx/core/worker/executor"
	"github.com/xyzbit/minitaskx/core/worker/executor/docker"
	"github.com/xyzbit/minitaskx/core/worker/executor/goroutine"
	// "github.com/xyzbit/minitaskx/core/worker/executor/k8sjob"
	"github.com/xyzbit/minitaskx/pkg/util"
	"go.uber.org/zap/zapcore"

	example "github.com/xyzbit/minitaskx-example/pkg"
)

var (
	port int
	id   string
)

func init() {
	executor.RegisterExecutor("goroutine", goroutine.NewExecutor(newBizLogicFunction))
	executor.RegisterExecutor("docker", docker.NewExecutor())
	// executor.RegisterExecutor("k8sjob", k8sjob.NewExecutor())

	flag.StringVar(&id, "id", "", "worker id, if empty, will be auto set to discover instance id")
	flag.IntVar(&port, "port", 0, "worker port")
	flag.Parse()
}

type bizLogic struct {
	index int
}

func newBizLogicFunction() goroutine.BizLogic {
	logic := &bizLogic{}
	// return logic.Do
	return func(task *model.TaskExecParam) (bool, error) {
		logic.index++
		fmt.Printf("[%s] logic run step(%d) \n", task.TaskKey, logic.index)
		time.Sleep(2 * time.Second)
		if logic.index >= 15 {
			logic.index = 0
			return true, nil
		}
		return false, nil
	}
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

	// run http server to export metrics
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			log.Error("HTTP server error: %v", err)
		}
	}()

	// run worker
	worker := worker.NewWorker(
		id, ip, port,
		nacosDiscover, taskrepo,
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
