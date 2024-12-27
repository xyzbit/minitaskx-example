#!/bin/bash

# 检查 Docker 是否已安装
if ! command -v docker &> /dev/null; then
    echo "Docker is not installed"
    exit 1
fi

# 启动 MySQL 容器
if ! docker inspect -f '{{.State.Running}}' minitaskx-mysql 2>/dev/null; then
    docker run --name minitaskx-mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=123456 -v $(pwd)/script/init.sql:/docker-entrypoint-initdb.d/init.sql -d mysql:8.0
    if [ $? -ne 0 ]; then
        echo "Failed to start MySQL container"
        exit 1
    fi
fi

sleep 2s

# 启动 Nacos 容器
if ! docker inspect -f '{{.State.Running}}' minitaskx-nacos 2>/dev/null; then
    docker run --name minitaskx-nacos -d -p 8848:8848 -e MODE=standalone nacos/nacos-server
    if [ $? -ne 0 ]; then
        echo "Failed to start Nacos container"
        exit 1
    fi
fi