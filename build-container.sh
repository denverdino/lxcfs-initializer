#!/bin/bash
docker build -t registry.cn-hangzhou.aliyuncs.com/denverdino/lxcfs:3.0.4 lxcfs-image
docker push registry.cn-hangzhou.aliyuncs.com/denverdino/lxcfs:3.0.4

docker build -t registry.cn-hangzhou.aliyuncs.com/denverdino/lxcfs-initializer:0.0.4 .
docker push registry.cn-hangzhou.aliyuncs.com/denverdino/lxcfs-initializer:0.0.4
