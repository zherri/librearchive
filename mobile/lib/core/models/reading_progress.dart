import 'book.dart';

class ReadingProgress {
  const ReadingProgress({
    required this.bookId,
    required this.status,
    required this.currentPage,
    required this.progressPercent,
    required this.isDownloaded,
    this.book,
  });

  final int bookId;
  final String status;
  final int currentPage;
  final double progressPercent;
  final bool isDownloaded;
  final Book? book;

  factory ReadingProgress.fromJson(Map<String, dynamic> json) =>
      ReadingProgress(
        bookId: json['bookId'] as int,
        status: json['status'] as String,
        currentPage: json['currentPage'] as int? ?? 0,
        progressPercent: (json['progressPercent'] as num? ?? 0).toDouble(),
        isDownloaded: json['isDownloaded'] as bool? ?? false,
        book: json['book'] is Map<String, dynamic>
            ? Book.fromJson(json['book'] as Map<String, dynamic>)
            : null,
      );
}
