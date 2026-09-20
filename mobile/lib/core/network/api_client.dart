import 'dart:convert';

import 'package:file_picker/file_picker.dart';
import 'package:http/http.dart' as http;

import '../models/annotation.dart';
import '../models/book.dart';
import '../models/catalog_item.dart';
import '../models/collection.dart';
import '../models/paged_result.dart';
import '../models/reading_progress.dart';
import '../models/user.dart';

class ApiClient {
  ApiClient(this.serverUrl, {http.Client? client})
    : _client = client ?? http.Client();
  final String serverUrl;
  final http.Client _client;
  Uri _uri(String path) => Uri.parse('$serverUrl$path');
  Future<void> checkHealth() async {
    final response = await _client
        .get(_uri('/health'))
        .timeout(const Duration(seconds: 8));
    if (response.statusCode != 200) {
      throw ApiException('The server returned status ${response.statusCode}.');
    }
  }

  Future<Session> login(String username, String passphrase) async {
    final response = await _client
        .post(
          _uri('/api/v1/auth/login'),
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode({'username': username, 'passphrase': passphrase}),
        )
        .timeout(const Duration(seconds: 12));
    final body = _decodeMap(response);
    if (response.statusCode != 200) {
      throw ApiException(body['error'] as String? ?? 'Unable to sign in.');
    }
    return Session(
      accessToken: body['token'] as String,
      refreshToken: body['refreshToken'] as String,
      user: ArchiveUser.fromJson(body['user'] as Map<String, dynamic>),
    );
  }

  Future<Session> refresh(String refreshToken) async {
    final response = await _client
        .post(
          _uri('/api/v1/auth/refresh'),
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode({'refreshToken': refreshToken}),
        )
        .timeout(const Duration(seconds: 12));
    final body = _decodeMap(response);
    if (response.statusCode != 200) {
      throw ApiException(
        body['error'] as String? ?? 'Unable to refresh session.',
      );
    }
    return Session(
      accessToken: body['token'] as String,
      refreshToken: body['refreshToken'] as String,
      user: ArchiveUser.fromJson(body['user'] as Map<String, dynamic>),
    );
  }

  Future<ArchiveUser> fetchMe(String token) async => ArchiveUser.fromJson(
    _requireSuccess(await _get(token, '/api/v1/me'), 'Unable to load profile.'),
  );

  Future<PagedResult<ArchiveUser>> fetchUsersPage(
    String token, {
    int page = 1,
    int pageSize = 20,
  }) => _fetchPaged(
    token,
    '/api/v1/users',
    ArchiveUser.fromJson,
    page: page,
    pageSize: pageSize,
    fallback: 'Unable to load users.',
  );
  Future<List<ArchiveUser>> fetchUsers(String token) async =>
      (await fetchUsersPage(token)).items;

  Future<CreatedUser> createUser(
    String token,
    String username,
    String name,
  ) async {
    final body = _requireSuccess(
      await _send(
        'POST',
        token,
        '/api/v1/users',
        body: {'username': username, 'name': name},
      ),
      'Unable to create user.',
    );
    return CreatedUser(
      user: ArchiveUser.fromJson(body['user'] as Map<String, dynamic>),
      passphrase: body['passphrase'] as String,
    );
  }

  Future<ArchiveUser> updateUser(
    String token,
    int userId,
    Map<String, dynamic> changes,
  ) async => ArchiveUser.fromJson(
    _requireSuccess(
      await _send('PATCH', token, '/api/v1/users/$userId', body: changes),
      'Unable to update user.',
    ),
  );

  Future<CreatedUser> resetUserPassphrase(String token, int userId) async {
    final body = _requireSuccess(
      await _send('POST', token, '/api/v1/users/$userId/reset-passphrase'),
      'Unable to reset passphrase.',
    );
    return CreatedUser(
      user: ArchiveUser.fromJson(body['user'] as Map<String, dynamic>),
      passphrase: body['passphrase'] as String,
    );
  }

  Future<void> deleteUser(String token, int userId) async => _expectNoContent(
    await _send('DELETE', token, '/api/v1/users/$userId'),
    'Unable to delete user.',
  );

  Future<void> deleteBook(String token, int bookId) async => _expectNoContent(
    await _send('DELETE', token, '/api/v1/books/$bookId'),
    'Unable to delete book.',
  );

