import 'dart:async';
import 'dart:typed_data';
import 'package:record/record.dart';
import '../models/audio_config.dart';

/// Mirrors the web AudioRecorderService:
/// mic → PCM 16-bit 16kHz mono → gain → raw Int16 chunks → WebSocket binary
class AudioRecordingService {
  final AudioRecorder _recorder = AudioRecorder();
  StreamSubscription<Uint8List>? _streamSub;
  bool _isRecording = false;
  DateTime? _recordingStartTime;

  void Function(Uint8List chunk)? onAudioChunk;
  double gain = 1.0;

  bool get isRecording => _isRecording;
  DateTime? get recordingStartTime => _recordingStartTime;

  Future<bool> hasPermission() async {
    return await _recorder.hasPermission();
  }

  /// Apply gain to PCM Int16 LE samples in-place.
  Uint8List _applyGain(Uint8List raw) {
    if (gain == 1.0) return raw;
    final samples = raw.buffer.asInt16List(raw.offsetInBytes, raw.lengthInBytes ~/ 2);
    final out = Int16List(samples.length);
    for (int i = 0; i < samples.length; i++) {
      final amplified = (samples[i] * gain).round();
      out[i] = amplified.clamp(-32768, 32767);
    }
    return Uint8List.view(out.buffer);
  }

  Future<bool> startRecording(AudioConfig config) async {
    if (_isRecording) return false;

    final hasPermission = await _recorder.hasPermission();
    if (!hasPermission) return false;

    try {
      final stream = await _recorder.startStream(
        RecordConfig(
          encoder: AudioEncoder.pcm16bits,
          sampleRate: config.sampleRate,
          numChannels: config.channels,
        ),
      );

      _isRecording = true;
      _recordingStartTime = DateTime.now();

      _streamSub = stream.listen(
        (chunk) {
          if (_isRecording) {
            onAudioChunk?.call(_applyGain(chunk));
          }
        },
        onError: (_) {},
        cancelOnError: false,
      );

      return true;
    } catch (e) {
      return false;
    }
  }

  Future<void> stopRecording() async {
    if (!_isRecording) return;
    _isRecording = false;
    await _streamSub?.cancel();
    _streamSub = null;
    await _recorder.stop();
    _recordingStartTime = null;
  }

  Future<void> cancelRecording() async {
    if (!_isRecording) return;
    _isRecording = false;
    await _streamSub?.cancel();
    _streamSub = null;
    await _recorder.stop();
    _recordingStartTime = null;
  }

  Stream<Amplitude> get amplitudeStream =>
      _recorder.onAmplitudeChanged(const Duration(milliseconds: 100));

  Future<void> dispose() async {
    await cancelRecording();
    _recorder.dispose();
  }
}
