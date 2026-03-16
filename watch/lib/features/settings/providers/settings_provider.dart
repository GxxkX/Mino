import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

class SettingsProvider extends ChangeNotifier {
  static const String _languageKey = 'settings_language';
  static const String _gainKey = 'settings_recording_gain';
  static const String _whisperLanguageKey = 'settings_whisper_language';

  String _language = 'en-US';
  double _recordingGain = 1.0;
  String _whisperLanguage = ''; // "" = auto-detect

  String get language => _language;
  double get recordingGain => _recordingGain;
  String get whisperLanguage => _whisperLanguage;

  Future<void> load() async {
    final prefs = await SharedPreferences.getInstance();
    _language = prefs.getString(_languageKey) ?? 'en-US';
    _recordingGain = prefs.getDouble(_gainKey) ?? 1.0;
    _whisperLanguage = prefs.getString(_whisperLanguageKey) ?? '';
    notifyListeners();
  }

  Future<void> setLanguage(String lang) async {
    _language = lang;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_languageKey, lang);
    notifyListeners();
  }

  Future<void> setRecordingGain(double gain) async {
    _recordingGain = gain.clamp(0.0, 3.0);
    final prefs = await SharedPreferences.getInstance();
    await prefs.setDouble(_gainKey, _recordingGain);
    notifyListeners();
  }

  Future<void> setWhisperLanguage(String lang) async {
    _whisperLanguage = lang;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_whisperLanguageKey, lang);
    notifyListeners();
  }

  bool get isChinese => _language.startsWith('zh');

  /// Simple i18n helper — returns Chinese text when language is zh-CN.
  String t(String en, String zh) => isChinese ? zh : en;
}
