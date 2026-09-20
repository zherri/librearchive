import 'package:flutter/material.dart';

import '../../core/app_controller.dart';
import '../../core/models/book.dart';
import '../../core/models/paged_result.dart';
import '../../core/network/api_client.dart';
import '../admin/admin_screen.dart';
import 'book_detail_screen.dart';
import 'collections_screen.dart';
import 'favorites_screen.dart';
import 'reading_screen.dart';
import '../settings/settings_screen.dart';

class LibraryScreen extends StatefulWidget {
  const LibraryScreen({super.key, required this.controller});
  final AppController controller;
  @override
  State<LibraryScreen> createState() => _LibraryScreenState();
}

class _LibraryScreenState extends State<LibraryScreen> {
  late Future<PagedResult<Book>> _books;
  late final TextEditingController _searchController;
  String _search = '';
  @override
  void initState() {
    super.initState();
    _searchController = TextEditingController();
    _books = _load();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<PagedResult<Book>> _load({int page = 1}) => ApiClient(
    widget.controller.serverUrl!,
  ).fetchBooksPage(widget.controller.accessToken!, search: _search, page: page);
  void _reload() {
    setState(() {
      _books = _load();
    });
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(
      title: const Text('LibreArchive'),
      actions: [
        if (widget.controller.currentUser?.isAdministrator ?? false)
          IconButton(
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => AdminScreen(controller: widget.controller),
              ),
            ),
            icon: const Icon(Icons.admin_panel_settings_outlined),
            tooltip: 'Administration',
          ),
        IconButton(
          onPressed: () => Navigator.of(context).push(
            MaterialPageRoute(
              builder: (_) => ReadingScreen(controller: widget.controller),
            ),
          ),
          icon: const Icon(Icons.menu_book_outlined),
          tooltip: 'Reading',
        ),
        IconButton(
          onPressed: () => Navigator.of(context).push(
            MaterialPageRoute(
              builder: (_) => FavoritesScreen(controller: widget.controller),
            ),
          ),
          icon: const Icon(Icons.favorite_border),
          tooltip: 'Favorites',
        ),
        IconButton(
          onPressed: () => Navigator.of(context).push(
            MaterialPageRoute(
              builder: (_) => CollectionsScreen(controller: widget.controller),
            ),
          ),
          icon: const Icon(Icons.collections_bookmark_outlined),
          tooltip: 'Collections',
        ),
        IconButton(
          onPressed: () => Navigator.of(context).push(
            MaterialPageRoute(
              builder: (_) => SettingsScreen(controller: widget.controller),
            ),
          ),
          icon: const Icon(Icons.settings),
          tooltip: 'Settings',
        ),
      ],
    ),
    body: FutureBuilder<PagedResult<Book>>(
      future: _books,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        if (snapshot.hasError) {
          return RefreshIndicator(
            onRefresh: () async {
              _reload();
              await _books;
            },
            child: ListView(
              physics: const AlwaysScrollableScrollPhysics(),
              children: [
                SizedBox(
                  height: 360,
                  child: Center(
                    child: Padding(
                      padding: const EdgeInsets.all(28),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(Icons.cloud_off, size: 42),
                          const SizedBox(height: 12),
                          Text(
                            'The shelves could not be reached.\n${snapshot.error}',
                            textAlign: TextAlign.center,
                          ),
                          const SizedBox(height: 16),
                          OutlinedButton(
                            onPressed: _reload,
                            child: const Text('Try again'),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ],
            ),
          );
        }
        final result =
            snapshot.data ??
            const PagedResult<Book>(items: [], page: 1, pageSize: 20, total: 0);
        final books = result.items;
        return RefreshIndicator(
          onRefresh: () async {
            _reload();
            await _books;
          },
          child: ListView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.all(16),
            children: [
              Text(
                'The Library',
                style: Theme.of(context).textTheme.headlineMedium,
              ),
              const SizedBox(height: 6),
              Text(
                '${result.total} volumes available',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 16),
              TextField(
                controller: _searchController,
                decoration: InputDecoration(
                  labelText: 'Search the library',
                  prefixIcon: const Icon(Icons.search),
                  suffixIcon: _search.isEmpty
                      ? null
                      : IconButton(
                          tooltip: 'Clear search',
                          icon: const Icon(Icons.clear),
                          onPressed: () {
                            _searchController.clear();
                            _search = '';
                            _reload();
                          },
                        ),
                ),
                onSubmitted: (value) {
                  _search = value;
                  _reload();
                },
              ),
              const SizedBox(height: 12),
              if (books.isEmpty)
                SizedBox(
                  height: 260,
                  child: Center(
                    child: Text(
                      _search.trim().isEmpty
                          ? 'No books have been added to this archive yet.'
                          : 'No books match your search.',
                    ),
                  ),
                ),
              ...books.map(
                (book) => Card(
                  child: ListTile(
                    contentPadding: const EdgeInsets.all(16),
                    leading: ClipRRect(
                      borderRadius: BorderRadius.circular(4),
                      child: Image.network(
                        '${widget.controller.serverUrl}/api/v1/books/${book.id}/cover',
                        width: 44,
                        height: 62,
                        fit: BoxFit.cover,
                        headers: {
                          'Authorization':
                              'Bearer ${widget.controller.accessToken}',
                        },
                        errorBuilder: (_, _, _) => Container(
                          width: 44,
                          height: 62,
                          color: const Color(0xFF472E22),
                          child: const Icon(Icons.menu_book_rounded),
                        ),
                      ),
                    ),
                    title: Text(
                      book.title,
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    subtitle: Text(book.authors),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () async {
                      await Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (_) => BookDetailScreen(
                            controller: widget.controller,
                            book: book,
                          ),
                        ),
                      );
                      _reload();
                    },
                  ),
                ),
              ),
              if (result.hasNextPage)
                Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: OutlinedButton(
                    onPressed: () async {
                      final next = await _load(page: result.page + 1);
                      if (!mounted) return;
                      setState(() {
                        _books = Future.value(
                          PagedResult(
                            items: [...books, ...next.items],
                            page: next.page,
                            pageSize: next.pageSize,
                            total: next.total,
                          ),
                        );
                      });
                    },
                    child: const Text('Load more books'),
                  ),
                ),
            ],
          ),
        );
      },
    ),
  );
}
