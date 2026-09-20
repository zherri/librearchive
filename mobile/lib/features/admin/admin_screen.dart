import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../core/app_controller.dart';
import '../../core/models/book.dart';
import '../../core/models/catalog_item.dart';
import '../../core/models/paged_result.dart';
import '../../core/models/user.dart';
import '../../core/network/api_client.dart';

class AdminScreen extends StatefulWidget {
  const AdminScreen({super.key, required this.controller});
  final AppController controller;
  @override
  State<AdminScreen> createState() => _AdminScreenState();
}

class _AdminScreenState extends State<AdminScreen> {
  late final ApiClient _api;
  int _section = 0;
  @override
  void initState() {
    super.initState();
    _api = ApiClient(widget.controller.serverUrl!);
  }

  String get _token => widget.controller.accessToken!;

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Archive administration')),
    body: Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(16),
          child: SegmentedButton<int>(
            segments: const [
              ButtonSegment(
                value: 0,
                label: Text('Users'),
                icon: Icon(Icons.people_outline),
              ),
              ButtonSegment(
                value: 1,
                label: Text('Books'),
                icon: Icon(Icons.menu_book_outlined),
              ),
              ButtonSegment(
                value: 2,
                label: Text('Categories'),
                icon: Icon(Icons.category_outlined),
              ),
              ButtonSegment(
                value: 3,
                label: Text('Tags'),
                icon: Icon(Icons.sell_outlined),
              ),
            ],
            selected: {_section},
            onSelectionChanged: (value) =>
                setState(() => _section = value.first),
          ),
        ),
        Expanded(
          child: switch (_section) {
            0 => _UsersPanel(api: _api, token: _token),
            1 => _BooksPanel(api: _api, token: _token),
            2 => _CatalogPanel(api: _api, token: _token, tags: false),
            _ => _CatalogPanel(api: _api, token: _token, tags: true),
          },
        ),
      ],
    ),
  );
}

class _UsersPanel extends StatefulWidget {
  const _UsersPanel({required this.api, required this.token});
  final ApiClient api;
  final String token;
  @override
  State<_UsersPanel> createState() => _UsersPanelState();
}

class _UsersPanelState extends State<_UsersPanel> {
  late Future<PagedResult<ArchiveUser>> _users;
  @override
  void initState() {
    super.initState();
    _users = _load();
  }

  Future<PagedResult<ArchiveUser>> _load({int page = 1}) =>
      widget.api.fetchUsersPage(widget.token, page: page);

  void _reload() {
    setState(() {
      _users = _load();
    });
  }

