FROM golang:1.25.1 AS builder

# 设置工作目录
WORKDIR /app

# 设置国内 Go 代理
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

# 复制 go.mod 和 go.sum 下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制整个项目
COPY . .

# 静态编译 Go 二进制
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# -----------------------------
# 运行阶段
# -----------------------------
FROM debian:12-slim

# 设置工作目录
WORKDIR /app

# 复制 kubeconfig（也可以挂载宿主机 config）
COPY config /app

# 复制编译好的二进制
COPY --from=builder /app/server /app/server

# 设置 Gin debug 模式和 Go stacktrace 输出和 跳过 TLS 证书验证
#ENV GIN_MODE=debug
#ENV GOTRACEBACK=all
#ENV KUBE_INSECURE_SKIP_TLS_VERIFY=true
# 入口
CMD ["/app/server"]
