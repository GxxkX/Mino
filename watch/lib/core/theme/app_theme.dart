import 'package:flutter/material.dart';

/// Design system tokens aligned with design-system/mino-ai-assistant/MASTER.md
/// Color palette: Dark bg + green positive indicators (OLED optimized)
/// Typography: System fonts (no network dependency)
class AppTheme {
  // ── Color Palette ──
  static const Color primary = Color(0xFF0F172A);
  static const Color secondary = Color(0xFF1E293B);
  static const Color cta = Color(0xFF22C55E);
  static const Color background = Color(0xFF020617);
  static const Color text = Color(0xFFF8FAFC);
  static const Color textSecondary = Color(0xFF94A3B8);
  static const Color error = Color(0xFFEF4444);
  static const Color warning = Color(0xFFF59E0B);
  static const Color success = Color(0xFF22C55E);
  static const Color border = Color(0xFF1E293B);

  // ── Spacing ──
  static const double spaceXs = 4;
  static const double spaceSm = 8;
  static const double spaceMd = 16;
  static const double spaceLg = 24;
  static const double spaceXl = 32;
  static const double space2xl = 48;

  // ── Border Radius ──
  static const double radiusSm = 8;
  static const double radiusMd = 12;
  static const double radiusLg = 16;

  // ── Shadows ──
  static List<BoxShadow> get shadowSm => [
    BoxShadow(
      offset: const Offset(0, 1),
      blurRadius: 2,
      color: Colors.black.withOpacity(0.05),
    ),
  ];

  static List<BoxShadow> get shadowMd => [
    BoxShadow(
      offset: const Offset(0, 4),
      blurRadius: 6,
      color: Colors.black.withOpacity(0.1),
    ),
  ];

  static List<BoxShadow> get shadowLg => [
    BoxShadow(
      offset: const Offset(0, 10),
      blurRadius: 15,
      color: Colors.black.withOpacity(0.1),
    ),
  ];

  static List<BoxShadow> get shadowXl => [
    BoxShadow(
      offset: const Offset(0, 20),
      blurRadius: 25,
      color: Colors.black.withOpacity(0.15),
    ),
  ];

  // ── Transition Duration ──
  static const Duration transitionFast = Duration(milliseconds: 150);
  static const Duration transitionNormal = Duration(milliseconds: 200);
  static const Duration transitionSlow = Duration(milliseconds: 300);

  // ── Typography helpers ──
  static TextStyle headingStyle({
    double fontSize = 24,
    FontWeight fontWeight = FontWeight.w700,
    Color color = text,
  }) {
    return TextStyle(
      fontSize: fontSize,
      fontWeight: fontWeight,
      color: color,
    );
  }

  static TextStyle bodyStyle({
    double fontSize = 14,
    FontWeight fontWeight = FontWeight.w400,
    Color color = text,
  }) {
    return TextStyle(
      fontSize: fontSize,
      fontWeight: fontWeight,
      color: color,
    );
  }

  static ThemeData get darkTheme {
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      scaffoldBackgroundColor: background,
      colorScheme: const ColorScheme.dark(
        primary: cta,
        secondary: secondary,
        surface: primary,
        error: error,
        onPrimary: text,
        onSecondary: text,
        onSurface: text,
        onError: text,
      ),
      textTheme: const TextTheme(
        displayLarge: TextStyle(fontSize: 36, fontWeight: FontWeight.w700, color: text),
        displayMedium: TextStyle(fontSize: 28, fontWeight: FontWeight.w600, color: text),
        displaySmall: TextStyle(fontSize: 24, fontWeight: FontWeight.w500, color: text),
        headlineLarge: TextStyle(fontSize: 22, fontWeight: FontWeight.w600, color: text),
        headlineMedium: TextStyle(fontSize: 20, fontWeight: FontWeight.w600, color: text),
        headlineSmall: TextStyle(fontSize: 18, fontWeight: FontWeight.w600, color: text),
        titleLarge: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: text),
        titleMedium: TextStyle(fontSize: 14, fontWeight: FontWeight.w500, color: text),
        titleSmall: TextStyle(fontSize: 12, fontWeight: FontWeight.w500, color: text),
        bodyLarge: TextStyle(fontSize: 16, fontWeight: FontWeight.w400, color: text),
        bodyMedium: TextStyle(fontSize: 14, fontWeight: FontWeight.w400, color: text),
        bodySmall: TextStyle(fontSize: 12, fontWeight: FontWeight.w400, color: textSecondary),
        labelLarge: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: text),
        labelMedium: TextStyle(fontSize: 12, fontWeight: FontWeight.w500, color: text),
        labelSmall: TextStyle(fontSize: 10, fontWeight: FontWeight.w500, color: textSecondary),
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: background,
        foregroundColor: text,
        elevation: 0,
        centerTitle: true,
        titleTextStyle: TextStyle(
          fontSize: 22,
          fontWeight: FontWeight.w600,
          color: text,
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: cta,
          foregroundColor: Colors.white,
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(radiusSm),
          ),
          elevation: 0,
          textStyle: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: text,
          side: const BorderSide(color: secondary),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(radiusSm),
          ),
          textStyle: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: textSecondary,
          textStyle: const TextStyle(fontWeight: FontWeight.w500, fontSize: 14),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: primary,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(radiusSm),
          borderSide: const BorderSide(color: secondary),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(radiusSm),
          borderSide: const BorderSide(color: secondary),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(radiusSm),
          borderSide: const BorderSide(color: cta, width: 1.5),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(radiusSm),
          borderSide: const BorderSide(color: error),
        ),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: spaceMd,
          vertical: spaceMd * 0.75,
        ),
        hintStyle: const TextStyle(color: textSecondary, fontSize: 14),
      ),
      cardTheme: CardThemeData(
        color: background,
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(radiusMd),
        ),
      ),
      iconTheme: const IconThemeData(color: text, size: 24),
      snackBarTheme: SnackBarThemeData(
        backgroundColor: secondary,
        contentTextStyle: const TextStyle(color: text, fontSize: 14),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(radiusSm),
        ),
        behavior: SnackBarBehavior.floating,
      ),
      listTileTheme: ListTileThemeData(
        contentPadding: const EdgeInsets.symmetric(
          horizontal: spaceMd,
          vertical: spaceXs,
        ),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(radiusSm),
        ),
      ),
      dividerTheme: const DividerThemeData(
        color: secondary,
        thickness: 0.5,
      ),
    );
  }
}
