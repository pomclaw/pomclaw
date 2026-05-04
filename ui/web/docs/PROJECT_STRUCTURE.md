# GoClaw Web UI 项目目录结构

## 📁 核心功能模块 (src/pages/)

### ✅ agents/ - Agent 管理 (47 个组件)
**作用**: Agent 的完整生命周期管理
- agent-card.tsx - Agent 卡片展示
- agent-create-dialog.tsx - 创建 Agent 对话框
- agents-page.tsx - Agent 列表页面
- summoning-modal.tsx - Agent 召唤/创建进度模态框

**子目录**:
- **agent-detail/** - Agent 详情页面
  - **config-sections/** (10个组件) - 配置区块: 压缩、上下文裁剪、内存、沙箱、思考、工具策略、工作区共享、ChatGPT OAuth 路由等
  - **file-sections/** (8个组件) - 文件管理: 文件编辑器、侧边栏、系统提示预览、联系人搜索、重新生成对话框等
  - **overview-sections/** (12个组件) - 概览区块: 能力、引擎版本、Hooks 摘要、模型预算、编排、个性化、固定技能、提示设置、技能、TTS、演化、心跳卡片等
  - **general-sections/** (1个组件) - 通用配置: 工作区配置
  - **hooks/** (5个) - Agent 相关 Hooks: 详情、实例、技能、配置权限、联系人搜索

### ✅ chat/ - 聊天界面 (3 个组件)
**作用**: 与 Agent 的实时对话界面
- chat-page.tsx - 聊天主页面
- chat-sidebar.tsx - 会话列表侧边栏
- chat-thread.tsx - 消息线程显示
- **hooks/** (4个) - 聊天相关 Hooks: 消息、发送、会话、团队任务

### ✅ overview/ - 仪表板 (8 个组件)
**作用**: 系统概览和监控面板
- overview-page.tsx - 主仪表板页面
- system-health-card.tsx - 系统健康状态卡片
- connected-clients-card.tsx - 已连接客户端统计
- cron-jobs-card.tsx - 定时任务列表
- quota-usage-card.tsx - 配额使用情况
- recent-requests-card.tsx - 最近的 API 请求
- channel-attention-panel.tsx - 频道关注面板
- stat-card.tsx - 通用统计卡片组件
- **hooks/** (2个) - 实时运行时间、Sparklines 图表

### ✅ login/ - 登录认证 (6 个组件)
**作用**: 用户登录和租户选择
- login-page.tsx - 登录页面主体
- login-layout.tsx - 登录页面布局
- login-tabs.tsx - 登录方式切换 (Token/配对)
- token-form.tsx - Token 登录表单
- pairing-form.tsx - 浏览器配对登录表单
- tenant-selector.tsx - 多租户选择页面

### ✅ setup/ - 首次设置向导 (9 个组件)
**作用**: 新用户引导设置流程 (Provider → Model → Agent → Channel)
- setup-page.tsx - 设置向导主页
- setup-layout.tsx - 设置页面布局
- setup-stepper.tsx - 步骤指示器
- step-provider.tsx - Provider 配置步骤
- step-model.tsx - 模型选择步骤
- step-agent.tsx - Agent 创建步骤
- step-channel.tsx - 频道配置步骤 (可选)
- setup-complete-modal.tsx - 完成引导模态框
- info-tip.tsx - 提示信息组件

### ⚠️ providers/ - Provider 管理 (3 个组件)
**作用**: LLM Provider 配置和管理
- provider-cli-section.tsx - Claude CLI Provider 配置
- provider-oauth-section.tsx - ChatGPT OAuth Provider 配置
- provider-utils.tsx - Provider 工具函数
- **hooks/** (6个) - Provider 相关 Hooks: 模型列表、配额、状态、Pool 活动、验证、Provider 列表

### ⚠️ channels/ - 频道管理 (2 个组件)
**作用**: Telegram/WhatsApp/Zalo/Feishu 等频道配置
- channel-fields.tsx - 频道表单字段
- channels-status-view.tsx - 频道状态视图
- **hooks/** (1个) - 频道实例管理 Hook
- channel-schemas.ts - 频道验证规则
- channels-status-utils.ts - 状态工具函数

### ⚠️ hooks/ - Hooks 管理 (3 个组件)
**作用**: 系统 Hook (事件钩子) 管理界面
- components/hook-list-row.tsx - Hook 列表行
- components/hook-form-dialog.tsx - Hook 创建/编辑表单
- components/hook-test-panel.tsx - Hook 测试面板

### ❌ builtin-tools/ - 内置工具 (0 个 TSX)
**状态**: 已精简，只保留 Hook
- 只有 hooks/use-builtin-tools.ts
- **建议**: 可能功能已合并到其他页面，保留 Hook 供其他模块调用

### ❌ config/ - 系统配置 (0 个 TSX)
**状态**: 已精简，只保留 Hook
- 只有 hooks/use-config.ts, hooks/use-config-defaults.ts
- **建议**: 配置功能可能集成在其他页面 (如 Overview 或设置菜单)

### ❌ skills/ - 技能管理 (0 个 TSX)
**状态**: 已精简，只保留 Hook
- 只有 hooks/use-skills.ts, hooks/use-runtimes.ts
- **建议**: 技能管理可能集成在 Agent 详情页面的技能区块

### ❌ traces/ - 追踪日志 (0 个 TSX)
**状态**: 已精简，只保留 Hook
- 只有 hooks/use-traces.ts
- **建议**: 追踪功能可能未完全实现或已移除

### ❌ tts/ - 文本转语音 (0 个 TSX)
**状态**: 已精简，只保留 Hook
- 只有 hooks/use-tts-config.ts
- **建议**: TTS 配置可能集成在 Agent 配置或系统设置中

---

## 📁 支撑模块 (src/)

### components/ - UI 组件库
- **layout/** - 布局组件: AppLayout, Sidebar, Header
- **providers/** - Context Providers: ThemeProvider, WsProvider, AppProviders
- **shared/** - 共享组件: ErrorBoundary, RequireAuth, RequireSetup, ConfirmDialog, EmptyState, LoadingSkeleton
- **ui/** - 基础 UI 组件库 (基于 shadcn/ui): Button, Dialog, Input, Select, Tabs, Toast, 等

### hooks/ - 全局 Hooks
- use-ws.ts - WebSocket 和 HTTP 客户端 Hook
- use-ws-event.ts - WebSocket 事件监听
- use-query-invalidation.ts - React Query 缓存失效管理
- use-virtual-keyboard.ts - 移动端虚拟键盘高度处理
- use-min-loading.ts - 最小加载时间保证
- use-hooks.ts - Hook 管理 (创建、更新、删除、切换、列表)

### stores/ - Zustand 状态管理
- use-auth-store.ts - 认证状态: token, userId, senderID, connected, role, tenant 等
- use-ui-store.ts - UI 状态: theme, language, timezone, sidebar 折叠状态
- use-toast-store.ts - Toast 通知状态
- use-team-event-store.ts - 团队事件状态

### api/ - API 客户端
- ws-client.ts - WebSocket 客户端 (RPC 通信)
- http-client.ts - HTTP 客户端 (REST API)
- protocol.ts - 协议定义: Methods, Events, Frame 结构

### i18n/ - 国际化 (i18next)
- **locales/en/** - 英文翻译文件
- **locales/vi/** - 越南语翻译文件
- **locales/zh/** - 中文翻译文件
- 分模块组织: agents.json, chat.json, common.json, login.json, setup.json 等

### types/ - TypeScript 类型定义
- agent.ts - Agent 相关类型
- provider.ts - Provider 相关类型
- session.ts - 会话类型
- tenant.ts - 租户类型
- channel.ts - 频道类型
- skill.ts - 技能类型

### lib/ - 工具函数库
- routes.ts - 路由常量定义
- constants.ts - 全局常量: LocalStorage 键、语言列表
- format.ts - 格式化工具: 日期、时区、Token 数
- utils.ts - 通用工具: cn (classnames), 字符串处理
- timezone-utils.ts - 时区处理工具
- setup-skip.ts - 设置向导跳过状态
- slug.ts - Slug 验证

### schemas/ - Zod 数据验证
- agent.schema.ts - Agent 表单验证规则
- hooks.schema.ts - Hooks 表单验证规则
- provider.schema.ts - Provider 验证规则

### adapters/ - 适配器层
- provider-pool.adapter.ts - Provider Pool 适配器 (ChatGPT OAuth Pool 相关)

### data/ - 静态数据
- countries.ts - 国家/地区列表
- phone-prefixes.ts - 电话区号前缀

### constants/ - 常量定义
- agents.ts - Agent 相关常量

---

## 🎯 精简分析

### ✅ 已精简的核心功能 (保留 3 个核心模块)
1. **agents/** - Agent 管理 (47 个组件) ✅ 核心
2. **chat/** - 聊天界面 (3 个组件) ✅ 核心
3. **overview/** - 仪表板 (8 个组件) ✅ 核心

### ✅ 已精简的支撑功能 (保留必要模块)
4. **login/** - 登录认证 (6 个组件) ✅ 必需
5. **setup/** - 首次设置向导 (9 个组件) ✅ 必需

### ⚠️ 保留但组件数少的模块
6. **providers/** - 3 个组件 (功能重要，保留)
7. **channels/** - 2 个组件 (功能重要，保留)
8. **hooks/** - 3 个组件 (功能重要，保留)

### ❌ 已精简为仅保留 Hook 的模块 (5 个)
9. **builtin-tools/** - 0 个 TSX，只有 Hook
10. **config/** - 0 个 TSX，只有 Hook
11. **skills/** - 0 个 TSX，只有 Hook
12. **traces/** - 0 个 TSX，只有 Hook
13. **tts/** - 0 个 TSX，只有 Hook

**这些目录的功能已集成到其他页面，Hook 仍被其他模块引用，不能删除。**

---

## 📊 统计数据

### 组件分布
**总 TSX 文件数**: ~81 个

| 模块 | 组件数 | 占比 | 状态 |
|------|--------|------|------|
| agents/ | 47 | 58% | ✅ 核心 |
| setup/ | 9 | 11% | ✅ 必需 |
| overview/ | 8 | 10% | ✅ 核心 |
| login/ | 6 | 7% | ✅ 必需 |
| chat/ | 3 | 4% | ✅ 核心 |
| providers/ | 3 | 4% | ⚠️ 保留 |
| hooks/ | 3 | 4% | ⚠️ 保留 |
| channels/ | 2 | 2% | ⚠️ 保留 |
| 其他 (空目录) | 0 | 0% | ❌ 已精简 |

### 精简成果
- **核心 3 模块** (agents + chat + overview) 占 **72%**
- **前 5 模块** (+ login + setup) 占 **90%**
- **已精简 5 个目录**为仅保留 Hook (builtin-tools, config, skills, traces, tts)
- **删除的文件**: Evolution Tab, Agent Hooks Tab, Heartbeat Config Dialog, Codex Pool 相关组件

---

## 🔍 下一步精简建议

### 1. 检查是否还有未使用的 Hook
可以检查 5 个"空目录"中的 Hook 是否真的被使用：
```bash
# 检查 use-builtin-tools.ts 的引用
grep -r "use-builtin-tools" src/

# 检查 use-config.ts 的引用
grep -r "use-config" src/

# 检查 use-skills.ts 的引用
grep -r "use-skills" src/

# 检查 use-traces.ts 的引用
grep -r "use-traces" src/

# 检查 use-tts-config.ts 的引用
grep -r "use-tts-config" src/
```

### 2. 删除未引用的 Hook
如果某些 Hook 完全没有被引用，可以安全删除整个目录。

### 3. 检查 adapters/ 目录
`provider-pool.adapter.ts` 是否还在使用？如果没有，可以删除。

### 4. 检查 data/ 和 constants/ 目录
检查这些静态数据文件是否都在使用。

---

## ✨ 项目优点

1. **模块化清晰**: 功能按页面/特性分离
2. **TypeScript 完整**: 所有类型都有明确定义
3. **国际化完善**: 支持 3 种语言 (en/vi/zh)
4. **状态管理统一**: Zustand + React Query
5. **组件复用高**: shared 组件库丰富
6. **测试覆盖**: 关键组件有单元测试

## 🎯 架构特点

- **SPA 应用**: React Router 7
- **实时通信**: WebSocket + HTTP 双协议
- **响应式设计**: Tailwind CSS 4 + 移动端优化
- **UI 组件**: Radix UI + shadcn/ui
- **构建工具**: Vite 6 + pnpm
