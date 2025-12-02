# Thingsboard RxProbe 探针系统

一个用于监控 Thingsboard 平台基础设施组件状态的探针系统，支持通过 Web 页面配置监控目标，并在异常时通过企业微信进行告警。

## 功能特性

- 🔍 **多组件监控**：支持 PostgreSQL、Cassandra、Redis、Kafka、HTTP、TCP 等多种探针类型
- 🎯 **Web 配置界面**：通过页面手动配置需要监控的组件
- ⚡ **实时监控**：可配置的探测间隔和超时时间
- 🔔 **企业微信告警**：异常状态自动发送企业微信通知
- 📊 **仪表盘**：直观展示各组件健康状态
- 🔐 **用户认证**：JWT 认证，支持多用户

## 支持的探针类型

| 类型 | 说明 | 监控指标 |
|-----|------|---------|
| PostgreSQL | PostgreSQL 数据库 | 连接状态、活跃连接数、复制延迟、慢查询 |
| Cassandra | Cassandra 集群 | 节点状态、读写延迟、集群健康 |
| Redis | Redis 缓存 | 连接状态、内存使用、主从状态 |
| Kafka | Kafka 消息队列 | Broker 状态、消费延迟、分区状态 |
| HTTP | HTTP 服务 | 响应状态码、响应时间、内容检查 |
| TCP | TCP 端口 | 连接状态、响应时间 |

## 快速开始

### 1. 环境要求

- Go 1.21+
- SQLite（默认）或 PostgreSQL

### 2. 安装

```bash
# 克隆项目
git clone https://github.com/your-repo/thingsboard-rxprobe.git
cd thingsboard-rxprobe

# 下载依赖
go mod tidy

# 编译
make build

# 或直接运行
make run
```

### 3. 配置

编辑 `configs/config.yaml`：

```yaml
server:
  host: 0.0.0.0
  port: 8088

database:
  driver: sqlite
  dbname: rxprobe.db

alerter:
  wecom:
    enabled: true
    webhook_url: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"
```

### 4. 运行

```bash
# 直接运行
./build/rxprobe -config configs/config.yaml

# 或使用 Docker
docker-compose up -d
```

### 5. 访问

- API 地址：http://localhost:8088
- 默认账号：admin / admin123

## API 文档

### 认证

```bash
# 登录
curl -X POST http://localhost:8088/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 探测目标管理

```bash
# 获取支持的探针类型
curl http://localhost:8088/api/v1/probe/types \
  -H "Authorization: Bearer <token>"

# 获取探针配置 Schema
curl http://localhost:8088/api/v1/probe/schema/postgresql \
  -H "Authorization: Bearer <token>"

# 创建监控目标
curl -X POST http://localhost:8088/api/v1/targets \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ThingsBoard PostgreSQL",
    "type": "postgresql",
    "config": {
      "host": "localhost",
      "port": 5432,
      "username": "postgres",
      "password": "password",
      "database": "thingsboard"
    },
    "interval_seconds": 30,
    "timeout_seconds": 5,
    "enabled": true
  }'

# 测试探测目标
curl -X POST http://localhost:8088/api/v1/probe/test \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "redis",
    "config": {
      "mode": "standalone",
      "host": "localhost",
      "port": 6379
    },
    "timeout_seconds": 5
  }'

# 获取目标列表
curl http://localhost:8088/api/v1/targets \
  -H "Authorization: Bearer <token>"

# 获取仪表盘概览
curl http://localhost:8088/api/v1/dashboard/summary \
  -H "Authorization: Bearer <token>"
```

## 项目结构

```
thingsboard-rxprobe/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── configs/
│   └── config.yaml              # 配置文件
├── internal/
│   ├── api/                     # HTTP API
│   │   ├── handler/             # 请求处理器
│   │   ├── middleware/          # 中间件
│   │   └── router.go            # 路由配置
│   ├── config/                  # 配置管理
│   ├── model/                   # 数据模型
│   ├── repository/              # 数据访问层
│   ├── service/                 # 业务逻辑层
│   ├── prober/                  # 探针实现
│   ├── alerter/                 # 告警通道
│   └── scheduler/               # 调度引擎
├── pkg/
│   ├── logger/                  # 日志组件
│   └── database/                # 数据库组件
├── web/                         # 前端页面
├── Dockerfile
├── docker-compose.yaml
├── Makefile
└── README.md
```

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|-------|------|-------|
| DB_PASSWORD | 数据库密码 | - |
| JWT_SECRET | JWT 密钥 | rxprobe-secret-key-change-me |
| WECOM_WEBHOOK_URL | 企业微信 Webhook URL | - |

### 企业微信告警配置

1. 在企业微信群中添加群机器人
2. 复制机器人的 Webhook URL
3. 配置到 `configs/config.yaml` 或环境变量 `WECOM_WEBHOOK_URL`

## 开发

```bash
# 安装依赖
make deps

# 运行开发模式
make run

# 运行测试
make test

# 代码格式化
make fmt
```

## License

Apache License 2.0

