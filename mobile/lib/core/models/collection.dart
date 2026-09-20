import 'book.dart';

class LibraryCollection {
  const LibraryCollection({required this.id, required this.name});

  final int id;
  final String name;

  factory LibraryCollection.fromJson(Map<String, dynamic> json) =>
      LibraryCollection(id: json['id'] as int, name: json['name'] as String);
}

class CollectionDetails {
  const CollectionDetails({
    required this.collection,
    required this.books,
    required this.page,
    required this.pageSize,
    required this.total,
  });

  final LibraryCollection collection;
  final List<Book> books;
  final int page;
  final int pageSize;
  final int total;

  bool get hasNextPage => page * pageSize < total;

  factory CollectionDetails.fromJson(Map<String, dynamic> json) {
    final items = (json['items'] as List? ?? []).cast<Map<String, dynamic>>();
    return CollectionDetails(
      collection: LibraryCollection.fromJson(
        json['collection'] as Map<String, dynamic>,
      ),
      books: items
          .map((item) => Book.fromJson(item['book'] as Map<String, dynamic>))
          .toList(),
      page:
          ((json['offset'] as int? ?? 0) ~/ (json['limit'] as int? ?? 20)) + 1,
      pageSize: json['limit'] as int? ?? 20,
      total: json['total'] as int? ?? items.length,
    );
  }
}
