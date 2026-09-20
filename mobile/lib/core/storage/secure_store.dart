import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SecureStore {
  final FlutterSecureStorage _storage = const FlutterSecureStorage();
  static const _serverUrlKey = 'server_url';
  static const _accessTokenKey = 'access_token';
  static const _refreshTokenKey = 'refresh_token';
  Future<String?> readServerUrl() => _storage.read(key: _serverUrlKey);
  Future<String?> readAccessToken() => _storage.read(key: _accessTokenKey);
  Future<String?> readRefreshToken() => _storage.read(key: _refreshTokenKey);
  Future<void> saveServerUrl(String value) =>
      _storage.write(key: _serverUrlKey, value: value);
  Future<void> saveSession({
    required String accessToken,
    required String refreshToken,
  }) async {
    await _storage.write(key: _accessTokenKey, value: accessToken);
    await _storage.write(key: _refreshTokenKey, value: refreshToken);
  }

  Future<void> clearSession() => _storage
      .delete(key: _accessTokenKey)
      .then((_) => _storage.delete(key: _refreshTokenKey));
  Future<void> clearAll() => _storage.deleteAll();
}
