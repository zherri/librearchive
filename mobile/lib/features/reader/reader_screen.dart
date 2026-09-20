import 'package:flutter/material.dart';
import 'package:pdfrx/pdfrx.dart';

import '../../core/app_controller.dart';
import '../../core/models/annotation.dart';
import '../../core/models/book.dart';
import '../../core/models/paged_result.dart';
import '../../core/models/reading_progress.dart';
import '../../core/network/api_client.dart';

class ReaderScreen extends StatefulWidget {
  const ReaderScreen({
    super.key,
    required this.controller,
    required this.book,
    required this.progress,
  });
  final AppController controller;
  final Book book;
  final ReadingProgress progress;

  @override
  State<ReaderScreen> createState() => _ReaderScreenState();
}

class _ReaderScreenState extends State<ReaderScreen> {
  final PdfViewerController _viewer = PdfViewerController();
  PdfTextSearcher? _searcher;
  late final ApiClient _api;
  var _page = 1;
  var _pageCount = 0;
  var _saving = false;
  var _darkReader = false;
  final Map<int, List<_OverlayHighlight>> _overlays = {};
  final Map<int, List<Highlight>> _highlightsByPage = {};
  final Set<int> _loadingOverlayPages = {};

  String get _token => widget.controller.accessToken!;

  @override
  void initState() {
    super.initState();
    _api = ApiClient(widget.controller.serverUrl!);
    _page = widget.progress.currentPage > 0 ? widget.progress.currentPage : 1;
  }

  @override
  void dispose() {
    _saveProgress();
    _searcher?.dispose();
    super.dispose();
  }

  Future<void> _saveProgress() async {
    if (_saving) return;
    _saving = true;
    try {
      await _api.saveProgress(
        _token,
        widget.book.id,
        status: _pageCount > 0 && _page >= _pageCount
            ? 'finished'
            : 'in_progress',
        currentPage: _page,
        progressPercent: _pageCount == 0
            ? 0
            : (_page / _pageCount * 100).clamp(0, 100),
        isDownloaded: widget.progress.isDownloaded,
      );
    } catch (_) {
      // Reading must remain possible when a transient save fails.
    } finally {
      _saving = false;
    }
  }

  Future<void> _saveSelection() async {
    final selection = _viewer.textSelectionDelegate;
    if (!selection.hasSelectedText) {
      _message('Select text in the PDF first.');
      return;
    }
    final ranges = await selection.getSelectedTextRanges();
    if (ranges.isEmpty) return;
    final color = await _pickColor();
    if (color == null) return;
    try {
      Highlight? last;
      for (final range in ranges) {
        last = await _api.createHighlight(
          _token,
          widget.book.id,
          selectedText: range.text,
          color: color,
          page: range.pageNumber,
          startOffset: range.start,
          endOffset: range.end,
        );
      }
      await selection.clearTextSelection();
      await _loadStoredHighlights();
      if (!mounted || last == null) return;
      _message('Highlight saved.');
      await _addNote(last);
    } on ApiException catch (error) {
      if (mounted) _message(error.message);
    }
  }

