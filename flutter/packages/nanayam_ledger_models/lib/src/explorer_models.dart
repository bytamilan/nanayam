/// One block summary, as returned by `GET /v1/LedgerBlocks`.
class LedgerBlockSummary {
  const LedgerBlockSummary({
    required this.number,
    required this.hash,
    required this.prevHash,
    required this.txCount,
    required this.timestamp,
    required this.dataHash,
  });

  factory LedgerBlockSummary.fromJson(Map<String, dynamic> json) {
    return LedgerBlockSummary(
      number: (json['number'] as num?)?.toInt() ?? 0,
      hash: json['hash'] as String? ?? '',
      prevHash: json['prevHash'] as String? ?? '',
      txCount: (json['txCount'] as num?)?.toInt() ?? 0,
      timestamp: json['timestamp'] as String? ?? '',
      dataHash: json['dataHash'] as String? ?? '',
    );
  }

  final int number;
  final String hash;
  final String prevHash;
  final int txCount;
  final String timestamp;
  final String dataHash;
}

/// Response of `GET /v1/LedgerActivity`: a lightweight liveness/health
/// snapshot suitable for a dashboard widget.
class LedgerActivity {
  const LedgerActivity({
    required this.height,
    required this.complaintCount,
    required this.channel,
    required this.chaincode,
  });

  factory LedgerActivity.fromJson(Map<String, dynamic> json) {
    return LedgerActivity(
      height: (json['height'] as num?)?.toInt() ?? 0,
      complaintCount: (json['complaintCount'] as num?)?.toInt() ?? 0,
      channel: json['channel'] as String? ?? '',
      chaincode: json['chaincode'] as String? ?? '',
    );
  }

  final int height;
  final int complaintCount;
  final String channel;
  final String chaincode;
}

/// A network organization entry within [ChannelInfo.organizations].
class ChannelOrganization {
  const ChannelOrganization({
    required this.mspId,
    required this.name,
    required this.peers,
    required this.role,
  });

  factory ChannelOrganization.fromJson(Map<String, dynamic> json) {
    return ChannelOrganization(
      mspId: json['mspId'] as String? ?? '',
      name: json['name'] as String? ?? '',
      peers: (json['peers'] as List<dynamic>? ?? const <dynamic>[])
          .map((Object? e) => e as String)
          .toList(),
      role: json['role'] as String? ?? '',
    );
  }

  final String mspId;
  final String name;
  final List<String> peers;
  final String role;
}

/// Response of `GET /v1/ChannelInfo`.
class ChannelInfo {
  const ChannelInfo({
    required this.channel,
    required this.chaincode,
    required this.mspId,
    required this.organizations,
    required this.orderers,
  });

  factory ChannelInfo.fromJson(Map<String, dynamic> json) {
    return ChannelInfo(
      channel: json['channel'] as String? ?? '',
      chaincode: json['chaincode'] as String? ?? '',
      mspId: json['mspId'] as String? ?? '',
      organizations: (json['organizations'] as List<dynamic>? ??
              const <dynamic>[])
          .map((Object? e) =>
              ChannelOrganization.fromJson(e! as Map<String, dynamic>))
          .toList(),
      orderers: (json['orderers'] as List<dynamic>? ?? const <dynamic>[])
          .map((Object? e) => e as String)
          .toList(),
    );
  }

  final String channel;
  final String chaincode;
  final String mspId;
  final List<ChannelOrganization> organizations;
  final List<String> orderers;
}
