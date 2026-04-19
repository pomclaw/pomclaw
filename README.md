# PomClaw

**企业级分布式 AI Agent 平台**

<p>
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/数据库-PostgreSQL%2FOracle-336791?style=for-the-badge&logo=postgresql&logoColor=white" alt="Database">
  <img src="https://img.shields.io/badge/执行环境-SSH%20Sandbox-FF6600?style=for-the-badge" alt="SSH Sandbox">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Version-1.0.0-blue?style=for-the-badge" alt="Version">
</p>

[English](README.en.md) | [中文](#overview)

---

## 🎯 项目概述

PomClaw 是一个企业级平台，用最少的基础设施成本大规模部署 AI Agent。与个人版本需要为每个 Agent 配置一个独立 VM 不同，PomClaw 通过以下两个核心创新实现**无限 Agent 共享基础设施**：

- **分布式记忆存储**：所有 Agent 的记忆、对话和状态统一存储在数据库中
- **SSH 沙盒执行**：无需独立 VM，通过 SSH 沙盒安全隔离执行环境
- **多租户隔离**：支持数千个 Agent 的精细权限管理
- **成本降低 90%**：用 M 个计算节点（M ≈ N/10）服务 N 个 Agent

### 快速对比

| 方面 | 传统方案 | PomClaw |
|------|---------|---------|
| **架构** | 1 个 Agent = 1 个 VM | 共享基础设施 |
| **100个 Agent 成本** | 100 × $10/月 = $1000 | 10 × $10/月 = $100 |
| **存储** | 本地文件 | 分布式数据库 |
| **执行** | 本地计算 | SSH 沙盒池 |
| **可扩展性** | 随 Agent 线性增长 | 随数据集线性增长 |
| **管理** | 独立管理每个 VM | 统一中央平台 |

---

## ✨ 核心功能

### 🗄️ 分布式记忆存储
- **统一后端**：支持 PostgreSQL、Oracle 或任何 SQL 数据库
- **向量检索**：内置 pgvector 支持语义搜索
- **多租户隔离**：自动隔离不同组织/Agent 的数据
- **完整持久化**：保留所有对话历史、状态和元数据

### 🏗️ SSH 沙盒执行
- **安全隔离**：在隔离环境中执行代码，无需 VM 开销
- **灵活部署**：将任何 Linux/Unix 服务器连接为执行节点
- **负载均衡**：自动跨多个沙盒节点分配任务
- **资源控制**：内置超时和资源限制机制

### 💰 企业经济学
- **基础设施整合**：在同一硬件上运行数百个 Agent
- **按需扩展**：添加 SSH 节点而不是 Agent 节点
- **运维简化**：统一的日志、监控和升级管理
- **遗留系统集成**：与现有本地基础设施兼容

### 🔒 安全与合规
- **多租户 RBAC**：组织级和 Agent 级的访问控制
- **审计日志**：完整的操作审计跟踪
- **网络隔离**：支持 VPC、SSH 密钥管理、堡垒机
- **数据加密**：传输层和存储层加密

### 📊 可观测性
- **统一仪表板**：从一个地方监控所有 Agent
- **实时日志**：流式输出 Agent 执行日志和错误
- **性能指标**：CPU、内存、执行时间追踪
- **分布式追踪**：完整的系统端到端追踪

---

## 🚀 快速开始（10 分钟）

### 前置要求
- **Go 1.24+**
- **PostgreSQL 13+**（或 Oracle 数据库）
- **SSH 访问沙盒节点**

### 1. 克隆和编译

```bash
git clone https://github.com/pomclaw/pomclaw.git
cd pomclaw
make build
```

### 2. 配置数据库

```bash
# 创建数据库
createdb pomclaw

# 设置环境变量
export POM_STORAGE_TYPE=postgres
export POM_POSTGRES_HOST=localhost
export POM_POSTGRES_PORT=5432
export POM_POSTGRES_DATABASE=pomclaw
export POM_POSTGRES_USER=postgres
export POM_POSTGRES_PASSWORD=yourpassword
```

### 3. 初始化 Schema

```bash
./build/pomclaw setup-database
```

### 4. 配置 SSH 节点

```bash
# 添加 SSH 沙盒节点
export SSH_NODE_1=user@sandbox-1.example.com:22
```

### 5. 启动 Gateway

```bash
./build/pomclaw gateway

# Gateway 运行在 http://localhost:18790
```

### 6. 创建第一个 Agent

```bash
curl -X POST http://localhost:18790/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "agent-001",
    "organization": "acme-corp",
    "model": "gpt-4",
    "provider": "openai"
  }'
```

---

## 📋 架构设计

```
┌──────────────────────────────────────────────────────────┐
│      分布式数据库（PostgreSQL/Oracle）                    │
│  - 记忆、对话、状态（多租户）                             │
│  - pgvector 向量嵌入                                      │
└──────────────────────────────────────────────────────────┘
                          ↑
                ┌─────────┼─────────┐
                ↓         ↓         ↓
         ┌──────────┐┌──────────┐┌──────────┐
         │SSH Node1 ││SSH Node2 ││SSH Node3 │
         │（沙盒）  ││（沙盒）  ││（沙盒）  │
         └──────────┘└──────────┘└──────────┘
                ↑         ↑         ↑
                └─────────┼─────────┘
                          │
    ┌─────────────────────┴─────────────────────┐
    │    PomClaw Gateway API + WebSocket         │
    │  （单一控制平面服务所有 Agent）            │
    └─────────────────────┬─────────────────────┘
         ↑                 ↑                ↑
    ┌────────────┐   ┌────────────┐   ┌────────────┐
    │  Agent-1   │   │  Agent-2   │   │  Agent-N   │
    └────────────┘   └────────────┘   └────────────┘
```

---

## 🔧 配置

### 数据库配置

```json
{
  "storage_type": "postgres",
  "postgres": {
    "enabled": true,
    "host": "db.example.com",
    "port": 5432,
    "database": "pomclaw",
    "user": "pomclaw",
    "password": "${POSTGRES_PASSWORD}",
    "ssl_mode": "require",
    "pool_max_open": 25,
    "pool_max_idle": 5
  }
}
```

### SSH 沙盒节点配置

```json
{
  "sandbox": {
    "nodes": [
      {
        "name": "sandbox-1",
        "host": "sandbox-1.example.com",
        "port": 22,
        "user": "pomclaw",
        "key_path": "/etc/pomclaw/keys/sandbox-1",
        "max_concurrent": 10,
        "timeout_seconds": 300
      },
      {
        "name": "sandbox-2",
        "host": "sandbox-2.example.com",
        "port": 22,
        "user": "pomclaw",
        "key_path": "/etc/pomclaw/keys/sandbox-2",
        "max_concurrent": 10,
        "timeout_seconds": 300
      }
    ],
    "load_balance_strategy": "round-robin"
  }
}
```

---

## 📚 应用场景

### 🏢 企业 AI 客服
从 10 个扩展到 1000+ 个支持 Agent，成本增长无关

### 🤖 工作流自动化平台
用于 RPA、数据处理和业务逻辑自动化的分布式任务执行引擎

### 📊 大规模数据分析
为每个用户/组织提供隔离、安全工作区的多租户分析平台

### 🔬 科研计算
用于科学模拟和数据处理的高可用计算集群

### 🎓 教育平台
为数千名学生管理 AI 助手，拥有隔离且安全的工作区

---

## 📊 性能与扩展

### 容量规划

| 配置 | Agent 数 | 内存/Agent | CPU | 数据库 |
|------|---------|-----------|-----|--------|
| 小型 | 100 | 256MB | 2-4 核 | PostgreSQL 13 |
| 中型 | 1,000 | 256MB | 8-16 核 | PostgreSQL 14 |
| 大型 | 10,000 | 256MB | 32+ 核 | PostgreSQL 14+ 或 Oracle 21c |
| 企业 | 100,000+ | 256MB | 多节点 | 分布式数据库 |

### 存储需求

- **每个 Agent**：~1MB 元数据 + 10MB 对话（因使用情况而异）
- **向量存储**：~1,500 字节/条记忆（384 维嵌入）

---

## 🔒 安全

### 身份认证与授权
- SSO/OAuth2 支持企业目录集成
- 组织级和 Agent 级的 RBAC
- API 密钥管理和轮换

### 网络安全
- SSH 密钥认证（无密码）
- 所有通信都使用 TLS 1.3
- VPC/网络隔离支持
- 堡垒机兼容

### 数据保护
- 存储层加密（数据库级）
- 传输层加密（TLS）
- 所有操作的审计日志
- 数据保留和合规策略

---

## 🛠️ 开发

### 从源代码构建

```bash
git clone https://github.com/pomclaw/pomclaw.git
cd pomclaw
make build
make test
```

### 运行测试

```bash
# 单元测试
make test

# 集成测试（需要 Docker）
make test-integration

# 包含覆盖率的所有测试
make test-coverage
```

### Docker 部署

```bash
docker-compose up -d
# 启动 PostgreSQL、Redis 和 PomClaw Gateway
```

---

## 📖 文档

- [架构指南](docs/STORAGE_ARCHITECTURE.md)
- [PostgreSQL 设置](docs/POSTGRESQL_SUPPORT.md)
- [API 参考](docs/API.md)
- [部署指南](docs/DEPLOYMENT.md)
- [安全指南](docs/SECURITY.md)

---

## 🤝 贡献

欢迎贡献代码！请：

1. Fork 仓库
2. 创建功能分支（`git checkout -b feature/amazing-feature`）
3. 提交更改（`git commit -m 'Add amazing feature'`）
4. Push 到分支（`git push origin feature/amazing-feature`）
5. 开启 Pull Request

---

## 📜 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

---

## 🔗 相关项目

- [PicoClaw](https://github.com/jasperan/pomclaw) - 轻量级 AI Agent 框架

---

## 📞 支持

- **问题反馈**: [GitHub Issues](https://github.com/pomclaw/pomclaw/issues)
- **讨论**: [GitHub Discussions](https://github.com/pomclaw/pomclaw/discussions)
- **企业支持**: contact@pomclaw.com

---

## 🎉 致谢

PomClaw 基于以下优秀开源项目：
- PicoClaw 社区
- 开源数据库和 SSH 社区
- Go 生态系统贡献者
