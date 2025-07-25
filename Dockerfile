# 使用官方Go镜像作为构建环境
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o qwenmtapi .

# 使用轻量级的alpine镜像作为运行环境
FROM alpine:latest

# 安装ca-certificates以支持HTTPS请求
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/qwenmtapi .

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 设置默认环境变量
ENV AUTH_ENABLED=false
ENV API_KEY=""
ENV GIN_MODE=release

# 运行应用
CMD ["./qwenmtapi"]