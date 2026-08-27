/// A generic ledger asset, matching the shape produced by the
/// `asset-transfer-basic` sample chaincode (`ID`, `Color`, `Size`, `Owner`,
/// `AppraisedValue`) that the Nanayam gateway's `CreateAsset` / `QueryAsset` /
/// `ListAssets` RPCs operate on.
///
/// Higher-level domain packages (for example `nanayam_voucher_core`) build
/// richer models on top of this generic envelope rather than requiring a
/// bespoke chaincode + gateway RPC per use case — see that package's
/// `VoucherLedgerRepository` for the pattern.
class LedgerAsset {
  const LedgerAsset({
    required this.id,
    required this.color,
    required this.size,
    required this.owner,
    required this.appraisedValue,
  });

  /// Accepts the common key-casing variants the gateway/chaincode may emit
  /// (`ID`/`id`/`assetId`, `AppraisedValue`/`appraisedValue`, ...).
  factory LedgerAsset.fromJson(Map<String, dynamic> json) {
    return LedgerAsset(
      id: (json['ID'] ?? json['id'] ?? json['assetId']) as String? ?? '',
      color: (json['Color'] ?? json['color']) as String? ?? '',
      size: _asInt(json['Size'] ?? json['size']),
      owner: (json['Owner'] ?? json['owner']) as String? ?? '',
      appraisedValue:
          _asInt(json['AppraisedValue'] ?? json['appraisedValue']),
    );
  }

  final String id;
  final String color;
  final int size;
  final String owner;
  final int appraisedValue;

  Map<String, dynamic> toJson() => <String, dynamic>{
        'assetId': id,
        'color': color,
        'size': size,
        'owner': owner,
        'appraisedValue': appraisedValue,
      };
}

int _asInt(Object? value) {
  if (value is int) return value;
  if (value is num) return value.toInt();
  if (value is String) return int.tryParse(value) ?? 0;
  return 0;
}

/// Response of `POST /v1/CreateAsset`.
class CreateAssetResult {
  const CreateAssetResult({required this.success, this.error});

  factory CreateAssetResult.fromJson(Map<String, dynamic> json) {
    return CreateAssetResult(
      success: json['success'] as bool? ?? false,
      error: json['error'] as String?,
    );
  }

  final bool success;
  final String? error;
}
