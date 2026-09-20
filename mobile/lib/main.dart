import 'package:flutter/widgets.dart';

import 'app.dart';
import 'core/app_controller.dart';
import 'core/storage/secure_store.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final controller = AppController(SecureStore());
  await controller.initialize();
  runApp(LibreArchiveApp(controller: controller));
}
