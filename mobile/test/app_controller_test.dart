import 'package:flutter_test/flutter_test.dart';
import 'package:librearchive/core/app_controller.dart';

void main() {
  test(
    'normalizes a self-hosted server URL',
    () => expect(
      normalizeServerUrl(' https://books.example.test/ '),
      'https://books.example.test',
    ),
  );
  test(
    'rejects a URL containing an API path',
    () => expect(
      () => normalizeServerUrl('https://books.example.test/api/v1'),
      throwsA(isA<Exception>()),
    ),
  );
}
