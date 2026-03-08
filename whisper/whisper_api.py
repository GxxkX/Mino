import whisper
import io
import os
import torch
import wave
import tempfile
from fastapi import FastAPI, UploadFile, File, HTTPException, Header, Form
from fastapi.responses import JSONResponse
from typing import Optional
import logging

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="Whisper API", version="1.0.0")

# 从环境变量读取配置
API_KEY = os.getenv("WHISPER_API_KEY", "")
DEFAULT_MODEL = os.getenv("WHISPER_DEFAULT_MODEL", "turbo")

# 检查GPU可用性
device = "cuda" if torch.cuda.is_available() else "cpu"
logger.info(f"Using device: {device}")

# 缓存已加载的模型
loaded_models = {}


def _ensure_model_downloaded(model_name: str):
    """确保模型文件已下载到本地缓存，启动时调用以避免首次转写延迟。"""
    import whisper as _whisper
    download_root = os.path.join(os.path.expanduser("~"), ".cache", "whisper")
    os.makedirs(download_root, exist_ok=True)

    # whisper 内部维护了一个 name→url 映射，利用它来检查文件是否存在
    url = _whisper._MODELS.get(model_name)
    if url is None:
        logger.warning(f"Unknown model name '{model_name}', skipping pre-download")
        return

    expected_path = os.path.join(download_root, os.path.basename(url))
    if os.path.isfile(expected_path):
        logger.info(f"Model '{model_name}' already cached at {expected_path}")
    else:
        logger.info(f"Model '{model_name}' not found locally, downloading...")
        # whisper._download 会下载并校验 SHA256
        _whisper._download(url, download_root, in_memory=False)
        logger.info(f"Model '{model_name}' downloaded to {expected_path}")


def get_model(model_name: str):
    """获取或加载指定的模型"""
    if model_name not in loaded_models:
        try:
            logger.info(f"Loading model: {model_name}")
            loaded_models[model_name] = whisper.load_model(model_name, device=device)
            logger.info(f"Model {model_name} loaded successfully")
        except Exception as e:
            logger.error(f"Failed to load model {model_name}: {e}")
            raise HTTPException(status_code=500, detail=f"Failed to load model {model_name}: {str(e)}")
    return loaded_models[model_name]


# ── 启动时预下载并加载默认模型 ──────────────────────────────────
# 先确保模型文件存在（下载），再加载到内存，避免首次转写请求时的长时间等待。
logger.info(f"Pre-downloading default model '{DEFAULT_MODEL}' if needed...")
try:
    _ensure_model_downloaded(DEFAULT_MODEL)
    get_model(DEFAULT_MODEL)
except Exception as e:
    logger.warning(f"Failed to preload default model {DEFAULT_MODEL}: {e}")
    # 尝试加载 base 作为后备
    try:
        DEFAULT_MODEL = "base"
        _ensure_model_downloaded(DEFAULT_MODEL)
        get_model(DEFAULT_MODEL)
    except Exception as e2:
        logger.error(f"Failed to load fallback base model: {e2}")
        raise

def verify_api_key(authorization: Optional[str] = Header(None)):
    """验证 API Key（如果配置了的话）"""
    if API_KEY and API_KEY != "":
        if not authorization:
            raise HTTPException(status_code=401, detail="Missing authorization header")
        
        # 支持 "Bearer <token>" 格式
        token = authorization
        if authorization.startswith("Bearer "):
            token = authorization[7:]
        
        if token != API_KEY:
            raise HTTPException(status_code=401, detail="Invalid API key")

def save_audio_as_wav(audio_data: bytes, sample_rate: int = 16000, channels: int = 1, sample_width: int = 2) -> str:
    """
    将音频数据保存为 WAV 文件
    
    参数：
    - audio_data: 原始音频数据（可能是 PCM 或其他格式）
    - sample_rate: 采样率（默认 16000 Hz）
    - channels: 声道数（默认 1 = 单声道）
    - sample_width: 采样宽度（默认 2 = 16-bit）
    
    返回：临时 WAV 文件路径
    """
    temp_file = tempfile.NamedTemporaryFile(suffix=".wav", delete=False)
    temp_path = temp_file.name
    
    try:
        # 尝试直接写入（如果已经是有效的音频格式）
        temp_file.write(audio_data)
        temp_file.close()
        
        # 验证是否是有效的 WAV 文件
        try:
            with wave.open(temp_path, 'rb') as wf:
                wf.getnframes()  # 尝试读取帧数
            return temp_path  # 已经是有效的 WAV 文件
        except:
            # 不是有效的 WAV 文件，需要添加 WAV 文件头
            pass
        
        # 删除临时文件，重新创建带 WAV 头的文件
        os.unlink(temp_path)
        
        # 将原始 PCM 数据转换为 WAV 格式
        with wave.open(temp_path, 'wb') as wav_file:
            wav_file.setnchannels(channels)
            wav_file.setsampwidth(sample_width)
            wav_file.setframerate(sample_rate)
            wav_file.writeframes(audio_data)
        
        return temp_path
        
    except Exception as e:
        # 清理失败的临时文件
        try:
            temp_file.close()
            os.unlink(temp_path)
        except:
            pass
        raise e

