import 'package:flutter/foundation.dart';

enum ConnectionStatus {
  connected,
  connecting,
  disconnected,
  error,
}

class ConnectionProvider extends ChangeNotifier {
  ConnectionStatus _status = ConnectionStatus.disconnected;
  String? _errorMessage;

  ConnectionStatus get status => _status;
  String? get errorMessage => _errorMessage;
  bool get isConnected => _status == ConnectionStatus.connected;

  void setConnected() {
    _status = ConnectionStatus.connected;
    _errorMessage = null;
    notifyListeners();
  }

  void setConnecting() {
    _status = ConnectionStatus.connecting;
    notifyListeners();
  }

  void setDisconnected() {
    _status = ConnectionStatus.disconnected;
    _errorMessage = null;
    notifyListeners();
  }

  void setError(String message) {
    _status = ConnectionStatus.error;
    _errorMessage = message;
    notifyListeners();
  }
}
