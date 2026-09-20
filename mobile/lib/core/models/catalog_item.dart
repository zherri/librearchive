class CatalogItem {
  const CatalogItem({required this.id, required this.name});
  final int id;
  final String name;
  factory CatalogItem.fromJson(Map<String, dynamic> json) =>
      CatalogItem(id: json['id'] as int, name: json['name'] as String);
}
