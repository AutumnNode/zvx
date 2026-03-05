#!/bin/bash

set -e

IMAGE_NAME="zvx_backend"
CONTAINER_NAME="zvx_backend"
VERSION=$(date +v%Y%m%d%H%M)

echo "==== Step 1: 检查 kubeconfig ===="
if [ ! -f /root/.kube/config ]; then
    echo "❌ kubeconfig 不存在: /root/.kube/config"
    exit 1
fi

echo "复制 kubeconfig..."
cp /root/.kube/config ./config

echo "==== Step 2: 构建 Docker 镜像 ===="
echo "构建镜像: ${IMAGE_NAME}:${VERSION}"

docker build -t ${IMAGE_NAME}:${VERSION} .

echo "构建成功 ✔"

echo "==== Step 3: 删除旧容器 ===="
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "正在删除旧容器 ${CONTAINER_NAME}..."
    docker rm -f ${CONTAINER_NAME}
fi

echo "==== Step 4: 启动新容器 ===="
docker run -d \
    --name ${CONTAINER_NAME} \
    --network host \
    -v /root/.kube/config:/root/.kube/config:ro \
    ${IMAGE_NAME}:${VERSION}

echo "==== 完成部署 ===="
echo "镜像: ${IMAGE_NAME}:${VERSION}"
echo "容器名: ${CONTAINER_NAME}"
