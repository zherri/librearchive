import 'package:flutter/material.dart';

import '../../core/app_controller.dart';
import '../../core/models/paged_result.dart';
import '../../core/models/reading_progress.dart';
import '../../core/network/api_client.dart';
import 'book_detail_screen.dart';

class ReadingScreen extends StatefulWidget {
  const ReadingScreen({super.key, required this.controller});
  final AppController controller;
  @override
  State<ReadingScreen> createState() => _ReadingScreenState();
}

class _ReadingScreenState extends State<ReadingScreen> {
  late final ApiClient _api;
  late Future<PagedResult<ReadingProgress>> _items;
  @override
  void initState() {
    super.initState();
    _api = ApiClient(widget.controller.serverUrl!);
    _items = _load();
  }

  Future<PagedResult<ReadingProgress>> _load({int page = 1}) =>
      _api.fetchReadingProgressPage(
        widget.controller.accessToken!,
        status: 'in_progress',
        page: page,
      );
  Future<void> _refresh() async {
    final future = _load();
    setState(() {
      _items = future;
    });
    await future;
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Reading')),
    body: FutureBuilder<PagedResult<ReadingProgress>>(
      future: _items,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        final result =
            snapshot.data ??
            const PagedResult(
              items: <ReadingProgress>[],
              page: 1,
              pageSize: 20,
              total: 0,
            );
        final items = result.items.where((item) => item.book != null).toList();
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
                          'Unable to load reading.\n${snapshot.error}',
                        ),
                      ),
                    ),
                  ]
                : items.isEmpty
                ? const [
                    SizedBox(
                      height: 360,
                      child: Center(
                        child: Text('No books currently being read.'),
                      ),
                    ),
                  ]
                : [
                    ...items.map(
                      (item) => Card(
                        child: ListTile(
                          title: Text(item.book!.title),
                          subtitle: Text(
                            '${item.book!.authors} · ${item.progressPercent.toStringAsFixed(0)}%',
                          ),
                          trailing: const Icon(Icons.chevron_right),
                          onTap: () async {
                            await Navigator.push(
                              context,
                              MaterialPageRoute(
                                builder: (_) => BookDetailScreen(
                                  controller: widget.controller,
                                  book: item.book!,
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
                                items: [...result.items, ...next.items],
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
