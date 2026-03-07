class AudioConfig {
  final int sampleRate;
  final int bitRate;
  final int channels;
  final String encoder;

  const AudioConfig({
    this.sampleRate = 16000,
    this.bitRate = 32000,
    this.channels = 1,
    this.encoder = 'pcm16bits',
  });

  Map<String, dynamic> toJson() => {
    'sampleRate': sampleRate,
    'bitRate': bitRate,
    'channels': channels,
    'encoder': encoder,
  };
}

enum RecordingState {
  idle,
  connecting,
  recording,
  processing,
  completed,
  error,
}

enum ConnectionState {
  disconnected,
  connecting,
  connected,
  reconnecting,
  error,
}
