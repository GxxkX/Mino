# Whisper API 服务

本地 Whisper 语音转文字 API 服务，支持实时流式转录和完整文件转录。

## 功能特性

- ✅ 支持多种 Whisper 模型（tiny/base/small/medium/large/turbo）
- ✅ 动态模型加载和缓存
- ✅ API Key 认证（可选）
- ✅ 实时流式转录 (`/transcribe_stream`)
- ✅ 完整文件转录 (`/transcribe`)
- ✅ GPU 加速支持

## 快速开始

### 1. 构建镜像

```bash
cd whisper
docker build -t whisper-api .
```

### 2. 运行容器

**基础运行（无认证）：**
```bash
docker run --rm -p 9000:9000 whisper-api
```

**带 API Key 认证：**
```bash
docker run --rm -p 9000:9000 \
  -e WHISPER_API_KEY="your-secret-key" \
  whisper-api
```

**自定义默认模型：**
```bash
docker run --rm -p 9000:9000 \
  -e WHISPER_DEFAULT_MODEL="base" \
  whisper-api
```

**GPU 加速（需要 nvidia-docker）：**
```bash
docker run --rm -p 9000:9000 \
  --gpus all \
  whisper-api
```

**完整配置示例：**
```bash
docker run --rm -d \
  --name whisper-api \
  -p 9000:9000 \
  -e WHISPER_API_KEY="my-secret-key" \
  -e WHISPER_DEFAULT_MODEL="turbo" \
  --gpus all \
  whisper-api
```

## API 使用

### 健康检查

```bash
curl http://localhost:9000/health
```

### 查看服务信息

```bash
curl http://localhost:9000/
```

返回示例：
```json
{
  "message": "Whisper API is running",
  "default_model": "turbo",
  "device": "cuda",
  "loaded_models": ["turbo"],
  "available_models": ["tiny", "base", "small", "medium", "large", "turbo"]
}
```

### 转录音频文件

**使用默认模型：**
```bash
curl -X POST http://localhost:9000/transcribe \
  -F "file=@audio.wav"
```

**指定模型：**
```bash
curl -X POST http://localhost:9000/transcribe \
  -F "file=@audio.wav" \
  -F "model_name=base"
```

**带 API Key 认证：**
```bash
curl -X POST http://localhost:9000/transcribe \
  -H "Authorization: Bearer your-secret-key" \
  -F "file=@audio.wav"
```

### 实时流式转录

```bash
curl -X POST http://localhost:9000/transcribe_stream \
  -F "file=@audio.wav" \
  -F "model_name=turbo"
```

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `WHISPER_API_KEY` | `""` | API 认证密钥，为空则不启用认证 |
| `WHISPER_DEFAULT_MODEL` | `"turbo"` | 默认使用的模型 |
| `PYTHONUNBUFFERED` | `1` | Python 输出不缓冲 |

## 模型说明

| 模型 | 大小 | 速度 | 准确度 | 推荐场景 |
|------|------|------|--------|----------|
| tiny | ~39MB | 最快 | 较低 | 快速测试 |
| base | ~74MB | 快 | 一般 | 开发环境 |
| small | ~244MB | 中等 | 良好 | 平衡性能 |
| medium | ~769MB | 慢 | 很好 | 高质量转录 |
| large | ~1550MB | 很慢 | 最好 | 最高质量 |
| turbo | ~809MB | 快 | 很好 | **推荐生产环境** |

## 与 Mino 后端集成

在 `backend/.env` 中配置：

```env
STT_PROVIDER=whisper
STT_WHISPER_API_URL=http://localhost:9000
STT_WHISPER_API_KEY=your-secret-key
STT_WHISPER_MODEL=turbo
```

## 性能优化

1. **使用 GPU**：显著提升转录速度
2. **选择合适模型**：turbo 模型在速度和准确度间取得最佳平衡
3. **模型缓存**：首次加载后模型会缓存在内存中
4. **批量处理**：实时流式转录会自动批量处理音频块

## 故障排查

**问题：模型加载失败**
- 检查网络连接（首次运行需下载模型）
- 检查磁盘空间
- 尝试使用更小的模型（如 base）

**问题：GPU 不可用**
- 确认已安装 nvidia-docker
- 检查 CUDA 驱动版本
- 使用 `--gpus all` 参数

**问题：认证失败**
- 确认 API Key 正确
- 检查 Authorization header 格式：`Bearer <key>`

## 开发调试

查看日志：
```bash
docker logs -f whisper-api
```

进入容器：
```bash
docker exec -it whisper-api bash
```

## 许可证

本项目使用 MIT 许可证。
