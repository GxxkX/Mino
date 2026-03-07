# 手表客户端

本文档介绍 Mino 智能手表应用程序，该应用使用 Flutter 构建，
适用于 Wear OS 设备。手表应用专注于快速语音采集和实时转录显示。

## 技术栈

手表应用使用以下技术。

| 技术 | 用途 |
|------|------|
| Flutter (Dart 3.0+) | 跨平台 UI 框架 |
| Provider | 状态管理（ChangeNotifier 模式） |
| web_socket_channel | 实时音频流传输 |
| record | 音频录制（16kHz 单声道） |
| sqflite | 本地 SQLite 数据库，用于离线存储 |
| shared_preferences | Token 和设置持久化 |

## 项目结构

```
watch/
├── android/
│   └── app/
│       ├── build.gradle.kts       # Android 配置 (SDK 25-34)
│       └── src/main/
│           ├── AndroidManifest.xml # 权限声明
│           └── kotlin/.../
│               └── MainActivity.kt
├── assets/
│   └── icons/
├── lib/
│   ├── main.dart                  # 应用入口，MultiProvider 设置
│   ├── app/
│   │   ├── screens/
│   │   │   ├── home_screen.dart   # 录音界面 + 转录文本
│   │   │   ├── login_screen.dart  # 身份认证
│   │   │   ├── permission_screen.dart
│   │   │   └── settings_screen.dart
│   │   └── widgets/
│   │       ├── recording_button.dart
│   │       └── transcript_display.dart
│   ├── core/
│   │   ├── constants/
│   │   │   └── app_config.dart    # API 地址、音频参数
│   │   ├── providers/
│   │   │   └── auth_provider.dart
│   │   ├── services/
│   │   │   ├── auth_service.dart
│   │   │   └── database_service.dart
│   │   └── theme/
│   │       └── app_theme.dart
│   ├── features/
│   │   ├── recording/
│   │   │   ├── models/
│   │   │   │   └── audio_config.dart
│   │   │   ├── providers/
│   │   │   │   └── recording_provider.dart
│   │   │   └── services/
│   │   │       ├── audio_recording_service.dart
│   │   │       └── network_service.dart
│   │   └── settings/
│   │       └── providers/
│   │           └── settings_provider.dart
│   └── shared/
│       ├── providers/
│       │   └── connection_provider.dart
│       └── widgets/
│           └── battery_indicator.dart
└── pubspec.yaml
```

## 应用流程

手表应用遵循针对小屏幕优化的线性流程。

1. **登录**：输入用户名和密码，通过后端进行身份认证。
   Token 存储在 `SharedPreferences` 中。
2. **权限检查**：如果尚未授予麦克风权限，则请求该权限。
3. **主页（录音）**：主屏幕包含一个大型录音按钮和实时转录
   显示区域。
4. **设置**：账户信息、录音偏好设置和退出登录。

## 身份认证

手表应用使用与其他客户端相同的 REST API 进行身份认证。
它调用 `POST /v1/auth/signin` 接口，传入用户名和密码，
然后将访问令牌和刷新令牌存储在 `SharedPreferences` 中。

`AuthProvider` 管理认证状态，并通过 Provider 模式将当前
用户信息暴露给组件树。

## 录音

录音功能是手表应用的核心功能。

### 音频配置

音频采集使用以下设置，定义在
`lib/features/recording/models/audio_config.dart` 中。

| 参数 | 值 |
|------|-----|
| 采样率 | 16,000 Hz |
| 声道 | 单声道 |
| 编码格式 | PCM16 / AAC-LC |
| 比特率 | 32 kbps |
| 分片间隔 | 100 ms |

### 录音流程

`RecordingProvider` 管理完整的录音生命周期。

1. 与后端建立 WebSocket 连接，地址为
   `/v1/ws/audio?token=<jwt_token>`。
2. 使用 `record` 包开始音频采集。
3. 每 100 毫秒通过 WebSocket 发送音频分片。
4. 接收并显示来自后端的实时转录更新。
5. 当用户点击按钮或达到最大录音时长时停止录音。
6. 接收最终处理结果（标题、摘要、待办事项、记忆）。

### 录音状态

应用通过 `RecordingState` 枚举跟踪录音状态。

| 状态 | 描述 |
|------|------|
| `idle` | 准备录音 |
| `connecting` | 正在建立 WebSocket 连接 |
| `recording` | 正在录音和流式传输 |
| `processing` | 录音已停止，等待 AI 处理 |
| `completed` | 处理完成 |
| `error` | 发生错误 |

## 离线支持

当网络不可用时，手表应用使用 SQLite 在本地存储录音。

`DatabaseService` 管理两个 SQLite 表。

| 表名 | 用途 |
|------|------|
| `conversations` | 缓存的会话元数据 |
| `offline_recordings` | 待上传的录音 |

当网络恢复时，`SyncProvider` 检测到网络变化并将待上传的
录音上传到后端。上传成功后，本地记录将被标记为已同步。

## 设置

`SettingsProvider` 管理存储在 `SharedPreferences` 中的
用户偏好设置。

| 设置项 | 描述 |
|--------|------|
| 音频质量 | 录音质量级别 |
| 最大时长 | 最大录音长度 |
| 离线模式 | 启用/禁用离线录音 |
| 自动同步 | 联网时自动同步 |

## UI 组件

### 录音按钮

一个圆形动画按钮（`recording_button.dart`），在麦克风图标
和停止图标之间切换。录音过程中按钮会播放动画以提供视觉反馈。

### 转录显示

一个可滚动的文本区域（`transcript_display.dart`），在用户
说话时显示实时转录文本。文本更新通过 WebSocket 连接传入。

### 电量指示器

一个小型组件（`battery_indicator.dart`），以颜色编码
（绿色、黄色、红色）显示当前电池电量。

## 主题

手表应用使用定义在 `lib/core/theme/app_theme.dart` 中的
深色 OLED 主题。配色方案与项目设计系统一致，采用深色背景
以提高 OLED 屏幕的电源效率。

## Android 配置

手表应用面向 Wear OS 设备，使用以下 Android 配置。

| 设置项 | 值 |
|--------|-----|
| Namespace | `com.mino.watch` |
| Min SDK | 25 (Android 7.1) |
| Target SDK | 34 (Android 14) |
| 屏幕尺寸 | 400 x 492 dp |

### 所需权限

`AndroidManifest.xml` 声明了以下权限。

- `INTERNET`：用于网络通信
- `RECORD_AUDIO`：用于麦克风访问
- `BLUETOOTH` 和 `BLUETOOTH_ADMIN`：用于蓝牙连接
- `READ_EXTERNAL_STORAGE` 和 `WRITE_EXTERNAL_STORAGE`：
  用于离线文件存储

## 后端连接

手表应用使用 `lib/core/constants/app_config.dart` 中配置的
URL 连接后端。

| 端点 | URL |
|------|-----|
| REST API | `http://<host>:8000/v1` |
| WebSocket | `ws://<host>:8000/v1/ws/audio` |

在构建应用之前，请更新这些值以指向你的后端服务器。

## 开发

### 前置条件

安装 Flutter 3.0 或更高版本，以及 Dart SDK 3.0+。

### 在设备上运行

```bash
cd watch
flutter run
```

### 构建发布版 APK

```bash
flutter build apk --release
```

发布版 APK 可以直接安装到 Wear OS 设备上，也可以通过你
偏好的渠道进行分发。

## 后续步骤

有关后端 API 的详细信息，请参阅 [API 参考](api-reference.md)。
有关设置手表连接的后端，请参阅
[快速入门](getting-started.md) 指南。