  Future<Book> updateBook(
    String token,
    int bookId,
    Map<String, dynamic> changes,
  ) async => Book.fromJson(
    _requireSuccess(
      await _send('PATCH', token, '/api/v1/books/$bookId', body: changes),
      'Unable to update book.',
    ),
  );

  Future<Book> uploadBook(
    String token, {
    required String title,
    required String authors,
    required PlatformFile pdf,
    PlatformFile? cover,
    String? description,
    String? publisher,
    int? publishedYear,
  }) async {
    final request = http.MultipartRequest('POST', _uri('/api/v1/books'))
      ..headers['Authorization'] = 'Bearer $token'
      ..fields['title'] = title
      ..fields['authors'] = authors;
    if (description?.trim().isNotEmpty ?? false) {
      request.fields['description'] = description!.trim();
    }
    if (publisher?.trim().isNotEmpty ?? false) {
      request.fields['publisher'] = publisher!.trim();
    }
    if (publishedYear != null) {
      request.fields['publishedYear'] = '$publishedYear';
    }
    request.files.add(await _multipartFile('file', pdf));
    if (cover != null) request.files.add(await _multipartFile('cover', cover));
    final response = await http.Response.fromStream(await _client.send(request))
        .timeout(const Duration(seconds: 60));
    return Book.fromJson(_requireSuccess(response, 'Unable to upload book.'));
  }

  Future<PagedResult<CatalogItem>> fetchCatalogItemsPage(
    String token, {
    required bool tags,
    int page = 1,
    int pageSize = 20,
  }) => _fetchPaged(
    token,
    tags ? '/api/v1/tags' : '/api/v1/categories',
    CatalogItem.fromJson,
    page: page,
    pageSize: pageSize,
    fallback: 'Unable to load catalog metadata.',
  );
  Future<List<CatalogItem>> fetchCatalogItems(
    String token, {
    required bool tags,
  }) async => (await fetchCatalogItemsPage(token, tags: tags)).items;

  Future<CatalogItem> createCatalogItem(
    String token, {
    required bool tags,
    required String name,
  }) async => CatalogItem.fromJson(
    _requireSuccess(
      await _send(
        'POST',
        token,
        tags ? '/api/v1/tags' : '/api/v1/categories',
        body: {'name': name},
      ),
      'Unable to create catalog metadata.',
    ),
  );

  Future<void> deleteCatalogItem(
    String token, {
    required bool tags,
    required int itemId,
  }) async => _expectNoContent(
    await _send(
      'DELETE',
      token,
      '${tags ? '/api/v1/tags' : '/api/v1/categories'}/$itemId',
    ),
    'Unable to delete catalog metadata.',
  );

  Future<PagedResult<Book>> fetchBooksPage(
    String token, {
    String? search,
    int page = 1,
    int pageSize = 20,
  }) => _fetchPaged(
    token,
    '/api/v1/books',
    Book.fromJson,
    page: page,
    pageSize: pageSize,
    extraQuery: search == null || search.trim().isEmpty
        ? null
        : {'search': search.trim()},
    fallback: 'Unable to load books.',
  );
  Future<List<Book>> fetchBooks(String token, {String? search}) async =>
      (await fetchBooksPage(token, search: search)).items;

  Future<Book> fetchBook(String token, int bookId) async {
    final response = await _get(token, '/api/v1/books/$bookId');
    return Book.fromJson(_requireSuccess(response, 'Unable to load the book.'));
  }

  Future<ReadingProgress> fetchProgress(String token, int bookId) async {
    final response = await _get(
      token,
      '/api/v1/books/$bookId/reading-progress',
    );
    return ReadingProgress.fromJson(
      _requireSuccess(response, 'Unable to load reading progress.'),
    );
  }

  Future<ReadingProgress> saveProgress(
    String token,
    int bookId, {
    required String status,
    required int currentPage,
    required double progressPercent,
    required bool isDownloaded,
  }) async {
    final response = await _send(
      'PUT',
      token,
      '/api/v1/books/$bookId/reading-progress',
      body: {
        'status': status,
        'currentPage': currentPage,
        'progressPercent': progressPercent,
        'isDownloaded': isDownloaded,
      },
    );
    return ReadingProgress.fromJson(
      _requireSuccess(response, 'Unable to save reading progress.'),
    );
  }

