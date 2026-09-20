import 'package:flutter/material.dart';

import '../../core/app_controller.dart';
import '../../core/models/book.dart';
import '../../core/models/paged_result.dart';
import '../../core/network/api_client.dart';
import 'book_detail_screen.dart';

class FavoritesScreen extends StatefulWidget {
  const FavoritesScreen({super.key, required this.controller});
  final AppController controller;
  @override
  State<FavoritesScreen> createState() => _FavoritesScreenState();
}

class _FavoritesScreenState extends State<FavoritesScreen> {
  late final ApiClient _api;
  late Future<PagedResult<Book>> _items;
  @override
  void initState() {
    super.initState();
    _api = ApiClient(widget.controller.serverUrl!);
    _items = _load();
  }

  Future<PagedResult<Book>> _load({int page = 1}) =>
      _api.fetchFavoritesPage(widget.controller.accessToken!, page: page);

  Future<void> _refresh() async {
    final future = _load();
    setState(() {
      _items = future;
    });
    await future;
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Favorites')),
    body: FutureBuilder<PagedResult<Book>>(
      future: _items,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        final result =
            snapshot.data ??
            const PagedResult<Book>(items: [], page: 1, pageSize: 20, total: 0);
        final books = result.items;
        return RefreshIndicator(
          onRefresh: _refresh,
          child: ListView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.all(16),
            children: snapshot.hasError
                ? [
                    SizedBox(
                      height: 360,
                      child: Center(
                        child: Text(
                          'Unable to load favorites.\n${snapshot.error}',
                        ),
                      ),
                    ),
                  ]
                : books.isEmpty
                ? const [
                    SizedBox(
                      height: 360,
                      child: Center(child: Text('No favorite books yet.')),
                    ),
                  ]
                : [
                    ...books.map(
                      (book) => Card(
                        child: ListTile(
                          title: Text(book.title),
                          subtitle: Text(book.authors),
                          trailing: const Icon(Icons.chevron_right),
                          onTap: () async {
                            await Navigator.push(
                              context,
                              MaterialPageRoute(
                                builder: (_) => BookDetailScreen(
                                  controller: widget.controller,
                                  book: book,
                                ),
                              ),
                            );
                            _refresh();
                          },
                        ),
                      ),
                    ),
                    if (result.hasNextPage)
                      OutlinedButton(
                        onPressed: () async {
                          final next = await _load(page: result.page + 1);
                          if (!mounted) return;
                          setState(() {
                            _items = Future.value(
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
                  ],
          ),
        );
      },
    ),
  );
}
