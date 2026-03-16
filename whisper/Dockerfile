FROM python:3.10-slim

# 设置工作目录
WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    ffmpeg \
    libsndfile1 \
    && rm -rf /var/lib/apt/lists/*

# 安装Python依赖
# pyannote.audio>=3.3.2 uses the modern huggingface_hub `token=` API (no more use_auth_token)
# huggingface_hub>=0.24 removed the deprecated use_auth_token parameter
RUN pip install --no-cache-dir \
    openai-whisper \
    fastapi \
    uvicorn[standard] \
    python-multipart \
    "pyannote.audio>=3.3.2" \
    "huggingface_hub>=0.25" \
    torchaudio \
    numpy \
    omegaconf

# 复制API文件
COPY whisper_api.py .

# 暴露端口
EXPOSE 9000

# 设置环境变量
ENV PYTHONUNBUFFERED=1
ENV WHISPER_API_KEY="my-secret"
ENV WHISPER_DEFAULT_MODEL="tiny"
ENV PYANNOTE_ENABLED="false"
ENV PYANNOTE_HF_TOKEN=""

# 启动命令
CMD ["python", "whisper_api.py"]
