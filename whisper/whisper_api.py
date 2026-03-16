import whisper
import io
import os
import torch
import wave
import tempfile
import numpy as np
from fastapi import FastAPI, UploadFile, File, HTTPException, Header, Form
from fastapi.responses import JSONResponse
from typing import Optional, List, Dict
import logging

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="Whisper API", version="2.0.0")

# 从环境变量读取配置
API_KEY = os.getenv("WHISPER_API_KEY", "")
DEFAULT_MODEL = os.getenv("WHISPER_DEFAULT_MODEL", "turbo")
PYANNOTE_AUTH_TOKEN = os.getenv("PYANNOTE_HF_TOKEN", "") or os.getenv("PYANNOTE_AUTH_TOKEN", "")
PYANNOTE_ENABLED = os.getenv("PYANNOTE_ENABLED", "false").lower() in ("true", "1", "yes")

# Set HF_TOKEN env var so huggingface_hub auto-authenticates regardless of API version.
# This avoids the use_auth_token vs token parameter incompatibility across pyannote/hf_hub versions.
if PYANNOTE_AUTH_TOKEN:
    os.environ["HF_TOKEN"] = PYANNOTE_AUTH_TOKEN
    
# 检查GPU可用性
device = "cuda" if torch.cuda.is_available() else "cpu"
logger.info(f"Using device: {device}")

# 缓存已加载的模型
loaded_models = {}

# Pyannote pipeline (lazy-loaded)
_diarization_pipeline = None
_embedding_model = None


def _ensure_model_downloaded(model_name: str):
    """确保模型文件已下载到本地缓存，启动时调用以避免首次转写延迟。"""
    import whisper as _whisper
    download_root = os.path.join(os.path.expanduser("~"), ".cache", "whisper")
    os.makedirs(download_root, exist_ok=True)

    url = _whisper._MODELS.get(model_name)
    if url is None:
        logger.warning(f"Unknown model name '{model_name}', skipping pre-download")
        return

    expected_path = os.path.join(download_root, os.path.basename(url))
    if os.path.isfile(expected_path):
        logger.info(f"Model '{model_name}' already cached at {expected_path}")
    else:
        logger.info(f"Model '{model_name}' not found locally, downloading...")
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


def get_diarization_pipeline():
    """懒加载 Pyannote 说话人分离 pipeline"""
    global _diarization_pipeline
    if _diarization_pipeline is None:
        if not PYANNOTE_AUTH_TOKEN:
            raise HTTPException(status_code=500, detail="PYANNOTE_AUTH_TOKEN not configured")
        try:
            from pyannote.audio import Pipeline
            logger.info("Loading pyannote speaker-diarization-3.1 pipeline...")
            # Auth is handled via HF_TOKEN env var set at startup — no explicit token param needed.
            _diarization_pipeline = Pipeline.from_pretrained(
                "pyannote/speaker-diarization-3.1",
            )
            if device == "cuda":
                _diarization_pipeline = _diarization_pipeline.to(torch.device("cuda"))
            logger.info("Pyannote diarization pipeline loaded successfully")
        except Exception as e:
            logger.error(f"Failed to load pyannote pipeline: {e}")
            raise HTTPException(status_code=500, detail=f"Failed to load pyannote pipeline: {str(e)}")
    return _diarization_pipeline


def get_embedding_model():
    """懒加载 Pyannote 声纹嵌入模型"""
    global _embedding_model
    if _embedding_model is None:
        if not PYANNOTE_AUTH_TOKEN:
            raise HTTPException(status_code=500, detail="PYANNOTE_AUTH_TOKEN not configured")
        try:
            from pyannote.audio import Model, Inference
            logger.info("Loading pyannote embedding model...")
            # Auth is handled via HF_TOKEN env var set at startup.
            model = Model.from_pretrained(
                "pyannote/embedding",
            )
            _embedding_model = Inference(model, window="whole")
            if device == "cuda":
                _embedding_model.to(torch.device("cuda"))
            logger.info("Pyannote embedding model loaded successfully")
        except Exception as e:
            logger.error(f"Failed to load pyannote embedding model: {e}")
            raise HTTPException(status_code=500, detail=f"Failed to load embedding model: {str(e)}")
    return _embedding_model


