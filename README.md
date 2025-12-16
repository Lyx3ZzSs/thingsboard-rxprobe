# RxProbe 生产环境通用监控工具

一个用于监控生产环境基础设施组件状态的探针系统，支持通过 Web 页面配置监控目标，并在异常时通过多种通知渠道进行告警。

## 功能特性

- **多组件监控**：支持 PostgreSQL、Cassandra、Redis、Kafka、HTTP、TCP、Ping、CPU 等多种探针类型
- **Web 配置界面**：通过直观的 Web 界面配置和管理监控目标
- **实时监控**：可配置的探测间隔和超时时间，实时监控服务健康状态
- **多渠道告警**：支持企业微信等多种通知渠道，可配置告警消息模板
- **可视化仪表盘**：直观展示各组件健康状态、历史数据和统计信息
- **用户认证**：JWT 认证，支持多用户管理
- **历史记录**：保存探测结果历史，支持数据分析和趋势查看
- **自动清理**：自动清理过期的探测结果和告警记录

## 支持的探针类型

| 类型 | 说明 | 监控指标 | 特殊说明 |
|-----|------|---------|---------|
| PostgreSQL | PostgreSQL 数据库 | 连接状态、查询响应时间 | 支持 SSL 连接 |
| Cassandra | Cassandra 集群 | 节点状态、连接可用性 | 支持认证 |
| Redis | Redis 缓存 | 连接状态、PING 响应 | 支持单机、哨兵、集群模式 |
| Kafka | Kafka 消息队列 | Broker 状态、集群可用性 | 支持 SASL/SCRAM 和 TLS |
| HTTP | HTTP 服务 | 响应状态码、响应时间 | 支持自定义请求方法和 Headers |
| TCP | TCP 端口 | 连接状态、响应时间 | 基础端口连通性检查 |
| Ping | 网络连通性 | 丢包率、延迟统计 | 固定发送 4 个包，每个超时 3 秒 |
| CPU | CPU 监控 | CPU 占用率 | 实时采样服务器 CPU 使用率 |

## 快速开始

### 1. 环境要求

- Go 1.21+
- Node.js 18+（用于构建前端）
- SQLite（默认）或 PostgreSQL

### 2. 安装

```bash
# 克隆项目
git clone https://github.com/your-repo/thingsboard-rxprobe.git
cd thingsboard-rxprobe

# 后端依赖
go mod tidy

# 前端依赖
cd web
npm install
cd ..

# 构建前端
cd web
npm run build
cd ..

# 编译后端
make build

# 或直接运行（开发模式）
make run
```

### 3. 配置

编辑 `configs/config.yaml`：

```yaml
server:
  host: 0.0.0.0
  port: 8088
  mode: release  # debug/release

database:
  driver: sqlite   # postgres/sqlite
  dbname: rxprobe.db  # SQLite 使用文件名，PostgreSQL 使用数据库名

scheduler:
  default_interval: 30        # 默认探测间隔（秒）
  default_timeout: 5          # 默认超时时间（秒）
  alert_threshold: 3          # 告警阈值（连续失败次数）
  result_retention_days: 30   # 探测结果保留天数

log:
  level: info     # debug/info/warn/error
  format: console # console/json

auth:
  jwt_secret: "your-secret-key-change-in-production"  # JWT密钥，生产环境请修改
  jwt_expiry: 168h  # JWT过期时间，默认7天
```

### 4. 运行

