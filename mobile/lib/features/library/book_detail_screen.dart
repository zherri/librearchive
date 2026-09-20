import 'package:flutter/material.dart';

import '../../core/app_controller.dart';
import '../../core/models/annotation.dart';
import '../../core/models/book.dart';
import '../../core/models/collection.dart';
import '../../core/models/paged_result.dart';
import '../../core/models/reading_progress.dart';
import '../../core/network/api_client.dart';
import '../reader/reader_screen.dart';

class BookDetailScreen extends StatefulWidget {
  const BookDetailScreen({
    super.key,
    required this.controller,
    required this.book,
  });

  final AppController controller;
  final Book book;

  @override
  State<BookDetailScreen> createState() => _BookDetailScreenState();
}

class _BookDetailScreenState extends State<BookDetailScreen> {
  late final ApiClient _api;
  late Future<_BookDetails> _details;
  bool _favorite = false;

  @override
  void initState() {
    super.initState();
    _api = ApiClient(widget.controller.serverUrl!);
    _details = _load();
  }

  String get _token => widget.controller.accessToken!;

  Future<_BookDetails> _load({int highlightPage = 1}) async {
    final progress = await _api.fetchProgress(_token, widget.book.id);
    final highlights = await _api.fetchHighlightsPage(
      _token,
      widget.book.id,
      page: highlightPage,
    );
    final favorites = await _api.fetchFavorites(_token);
    final isFavorite = favorites.any((book) => book.id == widget.book.id);
    if (mounted) {
      setState(() => _favorite = isFavorite);
    }
    return _BookDetails(progress: progress, highlights: highlights);
  }

  void _reload() {
    setState(() {
      _details = _load();
    });
  }

