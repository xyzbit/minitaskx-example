# 获取项目路径
PROJECT_PATH=$(shell pwd)

.PHONY: init
init:
	sh ./script/init.sh

.PHONY: clean
clean:
	sh ./script/cleanup.sh

.PHONY: worker
worker:
	go build -o miniworker ${PROJECT_PATH}/worker/*.go && ./miniworker -port 9090 -id=worker-1

.PHONY: scheduler
scheduler:
	go build -o minischeduler ${PROJECT_PATH}/scheduler/*.go && ./minischeduler -port 8080

.PHONY: build
build:
	go build -o minitaskx -ldflags "-s -w -X main.Version=$(git show -s --format=%h) -X main.Build=$(date -u +%FT%TZ)" main.go