# ── 启动时预下载并加载默认模型 ──────────────────────────────────
logger.info(f"Pre-downloading default model '{DEFAULT_MODEL}' if needed...")
try:
    _ensure_model_downloaded(DEFAULT_MODEL)
    get_model(DEFAULT_MODEL)
except Exception as e:
    logger.warning(f"Failed to preload default model {DEFAULT_MODEL}: {e}")
    try:
        DEFAULT_MODEL = "base"
        _ensure_model_downloaded(DEFAULT_MODEL)
        get_model(DEFAULT_MODEL)
    except Exception as e2:
        logger.error(f"Failed to load fallback base model: {e2}")
        raise

# ── 启动时预加载 Pyannote 模型（避免首次请求延迟）──────────────
if PYANNOTE_ENABLED and PYANNOTE_AUTH_TOKEN:
    logger.info("PYANNOTE_ENABLED=true, preloading diarization pipeline and embedding model...")
    try:
        get_diarization_pipeline()
        logger.info("Pyannote diarization pipeline preloaded")
    except Exception as e:
        logger.warning(f"Failed to preload pyannote diarization pipeline: {e}")
    try:
        get_embedding_model()
        logger.info("Pyannote embedding model preloaded")
    except Exception as e:
        logger.warning(f"Failed to preload pyannote embedding model: {e}")
else:
    if PYANNOTE_ENABLED:
        logger.warning("PYANNOTE_ENABLED=true but PYANNOTE_AUTH_TOKEN is empty, skipping preload")


def verify_api_key(authorization: Optional[str] = Header(None)):
    """验证 API Key（如果配置了的话）"""
    if API_KEY and API_KEY != "":
        if not authorization:
            raise HTTPException(status_code=401, detail="Missing authorization header")
        token = authorization
        if authorization.startswith("Bearer "):
            token = authorization[7:]
        if token != API_KEY:
            raise HTTPException(status_code=401, detail="Invalid API key")


def save_audio_as_wav(audio_data: bytes, sample_rate: int = 16000, channels: int = 1, sample_width: int = 2) -> str:
    """将音频数据保存为 WAV 文件"""
    temp_file = tempfile.NamedTemporaryFile(suffix=".wav", delete=False)
    temp_path = temp_file.name

    try:
        temp_file.write(audio_data)
        temp_file.close()

        try:
            with wave.open(temp_path, 'rb') as wf:
                wf.getnframes()
            return temp_path
        except:
            pass

        os.unlink(temp_path)

        with wave.open(temp_path, 'wb') as wav_file:
            wav_file.setnchannels(channels)
            wav_file.setsampwidth(sample_width)
            wav_file.setframerate(sample_rate)
            wav_file.writeframes(audio_data)

        return temp_path

    except Exception as e:
        try:
            temp_file.close()
            os.unlink(temp_path)
        except:
            pass
        raise e


def align_diarization_with_transcription(
    diarization_result, whisper_segments: List[Dict]
) -> List[Dict]:
    """
    将 Pyannote 说话人分离结果与 Whisper 转写 segments 对齐。
    对每个 whisper segment，找到时间重叠最多的说话人标签。
    """
    aligned = []
    for seg in whisper_segments:
        seg_start = seg["start"]
        seg_end = seg["end"]
        seg_text = seg["text"].strip()
        if not seg_text:
            continue

        # 计算每个说话人在该 segment 时间范围内的重叠时长
        speaker_overlap = {}
        for turn, _, speaker in diarization_result.itertracks(yield_label=True):
            overlap_start = max(seg_start, turn.start)
            overlap_end = min(seg_end, turn.end)
            overlap = max(0, overlap_end - overlap_start)
            if overlap > 0:
                speaker_overlap[speaker] = speaker_overlap.get(speaker, 0) + overlap

        # 选择重叠最多的说话人
        if speaker_overlap:
            best_speaker = max(speaker_overlap, key=speaker_overlap.get)
        else:
            best_speaker = "SPEAKER_UNKNOWN"

        aligned.append({
            "speaker": best_speaker,
            "text": seg_text,
            "start": round(seg_start, 3),
            "end": round(seg_end, 3),
        })

    return aligned


