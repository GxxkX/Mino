import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../../../core/constants/app_config.dart';
import '../models/audio_config.dart';

/// WebSocket client mirroring the web AudioWebSocket:
/// - Control messages → JSON text frames
/// - Audio data → raw binary frames (Int16 PCM)
/// - Server responses → JSON text frames
class NetworkService {
  WebSocketChannel? _channel;
  final _messageController = StreamController<Map<String, dynamic>>.broadcast();
  final _connectionController = StreamController<ConnectionState>.broadcast();

  ConnectionState _connectionState = ConnectionState.disconnected;

  Stream<Map<String, dynamic>> get messages => _messageController.stream;
  Stream<ConnectionState> get connectionState => _connectionController.stream;
  ConnectionState get currentState => _connectionState;

  Future<bool> connect(String token) async {
    _updateConnectionState(ConnectionState.connecting);

    try {
      final uri = Uri.parse('${AppConfig.wsUrl}?token=$token');
      _channel = WebSocketChannel.connect(uri);

      await _channel!.ready;
      _updateConnectionState(ConnectionState.connected);

      _channel!.stream.listen(
        _handleMessage,
        onError: _handleError,
        onDone: _handleDone,
        cancelOnError: false,
      );

      return true;
    } catch (e) {
      _updateConnectionState(ConnectionState.error);
      return false;
    }
  }

  void _handleMessage(dynamic data) {
    if (data is String) {
      try {
        final message = jsonDecode(data) as Map<String, dynamic>;
        _messageController.add(message);
      } catch (_) {}
    }
  }

  void _handleError(dynamic error) {
    _updateConnectionState(ConnectionState.error);
    _messageController.add({'type': 'error', 'message': error.toString()});
  }

  void _handleDone() {
    _updateConnectionState(ConnectionState.disconnected);
  }

  void _updateConnectionState(ConnectionState state) {
    _connectionState = state;
    _connectionController.add(state);
  }

  /// Send raw PCM binary frame (mirrors web sendAudioBinary)
  void sendAudioBinary(Uint8List data) {
    if (_channel != null && _connectionState == ConnectionState.connected) {
      _channel!.sink.add(data);
    }
  }

  /// Send a control JSON frame with timestamp and optional extra fields.
  Future<void> sendControl(String action, {Map<String, dynamic>? extra}) async {
    if (_channel == null || _connectionState != ConnectionState.connected) return;
    final msg = <String, dynamic>{
      'type': 'control',
      'action': action,
      'timestamp': DateTime.now().millisecondsSinceEpoch,
    };
    if (extra != null) msg.addAll(extra);
    _channel!.sink.add(jsonEncode(msg));
  }

  Future<void> disconnect() async {
    await _channel?.sink.close();
    _channel = null;
    _updateConnectionState(ConnectionState.disconnected);
  }

  void dispose() {
    disconnect();
    _messageController.close();
    _connectionController.close();
  }
}
