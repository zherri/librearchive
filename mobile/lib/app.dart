import 'package:flutter/material.dart';

import 'core/app_controller.dart';
import 'core/theme/archive_theme.dart';
import 'features/auth/login_screen.dart';
import 'features/library/library_screen.dart';
import 'features/setup/setup_screen.dart';

class LibreArchiveApp extends StatelessWidget {
  const LibreArchiveApp({super.key, required this.controller});
  final AppController controller;

  @override
  Widget build(BuildContext context) => AnimatedBuilder(
    animation: controller,
    builder: (context, _) => MaterialApp(
      title: 'LibreArchive',
      debugShowCheckedModeBanner: false,
      theme: ArchiveTheme.dark,
      home: switch (controller.stage) {
        AppStage.setup => SetupScreen(controller: controller),
        AppStage.login => LoginScreen(controller: controller),
        AppStage.library => LibraryScreen(controller: controller),
      },
    ),
  );
}
