# 多租户Workspace隔离方案选型

## 目录

- [1. 背景与需求](#1-背景与需求)
- [2. 隔离层级](#2-隔离层级)
- [3. 方案对比](#3-方案对比)
- [4. 详细方案设计](#4-详细方案设计)
- [5. 成本分析](#5-成本分析)
- [6. 实现路线图](#6-实现路线图)
- [7. 选型建议](#7-选型建议)
- [附录](#附录)

---

## 1. 背景与需求

### 1.1 核心问题

PomClaw作为多租户AI Agent平台，需要支持多用户在共享基础设施上安全隔离运行：

```
N个用户 → 共享M个计算节点（M ≈ N/10）
```

**隔离目标：**
- ✅ 不同用户的代码执行不相互影响
- ✅ 不同用户的文件/数据不可见
- ✅ 资源配额独立（CPU/内存/网络）
- ✅ 性能开销最小化

### 1.2 当前状态

PomClaw已实现：
- ✅ **数据库级隔离**（ENTERPRISE_ARCHITECTURE.md）：org_id、user_id字段
- ❌ **执行环境隔离**：SSH沙箱只过滤危险命令，不隔离用户
- ❌ **文件系统隔离**：无workspace隔离

---

## 2. 隔离层级

多租户隔离必须在三层实现：

```
┌─────────────────────────────────────────┐
│ 第1层：数据库隔离（必须）                │
│ - 所有查询加org_id/user_id过滤           │
│ - 行级安全策略(RLS)                     │
│ 状态：✅ 已规划                          │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│ 第2层：文件系统隔离（必须）              │
│ - 每用户独立workspace目录                │
│ - 用户A不可访问用户B的文件               │
│ 状态：❌ 本文档关键                      │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│ 第3层：执行环境隔离（优化）              │
│ - 限制资源使用（CPU/内存）               │
│ - 命令安全过滤                          │
│ 状态：✅ shell.go已实现                  │
└─────────────────────────────────────────┘
```

**关键认知：** SSH沙箱只能做第3层，第2层必须由其他机制完成

---

## 3. 方案对比

### 3.1 快速对比表

| 方案 | 隔离强度 | 实现难度 | 性能开销 | 部署复杂度 | 推荐场景 |
|------|--------|--------|--------|----------|---------|
| **方案A：系统用户隔离** | 中强 | ⭐ | 极低 | ⭐ | **现在做** ✅ |
| **方案B：chroot隔离** | 中强 | ⭐⭐ | 低 | ⭐⭐ | 可选增强 |
| **方案C：容器隔离(单机)** | 强 | ⭐⭐ | 中等 | ⭐⭐ | 用户500+ |
| **方案D：K8s容器编排** | 非常强 | ⭐⭐⭐ | 中等 | ⭐⭐⭐ | 上云部署 |

### 3.2 详细数据对比

```
并发用户：1000
执行命令数/天：10000
单次执行时长：平均5秒

┌──────────────────┬──────────┬──────────┬──────────┬──────────┐
│ 指标             │ 系统用户 │ chroot   │ 容器(单) │ K8s      │
├──────────────────┼──────────┼──────────┼──────────┼──────────┤
│ 命令启动时间     │ 1-10ms   │ 5-20ms   │ 100-500m │ 1-2s     │
│ 基础内存占用     │ 100MB    │ 150MB    │ 500MB    │ 1GB+     │
│ 单用户开销       │ <1MB     │ 2-5MB    │ 50-200MB │ 100MB    │
│ 1000用户总内存   │ ~1GB     │ ~2GB     │ ~50-100G │ ~100GB+  │
│ 镜像大小         │ NA       │ NA       │ 100-500M │ 100-500M │
│ 镜像拉取时间     │ NA       │ NA       │ 初次30-60│ 初次30-60│
│ 并发容限         │ 1000+    │ 1000+    │ 10-50    │ 1000+    │
│ 故障影响范围     │ 单命令   │ 单命令   │ 单命令   │ 单Pod    │
│ 跨主机扩展       │ 困难     │ 困难     │ 困难     │ 容易     │
└──────────────────┴──────────┴──────────┴──────────┴──────────┘

结论：
- <500用户 → 系统用户方案性价比最高
- 500-5000用户 → 考虑chroot或容器
- 5000+用户 → 必须K8s容器编排
```

---

## 4. 详细方案设计

### 方案A：系统用户隔离（推荐现在实现）

#### 4.1.1 架构图

```
┌────────────────────────────────────────────────┐
│         API Gateway (Go)                       │
│  从JWT提取 user_id，传递给ExecTool             │
└────────────────────┬─────────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────────┐
│      ExecTool (Modified)                       │
│  - 传入：userID "user-abc123"                   │
│  - 查找/创建系统用户：exec_user_abc123         │
│  - 设置工作目录：/workspace/user-abc123         │
│  - 执行命令：sudo -u exec_user_abc123 <cmd>   │
└────────────────────┬─────────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────────┐
│    系统级隔离（Linux/Unix）                     │
│  ┌─────────────────────────────────────┐       │
│  │ User: exec_user_abc123              │       │
│  │ UID: 2001                            │       │
│  │ GID: 2000 (exec_group)               │       │
│  │ HOME: /workspace/user-abc123         │       │
│  │ File permissions: drwx------ (700)   │       │
│  │ Resource limits (ulimit):            │       │
│  │  - CPU: 2 cores                      │       │
│  │  - Memory: 512MB                     │       │
│  │  - Processes: 50                     │       │
│  └─────────────────────────────────────┘       │
└────────────────────────────────────────────────┘
```

#### 4.1.2 实现步骤

**Step 1: 初始化系统用户（容器启动时或systemd服务）**

```bash
#!/bin/bash
# scripts/init-sandbox-users.sh

# 创建执行用户组
groupadd -f exec_group

# 为每个潜在用户预创建账户（或动态创建）
# 这里简化为100个预留账户
for i in {0..100}; do
    username="exec_user_$(printf '%06d' $i)"
    uid=$((2000 + i))
    useradd -u $uid -g exec_group -s /bin/bash -m -d /workspace/$username $username 2>/dev/null || true
    # 设置严格权限
    chmod 700 /workspace/$username
    chown $username:exec_group /workspace/$username
done
```

**Step 2: 修改ExecTool**

```go
// pkg/tools/shell.go

type ExecTool struct {
    workingDir          string
    userID              string              // 新增：平台用户ID (UUID格式)
    timeout             time.Duration
    denyPatterns        []*regexp.Regexp
    allowPatterns       []*regexp.Regexp
    restrictToWorkspace bool
    sandboxUser         string              // 新增：系统用户名缓存
}

func NewExecTool(workingDir, userID string, restrict bool) *ExecTool {
    return &ExecTool{
        workingDir:          workingDir,
        userID:              userID,
        timeout:             60 * time.Second,
        denyPatterns:        defaultDenyPatterns,
        allowPatterns:       nil,
        restrictToWorkspace: restrict,
        sandboxUser:         deriveSandboxUsername(userID),  // UUID -> 系统用户名
    }
}

// deriveSandboxUsername 将平台userID转换为系统用户名
// userID: "user-abc123def456" (36 chars UUID)
// 返回: "exec_user_abc12345" (系统用户名有长度限制32 chars)
func deriveSandboxUsername(userID string) string {
    // 提取UUID的关键部分（去掉user-前缀和连字符）
    hash := md5.Sum([]byte(userID))
    hashStr := hex.EncodeToString(hash[:])[:8]
    return fmt.Sprintf("exec_user_%s", hashStr)
}

func (t *ExecTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
    command, ok := args["command"].(string)
    if !ok {
        return ErrorResult("command is required")
    }

    cwd := t.workingDir
    if wd, ok := args["working_dir"].(string); ok && wd != "" {
        cwd = wd
    }

    if guardError := t.guardCommand(command, cwd); guardError != "" {
        return ErrorResult(guardError)
    }

    // ==================== 新增：用户隔离 ====================
    var cmd *exec.Cmd
    if runtime.GOOS == "windows" {
        // Windows不支持sudo，不隔离（或用其他机制）
        cmd = exec.CommandContext(cmdCtx, "powershell", "-NoProfile", "-NonInteractive", "-Command", command)
    } else {
        // Unix/Linux：使用sudo以隔离用户身份执行
        // sudo -u exec_user_abc123 -g exec_group --
        //   sh -c "export HOME=/workspace/user-abc123; <command>"
        
        env := fmt.Sprintf("HOME=%s", cwd)
        fullCmd := fmt.Sprintf("export %s; %s", env, command)
        
        cmd = exec.CommandContext(
            cmdCtx,
            "sudo",
            "-u", t.sandboxUser,
            "-g", "exec_group",
            "--",
            "sh", "-c", fullCmd,
        )
    }
    
    if cwd != "" {
        cmd.Dir = cwd
    }
    
    // ==================== 资源限制 ====================
    if runtime.GOOS != "windows" {
        cmd.SysProcAttr = &syscall.SysProcAttr{
            // 进程信号处理
            Pdeathsig: syscall.SIGKILL,
        }
        
        // 使用 ulimit 限制资源
        // CPU: 2 cores × 60sec = 120 CPU seconds
        // Memory: 512MB
        // Processes: 50
        limitsCmd := fmt.Sprintf(
            "ulimit -t 120 -v 524288 -u 50 -n 1024; %s",
            fullCmd,
        )
        cmd = exec.CommandContext(
            cmdCtx,
            "sudo",
            "-u", t.sandboxUser,
            "-g", "exec_group",
            "--",
            "sh", "-c", limitsCmd,
        )
    }

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    // ... 其余逻辑同现有代码
    
    return &ToolResult{
        ForLLM:  output,
        ForUser: output,
        IsError: err != nil,
    }
}
```

**Step 3: API层集成**

```go
// pkg/api/handlers/chat.go
func (h *ChatHandler) Execute(w http.ResponseWriter, r *http.Request) {
    // 从JWT获取用户信息
    claims := GetClaims(r.Context())
    if claims == nil {
        RespondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    var req struct {
        Command string `json:"command"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // 创建隔离的ExecTool（关键：传入userID）
    userWorkspace := fmt.Sprintf("/workspace/%s", claims.UserID)
    execTool := tools.NewExecTool(userWorkspace, claims.UserID, true)
    
    result := execTool.Execute(r.Context(), map[string]interface{}{
        "command": req.Command,
    })

    RespondJSON(w, http.StatusOK, result)
}
```

#### 4.1.3 安全性检查清单

```yaml
初始化检查：
  ✅ 系统用户不能登录shell（/sbin/nologin）
  ✅ 系统用户只有exec_group组成员身份
  ✅ workspace目录权限为700（只有该用户可读写）
  ✅ 禁止sudo权限提升（sudoers中不包含exec_user_*）

运行时检查：
  ✅ 命令前过滤（shell.go现有deny-list）
  ✅ 工作目录校验（restrictToWorkspace）
  ✅ 资源限制生效（ulimit）
  ✅ 超时控制（context timeout）

监控检查：
  ✅ 审计日志（sudo -l, /var/log/auth.log）
  ✅ 资源使用监控（CPU, 内存, 文件描述符）
  ✅ 异常检测（shell逃逸尝试）
```

---

### 方案B：chroot隔离（可选增强）

#### 4.2.1 架构特点

```
比方案A强化：
- 用户A无法访问/etc/passwd查看其他用户
- 虚拟根目录="/workspace/user-abc123"
- 命令只能看到该用户的文件树
- 隔离更彻底但复杂度增加
```

#### 4.2.2 实现概要

```go
cmd.SysProcAttr = &syscall.SysProcAttr{
    Chroot: fmt.Sprintf("/workspace/%s", userID),
    // chroot后系统看不到父目录
}
```

**缺点：**
- 必须是root才能chroot
- 需要完整的最小根文件系统（busybox等）
- 性能略低于方案A
- 维护复杂

**建议：** 如无特殊安全需求，不如用容器

---

### 方案C：Docker容器隔离（单机版，用户500+时考虑）

#### 4.3.1 架构图

```
┌─────────────────────────────────────────┐
│    API Gateway                          │
│ 收到请求 → 创建临时容器 → 执行 → 销毁  │
└─────────────────────────────────────────┘
               ↓
┌─────────────────────────────────────────┐
│  Docker Daemon (Host或SSH Node)        │
│  ┌────────────────────────────────────┐ │
│  │ Container: user-abc123-exec-1234   │ │
│  │ Image: pomclaw-executor:latest     │ │
│  │ ├─ Resource limits:                │ │
│  │ │  ├─ Memory: 256MB               │ │
│  │ │  ├─ CPU: 1 core                 │ │
│  │ │  └─ Disk: 1GB                   │ │
│  │ ├─ Volume mount:                  │ │
│  │ │  └─ /workspace/user-abc123      │ │
│  │ └─ Security:                      │ │
│  │    ├─ readOnlyRootFilesystem      │ │
│  │    ├─ runAsNonRoot               │ │
│  │    └─ noNewPrivileges            │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

#### 4.3.2 Dockerfile

```dockerfile
# 轻量基础镜像
FROM busybox:1.36-musl AS base
# 或 alpine:3.19 (更完整，但仍<10MB)

# 创建非root用户
RUN adduser -D -u 1000 executor

# 安装必要工具
RUN apk add --no-cache bash curl wget git openssh-client

WORKDIR /workspace

# 切换非root用户
USER executor

ENTRYPOINT ["/bin/bash"]
```

#### 4.3.3 Go实现

```go
// pkg/executor/docker_executor.go
package executor

import (
    "context"
    "fmt"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)

type DockerExecutor struct {
    client      *client.Client
    image       string
    memoryLimit int64  // bytes
}

func NewDockerExecutor() (*DockerExecutor, error) {
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        return nil, err
    }
    return &DockerExecutor{
        client:      cli,
        image:       "pomclaw-executor:latest",
        memoryLimit: 256 * 1024 * 1024, // 256MB
    }, nil
}

func (e *DockerExecutor) Execute(ctx context.Context, userID string, command string, timeout time.Duration) (string, error) {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    // 容器配置
    config := &container.Config{
        Image: e.image,
        Cmd:   []string{"bash", "-c", command},
        Env: []string{
            fmt.Sprintf("USER_ID=%s", userID),
            "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
        },
        WorkingDir: "/workspace",
    }

    // 资源限制配置
    hostConfig := &container.HostConfig{
        Memory:     e.memoryLimit,
        CPUQuota:   100000,  // 1 core = 100000
        CPUPeriod:  100000,
        PidsLimit:  &[]int64{50}[0],
        ReadonlyRootfs: true,
        SecurityOpt: []string{"no-new-privileges=true"},
        Mounts: []mount.Mount{
            {
                Type:     mount.TypeBind,
                Source:   fmt.Sprintf("/workspace/%s", userID),
                Target:   "/workspace",
                ReadOnly: false,
            },
        },
    }

    // 创建容器
    resp, err := e.client.ContainerCreate(
        ctx,
        config,
        hostConfig,
        nil,
        nil,
        fmt.Sprintf("exec-%s-%d", userID, time.Now().UnixNano()),
    )
    if err != nil {
        return "", err
    }
    containerID := resp.ID

    defer func() {
        // 清理：停止并删除容器
        e.client.ContainerStop(context.Background(), containerID, container.StopOptions{})
        e.client.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})
    }()

    // 启动容器
    if err := e.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
        return "", err
    }

    // 等待完成
    statusCh, errCh := e.client.ContainerWait(ctx, containerID, container.WaitConditionNextExit)
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    case <-errCh:
        return "", fmt.Errorf("container error")
    case <-statusCh:
        // 获取日志
        out, err := e.client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
            ShowStdout: true,
            ShowStderr: true,
        })
        // ... 读取日志内容
        return string(logBytes), nil
    }
}
```

#### 4.3.4 性能注意事项

```
启动时间：100-500ms（受镜像拉取、容器初始化影响）
  - 第一次：需要拉取镜像（30-60秒）
  - 后续：使用缓存镜像（100-200ms）

内存开销：
  - 单个容器：50-200MB
  - 1000容器并发：不现实，需要容器复用池

建议：容器复用池
  ```go
  // 而不是每次创建/销毁
  type ContainerPool struct {
      containers chan *Container  // 复用池
      maxSize    int              // 最大容器数
  }
  
  func (p *ContainerPool) Acquire() *Container {
      select {
      case c := <-p.containers:
          return c
      default:
          return p.create()  // 创建新容器
      }
  }
  
  func (p *ContainerPool) Release(c *Container) {
      c.Clean()  // 清理状态
      select {
      case p.containers <- c:  // 放回池
      default:
          c.Destroy()  // 池满了就销毁
      }
  }
  ```
```

---

### 方案D：Kubernetes容器编排（云部署，用户5000+）

#### 4.4.1 架构设计

```
Master Cluster
├─ API Server
├─ Controller Manager
└─ Scheduler

Worker Nodes (N个)
├─ Node 1
│  ├─ Pod: pomclaw-gateway-1
│  └─ Pod: pomclaw-gateway-2
├─ Node 2
│  ├─ Pod: user-1-exec-job
│  ├─ Pod: user-2-exec-job
│  └─ ...
└─ Node 3
   └─ Pod: postgres-1
```

#### 4.4.2 Kubernetes配置示例

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: pomclaw

---
# ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pomclaw-executor
  namespace: pomclaw

---
# 为每个用户执行创建Job
apiVersion: batch/v1
kind: Job
metadata:
  name: exec-user-abc123-{{ .ExecutionID }}
  namespace: pomclaw
spec:
  ttlSecondsAfterFinished: 60  # 完成后60秒删除Job
  backoffLimit: 0  # 失败不重试
  template:
    spec:
      serviceAccountName: pomclaw-executor
      restartPolicy: Never
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: executor
        image: pomclaw-executor:latest
        imagePullPolicy: IfNotPresent
        command: ["/bin/bash", "-c"]
        args:
        - |
          user_id="user-abc123"
          exec_timeout=60
          timeout $exec_timeout bash -c "{{ .Command }}"
        
        resources:
          limits:
            memory: "256Mi"
            cpu: "500m"
            ephemeralStorage: "1Gi"
          requests:
            memory: "128Mi"
            cpu: "250m"
            ephemeralStorage: "500Mi"
        
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
        
        volumeMounts:
        - name: workspace
          mountPath: /workspace
        - name: tmp
          mountPath: /tmp
        - name: home
          mountPath: /home/executor
      
      volumes:
      - name: workspace
        persistentVolumeClaim:
          claimName: user-abc123-workspace-pvc
      - name: tmp
        emptyDir:
          sizeLimit: 500Mi
      - name: home
        emptyDir:
          sizeLimit: 100Mi

---
# 用户workspace PVC
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: user-abc123-workspace-pvc
  namespace: pomclaw
spec:
  accessModes:
  - ReadWriteOnce
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 10Gi
```

#### 4.4.3 Go中的Kubernetes集成

```go
// pkg/executor/k8s_executor.go
package executor

import (
    "context"
    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

type K8sExecutor struct {
    clientset *kubernetes.Clientset
    namespace string
}

func (e *K8sExecutor) Execute(ctx context.Context, userID string, command string) (string, error) {
    // 创建Job
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name: fmt.Sprintf("exec-%s-%d", userID, time.Now().UnixNano()),
        },
        Spec: batchv1.JobSpec{
            TTLSecondsAfterFinished: &[]int32{60}[0],
            BackoffLimit:            &[]int32{0}[0],
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    Containers: []corev1.Container{
                        {
                            Name:  "executor",
                            Image: "pomclaw-executor:latest",
                            Command: []string{"/bin/bash", "-c"},
                            Args: []string{command},
                            Resources: corev1.ResourceRequirements{
                                Limits: corev1.ResourceList{
                                    corev1.ResourceMemory: resource.MustParse("256Mi"),
                                    corev1.ResourceCPU:    resource.MustParse("500m"),
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    // 提交Job
    createdJob, err := e.clientset.BatchV1().Jobs(e.namespace).Create(ctx, job, metav1.CreateOptions{})
    if err != nil {
        return "", err
    }

    // 等待完成（通过watch）
    // ... 实现省略
    
    return "", nil
}
```

---

## 5. 成本分析

### 5.1 基础设施成本（月度）

| 方案 | 1000用户 | 5000用户 | 10000用户 | 备注 |
|------|---------|---------|-----------|------|
| **系统用户** | $50 | $200 | $400 | 单个Node |
| **Docker** | $100 | $500 | $1000 | 容器引擎开销 |
| **K8s** | $300 | $1000 | $2000+ | 管理/网络开销 |

### 5.2 开发成本估算

```
系统用户方案：
  初始实现：3-5天（1人）
  测试/调试：2-3天（1人）
  文档/交付：1-2天（1人）
  ────────────────────────
  总计：7-10天工作量

Docker方案：
  初始实现：5-7天（1人）
  镜像优化：2-3天（1人）
  K8s集成：3-5天（1人）
  ────────────────────────
  总计：10-15天工作量

K8s方案：
  基础设施：5-10天（DevOps）
  应用集成：5-7天（开发）
  运维培训：2-3天（DevOps）
  ────────────────────────
  总计：12-20天工作量
```

### 5.3 总体拥有成本(TCO)

```
3个月成本对比（假设月服务器$100）：

系统用户方案：
  开发：$5000（1人×10天×$500/天）
  基础设施：$300（3月）
  ─────────
  总计：$5,300

Docker方案：
  开发：$7,500（1.5人×10天）
  基础设施：$300（3月）
  ─────────
  总计：$7,800

K8s方案：
  开发：$10,000（2人×10天）
  基础设施：$900（3月）
  ─────────
  总计：$10,900

结论：系统用户方案成本最低，适合MVP阶段
```

---

## 6. 实现路线图

### 6.1 推荐阶段划分

#### Phase 0：数据库隔离（现在）
```
状态：✅ 规划中（ENTERPRISE_ARCHITECTURE.md）

任务：
  ☐ 在现有表添加 user_id/org_id 字段
  ☐ 所有查询加租户过滤
  ☐ 实现 RLS（行级安全）
  ☐ 测试租户隔离

预计工作量：3-5 天（1 后端工程师）
```

#### Phase 1：系统用户文件隔离（下2周）
```
状态：☐ 待开始

任务：
  ☐ 修改 ExecTool 支持 userID 参数
  ☐ 实现 workspace 目录隔离
  ☐ 创建系统用户管理脚本
  ☐ 集成到 API Gateway
  ☐ 单元测试 + 集成测试
  ☐ 安全审计

预计工作量：1 周（1 后端 + 1 DevOps）

Git分支：feature/user-isolation-v1
PR模板：支持多用户文件隔离
```

#### Phase 2：可选增强（1个月后）
```
状态：☐ 条件启动

条件：
  - 系统用户方案运行满意（2周+）
  - 或用户数突破 500+

选项A：chroot 增强
  - 额外隔离 / etc 等系统目录
  - 实现难度：中等
  - 收益：20% 安全性提升

选项B：容器化试点
  - 在高风险场景使用容器
  - 与系统用户方案并存
  - 实现难度：高
  - 收益：隔离更强但成本增加
```

#### Phase 3：Kubernetes 编排（6个月+）
```
状态：☐ 远期规划

触发条件：
  - 用户突破 5000+
  - 需要跨地域部署
  - 现有方案性能瓶颈

工作量：3-4 周（专业 DevOps 团队）
```

### 6.2 风险与缓解

| 风险 | 概率 | 影响 | 缓解方案 |
|------|------|------|---------|
| 系统用户数突破预留上限 | 低 | 中 | 动态创建用户/容器迁移 |
| 新增用户创建失败 | 低 | 高 | 降级为Docker容器 |
| Linux/Windows兼容性 | 中 | 中 | Windows用分离实现 |
| 权限配置复杂 | 中 | 中 | 提供配置验证脚本 |

---

## 7. 选型建议

### 7.1 决策树

```
                    ┌─ 现在还是未来？
                    │
        ┌───────────┴───────────┐
        │                       │
      现在(MVP)              未来(扩展)
        │                       │
        ▼                       ▼
    <500用户?               <5000用户?
        │                       │
    是/否                     是/否
    /   \                    /   \
   是   否                  是    否
   │    │                   │     │
   ▼    ▼                   ▼     ▼
 系统  Docker            Docker  K8s
 用户  容器              容器    编排
```

### 7.2 我的强烈推荐

```
🎯 立即采用：系统用户隔离（方案A）

理由：
  ✅ 实现快（1周）
  ✅ 成本低（$5K开发+$50/月）
  ✅ 安全足够（中等隔离强度）
  ✅ 易维护（基于Unix标准）
  ✅ 可升级（Docker/K8s兼容）

后续迁移路径：
  系统用户 → Docker容器（500+用户）
           → K8s编排（5000+用户）

关键点：不要跳级。先用系统用户方案获得可验证的隔离，
再根据实际需求升级到容器方案。
```

---

## 附录

### A. 系统用户隔离：完整配置文件

#### docker-compose.yml with Sandbox Users

```yaml
version: '3.8'

services:
  postgres:
    image: pgvector/pgvector:pg16
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data

  pomclaw:
    build: .
    depends_on:
      - postgres
    environment:
      POM_STORAGE_TYPE: postgres
      POM_POSTGRES_HOST: postgres
    volumes:
      - /workspace:/workspace  # 必须挂载主机的workspace目录
    cap_drop:
      - ALL
    cap_add:
      - SETUID       # 允许切换用户（sudo）
      - SETGID       # 允许切换组
    init: true
    command: >
      sh -c "
        /scripts/init-sandbox-users.sh &&
        /build/pomclaw gateway
      "

volumes:
  postgres-data:
```

#### init-sandbox-users.sh

```bash
#!/bin/bash
set -e

echo "Initializing sandbox users..."

# 创建组
if ! getent group exec_group > /dev/null; then
    groupadd -f exec_group
fi

# 创建workspace基础目录
mkdir -p /workspace
chmod 755 /workspace

# 创建N个预留用户账户（示例：10个）
for i in {0..9}; do
    hash=$(printf '%06d' $i)
    username="exec_user_${hash}"
    uid=$((2000 + i))
    userdir="/workspace/user_${hash}"
    
    # 创建用户
    if ! id "$username" > /dev/null 2>&1; then
        useradd \
            -u $uid \
            -g exec_group \
            -s /sbin/nologin \
            -m \
            -d $userdir \
            $username 2>/dev/null || true
        echo "Created user: $username (UID: $uid)"
    fi
    
    # 设置目录权限
    mkdir -p $userdir
    chmod 700 $userdir
    chown $username:exec_group $userdir
    
    # 为用户创建基础shell环境
    su - $username -s /bin/bash -c "mkdir -p ~/.ssh; chmod 700 ~/.ssh" 2>/dev/null || true
done

echo "Sandbox users initialized successfully"
```

### B. Docker 隔离：完整 Dockerfile

```dockerfile
# pomclaw-executor:latest
FROM alpine:3.19

RUN apk add --no-cache \
    bash \
    curl \
    wget \
    git \
    openssh-client \
    ca-certificates \
    jq

# 创建非 root 用户
RUN addgroup -S executor && \
    adduser -S executor -G executor

# 创建工作目录
RUN mkdir -p /workspace && \
    chown -R executor:executor /workspace

WORKDIR /workspace
USER executor

ENTRYPOINT ["/bin/bash"]
```

### C. 监控与审计脚本

```bash
#!/bin/bash
# scripts/monitor-isolation.sh

echo "=== Sandbox User Status ==="
getent group exec_group | cut -d: -f4 | tr ',' '\n' | while read user; do
    if [ -n "$user" ]; then
        uid=$(id -u $user 2>/dev/null)
        printf "%-25s UID: %d\n" "$user" "$uid"
    fi
done

echo ""
echo "=== Workspace Permissions ==="
ls -lhd /workspace/* 2>/dev/null | awk '{print $1, $3, $4, $9}'

echo ""
echo "=== Process Isolation ==="
ps aux | grep -E "exec_user_" | grep -v grep | wc -l
echo "Active sandbox processes"

echo ""
echo "=== Resource Limits ==="
for user in $(getent group exec_group | cut -d: -f4 | tr ',' ' '); do
    echo "User: $user"
    su - $user -c "ulimit -a" 2>/dev/null | grep -E "^(max memory|max processes|open files)"
done
```

### D. 迁移清单（系统用户 → Docker）

```markdown
## 从系统用户隔离迁移到 Docker 隔离

### 触发条件
- [ ] 并发用户数 >= 500
- [ ] 系统用户数超过 100+
- [ ] 内存占用 > 2GB
- [ ] 运维难度反馈

### 迁移步骤

1. **准备阶段（1周）**
   - [ ] 创建 pomclaw-executor Docker 镜像
   - [ ] 编写 Docker 隔离实现（DockerExecutor）
   - [ ] 配置容器资源限制
   - [ ] 灾难恢复演练

2. **切换阶段（1天）**
   - [ ] 配置特性开关（ExecMode: "docker"）
   - [ ] 新请求使用 Docker 隔离
   - [ ] 旧请求用系统用户隔离（兼容）
   - [ ] 监控错误率 + 性能

3. **验证阶段（3天）**
   - [ ] 性能对标测试
   - [ ] 隔离强度验证
   - [ ] 数据一致性检查
   - [ ] 用户反馈收集

4. **完全迁移（1周）**
   - [ ] 全量用户切换到 Docker
   - [ ] 清理系统用户账户
   - [ ] 关闭 exec_group
   - [ ] 更新文档
```

---

## 文档版本

- **版本**：1.0
- **更新日期**：2026-04-19
- **作者**：PomClaw架构组
- **审核状态**：待审核
- **下次审核**：2026-05-19（或根据实施情况）

## 反馈与改进

本文档将根据以下事件更新：
- 系统用户隔离方案实施反馈
- 性能数据验证
- 安全审计发现
- 用户需求变化