  Future<PagedResult<ReadingProgress>> fetchReadingProgressPage(
    String token, {
    String? status,
    int page = 1,
    int pageSize = 20,
  }) => _fetchPaged(
    token,
    '/api/v1/reading-progress',
    ReadingProgress.fromJson,
    page: page,
    pageSize: pageSize,
    extraQuery: status == null ? null : {'status': status},
    fallback: 'Unable to load reading progress.',
  );
  Future<List<ReadingProgress>> fetchReadingProgress(
    String token, {
    String? status,
  }) async => (await fetchReadingProgressPage(token, status: status)).items;

  Future<PagedResult<Book>> fetchFavoritesPage(
    String token, {
    int page = 1,
    int pageSize = 20,
  }) => _fetchPaged(
    token,
    '/api/v1/favorites',
    (item) => Book.fromJson(item['book'] as Map<String, dynamic>),
    page: page,
    pageSize: pageSize,
    fallback: 'Unable to load favorites.',
  );
  Future<List<Book>> fetchFavorites(String token) async =>
      (await fetchFavoritesPage(token)).items;

  Future<void> setFavorite(String token, int bookId, bool isFavorite) async {
    final response = await _send(
      isFavorite ? 'PUT' : 'DELETE',
      token,
      '/api/v1/books/$bookId/favorite',
    );
    if (response.statusCode != 200 && response.statusCode != 204) {
      _requireSuccess(response, 'Unable to update favorite.');
    }
  }

  Future<PagedResult<LibraryCollection>> fetchCollectionsPage(
    String token, {
    int page = 1,
    int pageSize = 20,
  }) => _fetchPaged(
    token,
    '/api/v1/collections',
    LibraryCollection.fromJson,
    page: page,
    pageSize: pageSize,
    fallback: 'Unable to load collections.',
  );
  Future<List<LibraryCollection>> fetchCollections(String token) async =>
      (await fetchCollectionsPage(token)).items;

  Future<LibraryCollection> createCollection(String token, String name) async {
    final response = await _send(
      'POST',
      token,
      '/api/v1/collections',
      body: {'name': name},
    );
    return LibraryCollection.fromJson(
      _requireSuccess(response, 'Unable to create collection.'),
    );
  }

  Future<CollectionDetails> fetchCollection(
    String token,
    int collectionId, {
    int page = 1,
    int pageSize = 20,
  }) async {
    final response = await _get(
      token,
      '/api/v1/collections/$collectionId?offset=${(page - 1) * pageSize}&limit=$pageSize',
    );
    return CollectionDetails.fromJson(
      _requireSuccess(response, 'Unable to load collection.'),
    );
  }

  Future<void> addBookToCollection(
    String token,
    int collectionId,
    int bookId,
  ) async {
    final response = await _send(
      'PUT',
      token,
      '/api/v1/collections/$collectionId/books/$bookId',
    );
    _requireSuccess(response, 'Unable to add book to collection.');
  }

  Future<void> removeBookFromCollection(
    String token,
    int collectionId,
    int bookId,
  ) async {
    final response = await _send(
      'DELETE',
      token,
      '/api/v1/collections/$collectionId/books/$bookId',
    );
    _requireSuccess(response, 'Unable to remove book from collection.');
  }

  Future<PagedResult<Highlight>> fetchHighlightsPage(
    String token,
    int bookId, {
    int page = 1,
    int pageSize = 20,
  }) => _fetchPaged(
    token,
    '/api/v1/books/$bookId/annotations',
    Highlight.fromJson,
    page: page,
    pageSize: pageSize,
    extraQuery: {'type': 'highlight'},
    fallback: 'Unable to load annotations.',
  );
  Future<List<Highlight>> fetchHighlights(String token, int bookId) async =>
      (await fetchHighlightsPage(token, bookId)).items;

  Future<Highlight> createHighlight(
    String token,
    int bookId, {
    required String selectedText,
    required String color,
    required int page,
    int? startOffset,
    int? endOffset,
  }) async => Highlight.fromJson(
    _requireSuccess(
      await _send(
        'POST',
        token,
        '/api/v1/books/$bookId/highlights',
        body: {
          'selectedText': selectedText,
          'color': color,
          'page': page,
          'startOffset': ?startOffset,
          'endOffset': ?endOffset,
        },
      ),
      'Unable to save highlight.',
    ),
  );

  Future<Note> createNote(
    String token,
    int highlightId,
    String content,
  ) async => Note.fromJson(
    _requireSuccess(
      await _send(
        'POST',
        token,
        '/api/v1/highlights/$highlightId/notes',
        body: {'content': content},
      ),
      'Unable to save note.',
    ),
  );

