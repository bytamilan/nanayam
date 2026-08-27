/// A console/gateway user account, as returned by `/v1/Register` and `/v1/Me`.
///
/// This is separate from the underlying Fabric identity (MSP cert) — it is
/// the lightweight application-level account the gateway's [AuthStore] uses
/// to issue JWTs.
class LedgerUser {
  const LedgerUser({
    required this.id,
    required this.username,
    required this.org,
    required this.role,
    required this.createdAt,
  });

  factory LedgerUser.fromJson(Map<String, dynamic> json) {
    return LedgerUser(
      id: json['id'] as String? ?? '',
      username: json['username'] as String? ?? '',
      org: json['org'] as String? ?? '',
      role: json['role'] as String? ?? 'user',
      createdAt: DateTime.tryParse(json['createdAt'] as String? ?? '') ??
          DateTime.fromMillisecondsSinceEpoch(0, isUtc: true),
    );
  }

  final String id;
  final String username;
  final String org;
  final String role;
  final DateTime createdAt;

  bool get isAdmin => role == 'admin';

  Map<String, dynamic> toJson() => <String, dynamic>{
        'id': id,
        'username': username,
        'org': org,
        'role': role,
        'createdAt': createdAt.toIso8601String(),
      };
}

/// Result of a successful `/v1/Login` call: a bearer token to send as
/// `Authorization: Bearer <token>` on every subsequent request.
class LedgerSession {
  const LedgerSession({required this.token});

  factory LedgerSession.fromJson(Map<String, dynamic> json) {
    return LedgerSession(token: json['token'] as String? ?? '');
  }

  final String token;
}

/// Public gateway configuration returned by `/v1/Config` (no auth required).
/// Useful for an app's splash/login screen to know whether self-registration
/// is available before rendering a "sign up" affordance.
class GatewayConfig {
  const GatewayConfig({
    required this.signupEnabled,
    required this.channel,
    required this.chaincode,
  });

  factory GatewayConfig.fromJson(Map<String, dynamic> json) {
    return GatewayConfig(
      signupEnabled: json['signupEnabled'] as bool? ?? false,
      channel: json['channel'] as String? ?? '',
      chaincode: json['chaincode'] as String? ?? '',
    );
  }

  final bool signupEnabled;
  final String channel;
  final String chaincode;
}
