# PomClaw

**Enterprise-Grade Distributed AI Agent Platform**

<p>
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Database-PostgreSQL%2FOracle-336791?style=for-the-badge&logo=postgresql&logoColor=white" alt="Database">
  <img src="https://img.shields.io/badge/Execution-SSH%20Sandbox-FF6600?style=for-the-badge" alt="SSH Sandbox">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Version-1.0.0-blue?style=for-the-badge" alt="Version">
</p>

[English](#-overview) | [中文](README.md)

---

## 🎯 Overview

PomClaw is an enterprise-grade platform designed to deploy AI Agents at scale with minimal infrastructure costs. Unlike personal-use solutions that require one VM per Agent, PomClaw enables **unlimited Agents on shared infrastructure** through:

- **Distributed Memory Storage**: Unified database for all Agent memories, conversations, and state
- **SSH Sandbox Execution**: Secure, isolated workspace execution without individual VMs
- **Multi-Tenant Isolation**: Support thousands of Agents with fine-grained security controls
- **90% Cost Reduction**: Serve N agents with M compute nodes (M ≈ N/10)

### Quick Comparison

| Aspect | Traditional | PomClaw |
|--------|-----------|---------|
| **Architecture** | 1 VM per Agent | Shared infrastructure |
| **Cost for 100 Agents** | 100 × $10/mo = $1000 | 10 × $10/mo = $100 |
| **Storage** | Local files | Distributed database |
| **Execution** | Local compute | SSH sandbox pool |
| **Scalability** | Linear with agents | Linear with dataset |
| **Management** | Individual VMs | Centralized platform |

---

## ✨ Core Features

### 🗄️ Distributed Memory Storage
- **Unified Backend**: PostgreSQL, Oracle, or any SQL database
- **Vector Search**: Built-in pgvector support for semantic search
- **Multi-Tenant**: Automatic isolation of data across organizations/agents
- **Persistence**: Complete conversation history, state, and metadata

### 🏗️ SSH Sandbox Execution
- **Secure Isolation**: Execute code in isolated environments without VM overhead
- **Flexible Deployment**: Connect any Linux/Unix server as execution node
- **Load Balancing**: Automatic distribution across multiple sandbox nodes
- **Resource Control**: Built-in timeout and resource limits

### 💰 Enterprise Economics
- **Infrastructure Consolidation**: Run 100s of agents on same hardware
- **On-Demand Scaling**: Add SSH nodes as needed, not agents
- **Reduced Operational Burden**: Centralized logging, monitoring, and updates
- **Legacy Integration**: Works with existing on-premises infrastructure

### 🔒 Security & Compliance
- **Multi-Tenant RBAC**: Organization and agent-level access control
- **Audit Logging**: Complete operational audit trail
- **Network Isolation**: VPC support, SSH key management, bastion host compatible
- **Data Encryption**: In-transit and at-rest encryption options

### 📊 Observability
- **Unified Dashboard**: Monitor all agents from one place
- **Real-Time Logs**: Stream agent execution logs and errors
- **Performance Metrics**: CPU, memory, execution time tracking
- **Distributed Tracing**: Full request tracing across system

---

## 🚀 Quick Start (10 minutes)

### Prerequisites
- **Go 1.24+**
- **Node.js 18+** (for frontend build)
- **PostgreSQL 13+** (or Oracle Database)
- **SSH access to sandbox nodes**

### 1. Clone and Build

```bash
git clone https://github.com/pomclaw/pomclaw.git
cd pomclaw
make build  # Automatically builds both backend and frontend UI
```

> **Note**: `make build` automatically:
> - Compiles the frontend UI (using `npm run build`)
> - Compiles the backend binary
> - Packages frontend into `dist/control-ui/` directory

### 2. Configure Database

```bash
# Create database
createdb pomclaw

# Set environment variables
export POM_STORAGE_TYPE=postgres
export POM_POSTGRES_HOST=localhost
export POM_POSTGRES_PORT=5432
export POM_POSTGRES_DATABASE=pomclaw
export POM_POSTGRES_USER=postgres
export POM_POSTGRES_PASSWORD=yourpassword
```

### 3. Initialize Schema

```bash
./build/pomclaw setup-database
```

### 4. Configure SSH Nodes

```bash
# Add an SSH sandbox node
export SSH_NODE_1=user@sandbox-1.example.com:22
```

### 5. Start Gateway

```bash
./build/pomclaw gateway

# Gateway starts on http://localhost:18790
# Frontend UI automatically served from: http://localhost:18790 (using dist/control-ui)
```

**Gateway Web UI:**

![PomClaw Gateway Chat UI](docs/screenshots/chat.jpg)

---

## 🎨 Frontend & Backend Integration

PomClaw uses a **decoupled frontend-backend architecture** while providing fully integrated deployment:

### Build & Deployment

**Backend**: Distributed AI Agent platform written in Go
- WebSocket and HTTP API endpoints
- Agent lifecycle, memory, and execution management

**Frontend**: Modern web UI built with TypeScript + React
- Session management and real-time chat interface
- Multi-language and theme customization

### Integrated Deployment

Running `make build` automatically builds the complete application:
- ✅ Backend binary: `build/pomclaw-*`
- ✅ Frontend assets: `dist/control-ui/` (compiled from `ui/` directory)

Starting Gateway automatically serves the Web UI:
```bash
./build/pomclaw gateway
# Access http://localhost:18790 to use the complete application
```

**Configuration**:
- Gateway's `ui_path` config defaults to `dist/control-ui`
- Customize UI path by modifying the configuration
- Frontend communicates with backend via WebSocket in real-time

---

## 📋 Architecture

```
┌──────────────────────────────────────────────────────────┐
│           Distributed Database (PostgreSQL/Oracle)       │
│  - Memories, conversations, state (multi-tenant)         │
│  - Vector embeddings with pgvector                       │
└──────────────────────────────────────────────────────────┘
                          ↑
                ┌─────────┼─────────┐
                ↓         ↓         ↓
         ┌──────────┐┌──────────┐┌──────────┐
         │SSH Node1 ││SSH Node2 ││SSH Node3 │
         │(Sandbox) ││(Sandbox) ││(Sandbox) │
         └──────────┘└──────────┘└──────────┘
                ↑         ↑         ↑
                └─────────┼─────────┘
                          │
    ┌─────────────────────┴─────────────────────┐
    │     PomClaw Gateway API + WebSocket        │
    │  (Single control plane for all agents)     │
    └─────────────────────┬─────────────────────┘
         ↑                 ↑                ↑
    ┌────────────┐   ┌────────────┐   ┌────────────┐
    │  Agent-1   │   │  Agent-2   │   │  Agent-N   │
    └────────────┘   └────────────┘   └────────────┘
```

---

## 🔧 Configuration

### Database Configuration

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

### SSH Sandbox Nodes

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

## 📚 Use Cases

### 🏢 Enterprise AI Customer Support
Scale from 10 to 1000+ support agents without proportional cost increase

### 🤖 Workflow Automation Platform
Distributed task execution engine for RPA, data processing, and business logic automation

### 📊 Data Analysis at Scale
Multi-tenant analytics platform with isolated workspaces for each user/organization

### 🔬 Research Computing
High-availability compute clusters for scientific simulations and data processing

### 🎓 Educational Platform
Manage AI assistants for thousands of students with isolated, secure workspaces

---

## 📊 Performance & Scaling

### Capacity Planning

| Configuration | Agents | Memory/Agent | CPU | Database |
|--------------|--------|-------------|-----|----------|
| Small | 100 | 256MB | 2-4 core | PostgreSQL 13 |
| Medium | 1,000 | 256MB | 8-16 core | PostgreSQL 14 |
| Large | 10,000 | 256MB | 32+ core | PostgreSQL 14+ or Oracle 21c |
| Enterprise | 100,000+ | 256MB | Multi-node | Distributed DB |

### Storage Requirements

- **Per Agent**: ~1MB metadata + 10MB conversations (varies by usage)
- **Vector Storage**: ~1,500 bytes per memory (384-dim embedding)

---

## 🔒 Security

### Authentication & Authorization
- OAuth2/SSO support for enterprise directories
- Organization-level and agent-level RBAC
- API key management with rotation

### Network Security
- SSH key-based authentication (no passwords)
- TLS 1.3 for all communications
- VPC/network isolation support
- Bastion host support for air-gapped deployments

### Data Protection
- Encryption at rest (database-level)
- Encryption in transit (TLS)
- Audit logging for all operations
- Data retention and compliance policies

---

## 🛠️ Development

### Build from Source

```bash
git clone https://github.com/pomclaw/pomclaw.git
cd pomclaw
make build      # Build backend + frontend
make test
```

### Frontend Development

```bash
cd ui
npm install
npm run dev      # Development server (hot reload)
npm run build    # Production build
npm run preview  # Preview production build
```

### Backend Development

```bash
make run ARGS=gateway  # Quick build and run Gateway
```

### Run Tests

```bash
# Unit tests
make test

# Integration tests (requires Docker)
make test-integration

# All tests with coverage
make test-coverage
```

### Docker Deployment

```bash
docker-compose up -d
# Starts PostgreSQL, Redis, and PomClaw Gateway (with UI)
```

---

## 📖 Documentation

- [Architecture Guide](docs/STORAGE_ARCHITECTURE.md)
- [PostgreSQL Setup](docs/POSTGRESQL_SUPPORT.md)
- [API Reference](docs/API.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [Security Guide](docs/SECURITY.md)

---

## 🤝 Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details

---

## 🔗 Related Projects

- [PicoClaw](https://github.com/jasperan/pomclaw) - Lightweight AI agent framework

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/pomclaw/pomclaw/issues)
- **Discussions**: [GitHub Discussions](https://github.com/pomclaw/pomclaw/discussions)
- **Enterprise Support**: contact@pomclaw.com

---

## 🎉 Acknowledgments

PomClaw builds on the excellent work of:
- PicoClaw community
- Open source database and SSH communities
- Go ecosystem contributors