def extract_speaker_embeddings(
    audio_path: str, diarization_result
) -> Dict[str, List[float]]:
    """
    为每个说话人提取声纹嵌入向量。
    从每个说话人的最长片段中提取嵌入。
    """
    try:
        import torchaudio
        from pyannote.core import Segment
    except ImportError:
        logger.warning("torchaudio not available, skipping embedding extraction")
        return {}

    embedding_model = get_embedding_model()

    # 收集每个说话人的所有片段
    speaker_segments: Dict[str, List] = {}
    for turn, _, speaker in diarization_result.itertracks(yield_label=True):
        if speaker not in speaker_segments:
            speaker_segments[speaker] = []
        speaker_segments[speaker].append((turn.start, turn.end, turn.end - turn.start))

    # 加载音频
    waveform, sample_rate = torchaudio.load(audio_path)

    embeddings = {}
    for speaker, segments in speaker_segments.items():
        # 选择最长的片段（最多取30秒）来提取嵌入
        segments.sort(key=lambda x: x[2], reverse=True)
        best_start, best_end, _ = segments[0]
        duration = min(best_end - best_start, 30.0)
        best_end = best_start + duration

        # 裁剪音频
        start_sample = int(best_start * sample_rate)
        end_sample = int(best_end * sample_rate)
        if end_sample > waveform.shape[1]:
            end_sample = waveform.shape[1]
        if start_sample >= end_sample:
            continue

        segment_waveform = waveform[:, start_sample:end_sample]

        # 保存临时片段
        seg_path = tempfile.NamedTemporaryFile(suffix=".wav", delete=False).name
        try:
            torchaudio.save(seg_path, segment_waveform, sample_rate)
            # 提取嵌入
            emb = embedding_model(seg_path)
            embeddings[speaker] = emb.flatten().tolist()
        except Exception as e:
            logger.warning(f"Failed to extract embedding for {speaker}: {e}")
        finally:
            try:
                os.unlink(seg_path)
            except:
                pass

    return embeddings


@app.get("/")
async def root():
    return {
        "message": "Whisper API is running",
        "default_model": DEFAULT_MODEL,
        "device": device,
        "loaded_models": list(loaded_models.keys()),
        "available_models": ["tiny", "base", "small", "medium", "large", "turbo"],
        "pyannote_enabled": PYANNOTE_ENABLED,
    }


@app.get("/health")
async def health_check():
    return {"status": "healthy", "device": device, "pyannote_enabled": PYANNOTE_ENABLED}


@app.post("/transcribe")
async def transcribe_audio(
    file: UploadFile = File(...),
    model_name: Optional[str] = Form(None),
    language: Optional[str] = Form(None),
    authorization: Optional[str] = Header(None)
):
    """转录音频文件（不含说话人分离）"""
    verify_api_key(authorization)

    if file.content_type and not (file.content_type.startswith('audio/') or file.content_type == 'application/octet-stream'):
        raise HTTPException(status_code=400, detail=f"Invalid content type: {file.content_type}")

    try:
        selected_model = model_name if model_name else DEFAULT_MODEL
        model = get_model(selected_model)

        audio_data = await file.read()
        temp_path = save_audio_as_wav(audio_data)

        try:
            # Use specified language or None for auto-detect
            transcribe_language = language if language else None
            logger.info(f"Transcribing file: {file.filename} with model: {selected_model}, language: {transcribe_language}")
            result = model.transcribe(temp_path, word_timestamps=True, language=transcribe_language)

            return JSONResponse(content={
                "text": result["text"],
                "language": result["language"],
                "segments": result["segments"],
                "model": selected_model
            })
        finally:
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
    language: Optional[str] = Form(None),
    authorization: Optional[str] = Header(None)
):
    """流式转录，适用于实时音频处理"""
    verify_api_key(authorization)

    if file.content_type and not (file.content_type.startswith('audio/') or file.content_type == 'application/octet-stream'):
        raise HTTPException(status_code=400, detail=f"Invalid content type: {file.content_type}")

    try:
        selected_model = model_name if model_name else DEFAULT_MODEL
        model = get_model(selected_model)

        audio_data = await file.read()
        temp_path = save_audio_as_wav(audio_data)

        try:
            transcribe_language = language if language else None
            logger.info(f"Stream transcribing with model: {selected_model}, language: {transcribe_language}")
            result = model.transcribe(
                temp_path, word_timestamps=True, language=transcribe_language,
                condition_on_previous_text=False
            )

            return JSONResponse(content={
                "text": result["text"],
                "language": result["language"],
                "segments": result["segments"],
                "word_timestamps": True,
                "model": selected_model
            })
        finally:
            try:
                os.unlink(temp_path)
            except:
                pass

    except Exception as e:
        logger.error(f"Stream transcription error: {e}")
        raise HTTPException(status_code=500, detail=f"Stream transcription failed: {str(e)}")


