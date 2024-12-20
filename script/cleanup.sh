#!/bin/bash

# 清理所有资源

docker stop minitaskx-mysql
docker rm minitaskx-mysql
docker stop minitaskx-nacos
docker rm minitaskx-nacos