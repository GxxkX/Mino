import 'dart:async';
import 'package:flutter/foundation.dart';
import '../../../core/providers/auth_provider.dart';
import '../models/audio_config.dart';
import '../services/audio_recording_service.dart';
import '../services/network_service.dart';

class RecordingProvider extends ChangeNotifier {
  final AudioRecordingService _audioService = AudioRecordingService();
  final NetworkService _networkService = NetworkService();

  RecordingState _state = RecordingState.idle;
  String _finalTranscript = '';
  String _partialTranscript = '';
  DateTime? _recordingStartTime;
  String _error = '';

  StreamSubscription? _messageSubscription;
  StreamSubscription? _connectionSubscription;
  Timer? _completedTimeout;

  RecordingState get state => _state;
  String get transcript => _partialTranscript.isNotEmpty
      ? (_finalTranscript.isNotEmpty
          ? '$_finalTranscript $_partialTranscript'
          : _partialTranscript)
      : _finalTranscript;
  String get error => _error;
  bool get isRecording => _state == RecordingState.recording;
  DateTime? get recordingStartTime => _recordingStartTime;

  RecordingProvider() {
    _messageSubscription = _networkService.messages.listen(_handleMessage);
    _connectionSubscription =
        _networkService.connectionState.listen(_handleConnectionState);
  }

  void _handleMessage(Map<String, dynamic> message) {
    final type = message['type'] as String?;

    switch (type) {
      case 'transcript':
        final text = message['text'] as String? ?? '';
        final isFinal = message['is_final'] as bool? ?? false;
        if (isFinal) {
          // Fold any accumulated partial text + this final delta into _finalTranscript
          final pending = _partialTranscript.isNotEmpty
              ? '$_partialTranscript $text'
              : text;
          if (_finalTranscript.isNotEmpty && pending.isNotEmpty) {
            _finalTranscript += ' ';
          }
          _finalTranscript += pending;
          _partialTranscript = '';
        } else {
          // Backend sends incremental deltas — accumulate them
          if (_partialTranscript.isNotEmpty && text.isNotEmpty) {
            _partialTranscript += text;
          } else {
            _partialTranscript = text;
          }
        }
        notifyListeners();
        break;

      case 'completed':
        // Only handle completed when we're NOT actively recording
        // (i.e. after user pressed stop and we're waiting for server)
        if (_state != RecordingState.recording &&
            _state != RecordingState.connecting) {
          final completedTranscript = message['transcript'] as String?;
          if (completedTranscript != null && completedTranscript.isNotEmpty) {
            // Append to existing transcript instead of overwriting
            if (_finalTranscript.isNotEmpty) {
              _finalTranscript += ' $completedTranscript';
            } else {
              _finalTranscript = completedTranscript;
            }
            _partialTranscript = '';
          }
          _state = RecordingState.completed;
          _cleanupAfterStop();
          notifyListeners();
        }
        break;

      case 'error':
        _error = message['message'] as String? ?? 'Unknown error';
        _state = RecordingState.error;
        _cleanupAfterStop();
        notifyListeners();
        break;
    }
  }

  void _handleConnectionState(ConnectionState state) {
    if (state == ConnectionState.connected &&
        _state == RecordingState.connecting) {
      _state = RecordingState.recording;
      notifyListeners();
    } else if (state == ConnectionState.error) {
      _error = 'Connection error';
      _state = RecordingState.error;
      notifyListeners();
    }
  }

  Future<bool> checkPermission() async {
    return await _audioService.hasPermission();
  }

  Future<bool> startRecording(AuthProvider authProvider,
      {double gain = 1.0, String language = 'zh-CN'}) async {
    if (_state == RecordingState.recording ||
        _state == RecordingState.connecting) return false;

    final token = await authProvider.getToken();
    if (token == null || token.isEmpty) {
      _error = 'Not authenticated';
      notifyListeners();
      return false;
    }

    _state = RecordingState.connecting;
    _finalTranscript = '';
    _partialTranscript = '';
    _error = '';
    notifyListeners();

    final connected = await _networkService.connect(token);
    if (!connected) {
      _state = RecordingState.error;
      _error = 'Failed to connect to server';
      notifyListeners();
      return false;
    }

    await _networkService.sendControl('start', extra: {'language': language});

    _audioService.gain = gain;
    _audioService.onAudioChunk = (chunk) {
      _networkService.sendAudioBinary(chunk);
    };

    final started = await _audioService.startRecording(const AudioConfig());
    if (!started) {
      _state = RecordingState.error;
      _error = 'Failed to start microphone';
      await _networkService.sendControl('stop');
      await _networkService.disconnect();
      notifyListeners();
      return false;
    }

    _recordingStartTime = DateTime.now();
    _state = RecordingState.recording;
    notifyListeners();
    return true;
  }

  /// Mirrors web stop(): immediately resets UI to idle, keeps WS alive
  /// to receive the final 'completed' message in the background.
  Future<bool> stopRecording() async {
    if (_state != RecordingState.recording) return false;

    await _audioService.stopRecording();
    _audioService.onAudioChunk = null;

    await _networkService.sendControl('stop');

    // Immediately return to idle (like web does) so the button resets
    _state = RecordingState.idle;
    _recordingStartTime = null;
    notifyListeners();

    // Keep WS alive to receive 'completed'; auto-cleanup after timeout
    _completedTimeout = Timer(const Duration(seconds: 120), () {
      _completedTimeout = null;
      _networkService.disconnect();
    });

    return true;
  }

  Future<void> cancelRecording() async {
    _completedTimeout?.cancel();
    await _audioService.cancelRecording();
    _audioService.onAudioChunk = null;
    await _networkService.sendControl('stop');
    await _networkService.disconnect();
    _state = RecordingState.idle;
    _finalTranscript = '';
    _partialTranscript = '';
    _error = '';
    _recordingStartTime = null;
    notifyListeners();
  }

  void _cleanupAfterStop() {
    _completedTimeout?.cancel();
    _completedTimeout = null;
    _networkService.disconnect();
    _recordingStartTime = null;
  }

  void reset() {
    _state = RecordingState.idle;
    _finalTranscript = '';
    _partialTranscript = '';
    _error = '';
    _recordingStartTime = null;
    notifyListeners();
  }

  @override
  void dispose() {
    _completedTimeout?.cancel();
    _messageSubscription?.cancel();
    _connectionSubscription?.cancel();
    _audioService.dispose();
    _networkService.dispose();
    super.dispose();
  }
}