  Future<void> _searchDocument() async {
    final searcher = _searcher;
    if (searcher == null) {
      _message('The PDF is still loading.');
      return;
    }
    final controller = TextEditingController(
      text: searcher.pattern?.toString() ?? '',
    );
    final query = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Search in book'),
        content: TextField(
          controller: controller,
          autofocus: true,
          textInputAction: TextInputAction.search,
          onSubmitted: (value) => Navigator.pop(context, value),
          decoration: const InputDecoration(labelText: 'Text to find'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, controller.text),
            child: const Text('Search'),
          ),
        ],
      ),
    );
    Future<void>.delayed(const Duration(milliseconds: 300), controller.dispose);
    if (query != null) {
      searcher.startTextSearch(query.trim(), searchImmediately: true);
    }
  }

  Future<void> _goToMatch(bool forward) async {
    final searcher = _searcher;
    if (searcher == null || searcher.matches.isEmpty) {
      _message('Search for text first.');
      return;
    }
    final result = forward
        ? await searcher.goToNextMatch()
        : await searcher.goToPrevMatch();
    if (result < 0 && mounted) {
      _message(forward ? 'No later match.' : 'No earlier match.');
    }
  }

  Future<void> _showHighlights() async {
    PagedResult<Highlight> result;
    try {
      result = await _api.fetchHighlightsPage(
        _token,
        widget.book.id,
        pageSize: 20,
      );
    } on ApiException catch (error) {
      _message(error.message);
      return;
    }
    if (!mounted) return;
    await showModalBottomSheet<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setSheetState) => SafeArea(
          child: ListView(
            children: [
              const ListTile(
                leading: Icon(Icons.format_quote_outlined),
                title: Text('Highlights'),
              ),
              if (result.items.isEmpty)
                const Padding(
                  padding: EdgeInsets.all(24),
                  child: Center(child: Text('No highlights in this book yet.')),
                ),
              ...result.items.map(
                (highlight) => ListTile(
                  leading: Container(
                    width: 14,
                    height: 36,
                    decoration: BoxDecoration(
                      color: _color(highlight.color),
                      borderRadius: BorderRadius.circular(3),
                    ),
                  ),
                  title: Text(
                    highlight.selectedText,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  subtitle: Text(
                    'Page ${highlight.page}${highlight.note == null ? '' : ' · Note'}',
                  ),
                  onTap: () async {
                    Navigator.pop(context);
                    await _viewer.goToPage(pageNumber: highlight.page);
                  },
                ),
              ),
              if (result.hasNextPage)
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: OutlinedButton(
                    onPressed: () async {
                      final next = await _api.fetchHighlightsPage(
                        _token,
                        widget.book.id,
                        page: result.page + 1,
                        pageSize: result.pageSize,
                      );
                      setSheetState(() {
                        result = PagedResult(
                          items: [...result.items, ...next.items],
                          page: next.page,
                          pageSize: next.pageSize,
                          total: next.total,
                        );
                      });
                    },
                    child: const Text('Load more highlights'),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _loadStoredHighlights() async {
    final highlights = <Highlight>[];
    var page = 1;
    while (true) {
      final result = await _api.fetchHighlightsPage(
        _token,
        widget.book.id,
        page: page,
        pageSize: 100,
      );
      highlights.addAll(result.items);
      if (!result.hasNextPage) break;
      page = result.page + 1;
    }
    final byPage = <int, List<Highlight>>{};
    for (final highlight in highlights) {
      if (highlight.startOffset != null && highlight.endOffset != null) {
        byPage.putIfAbsent(highlight.page, () => []).add(highlight);
      }
    }
    if (mounted) {
      setState(() {
        _highlightsByPage
          ..clear()
          ..addAll(byPage);
        _overlays.clear();
        _loadingOverlayPages.clear();
      });
      _viewer.invalidate();
    }
  }

  void _ensurePageOverlays(PdfPage pdfPage) {
    final highlights = _highlightsByPage[pdfPage.pageNumber];
    if (highlights == null ||
        _overlays.containsKey(pdfPage.pageNumber) ||
        _loadingOverlayPages.contains(pdfPage.pageNumber)) {
      return;
    }
    _loadingOverlayPages.add(pdfPage.pageNumber);
    Future<void>(() async {
      try {
        final text = await pdfPage.loadStructuredText();
        final items = <_OverlayHighlight>[];
        for (final highlight in highlights) {
          final start = highlight.startOffset!.clamp(0, text.fullText.length);
          final end = highlight.endOffset!.clamp(start, text.fullText.length);
          if (start == end) continue;
          final range = PdfPageTextRange(
            pageText: text,
            start: start,
            end: end,
          );
          for (final fragment in range.enumerateFragmentBoundingRects()) {
            final bounds = fragment.bounds;
            items.add(
              _OverlayHighlight(
                highlight,
                Rect.fromLTWH(
                  bounds.left / pdfPage.width,
                  (pdfPage.height - bounds.top) / pdfPage.height,
                  bounds.width / pdfPage.width,
                  bounds.height / pdfPage.height,
                ),
              ),
            );
          }
        }
        if (mounted) {
          setState(() => _overlays[pdfPage.pageNumber] = items);
          _viewer.invalidate();
        }
      } finally {
        _loadingOverlayPages.remove(pdfPage.pageNumber);
      }
    });
  }

  Future<void> _openHighlight(Highlight highlight) async {
    final action = await showModalBottomSheet<String>(
      context: context,
      builder: (context) => SafeArea(
        child: ListView(
          shrinkWrap: true,
          children: [
            ListTile(
              title: Text(
                '“${highlight.selectedText}”',
                maxLines: 3,
                overflow: TextOverflow.ellipsis,
              ),
              subtitle: Text('Page ${highlight.page}'),
            ),
            if (highlight.note != null) ...[
              const Divider(),
              const ListTile(
                leading: Icon(Icons.sticky_note_2_outlined),
                title: Text('Note'),
              ),
              ListTile(title: Text(highlight.note!.content)),
            ],
            const Divider(),
            ListTile(
              leading: const Icon(Icons.edit_note_outlined),
              title: Text(highlight.note == null ? 'Add note' : 'Edit note'),
              onTap: () => Navigator.pop(context, 'note'),
            ),
            if (highlight.note != null)
              ListTile(
                leading: const Icon(
                  Icons.remove_circle_outline,
                  color: Colors.red,
                ),
                title: const Text(
                  'Remove note',
                  style: TextStyle(color: Colors.red),
                ),
                onTap: () => Navigator.pop(context, 'remove-note'),
              ),
            ListTile(
              leading: const Icon(Icons.palette_outlined),
              title: const Text('Change color'),
              onTap: () => Navigator.pop(context, 'color'),
            ),
            ListTile(
              leading: const Icon(Icons.delete_outline, color: Colors.red),
              title: const Text(
                'Remove highlight',
                style: TextStyle(color: Colors.red),
              ),
              onTap: () => Navigator.pop(context, 'delete'),
            ),
          ],
        ),
      ),
    );
    if (action == null) return;
    try {
      if (action == 'delete') {
        await _api.deleteHighlight(_token, highlight.id);
        if (mounted) _message('Highlight removed.');
      } else if (action == 'note') {
        await _editNote(highlight);
        return;
      } else if (action == 'remove-note') {
        await _api.deleteNote(_token, highlight.note!.id);
        await _loadStoredHighlights();
        if (mounted) _message('Note removed.');
        return;
      } else {
        final color = await _pickColor();
        if (color == null) return;
        await _api.updateHighlight(_token, highlight, color: color);
        if (mounted) _message('Highlight color updated.');
      }
      await _loadStoredHighlights();
    } on ApiException catch (error) {
      if (mounted) _message(error.message);
    }
  }

  Future<void> _editNote(Highlight highlight) async {
    final controller = TextEditingController(
      text: highlight.note?.content ?? '',
    );
    final content = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(highlight.note == null ? 'Add note' : 'Edit note'),
        content: TextField(
          controller: controller,
          autofocus: true,
          maxLines: 5,
          decoration: const InputDecoration(labelText: 'Note'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, controller.text.trim()),
            child: const Text('Save'),
          ),
        ],
      ),
    );
    Future<void>.delayed(const Duration(milliseconds: 300), controller.dispose);
    if (content == null || content.isEmpty) return;
    try {
      if (highlight.note == null) {
        await _api.createNote(_token, highlight.id, content);
      } else {
        await _api.updateNote(_token, highlight.note!.id, content);
      }
      await _loadStoredHighlights();
      if (mounted) _message('Note saved.');
    } on ApiException catch (error) {
      if (mounted) _message(error.message);
    }
  }

  Future<String?> _pickColor() => showModalBottomSheet<String>(
    context: context,
    builder: (context) => SafeArea(
      child: Wrap(
        children: [
          const ListTile(title: Text('Highlight color')),
          for (final color in const [
            'yellow',
            'orange',
            'green',
            'blue',
            'pink',
          ])
            ListTile(
              leading: CircleAvatar(backgroundColor: _color(color)),
              title: Text(color[0].toUpperCase() + color.substring(1)),
              onTap: () => Navigator.pop(context, color),
            ),
        ],
      ),
    ),
  );

  Future<void> _addNote(Highlight highlight) async {
    final controller = TextEditingController();
    final content = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Add a note'),
        content: TextField(
          controller: controller,
          autofocus: true,
          maxLines: 4,
          decoration: const InputDecoration(labelText: 'Optional note'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Skip'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, controller.text.trim()),
            child: const Text('Save note'),
          ),
        ],
      ),
    );
    Future<void>.delayed(const Duration(milliseconds: 300), controller.dispose);
    if (content == null || content.isEmpty) return;
    try {
      await _api.createNote(_token, highlight.id, content);
      await _loadStoredHighlights();
      if (mounted) _message('Note saved.');
    } on ApiException catch (error) {
      if (mounted) _message(error.message);
    }
  }

  void _message(String value) =>
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(value)));

  Color _color(String color) => switch (color) {
    'orange' => Colors.orange,
    'green' => Colors.green,
    'blue' => Colors.blue,
    'pink' => Colors.pink,
    _ => Colors.amber,
  };

  Color _pdfHighlightColor(String color) {
    final base = _color(color);
    if (!_darkReader) return base;
    return Color.from(
      alpha: base.a,
      red: 1 - base.r,
      green: 1 - base.g,
      blue: 1 - base.b,
      colorSpace: base.colorSpace,
    );
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(
      title: Text(widget.book.title),
      actions: [
        IconButton(
          onPressed: _saveSelection,
          tooltip: 'Highlight selected text',
          icon: const Icon(Icons.highlight_alt),
        ),
        IconButton(
          onPressed: () => setState(() => _darkReader = !_darkReader),
          tooltip: _darkReader ? 'Use light reader' : 'Use dark reader',
          icon: Icon(
            _darkReader ? Icons.light_mode_outlined : Icons.dark_mode_outlined,
          ),
        ),
        IconButton(
          onPressed: _showHighlights,
          tooltip: 'Book highlights',
          icon: const Icon(Icons.format_quote_outlined),
        ),
        IconButton(
          onPressed: _searchDocument,
          tooltip: 'Search in book',
          icon: const Icon(Icons.search),
        ),
        IconButton(
          onPressed: _searcher?.matches.isNotEmpty ?? false
              ? () => _goToMatch(false)
              : null,
          tooltip: 'Previous match',
          icon: const Icon(Icons.keyboard_arrow_up),
        ),
        IconButton(
          onPressed: _searcher?.matches.isNotEmpty ?? false
              ? () => _goToMatch(true)
              : null,
          tooltip: 'Next match',
          icon: const Icon(Icons.keyboard_arrow_down),
        ),
        IconButton(
          onPressed: _saveProgress,
          tooltip: 'Save progress',
          icon: const Icon(Icons.bookmark_added_outlined),
        ),
      ],
    ),
    body: Column(
      children: [
        Expanded(
          child: ColorFiltered(
            colorFilter: _darkReader
                ? const ColorFilter.matrix(<double>[
                    -1,
                    0,
                    0,
                    0,
                    255,
                    0,
                    -1,
                    0,
                    0,
                    255,
                    0,
                    0,
                    -1,
                    0,
                    255,
                    0,
                    0,
                    0,
                    1,
                    0,
                  ])
                : const ColorFilter.mode(Colors.transparent, BlendMode.dst),
            child: PdfViewer.uri(
              Uri.parse(
                '${widget.controller.serverUrl}/api/v1/books/${widget.book.id}/file',
              ),
              headers: {'Authorization': 'Bearer $_token'},
              controller: _viewer,
              initialPageNumber: _page,
              params: PdfViewerParams(
                backgroundColor: const Color(0xFFE8DDCA),
                pageDropShadow: const BoxShadow(color: Colors.transparent),
                textSelectionParams: const PdfTextSelectionParams(
                  enabled: true,
                  showContextMenuAutomatically: true,
                ),
                pagePaintCallbacks: [
                  if (_searcher != null) _searcher!.pageTextMatchPaintCallback,
                ],
                pageOverlaysBuilder: (context, pageRect, pdfPage) {
                  _ensurePageOverlays(pdfPage);
                  return [
                    for (final overlay
                        in _overlays[pdfPage.pageNumber] ?? const [])
                      Positioned.fromRect(
                        rect: Rect.fromLTWH(
                          overlay.rect.left * pageRect.width,
                          overlay.rect.top * pageRect.height,
                          overlay.rect.width * pageRect.width,
                          overlay.rect.height * pageRect.height,
                        ),
                        child: PdfOverlayInteractionRegion(
                          onTap: (_) {
                            _openHighlight(overlay.highlight);
                            return true;
                          },
                          child: DecoratedBox(
                            decoration: BoxDecoration(
                              color: _pdfHighlightColor(overlay.highlight.color)
                                  .withValues(alpha: 0.35),
                            ),
                          ),
                        ),
                      ),
                  ];
                },
                onPageChanged: (page) {
                  if (page == null || page == _page) return;
                  setState(() => _page = page);
                  _saveProgress();
                },
                onViewerReady: (document, _) {
                  _loadStoredHighlights();
                  _searcher ??= PdfTextSearcher(_viewer);
                  _searcher!.addListener(() {
                    WidgetsBinding.instance.addPostFrameCallback((_) {
                      if (mounted) setState(() {});
                    });
                  });
                  WidgetsBinding.instance.addPostFrameCallback((_) {
                    if (mounted) {
                      setState(() => _pageCount = document.pages.length);
                    }
                  });
                },
              ),
            ),
          ),
        ),
        SafeArea(
          top: false,
          child: Padding(
            padding: const EdgeInsets.all(10),
            child: Text(
              _pageCount == 0
                  ? 'Loading pages…'
                  : 'Page $_page of $_pageCount · ${_searcher?.matches.isEmpty ?? true ? 'Select text, then tap highlight.' : '${_searcher!.matches.length} search matches'}',
            ),
          ),
        ),
      ],
    ),
  );
}

class _OverlayHighlight {
  const _OverlayHighlight(this.highlight, this.rect);
  final Highlight highlight;
  final Rect rect;
}
