import 'package:flutter/material.dart';

import '../../core/app_controller.dart';
import '../../core/models/book.dart';
import '../../core/models/collection.dart';
import '../../core/models/paged_result.dart';
import '../../core/network/api_client.dart';
import 'book_detail_screen.dart';

class CollectionsScreen extends StatefulWidget {
  const CollectionsScreen({super.key, required this.controller});
  final AppController controller;
  @override
  State<CollectionsScreen> createState() => _CollectionsScreenState();
}

class _CollectionsScreenState extends State<CollectionsScreen> {
  late final ApiClient _api;
  late Future<PagedResult<LibraryCollection>> _collections;
  @override
  void initState() {
    super.initState();
    _api = ApiClient(widget.controller.serverUrl!);
    _collections = _load();
  }

  Future<PagedResult<LibraryCollection>> _load({int page = 1}) =>
      _api.fetchCollectionsPage(widget.controller.accessToken!, page: page);
  Future<void> _refresh() async {
    final future = _load();
    setState(() {
      _collections = future;
    });
    await future;
  }

  Future<void> _create() async {
    final name = TextEditingController();
    final create = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('New collection'),
        content: TextField(
          controller: name,
          autofocus: true,
          decoration: const InputDecoration(labelText: 'Collection name'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Create'),
          ),
        ],
      ),
    );
    if (create == true && name.text.trim().isNotEmpty) {
      try {
        await _api.createCollection(
          widget.controller.accessToken!,
          name.text.trim(),
        );
        await _refresh();
      } on ApiException catch (error) {
        if (mounted) {
          ScaffoldMessenger.of(context)
              .showSnackBar(SnackBar(content: Text(error.message)));
        }
      }
    }
    name.dispose();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Collections')),
    floatingActionButton: FloatingActionButton(
      onPressed: _create,
      child: const Icon(Icons.add),
    ),
    body: FutureBuilder<PagedResult<LibraryCollection>>(
      future: _collections,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        final result =
            snapshot.data ??
            const PagedResult(
              items: <LibraryCollection>[],
              page: 1,
              pageSize: 20,
              total: 0,
            );
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
                          'Unable to load collections.\n${snapshot.error}',
                        ),
                      ),
                    ),
                  ]
                : result.items.isEmpty
                ? const [
                    SizedBox(
                      height: 360,
                      child: Center(
                        child: Text(
                          'Create a collection for books you want to group together.',
                        ),
                      ),
                    ),
                  ]
                : [
                    ...result.items.map(
                      (collection) => Card(
                        child: ListTile(
                          title: Text(collection.name),
                          trailing: const Icon(Icons.chevron_right),
                          onTap: () async {
                            await Navigator.push(
                              context,
                              MaterialPageRoute(
                                builder: (_) => _CollectionDetailScreen(
                                  controller: widget.controller,
                                  collection: collection,
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
                            _collections = Future.value(
                              PagedResult(
                                items: [...result.items, ...next.items],
                                page: next.page,
                                pageSize: next.pageSize,
                                total: next.total,
                              ),
                            );
                          });
                        },
                        child: const Text('Load more collections'),
                      ),
                  ],
          ),
        );
      },
    ),
  );
}

class _CollectionDetailScreen extends StatefulWidget {
  const _CollectionDetailScreen({
    required this.controller,
    required this.collection,
  });
  final AppController controller;
  final LibraryCollection collection;
  @override
  State<_CollectionDetailScreen> createState() =>
      _CollectionDetailScreenState();
}

class _CollectionDetailScreenState extends State<_CollectionDetailScreen> {
  late final ApiClient _api;
  late Future<CollectionDetails> _details;
  @override
  void initState() {
    super.initState();
    _api = ApiClient(widget.controller.serverUrl!);
    _details = _load();
  }

  Future<CollectionDetails> _load({int page = 1}) => _api.fetchCollection(
    widget.controller.accessToken!,
    widget.collection.id,
    page: page,
  );
  Future<void> _refresh() async {
    final future = _load();
    setState(() {
      _details = future;
    });
    await future;
  }

  Future<void> _removeBook(Book book) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Remove book?'),
        content: Text('Remove "${book.title}" from this collection?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Remove'),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;

    try {
      await _api.removeBookFromCollection(
        widget.controller.accessToken!,
        widget.collection.id,
        book.id,
      );
      if (!mounted) return;
      await _refresh();
    } catch (error) {
      if (!mounted) return;
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text('Unable to remove book. $error')));
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: Text(widget.collection.name)),
    body: FutureBuilder<CollectionDetails>(
      future: _details,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        final details = snapshot.data;
        final books = details?.books ?? [];
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
                          'Unable to load collection.\n${snapshot.error}',
                        ),
                      ),
                    ),
                  ]
                : books.isEmpty
                ? const [
                    SizedBox(
                      height: 360,
                      child: Center(
                        child: Text('No books in this collection.'),
                      ),
                    ),
                  ]
                : [
                    ...books.map(
                      (book) => Card(
                        child: ListTile(
                          title: Text(book.title),
                          subtitle: Text(book.authors),
                          trailing: IconButton(
                            icon: const Icon(Icons.remove_circle_outline),
                            tooltip: 'Remove from collection',
                            onPressed: () => _removeBook(book),
                          ),
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
                    if (details!.hasNextPage)
                      OutlinedButton(
                        onPressed: () async {
                          final next = await _load(page: details.page + 1);
                          if (!mounted) return;
                          setState(() {
                            _details = Future.value(
                              CollectionDetails(
                                collection: details.collection,
                                books: [...details.books, ...next.books],
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
