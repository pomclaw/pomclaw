# 数据库初始化指南

## 概述

Pomclaw 支持两种数据库：
- **PostgreSQL** (推荐)
- **Oracle Database**

数据库表在首次启动时**自动创建**，无需手动执行SQL。

## PostgreSQL 配置（推荐）

### 1. 安装 PostgreSQL + pgvector

```bash
# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib

# macOS
brew install postgresql

# 启动PostgreSQL
sudo systemctl start postgresql  # Linux
brew services start postgresql   # macOS

# 安装 pgvector 扩展（用于向量搜索）
# 参考: https://github.com/pgvector/pgvector
```

### 2. 创建数据库和用户

```bash
# 连接到PostgreSQL
sudo -u postgres psql

# 在psql中执行
CREATE DATABASE pomclaw;
CREATE USER pomclaw WITH ENCRYPTED PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE pomclaw TO pomclaw;

# 启用pgvector扩展
\c pomclaw
CREATE EXTENSION IF NOT EXISTS vector;

# 退出
\q
```

### 3. 配置 config.json

在你的 `config.json` 中添加或修改：

```json
{
  "storage_type": "postgres",
  "postgres": {
    "enabled": true,
    "host": "localhost",
    "port": 5432,
    "database": "pomclaw",
    "user": "pomclaw",
    "password": "your_password",
    "ssl_mode": "disable",
    "pool_max_open": 10,
    "pool_max_idle": 2,
    "agent_id": "default",
    "embedding_provider": "api",
    "embedding_api_base": "http://localhost:11434/v1",
    "embedding_api_key": "",
    "embedding_model": "bge-m3:latest"
  },
  "agents": {
    "defaults": {
      "workspace": "~/.pomclaw/workspace",
      "restrict_to_workspace": true,
      "provider": "ollama",
      "model": "qwen3.5:35b-a3b",
      "max_tokens": 8192,
      "temperature": 0.7,
      "max_tool_iterations": 20
    }
  }
}
```

**重要字段说明：**
- `storage_type`: 设置为 `"postgres"`
- `postgres.enabled`: 必须设置为 `true`
- `postgres.agent_id`: 默认agent的ID（通常是 "default"）
- `embedding_provider`: "api" 或 "local"（推荐用API调用embedding模型）

### 4. 启动 Pomclaw

```bash
./pomclaw
```

**首次启动会自动创建表：**
```
✓ Connected to postgres Database
✓ Schema initialized (8 tables with POM_ prefix)
```

创建的表包括：
- `POM_SESSIONS` - 会话历史
- `POM_SESSION_SUMMARIES` - 会话摘要
- `POM_MEMORY_NOTES` - 记忆笔记
- `POM_VECTOR_MEMORIES` - 向量记忆（pgvector）
- `POM_AGENT_STATE` - Agent状态
- `POM_PROMPT_FILES` - 提示文件
- `POM_USERS` - 用户表（新增）
- `POM_AGENT_CONFIGS` - Agent配置表（新增）

## Oracle Database 配置

### 1. 准备Oracle数据库

可以使用：
- Oracle Database Free (23ai)
- Oracle Cloud Free Tier
- 本地Oracle实例

### 2. 配置 config.json

```json
{
  "storage_type": "oracle",
  "oracle": {
    "enabled": true,
    "mode": "freepdb",
    "host": "localhost",
    "port": 1521,
    "service": "FREEPDB1",
    "user": "pomclaw",
    "password": "your_password",
    "dsn": "",
    "walletPath": "",
    "poolMaxOpen": 10,
    "poolMaxIdle": 2,
    "onnxModel": "ALL_MINILM_L12_V2",
    "agentId": "default"
  }
}
```

### 3. 启动 Pomclaw

```bash
./pomclaw
```

同样会自动创建所有表。

## 验证数据库初始化

### PostgreSQL

```bash
psql -U pomclaw -d pomclaw

# 查看所有表
\dt

# 查看Agent配置表
SELECT * FROM POM_AGENT_CONFIGS;

# 退出
\q
```

### Oracle

```sql
-- 连接到数据库
sqlplus pomclaw/your_password@localhost:1521/FREEPDB1

-- 查看所有表
SELECT table_name FROM user_tables WHERE table_name LIKE 'POM_%';

-- 查看Agent配置表
SELECT * FROM POM_AGENT_CONFIGS;
```

## 常见问题

**Q: 启动时报连接数据库失败？**
- 检查数据库服务是否运行
- 确认 `config.json` 中的连接信息正确
- PostgreSQL: 检查 `pg_hba.conf` 是否允许连接
- 确保 `storage_type` 和对应的 `postgres/oracle.enabled` 一致

**Q: 表已经存在，如何重新初始化？**
```sql
-- PostgreSQL
DROP TABLE IF EXISTS POM_USERS CASCADE;
DROP TABLE IF EXISTS POM_AGENT_CONFIGS CASCADE;
-- ... 删除其他表

-- 重启pomclaw会重新创建
```

**Q: 不使用数据库可以吗？**
- 不可以。多Agent架构必须使用数据库存储配置和数据。
- 推荐使用PostgreSQL，开源且功能强大。

**Q: 需要安装pgvector吗？**
- 是的，如果使用PostgreSQL，pgvector用于向量相似度搜索（Remember/Recall功能）
- 安装方法：https://github.com/pgvector/pgvector

**Q: Embedding服务是什么？**
- 用于将文本转换为向量（Remember/Recall功能需要）
- 可以使用Ollama的embedding模型（如 bge-m3:latest）
- 配置 `embedding_api_base` 指向你的embedding服务

## 下一步

数据库初始化完成后，参考 `multi-agent-usage.md` 了解如何：
1. 创建新的Agent
2. 配置不同的Agent参数
3. 在消息中指定Agent ID