@app.post("/diarize")
async def diarize_and_transcribe(
    file: UploadFile = File(...),
    model_name: Optional[str] = Form(None),
    num_speakers: Optional[int] = Form(None),
    language: Optional[str] = Form(None),
    authorization: Optional[str] = Header(None)
):
    """
    说话人分离 + 转写。
    先用 Pyannote 做说话人分离，再用 Whisper 转写，最后按时间对齐合并。

    返回格式：
    {
      "text": "完整文本",
      "language": "zh",
      "segments": [
        {"speaker": "SPEAKER_00", "text": "你好", "start": 0.0, "end": 1.5},
        ...
      ],
      "speakers": {
        "SPEAKER_00": {"embedding": [0.1, 0.2, ...]},
        ...
      },
      "num_speakers": 2
    }
    """
    verify_api_key(authorization)

    if not PYANNOTE_ENABLED:
        raise HTTPException(status_code=400, detail="Pyannote diarization is not enabled")

    if file.content_type and not (file.content_type.startswith('audio/') or file.content_type == 'application/octet-stream'):
        raise HTTPException(status_code=400, detail=f"Invalid content type: {file.content_type}")

    try:
        selected_model = model_name if model_name else DEFAULT_MODEL
        model = get_model(selected_model)

        audio_data = await file.read()
        temp_path = save_audio_as_wav(audio_data)

        try:
            # Step 1: Pyannote 说话人分离
            logger.info("Running speaker diarization...")
            pipeline = get_diarization_pipeline()
            diarize_params = {}
            if num_speakers and num_speakers > 0:
                diarize_params["num_speakers"] = num_speakers
            diarization = pipeline(temp_path, **diarize_params)

            # Step 2: Whisper 转写
            transcribe_language = language if language else None
            logger.info(f"Transcribing with model: {selected_model}, language: {transcribe_language}")
            whisper_result = model.transcribe(
                temp_path, word_timestamps=True, language=transcribe_language
            )

            # Step 3: 对齐说话人标签与转写 segments
            aligned_segments = align_diarization_with_transcription(
                diarization, whisper_result["segments"]
            )

            # Step 4: 提取每个说话人的声纹嵌入
            speaker_embeddings = extract_speaker_embeddings(temp_path, diarization)

            # 构建 speakers 字典
            speakers = {}
            unique_speakers = set(seg["speaker"] for seg in aligned_segments)
            for spk in unique_speakers:
                speakers[spk] = {
                    "embedding": speaker_embeddings.get(spk, []),
                }

            return JSONResponse(content={
                "text": whisper_result["text"],
                "language": whisper_result["language"],
                "segments": aligned_segments,
                "speakers": speakers,
                "num_speakers": len(unique_speakers),
                "model": selected_model,
            })

        finally:
            try:
                os.unlink(temp_path)
            except:
                pass

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Diarization error: {e}")
        raise HTTPException(status_code=500, detail=f"Diarization failed: {str(e)}")


@app.post("/extract_embedding")
async def extract_embedding(
    file: UploadFile = File(...),
    authorization: Optional[str] = Header(None)
):
    """
    从音频中提取说话人声纹嵌入向量。
    假设音频中只有一个说话人。

    返回格式：
    {
      "embedding": [0.1, 0.2, ...],
      "dimension": 512
    }
    """
    verify_api_key(authorization)

    if not PYANNOTE_ENABLED:
        raise HTTPException(status_code=400, detail="Pyannote is not enabled")

    try:
        audio_data = await file.read()
        temp_path = save_audio_as_wav(audio_data)

        try:
            embedding_model = get_embedding_model()
            emb = embedding_model(temp_path)
            embedding_list = emb.flatten().tolist()

            return JSONResponse(content={
                "embedding": embedding_list,
                "dimension": len(embedding_list),
            })
        finally:
            try:
                os.unlink(temp_path)
            except:
                pass

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Embedding extraction error: {e}")
        raise HTTPException(status_code=500, detail=f"Embedding extraction failed: {str(e)}")


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=9000)