  Future<void> _create() async {
    final username = TextEditingController();
    final name = TextEditingController();
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Create reader'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: name,
              decoration: const InputDecoration(labelText: 'Name'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: username,
              decoration: const InputDecoration(labelText: 'Username'),
            ),
          ],
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
    if (confirmed == true) {
      try {
        final result = await widget.api.createUser(
          widget.token,
          username.text.trim(),
          name.text.trim(),
        );
        if (mounted) {
          await _showPassphrase(
            context,
            result.user.username,
            result.passphrase,
          );
        }
        _reload();
      } on ApiException catch (error) {
        if (mounted) _snack(error.message);
      }
    }
    username.dispose();
    name.dispose();
  }

  Future<void> _reset(ArchiveUser user) async {
    try {
      final result = await widget.api.resetUserPassphrase(
        widget.token,
        user.id,
      );
      if (mounted) {
        await _showPassphrase(context, user.username, result.passphrase);
      }
    } on ApiException catch (error) {
      if (mounted) _snack(error.message);
    }
  }

  void _snack(String message) =>
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(message)));
  @override
  Widget build(BuildContext context) => Scaffold(
    floatingActionButton: FloatingActionButton(
      onPressed: _create,
      child: const Icon(Icons.person_add_alt_1),
    ),
    body: FutureBuilder<PagedResult<ArchiveUser>>(
      future: _users,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        if (snapshot.hasError) return Center(child: Text('${snapshot.error}'));
        final result =
            snapshot.data ??
            const PagedResult<ArchiveUser>(
              items: [],
              page: 1,
              pageSize: 20,
              total: 0,
            );
        final users = result.items;
        return RefreshIndicator(
          onRefresh: () async => _reload(),
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              ...users.map(
                (user) => Card(
                  child: ListTile(
                    title: Text(user.name),
                    subtitle: Text(
                      '@${user.username} · ${user.role}${user.isActive ? '' : ' · inactive'}',
                    ),
                    leading: Icon(
                      user.isAdministrator
                          ? Icons.admin_panel_settings_outlined
                          : Icons.person_outline,
                    ),
                    trailing: PopupMenuButton<String>(
                      onSelected: (value) async {
                        if (value == 'reset') await _reset(user);
                        if (value == 'active') {
                          try {
                            await widget.api.updateUser(widget.token, user.id, {
                              'isActive': !user.isActive,
                            });
                            _reload();
                          } on ApiException catch (error) {
                            _snack(error.message);
                          }
                        }
                        if (value == 'delete') {
                          try {
                            await widget.api.deleteUser(widget.token, user.id);
                            _reload();
                          } on ApiException catch (error) {
                            _snack(error.message);
                          }
                        }
                      },
                      itemBuilder: (_) => [
                        const PopupMenuItem(
                          value: 'reset',
                          child: Text('Reset passphrase'),
                        ),
                        PopupMenuItem(
                          value: 'active',
                          child: Text(
                            user.isActive ? 'Deactivate' : 'Activate',
                          ),
                        ),
                        const PopupMenuItem(
                          value: 'delete',
                          child: Text('Delete user'),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              if (result.hasNextPage)
                OutlinedButton(
                  onPressed: () async {
                    final next = await _load(page: result.page + 1);
                    if (!mounted) return;
                    setState(() {
                      _users = Future.value(
                        PagedResult(
                          items: [...users, ...next.items],
                          page: next.page,
                          pageSize: next.pageSize,
                          total: next.total,
                        ),
                      );
                    });
                  },
                  child: const Text('Load more users'),
                ),
            ],
          ),
        );
      },
    ),
  );
}

class _BooksPanel extends StatefulWidget {
  const _BooksPanel({required this.api, required this.token});
  final ApiClient api;
  final String token;
  @override
  State<_BooksPanel> createState() => _BooksPanelState();
}

class _BooksPanelState extends State<_BooksPanel> {
  late Future<PagedResult<Book>> _books;
  late final TextEditingController _searchController;
  String _search = '';
  @override
  void initState() {
    super.initState();
    _searchController = TextEditingController();
    _books = _load();
  }

  Future<PagedResult<Book>> _load({int page = 1}) =>
      widget.api.fetchBooksPage(widget.token, search: _search, page: page);

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _reload() {
    setState(() {
      _books = _load();
    });
  }

  Future<void> _upload() async {
    FilePickerResult? pdf;
    try {
      pdf = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['pdf'],
        withData: true,
      );
    } on MissingPluginException {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'File selection is unavailable. Fully restart the app after installing its dependencies.',
            ),
          ),
        );
      }
      return;
    } on PlatformException catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Unable to open file selection: ${error.message ?? error.code}',
            ),
          ),
        );
      }
      return;
    }
    final selectedPdf = pdf;
    if (selectedPdf == null || !mounted) return;
    final cover = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['jpg', 'jpeg', 'png', 'webp'],
      withData: true,
    );
    if (!mounted) return;
    final title = TextEditingController();
    final authors = TextEditingController();
    final description = TextEditingController();
    final publisher = TextEditingController();
    final year = TextEditingController();
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Add PDF book'),
        content: SizedBox(
          width: 360,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'File: ${selectedPdf.files.single.name}',
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                Text('Cover: ${cover?.files.single.name ?? 'None selected'}'),
                const SizedBox(height: 12),
                TextField(
                  controller: title,
                  decoration: const InputDecoration(labelText: 'Title'),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: authors,
                  decoration: const InputDecoration(labelText: 'Author(s)'),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: description,
                  maxLines: 3,
                  decoration: const InputDecoration(labelText: 'Description'),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: publisher,
                  decoration: const InputDecoration(labelText: 'Publisher'),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: year,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(
                    labelText: 'Publication year',
                  ),
                ),
              ],
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Upload'),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      try {
        await widget.api.uploadBook(
          widget.token,
          title: title.text.trim(),
          authors: authors.text.trim(),
          pdf: selectedPdf.files.single,
          cover: cover?.files.single,
          description: description.text.trim(),
          publisher: publisher.text.trim(),
          publishedYear: int.tryParse(year.text),
        );
        _reload();
      } on ApiException catch (error) {
        if (mounted) {
          ScaffoldMessenger.of(context)
              .showSnackBar(SnackBar(content: Text(error.message)));
        }
      }
    }
    title.dispose();
    authors.dispose();
    description.dispose();
    publisher.dispose();
    year.dispose();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    floatingActionButton: FloatingActionButton(
      onPressed: _upload,
      child: const Icon(Icons.upload_file),
    ),
    body: FutureBuilder<PagedResult<Book>>(
      future: _books,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        if (snapshot.hasError) return Center(child: Text('${snapshot.error}'));
        return RefreshIndicator(
          onRefresh: () async {
            _reload();
            await _books;
          },
          child: ListView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.all(16),
            children: [
              TextField(
                controller: _searchController,
                decoration: InputDecoration(
                  labelText: 'Search books',
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
                textInputAction: TextInputAction.search,
                onSubmitted: (value) {
                  _search = value.trim();
                  _reload();
                },
              ),
              const SizedBox(height: 12),
              if (snapshot.data!.items.isEmpty)
                SizedBox(
                  height: 220,
                  child: Center(
                    child: Text(
                      _search.isEmpty
                          ? 'No books have been added yet.'
                          : 'No books match your search.',
                    ),
                  ),
                ),
              ...snapshot.data!.items.map(
                (book) => Card(
                  child: ListTile(
                    title: Text(book.title),
                    subtitle: Text(book.authors),
                    trailing: IconButton(
                      icon: const Icon(Icons.delete_outline),
                      onPressed: () async {
                        final messenger = ScaffoldMessenger.of(context);
                        try {
                          await widget.api.deleteBook(widget.token, book.id);
                          _reload();
                        } on ApiException catch (error) {
                          if (!mounted) return;
                          {
                            messenger.showSnackBar(
                              SnackBar(content: Text(error.message)),
                            );
                          }
                        }
                      },
                    ),
                  ),
                ),
              ),
              if (snapshot.data!.hasNextPage)
                OutlinedButton(
                  onPressed: () async {
                    final current = snapshot.data!;
                    final next = await _load(page: current.page + 1);
                    if (!mounted) return;
                    setState(() {
                      _books = Future.value(
                        PagedResult(
                          items: [...current.items, ...next.items],
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

class _CatalogPanel extends StatefulWidget {
  const _CatalogPanel({
    required this.api,
    required this.token,
    required this.tags,
  });
  final ApiClient api;
  final String token;
  final bool tags;
  @override
  State<_CatalogPanel> createState() => _CatalogPanelState();
}

class _CatalogPanelState extends State<_CatalogPanel> {
  late Future<PagedResult<CatalogItem>> _items;
  @override
  void initState() {
    super.initState();
    _items = _load();
  }

  Future<PagedResult<CatalogItem>> _load({int page = 1}) => widget.api
      .fetchCatalogItemsPage(widget.token, tags: widget.tags, page: page);

  void _reload() {
    setState(() {
      _items = _load();
    });
  }

  Future<void> _create() async {
    final name = TextEditingController();
    final create = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('New ${widget.tags ? 'tag' : 'category'}'),
        content: TextField(
          controller: name,
          autofocus: true,
          decoration: const InputDecoration(labelText: 'Name'),
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
    if (create == true) {
      try {
        await widget.api.createCatalogItem(
          widget.token,
          tags: widget.tags,
          name: name.text.trim(),
        );
        _reload();
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
    floatingActionButton: FloatingActionButton(
      onPressed: _create,
      child: const Icon(Icons.add),
    ),
    body: FutureBuilder<PagedResult<CatalogItem>>(
      future: _items,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        if (snapshot.hasError) return Center(child: Text('${snapshot.error}'));
        return RefreshIndicator(
          onRefresh: () async {
            _reload();
            await _items;
          },
          child: ListView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.all(16),
            children: [
              ...snapshot.data!.items.map(
                (item) => Card(
                  child: ListTile(
                    title: Text(item.name),
                    trailing: IconButton(
                      icon: const Icon(Icons.delete_outline),
                      onPressed: () async {
                        final messenger = ScaffoldMessenger.of(context);
                        try {
                          await widget.api.deleteCatalogItem(
                            widget.token,
                            tags: widget.tags,
                            itemId: item.id,
                          );
                          _reload();
                        } on ApiException catch (error) {
                          if (!mounted) return;
                          {
                            messenger.showSnackBar(
                              SnackBar(content: Text(error.message)),
                            );
                          }
                        }
                      },
                    ),
                  ),
                ),
              ),
              if (snapshot.data!.hasNextPage)
                OutlinedButton(
                  onPressed: () async {
                    final current = snapshot.data!;
                    final next = await _load(page: current.page + 1);
                    if (!mounted) return;
                    setState(() {
                      _items = Future.value(
                        PagedResult(
                          items: [...current.items, ...next.items],
                          page: next.page,
                          pageSize: next.pageSize,
                          total: next.total,
                        ),
                      );
                    });
                  },
                  child: Text(
                    'Load more ${widget.tags ? 'tags' : 'categories'}',
                  ),
                ),
            ],
          ),
        );
      },
    ),
  );
}

Future<void> _showPassphrase(
  BuildContext context,
  String username,
  String passphrase,
) => showDialog<void>(
  context: context,
  barrierDismissible: false,
  builder: (context) => AlertDialog(
    title: const Text('Store the passphrase'),
    content: Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Account: @$username'),
        const SizedBox(height: 12),
        const Text('This passphrase is displayed once only.'),
        const SizedBox(height: 12),
        SelectableText(passphrase),
      ],
    ),
    actions: [
      FilledButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('I stored it'),
      ),
    ],
  ),
);
