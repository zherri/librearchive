import 'package:flutter/material.dart';

import '../../core/app_controller.dart';

class SettingsScreen extends StatelessWidget {
  const SettingsScreen({super.key, required this.controller});
  final AppController controller;
  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Archive settings')),
    body: ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Card(
          child: ListTile(
            leading: const Icon(Icons.dns_outlined),
            title: const Text('Connected server'),
            subtitle: Text(controller.serverUrl ?? ''),
            trailing: const Icon(Icons.lock_outline),
          ),
        ),
        const SizedBox(height: 12),
        Card(
          child: Column(
            children: [
              ListTile(
                leading: const Icon(Icons.swap_horiz),
                title: const Text('Change server'),
                subtitle: const Text(
                  'Disconnect and enter a different self-hosted address.',
                ),
                onTap: () async {
                  await controller.changeServer();
                  if (context.mounted) Navigator.of(context).pop();
                },
              ),
              const Divider(height: 1),
              ListTile(
                leading: const Icon(Icons.logout),
                title: const Text('Sign out'),
                onTap: () async {
                  await controller.signOut();
                  if (context.mounted) Navigator.of(context).pop();
                },
              ),
            ],
          ),
        ),
      ],
    ),
  );
}
