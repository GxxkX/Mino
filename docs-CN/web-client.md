# Web 客户端

本文档介绍 Mino Web 应用程序，该应用基于 Next.js 14 和
React 18 构建。Web 客户端提供了完整的功能集，用于管理录音、
记忆、任务以及与 AI 助手进行对话。

## 技术栈

Web 应用程序使用以下技术。

| 技术 | 用途 |
|------|------|
| Next.js 14 | App Router、服务端渲染、API 代理 |
| React 18 | UI 组件 |
| Zustand 4.5 | 全局状态管理 |
| Tailwind CSS | 自定义暗色主题样式 |
| Lucide React | 图标库 |
| TypeScript | 类型安全 |

## 项目结构

Web 应用程序遵循 Next.js App Router 约定。

```
web/
├── app/                    # 页面和布局
│   ├── layout.tsx          # 根布局（zh-CN 语言环境）
│   ├── page.tsx            # 重定向到 /auth/login
│   ├── globals.css         # 全局样式和 CSS 变量
│   ├── auth/
│   │   └── login/
│   │       └── page.tsx    # 登录页面
│   └── dashboard/
│       ├── layout.tsx      # 仪表盘外壳（侧边栏 + 内容区）
│       ├── page.tsx        # 首页：统计数据、最近项目
│       ├── memories/       # 记忆管理
│       ├── tasks/          # 任务管理
│       ├── audio/          # 音频录制
│       ├── chat/           # AI 对话界面
│       ├── extensions/     # 扩展管理
│       └── settings/       # 设置（云同步、MCP）
├── components/
│   ├── ui/                 # 可复用基础组件
│   ├── layout/             # 侧边栏、头部
│   └── features/           # 业务领域组件
├── lib/
│   ├── api/                # 后端 API 客户端模块
│   ├── store.ts            # Zustand 全局 store
│   ├── utils.ts            # 工具函数
│   └── demo-data.ts        # 开发用演示数据
└── types/
    └── index.ts            # TypeScript 接口定义
```

## 页面

Web 应用程序提供以下页面。

### 登录页（`/auth/login`）

用户名和密码认证表单。登录成功后，access token 和
refresh token 将存储在 `localStorage` 中，用户随后被
重定向到仪表盘。

### 仪表盘（`/dashboard`）

首页展示数据概览：对话总数、记忆总数和任务总数。页面以
网格布局展示最近的记忆、最近的对话和待处理的任务。

### 记忆（`/dashboard/memories`）

列出所有提取的记忆，支持按类别（洞察、事实、偏好、事件）
筛选和文本搜索。每张记忆卡片显示内容、类别图标、重要程度
（以 5 点指示器展示）以及指向源对话音频的链接。

### 任务（`/dashboard/tasks`）

列出所有任务，支持按状态（待处理、进行中、已完成、已取消）
筛选。统计栏显示各状态的数量。任务显示标题、描述、优先级
标签、截止日期，以及用于快速更新的内联状态选择下拉菜单。

### 音频（`/dashboard/audio`）

列出所有录制的对话，支持搜索功能。每张对话卡片显示标题、
摘要、时长、标签以及音频播放/暂停按钮。统计栏显示录制
总数、总时长和已完成数量。

### 对话（`/dashboard/chat`）

完整的对话界面，左侧为会话侧边栏，右侧为消息区域。你可以
创建、重命名和删除对话会话。消息以用户/助手样式展示，包含
时间戳和链接到原始对话的来源引用。

### 扩展（`/dashboard/extensions`）

管理自定义扩展，支持启用/禁用切换。你可以添加新扩展和删除
现有扩展。每个扩展包含名称、描述、图标和 JSON 配置。

### 设置（`/dashboard/settings`）

设置中心提供主题、语言、通知偏好和 LLM 提供商选择的配置。
子页面包括云同步设置（MinIO 和 PostgreSQL 连接配置）以及
MCP 协议设置（启用/禁用及服务器配置）。

## API 层

Web 应用程序通过客户端 API 层与后端通信。所有 API 调用
通过 Next.js 重写代理进行，该代理将 `/api/v1/*` 请求
转发到后端。

