import 'package:flutter/foundation.dart';

import 'network/api_client.dart';
import 'models/user.dart';
import 'storage/secure_store.dart';

enum AppStage { setup, login, library }

class AppController extends ChangeNotifier {
  AppController(this._store);
  final SecureStore _store;
  AppStage _stage = AppStage.setup;
  String? _serverUrl;
  String? _accessToken;
  ArchiveUser? _currentUser;
  AppStage get stage => _stage;
  String? get serverUrl => _serverUrl;
  String? get accessToken => _accessToken;
  ArchiveUser? get currentUser => _currentUser;
  Future<void> initialize() async {
    _serverUrl = await _store.readServerUrl();
    _accessToken = await _store.readAccessToken();
    if (_serverUrl != null && _accessToken != null) {
      try {
        _currentUser = await ApiClient(_serverUrl!).fetchMe(_accessToken!);
        _stage = AppStage.library;
      } catch (_) {
        final refreshToken = await _store.readRefreshToken();
        if (refreshToken == null) {
          await _store.clearSession();
          _accessToken = null;
          _stage = AppStage.login;
        } else {
          try {
            final session = await ApiClient(_serverUrl!).refresh(refreshToken);
            await _store.saveSession(
              accessToken: session.accessToken,
              refreshToken: session.refreshToken,
            );
            _accessToken = session.accessToken;
            _currentUser = session.user;
            _stage = AppStage.library;
          } catch (_) {
            await _store.clearSession();
            _accessToken = null;
            _stage = AppStage.login;
          }
        }
      }
    } else {
      _stage = _serverUrl == null ? AppStage.setup : AppStage.login;
    }
    notifyListeners();
  }

  Future<void> configureServer(String rawUrl) async {
    final url = normalizeServerUrl(rawUrl);
    final client = ApiClient(url);
    await client.checkHealth();
    await _store.saveServerUrl(url);
    _serverUrl = url;
    _stage = AppStage.login;
    notifyListeners();
  }

  Future<void> signIn(String username, String passphrase) async {
    final session = await ApiClient(_serverUrl!).login(username, passphrase);
    await _store.saveSession(
      accessToken: session.accessToken,
      refreshToken: session.refreshToken,
    );
    _accessToken = session.accessToken;
    _currentUser = session.user;
    _stage = AppStage.library;
    notifyListeners();
  }

  Future<void> signOut() async {
    if (_serverUrl != null && _accessToken != null) {
      try {
        await ApiClient(_serverUrl!).logout(_accessToken!);
      } catch (_) {}
    }
    await _store.clearSession();
    _accessToken = null;
    _currentUser = null;
    _stage = _serverUrl == null ? AppStage.setup : AppStage.login;
    notifyListeners();
  }

  Future<void> changeServer() async {
    await _store.clearAll();
    _accessToken = null;
    _currentUser = null;
    _serverUrl = null;
    _stage = AppStage.setup;
    notifyListeners();
  }
}

String normalizeServerUrl(String raw) {
  final value = raw.trim().replaceAll(RegExp(r'/+$'), '');
  final uri = Uri.tryParse(value);
  if (uri == null ||
      !uri.hasScheme ||
      !uri.hasAuthority ||
      !(uri.scheme == 'http' || uri.scheme == 'https')) {
    throw ApiException('Enter a valid http or https server URL.');
  }
  if (uri.path.isNotEmpty && uri.path != '/') {
    throw ApiException('Enter the server URL without /api/v1.');
  }
  return value;
}
