import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../constants/app_config.dart';

class AuthService {
  static const String _refreshTokenKey = AppConfig.refreshTokenKey;
  static const String _userIdKey = AppConfig.userIdKey;
  static const String _usernameKey = AppConfig.usernameKey;

  Future<Map<String, dynamic>?> signIn(String username, String password) async {
    try {
      final response = await http.post(
        Uri.parse('${AppConfig.apiUrl}/auth/signin'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'username': username, 'password': password}),
      );

      if (response.statusCode == 200) {
        final body = jsonDecode(response.body) as Map<String, dynamic>;
        // Support both flat and wrapped {code, data} response shapes
        final data = (body['data'] as Map<String, dynamic>?) ?? body;
        await _saveTokens(data);
        return data;
      }
      return null;
    } catch (e) {
      return null;
    }
  }

  Future<bool> signOut() async {
    try {
      final token = await getToken();
      if (token != null) {
        await http.post(
          Uri.parse('${AppConfig.apiUrl}/auth/signout'),
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer $token',
          },
        );
      }
      await _clearTokens();
      return true;
    } catch (e) {
      await _clearTokens();
      return true;
    }
  }

  Future<bool> refreshToken() async {
    try {
      final refreshToken = await getRefreshToken();
      if (refreshToken == null) return false;

      final response = await http.post(
        Uri.parse('${AppConfig.apiUrl}/auth/refresh'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'refresh_token': refreshToken}),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        await _saveTokens(data);
        return true;
      }
      return false;
    } catch (e) {
      return false;
    }
  }

  Future<void> _saveTokens(Map<String, dynamic> data) async {
    final prefs = await SharedPreferences.getInstance();
    if (data['access_token'] != null) {
      await AppConfig.setToken(data['access_token'] as String);
    }
    if (data['refresh_token'] != null) {
      await prefs.setString(_refreshTokenKey, data['refresh_token'] as String);
    }
    if (data['user'] != null) {
      await prefs.setString(_userIdKey, data['user']['id'] ?? '');
      await prefs.setString(_usernameKey, data['user']['username'] ?? '');
    }
  }

  Future<void> saveCredentials(String username, String password, String apiUrl) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_usernameKey, username);
    await prefs.setString(AppConfig.passwordKey, password);
    await prefs.setString(AppConfig.apiUrlKey, apiUrl);
  }

  Future<Map<String, String?>> getSavedCredentials() async {
    final prefs = await SharedPreferences.getInstance();
    return {
      'username': prefs.getString(_usernameKey),
      'password': prefs.getString(AppConfig.passwordKey),
      'apiUrl': prefs.getString(AppConfig.apiUrlKey),
    };
  }

  Future<void> _clearTokens() async {
    final prefs = await SharedPreferences.getInstance();
    await AppConfig.setToken(null);
    await prefs.remove(_refreshTokenKey);
    await prefs.remove(_userIdKey);
    await prefs.remove(_usernameKey);
  }

  Future<String?> getToken() async {
    return AppConfig.token;
  }

  Future<String?> getRefreshToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_refreshTokenKey);
  }

  Future<String?> getUserId() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_userIdKey);
  }

  Future<String?> getUsername() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_usernameKey);
  }

  Future<bool> isLoggedIn() async {
    final token = await getToken();
    return token != null && token.isNotEmpty;
  }
}