@app.get("/")
async def root():
    return {
        "message": "Whisper API is running",
        "default_model": DEFAULT_MODEL,
        "device": device,
        "loaded_models": list(loaded_models.keys()),
        "available_models": ["tiny", "base", "small", "medium", "large", "turbo"]
    }

@app.get("/health")
async def health_check():
    return {"status": "healthy", "device": device}

@app.post("/transcribe")
async def transcribe_audio(
    file: UploadFile = File(...),
    model_name: Optional[str] = Form(None),
    authorization: Optional[str] = Header(None)
):
    """
    转录音频文件
    支持的格式：wav, mp3, flac, m4a, ogg等
    
    参数：
    - file: 音频文件
    - model_name: 模型名称 (tiny/base/small/medium/large/turbo)，默认使用环境变量配置
    """
    verify_api_key(authorization)
    
    # 允许 audio/* 或 application/octet-stream (用于原始 PCM 数据)
    if file.content_type and not (file.content_type.startswith('audio/') or file.content_type == 'application/octet-stream'):
        raise HTTPException(status_code=400, detail=f"Invalid content type: {file.content_type}. Expected audio/* or application/octet-stream")

    try:
        # 使用指定模型或默认模型
        selected_model = model_name if model_name else DEFAULT_MODEL
        model = get_model(selected_model)
        
        audio_data = await file.read()
        
        # 保存为 WAV 文件（自动处理 PCM 数据）
        temp_path = save_audio_as_wav(audio_data)
        
        try:
            logger.info(f"Transcribing file: {file.filename} with model: {selected_model}")
            result = model.transcribe(
                temp_path,
                word_timestamps=True,
                language=None  # 自动检测语言
            )
            
            return JSONResponse(content={
                "text": result["text"],
                "language": result["language"],
                "segments": result["segments"],
                "model": selected_model
            })
        finally:
            # 清理临时文件
            try:
                os.unlink(temp_path)
            except:
                pass

    except Exception as e:
        logger.error(f"Transcription error: {e}")
        raise HTTPException(status_code=500, detail=f"Transcription failed: {str(e)}")

@app.post("/transcribe_stream")
async def transcribe_stream(
    file: UploadFile = File(...),
    model_name: Optional[str] = Form(None),
    authorization: Optional[str] = Header(None)
):
    """
    流式转录，适用于实时音频处理
    
    参数：
    - file: 音频文件
    - model_name: 模型名称 (tiny/base/small/medium/large/turbo)，默认使用环境变量配置
    """
    verify_api_key(authorization)
    
    # 允许 audio/* 或 application/octet-stream (用于原始 PCM 数据)
    if file.content_type and not (file.content_type.startswith('audio/') or file.content_type == 'application/octet-stream'):
        raise HTTPException(status_code=400, detail=f"Invalid content type: {file.content_type}. Expected audio/* or application/octet-stream")

    try:
        # 使用指定模型或默认模型
        selected_model = model_name if model_name else DEFAULT_MODEL
        model = get_model(selected_model)
        
        audio_data = await file.read()
        
        # 保存为 WAV 文件（自动处理 PCM 数据）
        temp_path = save_audio_as_wav(audio_data)
        
        try:
            logger.info(f"Stream transcribing with model: {selected_model}")
            result = model.transcribe(
                temp_path,
                word_timestamps=True,
                language=None,
                condition_on_previous_text=False  # 流式处理时不依赖前文
            )
            
            return JSONResponse(content={
                "text": result["text"],
                "language": result["language"],
                "segments": result["segments"],
                "word_timestamps": True,
                "model": selected_model
            })
        finally:
            # 清理临时文件
            try:
                os.unlink(temp_path)
            except:
                pass

    except Exception as e:
        logger.error(f"Stream transcription error: {e}")
        raise HTTPException(status_code=500, detail=f"Stream transcription failed: {str(e)}")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=9000)
