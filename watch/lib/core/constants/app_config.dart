import 'package:shared_preferences/shared_preferences.dart';

class AppConfig {
  static const String appName = 'Mino';
  static const String appVersion = '1.0.0';

  static const String _defaultApiUrl = 'http://localhost:8000/v1';

  static const int audioSampleRate = 16000;
  static const int audioBitRate = 32000;
  static const int audioChunkDuration = 100;

  static const Duration connectionTimeout = Duration(seconds: 10);

  static const String tokenKey = 'access_token';
  static const String refreshTokenKey = 'refresh_token';
  static const String userIdKey = 'user_id';
  static const String usernameKey = 'username';
  static const String passwordKey = 'saved_password';
  static const String apiUrlKey = 'api_url';

  static String _cachedApiUrl = _defaultApiUrl;
  static String? _cachedToken;

  static String get apiUrl => _cachedApiUrl;
  static String? get token => _cachedToken;

  static String get wsUrl {
    final uri = Uri.tryParse(_cachedApiUrl);
    if (uri == null) return 'ws://localhost:8000/v1/ws/audio';
    final scheme = uri.scheme == 'https' ? 'wss' : 'ws';
    final port = uri.hasPort ? ':${uri.port}' : '';
    return '$scheme://${uri.host}$port/v1/ws/audio';
  }

  static Future<void> loadApiUrl() async {
    final prefs = await SharedPreferences.getInstance();
    _cachedApiUrl = prefs.getString(apiUrlKey) ?? _defaultApiUrl;
  }

  static Future<void> setApiUrl(String url) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(apiUrlKey, url);
    _cachedApiUrl = url;
  }

  static Future<void> loadToken() async {
    final prefs = await SharedPreferences.getInstance();
    _cachedToken = prefs.getString(tokenKey);
  }

  static Future<void> setToken(String? token) async {
    final prefs = await SharedPreferences.getInstance();
    if (token != null) {
      await prefs.setString(tokenKey, token);
    } else {
      await prefs.remove(tokenKey);
    }
    _cachedToken = token;
  }

  static String get defaultApiUrl => _defaultApiUrl;
}
