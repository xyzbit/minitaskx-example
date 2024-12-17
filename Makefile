# 获取项目路径
PROJECT_PATH=$(shell pwd)

.PHONY: init
init:
	sh ./tests/init.sh

.PHONY: clean
clean:
	sh ./tests/cleanup.sh

.PHONY: worker
worker:
	go run ${PROJECT_PATH}/worker/*.go -port 9090 -id=worker-1

.PHONY: scheduler
scheduler:
	go run ${PROJECT_PATH}/scheduler/*.go -port 8080