import 'package:flutter/material.dart';

import '../../core/app_controller.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key, required this.controller});
  final AppController controller;
  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _username = TextEditingController();
  final _passphraseWords = List.generate(12, (_) => TextEditingController());
  final _wordFocusNodes = List.generate(12, (_) => FocusNode());
  bool _loading = false;
  bool _enteringPassphrase = false;
  String? _error;
  @override
  void dispose() {
    _username.dispose();
    for (final controller in _passphraseWords) {
      controller.dispose();
    }
    for (final focusNode in _wordFocusNodes) {
      focusNode.dispose();
    }
    super.dispose();
  }

  void _continueToPassphrase() {
    if (!_formKey.currentState!.validate()) return;
    setState(() {
      _enteringPassphrase = true;
      _error = null;
    });
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _wordFocusNodes.first.requestFocus();
    });
  }

  Future<void> _submit() async {
    final words = _passphraseWords.map((controller) => controller.text.trim());
    if (words.any((word) => word.isEmpty)) {
      setState(() => _error = 'Enter all twelve passphrase words.');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await widget.controller.signIn(_username.text, words.join(' '));
    } catch (error) {
      if (mounted) setState(() => _error = error.toString());
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _handleWordChanged(int index, String value) {
    final words = value.trim().split(RegExp(r'\s+'));
    if (words.length < 2) return;
    for (
      var offset = 0;
      offset < words.length && index + offset < 12;
      offset++
    ) {
      final target = _passphraseWords[index + offset];
      if (target.text != words[offset]) target.text = words[offset];
    }
    final nextIndex = (index + words.length).clamp(0, 11).toInt();
    _wordFocusNodes[nextIndex].requestFocus();
  }

  void _focusNextWord(int index) {
    if (index < _wordFocusNodes.length - 1) {
      _wordFocusNodes[index + 1].requestFocus();
    } else {
      FocusScope.of(context).unfocus();
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(
      actions: [
        IconButton(
          onPressed: widget.controller.changeServer,
          tooltip: 'Change server',
          icon: const Icon(Icons.settings_ethernet),
        ),
      ],
    ),
    body: SafeArea(
      child: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(28),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: Card(
              child: Padding(
                padding: const EdgeInsets.all(28),
                child: Form(
                  key: _formKey,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _enteringPassphrase
                            ? 'Enter your passphrase'
                            : 'Enter the library',
                        style: Theme.of(context).textTheme.headlineMedium,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        widget.controller.serverUrl ?? '',
                        style: Theme.of(context).textTheme.bodyMedium,
                      ),
                      const SizedBox(height: 28),
                      if (!_enteringPassphrase)
                        TextFormField(
                          controller: _username,
                          textInputAction: TextInputAction.next,
                          decoration: const InputDecoration(
                            labelText: 'Username',
                          ),
                          validator: (value) =>
                              value == null || value.trim().isEmpty
                              ? 'Enter your username'
                              : null,
                          onFieldSubmitted: (_) => _continueToPassphrase(),
                        )
                      else ...[
                        Text(
                          'Enter the twelve words in order. You may paste the complete passphrase into the first field.',
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                        const SizedBox(height: 20),
                        Wrap(
                          spacing: 12,
                          runSpacing: 12,
                          children: List.generate(
                            12,
                            (index) => SizedBox(
                              width: 126,
                              child: TextField(
                                controller: _passphraseWords[index],
                                focusNode: _wordFocusNodes[index],
                                autocorrect: false,
                                enableSuggestions: false,
                                textInputAction: index == 11
                                    ? TextInputAction.done
                                    : TextInputAction.next,
                                decoration: InputDecoration(
                                  labelText: 'Word ${index + 1}',
                                ),
                                onChanged: (value) =>
                                    _handleWordChanged(index, value),
                                onSubmitted: (_) => _focusNextWord(index),
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(height: 8),
                        TextButton.icon(
                          onPressed: _loading
                              ? null
                              : () => setState(() {
                                  _enteringPassphrase = false;
                                  _error = null;
                                }),
                          icon: const Icon(Icons.arrow_back),
                          label: const Text('Change username'),
                        ),
                      ],
                      if (_error != null) ...[
                        const SizedBox(height: 14),
                        Text(
                          _error!,
                          style: TextStyle(
                            color: Theme.of(context).colorScheme.error,
                          ),
                        ),
                      ],
                      const SizedBox(height: 24),
                      FilledButton(
                        onPressed: _loading
                            ? null
                            : (_enteringPassphrase
                                  ? _submit
                                  : _continueToPassphrase),
                        child: Text(
                          _loading
                              ? 'Opening archive...'
                              : (_enteringPassphrase ? 'Sign in' : 'Continue'),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    ),
  );
}