  Future<Note> updateNote(String token, int noteId, String content) async =>
      Note.fromJson(
        _requireSuccess(
          await _send(
            'PATCH',
            token,
            '/api/v1/notes/$noteId',
            body: {'content': content},
          ),
          'Unable to update note.',
        ),
      );

  Future<void> deleteNote(String token, int noteId) async => _expectNoContent(
    await _send('DELETE', token, '/api/v1/notes/$noteId'),
    'Unable to remove note.',
  );

  Future<Highlight> updateHighlight(
    String token,
    Highlight highlight, {
    required String color,
  }) async => Highlight.fromJson(
    _requireSuccess(
      await _send(
        'PATCH',
        token,
        '/api/v1/highlights/${highlight.id}',
        body: {
          'selectedText': highlight.selectedText,
          'color': color,
          'page': highlight.page,
          'startOffset': ?highlight.startOffset,
          'endOffset': ?highlight.endOffset,
          'chapterRef': highlight.chapterRef ?? '',
        },
      ),
      'Unable to update highlight.',
    ),
  );

  Future<void> deleteHighlight(String token, int highlightId) async =>
      _expectNoContent(
        await _send('DELETE', token, '/api/v1/highlights/$highlightId'),
        'Unable to remove highlight.',
      );

  Future<void> logout(String token) async {
    await _client
        .post(
          _uri('/api/v1/auth/logout'),
          headers: {'Authorization': 'Bearer $token'},
        )
        .timeout(const Duration(seconds: 8));
  }

  void _expectNoContent(http.Response response, String fallback) {
    if (response.statusCode != 204) _requireSuccess(response, fallback);
  }

  Future<http.Response> _get(String token, String path) => _client
      .get(_uri(path), headers: {'Authorization': 'Bearer $token'})
      .timeout(const Duration(seconds: 12));

  Future<PagedResult<T>> _fetchPaged<T>(
    String token,
    String path,
    T Function(Map<String, dynamic>) fromJson, {
    required int page,
    required int pageSize,
    required String fallback,
    Map<String, String>? extraQuery,
  }) async {
    final query = {
      'offset': '${(page - 1) * pageSize}',
      'limit': '$pageSize',
      ...?extraQuery,
    };
    final separator = path.contains('?') ? '&' : '?';
    final response = await _get(
      token,
      '$path$separator${Uri(queryParameters: query).query}',
    );
    final body = _requireSuccess(response, fallback);
    final items = ((body['items'] as List?) ?? [])
        .cast<Map<String, dynamic>>()
        .map(fromJson)
        .toList();
    return PagedResult(
      items: items,
      page:
          ((body['offset'] as int? ?? ((page - 1) * pageSize)) ~/
              (body['limit'] as int? ?? pageSize)) +
          1,
      pageSize: body['limit'] as int? ?? pageSize,
      total: body['total'] as int? ?? items.length,
    );
  }

  Future<http.MultipartFile> _multipartFile(String field, PlatformFile file) {
    if (file.path != null) {
      return http.MultipartFile.fromPath(
        field,
        file.path!,
        filename: file.name,
      );
    }
    if (file.bytes != null) {
      return Future.value(
        http.MultipartFile.fromBytes(field, file.bytes!, filename: file.name),
      );
    }
    throw ApiException('The selected file cannot be read.');
  }

  Future<http.Response> _send(
    String method,
    String token,
    String path, {
    Map<String, dynamic>? body,
  }) async {
    final request = http.Request(method, _uri(path))
      ..headers.addAll({
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      });
    if (body != null) request.body = jsonEncode(body);
    return http.Response.fromStream(await _client.send(request))
        .timeout(const Duration(seconds: 12));
  }

  Map<String, dynamic> _requireSuccess(
    http.Response response,
    String fallback,
  ) {
    final body = _decodeMap(response);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw ApiException(body['error'] as String? ?? fallback);
    }
    return body;
  }

  Map<String, dynamic> _decodeMap(http.Response response) =>
      response.body.isEmpty
      ? {}
      : jsonDecode(response.body) as Map<String, dynamic>;
}

class ApiException implements Exception {
  ApiException(this.message);
  final String message;
  @override
  String toString() => message;
}

class Session {
  Session({
    required this.accessToken,
    required this.refreshToken,
    required this.user,
  });
  final String accessToken;
  final String refreshToken;
  final ArchiveUser user;
}
