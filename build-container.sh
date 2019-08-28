#!/bin/bash
docker build -t registry.cn-hangzhou.aliyuncs.com/denverdino/lxcfs:3.1.2 lxcfs-image
docker push registry.cn-hangzhou.aliyuncs.com/denverdino/lxcfs:3.1.2

docker build -t registry.cn-hangzhou.aliyuncs.com/denverdino/lxcfs-initializer:0.0.5 .
docker push registry.cn-hangzhou.aliyuncs.com/denverdino/lxcfs-initializer:0.0.5
