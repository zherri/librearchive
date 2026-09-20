class Highlight {
  const Highlight({
    required this.id,
    required this.selectedText,
    required this.color,
    required this.page,
    this.note,
    this.startOffset,
    this.endOffset,
    this.chapterRef,
  });

  final int id;
  final String selectedText;
  final String color;
  final int page;
  final int? startOffset;
  final int? endOffset;
  final String? chapterRef;
  final Note? note;

  factory Highlight.fromJson(Map<String, dynamic> json) => Highlight(
    id: json['id'] as int,
    selectedText: json['selectedText'] as String,
    color: json['color'] as String,
    page: json['page'] as int,
    startOffset: json['startOffset'] as int?,
    endOffset: json['endOffset'] as int?,
    chapterRef: json['chapterRef'] as String?,
    note: json['note'] is Map<String, dynamic>
        ? Note.fromJson(json['note'] as Map<String, dynamic>)
        : null,
  );
}

class Note {
  const Note({
    required this.id,
    required this.highlightId,
    required this.content,
  });

  final int id;
  final int highlightId;
  final String content;

  factory Note.fromJson(Map<String, dynamic> json) => Note(
    id: json['id'] as int,
    highlightId: json['highlightId'] as int,
    content: json['content'] as String,
  );
}
