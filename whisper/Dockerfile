FROM python:3.9-slim

# 设置工作目录
WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

# 安装Python依赖
RUN pip install --no-cache-dir \
    openai-whisper \
    fastapi \
    uvicorn[standard] \
    python-multipart

# 复制API文件
COPY whisper_api.py .

# 暴露端口
EXPOSE 9000

# 设置环境变量
ENV PYTHONUNBUFFERED=1
ENV WHISPER_API_KEY="my-secret"
ENV WHISPER_DEFAULT_MODEL="tiny"

# 启动命令
CMD ["python", "whisper_api.py"]
