class Book {
  const Book({
    required this.id,
    required this.title,
    required this.authors,
    required this.originalName,
    this.description,
    this.publisher,
    this.publishedYear,
    this.pageCount,
  });

  final int id;
  final String title;
  final String authors;
  final String originalName;
  final String? description;
  final String? publisher;
  final int? publishedYear;
  final int? pageCount;

  factory Book.fromJson(Map<String, dynamic> json) => Book(
    id: json['id'] as int,
    title: json['title'] as String,
    authors: json['authors'] as String,
    originalName: json['originalName'] as String? ?? '',
    description: json['description'] as String?,
    publisher: json['publisher'] as String?,
    publishedYear: json['publishedYear'] as int?,
    pageCount: json['pageCount'] as int?,
  );
}
