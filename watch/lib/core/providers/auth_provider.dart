import 'package:flutter/foundation.dart';
import '../constants/app_config.dart';
import '../services/auth_service.dart';

class AuthProvider extends ChangeNotifier {
  final AuthService _authService = AuthService();
  
  bool _isLoading = false;
  bool _isLoggedIn = false;
  String? _username;
  String? _userId;
  String? _error;

  bool get isLoading => _isLoading;
  bool get isLoggedIn => _isLoggedIn;
  String? get username => _username;
  String? get userId => _userId;
  String? get error => _error;

  Future<void> checkAuthStatus() async {
    _isLoading = true;
    notifyListeners();

    _isLoggedIn = await _authService.isLoggedIn();
    if (_isLoggedIn) {
      _username = await _authService.getUsername();
      _userId = await _authService.getUserId();
    }

    _isLoading = false;
    notifyListeners();
  }

  Future<bool> signIn(String username, String password) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    final result = await _authService.signIn(username, password);

    if (result != null) {
      _isLoggedIn = true;
      _username = username;
      _userId = result['user']?['id'];
      _isLoading = false;
      notifyListeners();
      return true;
    } else {
      _error = 'Invalid username or password';
      _isLoading = false;
      notifyListeners();
      return false;
    }
  }

  Future<void> signOut() async {
    _isLoading = true;
    notifyListeners();

    await _authService.signOut();
    _isLoggedIn = false;
    _username = null;
    _userId = null;
    _error = null;

    _isLoading = false;
    notifyListeners();
  }

  Future<String?> getToken() async => AppConfig.token;
}