```bash
# 直接运行
./build/rxprobe -config configs/config.yaml

# 或使用 Docker
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 5. 访问

- Web 界面：http://localhost:8088
- API 地址：http://localhost:8088/api/v1
- 默认账号：admin / admin123

首次登录后，系统会提示修改默认密码。

## 使用指南

### 配置监控目标

1. 登录 Web 界面
2. 进入「监控服务」页面
3. 点击「新增监控服务」
4. 选择探针类型，填写连接配置
5. 设置探测间隔和超时时间（部分类型如 CPU、Ping 有固定配置）
6. 选择通知渠道（可选）
7. 点击「测试连接」验证配置
8. 保存并启用监控

### 配置通知渠道

1. 进入「通知渠道」页面
2. 点击「新增通知渠道」
3. 选择渠道类型（目前支持企业微信）
4. 填写 Webhook URL
5. 配置消息模板（可选，支持模板变量）
6. 启用渠道

### 查看告警

1. 进入「告警记录」页面
2. 查看所有告警事件
3. 可以按状态筛选（告警中/已恢复）
4. 可以静默告警（临时屏蔽）

## API 文档

### 认证

```bash
# 登录
curl -X POST http://localhost:8088/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 响应
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin"
    }
  }
}
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
    "name": "生产环境 PostgreSQL",
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
│   │   ├── middleware/          # 中间件（认证、CORS）
│   │   └── router.go            # 路由配置
│   ├── config/                  # 配置管理
│   ├── model/                   # 数据模型
│   ├── repository/              # 数据访问层
│   ├── service/                 # 业务逻辑层
│   ├── prober/                  # 探针实现
│   ├── scheduler/               # 调度引擎
│   └── service/                 # 业务服务
│       ├── probe_service.go     # 探测服务
│       ├── alert_service.go     # 告警服务
│       ├── auth_service.go      # 认证服务
│       └── cleanup_service.go   # 清理服务
├── pkg/
│   ├── logger/                  # 日志组件
│   └── database/                # 数据库组件
├── web/                         # 前端页面
│   ├── src/
│   │   ├── api/                 # API 客户端
│   │   ├── components/          # Vue 组件
│   │   ├── views/               # 页面视图
│   │   ├── router/              # 路由配置
│   │   └── store/               # 状态管理
│   └── dist/                    # 构建产物
├── Dockerfile
├── docker-compose.yaml
├── Makefile
├── DESIGN.md                    # 设计文档
└── README.md
```

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|-------|------|-------|
| DB_PASSWORD | 数据库密码 | - |
| JWT_SECRET | JWT 密钥 | config.yaml 中的值 |
| WECOM_WEBHOOK_URL | 企业微信 Webhook URL | - |

### 企业微信告警配置

1. 在企业微信群中添加群机器人
2. 复制机器人的 Webhook URL
3. 在「通知渠道」页面添加企业微信通知渠道
4. 配置 Webhook URL 和消息模板
5. 在监控目标中选择该通知渠道

### 探针类型特殊说明

#### CPU 监控
- 采样时长即为检测耗时，无需设置探测间隔
- 超时时间自动设置为采样时长 + 5 秒

#### Ping 探测
- 固定发送 4 个数据包，每个超时 3 秒
- 总超时时间固定为 15 秒，无需手动设置

#### Kafka 监控
- 默认探测间隔 30 秒，超时 10 秒
- 支持 SASL/SCRAM 认证和 TLS 加密

## 开发

```bash
# 安装依赖
make deps

# 运行开发模式（后端）
make run

# 运行前端开发服务器
cd web
npm run dev

# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint

# 构建 Docker 镜像
make docker-build

# 运行 Docker 容器
make docker-run
```

## 技术栈

### 后端
- Go 1.21+
- Gin Web 框架
- GORM 数据库 ORM
- Cron 定时任务调度
- Zap 日志库

### 前端
- Vue 3
- Vite
- Tailwind CSS
- Chart.js（图表）
- Vue Router
- Axios

### 数据库
- SQLite（默认）
- PostgreSQL（可选）

## 许可证

Apache License 2.0

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更新日志

### v2.0.0
- 完整的 Web 管理界面
- 支持多种探针类型（PostgreSQL、Cassandra、Redis、Kafka、HTTP、TCP、Ping、CPU）
- 多渠道通知支持
- 告警记录和恢复机制
- 历史数据查看和统计
- 自动数据清理
