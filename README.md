# QWen Translation API

一个基于千问大模型的多格式翻译API服务，兼容DeepL和DeepLX API格式。

## 功能特性

- 🌐 支持多种API格式：DeepL、DeepLX、原生格式
- 🔐 可选的API密钥认证
- 🚀 高并发支持，内置限流和重试机制
- 🐳 Docker容器化部署
- 📊 健康检查和监控支持

## 支持的API格式

### DeepLX 格式
```bash
POST /translate
{
    "source_lang": "EN",
    "target_lang": "ZH", 
    "text": "Hello world"
}

# 响应
{
    "code": 200,
    "id": 1753422603925,
    "data": "你好，世界"
}
```

### DeepL 格式
```bash
POST /v2/translate
{
    "text": ["Hello world"],
    "source_lang": "EN",
    "target_lang": "ZH"
}

# 响应
{
    "translations": [
        {
            "detected_source_language": "EN",
            "text": "你好，世界"
        }
    ]
}
```

### 原生格式
```bash
POST /api/translate
{
    "text": ["Hello world"],
    "source_lang": "EN", 
    "target_lang": "ZH"
}
```

## 语言支持

| 语言代码 | 中文名称 | 英文名称 |
|---------|---------|----------|
| EN      | 英语    | English  |
| ZH      | 简体中文 | Chinese  |
| auto    | 自动检测 | Auto     |

## 认证方式

服务支持多种API密钥认证方式：

1. **Authorization头部** (DeepL格式)
   ```
   Authorization: DeepL-Auth-Key sk-your-api-key
   ```

2. **Bearer Token**
   ```
   Authorization: Bearer sk-your-api-key
   ```

3. **X-API-Key头部**
   ```
   X-API-Key: sk-your-api-key
   ```

4. **查询参数**
   ```
   POST /translate?api_key=sk-your-api-key
   ```

## 快速开始

### 使用Docker Compose（推荐）

1. 克隆项目
```bash
git clone <repository-url>
cd qwenmtapi
```

2. 配置环境变量（可选）
```bash
# 编辑docker-compose.yml中的环境变量
# 或创建.env文件
echo "AUTH_ENABLED=true" > .env
echo "API_KEY=sk-your-secret-key" >> .env
```

3. 启动服务
```bash
docker-compose up -d
```

4. 测试服务
```bash
# 健康检查
curl http://localhost:8080/health

# 翻译测试
curl -X POST http://localhost:8080/translate \
  -H "Content-Type: application/json" \
  -H "Authorization: DeepL-Auth-Key sk-your-secret-key" \
  -d '{"source_lang":"EN","target_lang":"ZH","text":"Hello world"}'
```

### 手动编译运行

1. 确保Go版本 >= 1.21
```bash
go version
```

2. 安装依赖
```bash
go mod download
```

3. 运行服务
```bash
# 不启用认证
go run main.go

# 启用认证
AUTH_ENABLED=true API_KEY=sk-test123 go run main.go
```

## 环境变量配置

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `AUTH_ENABLED` | `false` | 是否启用API密钥认证 |
| `API_KEY` | `""` | 单个API密钥 |
| `API_KEYS` | `""` | 多个API密钥（逗号分隔） |
| `GIN_MODE` | `debug` | Gin框架模式 (debug/release) |
| `TZ` | `UTC` | 时区设置 |

## API端点

| 端点 | 方法 | 格式 | 说明 |
|------|------|------|------|
| `/` | GET | - | 服务信息和端点列表 |
| `/health` | GET | - | 健康检查 |
| `/translate` | POST | DeepLX | DeepLX兼容格式 |
| `/v2/translate` | POST | DeepL | DeepL兼容格式 |
| `/api/translate` | POST | 原生 | 原生API格式 |

## 性能优化

- 内置并发限制（最大2个并发请求）
- 自动重试机制（最多3次重试）
- 连接池复用
- 请求去重和缓存（计划中）

## 部署建议

### 生产环境

1. 启用认证
```bash
AUTH_ENABLED=true
API_KEYS=sk-key1,sk-key2,sk-key3
```

2. 使用反向代理（Nginx/Traefik）
3. 设置合适的资源限制
4. 启用HTTPS
5. 配置日志收集

### 监控和日志

- 健康检查端点：`GET /health`
- 应用日志输出到stdout
- 支持结构化日志格式

## 开发

### 项目结构
```
.
├── main.go                 # 主入口
├── internal/
│   ├── controller/         # 控制器层
│   ├── service/           # 业务逻辑层
│   ├── model/             # 数据模型
│   └── middleware/        # 中间件
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
└── README.md
```

### 贡献指南

1. Fork项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交Issue。