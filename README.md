# Blockchain Asset API
web3
区块链资产查询API，提供以太坊区块链上账户余额、交易详情、区块信息等查询功能，并支持区块扫描和数据存储。

## 功能特性

- ✅ 查询ETH余额
- ✅ 查询ERC20代币余额  
- ✅ 查询交易详情
- ✅ 查询区块信息
- ✅ 区块扫描与数据存储
- ✅ RESTful API设计
- ✅ Swagger文档支持
- ✅ Redis缓存优化
- ✅ 请求频率限制

## 项目结构

```
.
├── cmd/api                 # API主程序入口
│   ├── docs               # Swagger文档
│   ├── config.yaml        # 配置文件
│   └── main.go            # 主程序
├── config                 # 配置管理
├── internal               # 核心业务逻辑
│   ├── handler            # HTTP处理器
│   ├── model              # 数据模型
│   ├── repository         # 数据访问层
│   ├── service            # 业务逻辑层
│   └── util               # 工具函数
└── web                    # 前端静态资源
```


## 环境要求

- Go 1.19+
- MySQL 5.7+
- Redis 6.0+
- Ethereum节点 (本地GETH或Infura等)

## 安装部署

### 1. 克隆项目

```bash
git clone <repository-url>
cd blockchain-asset-api
```


### 2. 配置环境

创建 `cmd/api/config.yaml` 配置文件：

```yaml
server:
  port: ":8080"
  timeout: 10s

eth:
  nodeURL: "http://localhost:8545"  # 以太坊节点地址

redis:
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
  expire: 5m

mysql:
  dsn: "username:password@tcp(127.0.0.1:3306)/blockchain_asset?charset=utf8mb4&parseTime=True&loc=Local"
```


### 3. 数据库初始化

创建MySQL数据库：
```sql
CREATE DATABASE blockchain_asset CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```


### 4. 安装依赖

```bash
go mod tidy
```


### 5. 运行程序

```bash
cd cmd/api
go run main.go
```


## API接口文档

服务启动后，可通过以下地址访问API文档：
- Swagger UI: http://localhost:8080/swagger/index.html
- JSON文档: http://localhost:8080/swagger/doc.json

### 主要API接口

| 接口 | 方法 | 描述 |
|------|------|------|
| `/api/v1/address/{addr}/balance` | GET | 查询ETH余额 |
| `/api/v1/address/{addr}/tokens` | GET | 查询ERC20代币余额 |
| `/api/v1/transaction/{txhash}` | GET | 查询交易详情 |
| `/api/v1/block/{blocknum}` | GET | 查询区块信息 |
| `/api/v1/scan` | GET | 扫描区块 |

## 部署方式

### 方式一：直接运行

```bash
cd cmd/api
go build -o blockchain-api main.go
./blockchain-api
```


### 方式二：Docker部署

创建 `Dockerfile`:
```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o blockchain-api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/blockchain-api .
COPY cmd/api/config.yaml .
EXPOSE 8080
CMD ["./blockchain-api"]
```


构建和运行：
```bash
docker build -t blockchain-api .
docker run -p 8080:8080 blockchain-api
```


## 性能优化

1. **Redis缓存**: 对常用查询结果进行缓存，减少链上请求
2. **请求限流**: 每个IP每分钟最多100次请求
3. **数据库连接池**: 优化数据库连接管理
4. **Goroutine并发**: 区块扫描采用并发处理提高效率

## 注意事项

项目只是本地节点学习demo，扩展功能考虑：
1. 扫描策略优化分批扫描而非全量扫描  
2. 数据过滤机制：只扫描感兴趣的地址或合约 
3. 监控告警：添加扫描进度监控、设置磁盘空间预警
4. 可根据实际需求调整缓存过期时间和限流策略