### API 客户端

基础 API 客户端（`lib/api/client.ts`）提供以下功能。

- 使用 `localStorage` 自动管理 JWT token
- 在收到 401 响应时自动刷新 token
- 认证失败时重定向到登录页面
- 类型化的请求/响应封装

### API 模块

每个后端资源都有专用的 API 模块。

| 模块 | 文件 | 端点 |
|------|------|------|
| Auth | `lib/api/auth.ts` | 登录、登出、修改密码 |
| Conversations | `lib/api/conversations.ts` | 列表、获取、删除 |
| Memories | `lib/api/memories.ts` | 列表、获取、更新、删除 |
| Tasks | `lib/api/tasks.ts` | 列表、创建、更新、删除 |
| Chat | `lib/api/chat.ts` | 会话 CRUD、消息 |
| Extensions | `lib/api/extensions.ts` | 完整 CRUD |
| Search | `lib/api/search.ts` | 搜索、重建索引 |

所有模块通过 `lib/api/index.ts` 以命名空间对象的形式
重新导出（`authApi`、`conversationsApi`、`memoriesApi`、
`tasksApi`、`chatApi`、`extensionsApi`、`searchApi`）。

## 状态管理

应用程序使用单一的 Zustand store（`lib/store.ts`）来管理
所有客户端状态。

| 状态 | 描述 |
|------|------|
| `user` | 当前已认证用户 |
| `conversations` | 对话列表 |
| `memories` | 记忆列表 |
| `tasks` | 任务列表 |
| `extensions` | 扩展列表 |
| `chatSessions` | 对话会话列表 |
| `activeSessionId` | 当前选中的对话会话 |
| `chatMessages` | 当前活跃会话的消息 |
| `settings` | 应用设置（主题、语言、LLM 配置） |
| `isRecording` | 是否正在录制 |
| `currentTranscript` | 录制过程中的实时转录文本 |

## 设计系统

Web 应用程序使用针对 OLED 显示屏优化的暗色主题。调色板
在 `tailwind.config.js` 和 `globals.css` 中定义。

| 令牌 | 值 | 用途 |
|------|------|------|
| Background | `#09090b` | 页面背景 |
| Surface | `#18181b` | 卡片背景 |
| Border | `#27272a` | 边框和分隔线 |
| Text primary | `#fafafa` | 主要文本 |
| Text secondary | `#a1a1aa` | 次要文本 |
| CTA / Accent | `#a3e635` | 按钮、激活状态 |

排版使用 Inter 作为正文字体，Newsreader 作为衬线装饰字体。
应用程序支持 `prefers-reduced-motion` 以提升无障碍体验。

## 全局搜索

头部包含一个全局搜索栏，可通过点击搜索图标或按下
Cmd+K（Windows/Linux 上为 Ctrl+K）激活。搜索查询通过
防抖机制发送到基于 Typesense 的搜索 API。结果以下拉列表
形式展示，显示来自对话和记忆的匹配项。

## 录制

头部包含一个录制切换按钮。当录制处于活跃状态时，头部下方
会出现一个 `RecordingBanner` 组件，显示实时转录内容。
录制状态和转录文本通过 Zustand store 进行管理。

## 开发

### 安装依赖

```bash
cd web
npm install
```

### 启动开发服务器

```bash
npm run dev
```

应用将在 `http://localhost:3000` 启动。

### 生产环境构建

```bash
npm run build
```

### 类型检查

```bash
npm run typecheck
```

### 代码检查

```bash
npm run lint
```

## 配置

Web 应用程序使用以下环境变量。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `API_URL` | `http://localhost:8000` | 代理重写使用的后端 API 地址 |
| `NEXT_PUBLIC_API_URL` | `/api/v1` | 客户端 API 基础 URL |

`next.config.js` 文件配置了一条重写规则，将 `/api/v1/*`
请求代理到后端，以避免开发过程中的 CORS 问题。

## 后续步骤

有关后端 API 的详细信息，请参阅 [API 参考](api-reference.md)。
有关 Web 应用的生产部署，请参阅[部署指南](deployment.md)。
