class ArchiveUser {
  const ArchiveUser({
    required this.id,
    required this.username,
    required this.name,
    required this.role,
    required this.isActive,
  });

  final int id;
  final String username;
  final String name;
  final String role;
  final bool isActive;

  bool get isAdministrator => role == 'admin';

  factory ArchiveUser.fromJson(Map<String, dynamic> json) => ArchiveUser(
    id: json['id'] as int,
    username: json['username'] as String,
    name: json['name'] as String,
    role: json['role'] as String,
    isActive: json['isActive'] as bool,
  );
}

class CreatedUser {
  const CreatedUser({required this.user, required this.passphrase});
  final ArchiveUser user;
  final String passphrase;
}
