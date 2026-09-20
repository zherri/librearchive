import 'package:flutter/material.dart';

abstract final class ArchiveTheme {
  static const ink = Color(0xFF171713);
  static const parchment = Color(0xFFE6D8B8);
  static const fadedGold = Color(0xFFC9A86A);
  static const moss = Color(0xFF708067);
  static final dark = ThemeData(
    useMaterial3: true,
    brightness: Brightness.dark,
    scaffoldBackgroundColor: ink,
    colorScheme: const ColorScheme.dark(
      primary: fadedGold,
      secondary: moss,
      surface: Color(0xFF25241D),
      onSurface: parchment,
      error: Color(0xFFE28C7C),
    ),
    textTheme: const TextTheme(
      headlineMedium: TextStyle(
        fontFamily: 'serif',
        fontWeight: FontWeight.w700,
        letterSpacing: .4,
      ),
      titleLarge: TextStyle(fontFamily: 'serif', fontWeight: FontWeight.w700),
      titleMedium: TextStyle(fontFamily: 'serif', fontWeight: FontWeight.w600),
      bodyLarge: TextStyle(height: 1.45),
      bodyMedium: TextStyle(height: 1.4),
    ).apply(bodyColor: parchment, displayColor: parchment),
    appBarTheme: const AppBarTheme(
      backgroundColor: ink,
      foregroundColor: parchment,
      elevation: 0,
    ),
    cardTheme: CardThemeData(
      color: const Color(0xFF25241D),
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: Color(0xFF5C513D)),
      ),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: const Color(0xFF25241D),
      labelStyle: const TextStyle(color: parchment),
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: const BorderSide(color: Color(0xFF6D604B)),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: const BorderSide(color: Color(0xFF6D604B)),
      ),
    ),
  );
}
