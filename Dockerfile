# 第一阶段：构建应用
FROM golang:1.22.12-alpine AS builder

# 安装 git
RUN apk update && apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 克隆仓库
RUN git clone https://github.com/a15355447898a/qwen-mt-api.git qwen-mt-api
WORKDIR /app/qwen-mt-api

# 示例：设置一个构建时或默认的运行时环境变量
# 这会成为镜像的一部分，并在容器运行时生效，除非被 -e 覆盖
ENV GIN_MODE="release"
ENV DEFAULT_LOG_LEVEL="info"

RUN go mod download
RUN go build -ldflags="-s -w" -o qwen-mt-api main.go

# 第二阶段：创建精简的运行时镜像
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/qwen-mt-api/qwen-mt-api .

EXPOSE 8080

ENTRYPOINT ["./qwen-mt-api"]