  Future<void> _editProgress(ReadingProgress current) async {
    final page = TextEditingController(text: current.currentPage.toString());
    final percentage = TextEditingController(
      text: current.progressPercent.toStringAsFixed(0),
    );
    var status = current.status;
    final saved = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Reading progress'),
        content: StatefulBuilder(
          builder: (context, setDialogState) => Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              DropdownButtonFormField<String>(
                initialValue: status,
                items: const [
                  DropdownMenuItem(value: 'unread', child: Text('Unread')),
                  DropdownMenuItem(
                    value: 'in_progress',
                    child: Text('Reading'),
                  ),
                  DropdownMenuItem(value: 'finished', child: Text('Finished')),
                ],
                onChanged: (value) => setDialogState(() => status = value!),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: page,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(labelText: 'Current page'),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: percentage,
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                decoration: const InputDecoration(
                  labelText: 'Progress percentage',
                ),
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Save'),
          ),
        ],
      ),
    );
    if (saved != true) return;
    try {
      await _api.saveProgress(
        _token,
        widget.book.id,
        status: status,
        currentPage: int.tryParse(page.text) ?? 0,
        progressPercent: double.tryParse(percentage.text) ?? 0,
        isDownloaded: current.isDownloaded,
      );
      _reload();
    } on ApiException catch (error) {
      if (mounted) _showError(error.message);
    } finally {
      page.dispose();
      percentage.dispose();
    }
  }

  Future<void> _addToCollection() async {
    try {
      final collections = await _loadAllCollections();
      if (!mounted) return;
      final collection = await showModalBottomSheet<LibraryCollection>(
        context: context,
        builder: (context) => SafeArea(
          child: ListView(
            shrinkWrap: true,
            children: [
              const ListTile(title: Text('Add to collection')),
              ...collections.map(
                (item) => ListTile(
                  title: Text(item.name),
                  onTap: () => Navigator.pop(context, item),
                ),
              ),
            ],
          ),
        ),
      );
      if (collection != null) {
        await _api.addBookToCollection(_token, collection.id, widget.book.id);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Added to ${collection.name}.')),
          );
        }
      }
    } on ApiException catch (error) {
      if (mounted) _showError(error.message);
    }
  }

  Future<List<LibraryCollection>> _loadAllCollections() async {
    final collections = <LibraryCollection>[];
    var page = 1;
    while (true) {
      final result = await _api.fetchCollectionsPage(_token, page: page);
      collections.addAll(result.items);
      if (!result.hasNextPage) return collections;
      page = result.page + 1;
    }
  }

  Future<void> _toggleFavorite() async {
    final desired = !_favorite;
    setState(() => _favorite = desired);
    try {
      await _api.setFavorite(_token, widget.book.id, desired);
      _reload();
    } on ApiException catch (error) {
      if (mounted) {
        setState(() => _favorite = !desired);
        _showError(error.message);
      }
    }
  }

  void _showError(String message) =>
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(message)));

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(
      title: Text(widget.book.title),
      actions: [
        IconButton(
          onPressed: _toggleFavorite,
          tooltip: 'Favorite',
          icon: Icon(_favorite ? Icons.favorite : Icons.favorite_border),
        ),
        IconButton(
          onPressed: _addToCollection,
          tooltip: 'Add to collection',
          icon: const Icon(Icons.bookmark_add_outlined),
        ),
      ],
    ),
    body: FutureBuilder<_BookDetails>(
      future: _details,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        if (snapshot.hasError) {
          return Center(
            child: Text('Unable to load book details.\n${snapshot.error}'),
          );
        }
        final details = snapshot.data!;
        return RefreshIndicator(
          onRefresh: () async => _reload(),
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(
                widget.book.authors,
                style: Theme.of(context).textTheme.titleLarge,
              ),
              if (widget.book.description?.isNotEmpty ?? false) ...[
                const SizedBox(height: 16),
                Text(widget.book.description!),
              ],
              const SizedBox(height: 24),
              Card(
                child: ListTile(
                  leading: const Icon(Icons.auto_stories_outlined),
                  title: Text(_statusLabel(details.progress.status)),
                  subtitle: Text(
                    'Page ${details.progress.currentPage} · ${details.progress.progressPercent.toStringAsFixed(0)}%',
                  ),
                  trailing: const Icon(Icons.edit_outlined),
                  onTap: () => _editProgress(details.progress),
                ),
              ),
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: FilledButton.icon(
                  onPressed: () async {
                    await Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => ReaderScreen(
                          controller: widget.controller,
                          book: widget.book,
                          progress: details.progress,
                        ),
                      ),
                    );
                    _reload();
                  },
                  icon: const Icon(Icons.chrome_reader_mode_outlined),
                  label: const Text('Read book'),
                ),
              ),
              const SizedBox(height: 24),
              Text(
                'Highlights & notes',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 8),
              if (details.highlights.items.isEmpty)
                const Text('No highlights have been saved for this book yet.')
              else
                ...details.highlights.items.map(_highlightCard),
              if (details.highlights.hasNextPage)
                OutlinedButton(
                  onPressed: () async {
                    final next = await _api.fetchHighlightsPage(
                      _token,
                      widget.book.id,
                      page: details.highlights.page + 1,
                    );
                    if (!mounted) return;
                    setState(() {
                      _details = Future.value(
                        _BookDetails(
                          progress: details.progress,
                          highlights: PagedResult(
                            items: [...details.highlights.items, ...next.items],
                            page: next.page,
                            pageSize: next.pageSize,
                            total: next.total,
                          ),
                        ),
                      );
                    });
                  },
                  child: const Text('Load more highlights'),
                ),
            ],
          ),
        );
      },
    ),
  );

  Widget _highlightCard(Highlight highlight) => Card(
    child: Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 14,
                height: 14,
                decoration: BoxDecoration(
                  color: _highlightColor(highlight.color),
                  borderRadius: BorderRadius.circular(3),
                ),
              ),
              const SizedBox(width: 8),
              Text('Page ${highlight.page}'),
            ],
          ),
          const SizedBox(height: 8),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
            color: _highlightColor(highlight.color).withValues(alpha: 0.35),
            child: Text('“${highlight.selectedText}”'),
          ),
          if (highlight.note != null) ...[
            const Divider(height: 24),
            Text(highlight.note!.content),
          ],
        ],
      ),
    ),
  );

  Color _highlightColor(String color) => switch (color) {
    'orange' => Colors.orange,
    'green' => Colors.green,
    'blue' => Colors.lightBlue,
    'pink' => Colors.pink,
    _ => Colors.amber,
  };

  String _statusLabel(String status) => switch (status) {
    'in_progress' => 'Currently reading',
    'finished' => 'Finished',
    _ => 'Unread',
  };
}

class _BookDetails {
  const _BookDetails({required this.progress, required this.highlights});
  final ReadingProgress progress;
  final PagedResult<Highlight> highlights;
}
