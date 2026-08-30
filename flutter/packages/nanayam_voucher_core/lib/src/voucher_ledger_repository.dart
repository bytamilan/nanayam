import 'dart:convert';
import 'dart:math';

import 'package:nanayam_ledger_client/nanayam_ledger_client.dart';

import 'voucher_exceptions.dart';
import 'voucher_models.dart';

const String _voucherPrefix = 'VCH-';
const String _redemptionPrefix = 'TXN-';

/// Maps voucher provisioning and redemption onto the Nanayam gateway's
/// generic asset ledger (`createAsset` / `queryAsset` / `listAssets`,
/// backed by the `asset-transfer-basic` sample chaincode).
///
/// ## Why a generic ledger, and not a dedicated `Voucher` chaincode?
///
/// The main Nanayam repo ships one sample chaincode (`asset-transfer-basic`,
/// fetched at setup time — it does not live in this repo) plus one bespoke
/// one (the anti-corruption complaint workflow, wired through its own gRPC
/// methods in `services/gateway`). Adding a third, voucher-specific
/// chaincode would mean designing and deploying new Fabric chaincode, which
/// is out of scope for a Flutter example. Instead, this repository shows how
/// **any** domain can be layered on top of the ledger's existing generic
/// asset primitives — a pattern worth knowing even when a dedicated
/// chaincode *is** available. This is explicitly a sample/demo mapping, not
/// a production data model.
///
/// ## Ledger encoding
///
/// Every write is a brand-new asset — there is no `UpdateAsset` RPC exposed
/// by the gateway, so this repository never mutates a ledger entry, only
/// appends new ones (an event-sourcing style that also happens to fit an
/// immutable ledger far better than in-place updates would).
///
/// | Event      | assetId                    | owner        | color (JSON)                                             | size          | appraisedValue        |
/// |------------|-----------------------------|--------------|-----------------------------------------------------------|---------------|------------------------|
/// | Provision  | `VCH-<code>`                | holder id    | `{"kind":"issue","category":...,"program":...,"issuedAt":...,"expiresAt":...}` | face value (cents) | face value (cents) |
/// | Redemption | `TXN-<code>-<unique suffix>`| merchant id  | `{"kind":"redeem","voucherCode":<code>}`                   | amount redeemed (cents) | balance remaining after this redemption (cents) |
///
/// A voucher's current balance and status are never stored directly; they
/// are always recomputed by folding a voucher's issuance record with every
/// redemption record that references it (see [Voucher.remainingCents] and
/// [Voucher.status]).
///
/// ## Known limitations (this is a demo)
///
/// - **No atomicity.** Two redemptions submitted concurrently against the
///   same voucher can both read the same "remaining balance" and both
///   succeed, over-redeeming the voucher. A production system would enforce
///   this in chaincode (e.g. a real `RedeemVoucher` transaction that reads
///   and writes the balance atomically), the same way `UpdateComplaint`
///   enforces the complaint workflow's state machine today.
/// - **O(n) discovery.** Finding "every voucher for holder X" or "every
///   redemption by merchant Y" lists every asset on the ledger and inspects
///   each one, because the gateway has no query-by-owner index. Fine for a
///   demo with a handful of vouchers; a real deployment would add a CouchDB
///   rich query or a dedicated index.
class VoucherLedgerRepository {
  VoucherLedgerRepository(this._client);

  final NanayamLedgerClient _client;

  static final Random _random = Random();

  String _voucherAssetId(String code) => '$_voucherPrefix$code';

  String _newRedemptionAssetId(String code) {
    final now = DateTime.now().toUtc().microsecondsSinceEpoch;
    final salt = _random.nextInt(0xFFFF).toRadixString(16).padLeft(4, '0');
    return '$_redemptionPrefix$code-$now-$salt';
  }

  /// Issues a new voucher of [faceValueCents] to [holderId], recorded as a
  /// single ledger asset.
  ///
  /// Throws [VoucherAlreadyExistsException] if [code] is already on the
  /// ledger.
  Future<Voucher> provisionVoucher({
    required String code,
    required String holderId,
    required String category,
    required String program,
    required int faceValueCents,
    required DateTime expiresAt,
  }) async {
    final issuedAt = DateTime.now().toUtc();
    final meta = <String, dynamic>{
      'kind': 'issue',
      'category': category,
      'program': program,
      'issuedAt': issuedAt.toIso8601String(),
      'expiresAt': expiresAt.toUtc().toIso8601String(),
    };

    final result = await _client.createAsset(
      assetId: _voucherAssetId(code),
      color: jsonEncode(meta),
      size: faceValueCents,
      owner: holderId,
      appraisedValue: faceValueCents,
    );

    if (!result.success) {
      final error = result.error ?? '';
      if (error.toLowerCase().contains('already exists')) {
        throw VoucherAlreadyExistsException(code);
      }
      throw VoucherLedgerRejectedException(error);
    }

    return Voucher(
      code: code,
      holderId: holderId,
      category: category,
      program: program,
      faceValueCents: faceValueCents,
      issuedAt: issuedAt,
      expiresAt: expiresAt,
    );
  }

  /// Reads a voucher and its full redemption history from the ledger.
  ///
  /// Throws [VoucherNotFoundException] if [code] does not exist.
  Future<Voucher> getVoucher(String code) async {
    final LedgerAsset issuance;
    try {
      issuance = await _client.queryAsset(_voucherAssetId(code));
    } on LedgerApiException catch (e) {
      if (e.isNotFound) throw VoucherNotFoundException(code);
      rethrow;
    }

    final meta = _decodeMeta(issuance.color);
    if (meta == null || meta['kind'] != 'issue') {
      throw VoucherNotFoundException(code);
    }

    final redemptions = await _redemptionsFor(code);

    return Voucher(
      code: code,
      holderId: issuance.owner,
      category: meta['category'] as String? ?? '',
      program: meta['program'] as String? ?? '',
      faceValueCents: issuance.appraisedValue,
      issuedAt: DateTime.tryParse(meta['issuedAt'] as String? ?? '') ??
          DateTime.now().toUtc(),
      expiresAt: DateTime.tryParse(meta['expiresAt'] as String? ?? '') ??
          DateTime.now().toUtc(),
      redemptions: redemptions,
    );
  }

  /// Lists every voucher provisioned to [holderId], each with its full
  /// redemption history attached.
  ///
  /// See the class doc for the `O(n)` caveat — this walks every asset on the
  /// ledger, so it is meant for demo-scale data, not production listings.
  Future<List<Voucher>> listVouchersForHolder(String holderId) async {
    final assets = await _client.listAssets();
    final codes = <String>[];
    for (final asset in assets) {
      final meta = _decodeMeta(asset.color);
      if (meta == null) continue;
      if (meta['kind'] == 'issue' &&
          asset.owner == holderId &&
          asset.id.startsWith(_voucherPrefix)) {
        codes.add(asset.id.substring(_voucherPrefix.length));
      }
    }

    final vouchers = <Voucher>[];
    for (final code in codes) {
      vouchers.add(await getVoucher(code));
    }
    return vouchers;
  }

  /// Redeems [amountCents] of voucher [code] on behalf of [merchantId].
  ///
  /// Throws [VoucherNotFoundException], [VoucherExpiredException], or
  /// [InsufficientVoucherBalanceException] as appropriate; see the class doc
  /// for why this check-then-write is not atomic under concurrent
  /// redemptions.
  Future<VoucherRedemption> redeemVoucher({
    required String code,
    required String merchantId,
    required int amountCents,
  }) async {
    if (amountCents <= 0) {
      throw ArgumentError.value(
        amountCents,
        'amountCents',
        'must be greater than zero',
      );
    }

    final voucher = await getVoucher(code);

    if (voucher.isExpired) {
      throw VoucherExpiredException(code, voucher.expiresAt);
    }
    if (amountCents > voucher.remainingCents) {
      throw InsufficientVoucherBalanceException(
        code: code,
        requestedCents: amountCents,
        remainingCents: voucher.remainingCents,
      );
    }

    final remainingAfter = voucher.remainingCents - amountCents;
    final assetId = _newRedemptionAssetId(code);
    final redeemedAt = DateTime.now().toUtc();

    final result = await _client.createAsset(
      assetId: assetId,
      color: jsonEncode(<String, dynamic>{
        'kind': 'redeem',
        'voucherCode': code,
      }),
      size: amountCents,
      owner: merchantId,
      appraisedValue: remainingAfter,
    );

    if (!result.success) {
      throw VoucherLedgerRejectedException(result.error ?? '');
    }

    return VoucherRedemption(
      transactionId: assetId,
      voucherCode: code,
      merchantId: merchantId,
      amountCents: amountCents,
      remainingAfterCents: remainingAfter,
      redeemedAt: redeemedAt,
    );
  }

  /// Lists every redemption made by [merchantId], most recent first.
  Future<List<VoucherRedemption>> listMerchantRedemptions(
    String merchantId,
  ) async {
    final assets = await _client.listAssets();
    final redemptions = <VoucherRedemption>[];
    for (final asset in assets) {
      final meta = _decodeMeta(asset.color);
      if (meta == null) continue;
      if (meta['kind'] == 'redeem' && asset.owner == merchantId) {
        redemptions.add(_redemptionFromAsset(asset, meta));
      }
    }
    redemptions.sort((a, b) => b.redeemedAt.compareTo(a.redeemedAt));
    return redemptions;
  }

  Future<List<VoucherRedemption>> _redemptionsFor(String code) async {
    final ids = await _client.listAssetIds();
    final prefix = '$_redemptionPrefix$code-';
    final redemptions = <VoucherRedemption>[];
    for (final id in ids) {
      if (!id.startsWith(prefix)) continue;
      final asset = await _client.queryAsset(id);
      final meta = _decodeMeta(asset.color);
      if (meta == null || meta['kind'] != 'redeem') continue;
      redemptions.add(_redemptionFromAsset(asset, meta));
    }
    redemptions.sort((a, b) => a.redeemedAt.compareTo(b.redeemedAt));
    return redemptions;
  }

  VoucherRedemption _redemptionFromAsset(
    LedgerAsset asset,
    Map<String, dynamic> meta,
  ) {
    return VoucherRedemption(
      transactionId: asset.id,
      voucherCode: meta['voucherCode'] as String? ?? '',
      merchantId: asset.owner,
      amountCents: asset.size,
      remainingAfterCents: asset.appraisedValue,
      redeemedAt: _timestampFromRedemptionAssetId(asset.id),
    );
  }

  /// The redemption asset ID embeds a microsecond epoch timestamp (see
  /// [_newRedemptionAssetId]); this recovers it without needing a separate
  /// timestamp field on the ledger record.
  DateTime _timestampFromRedemptionAssetId(String assetId) {
    final parts = assetId.split('-');
    if (parts.length >= 3) {
      final micros = int.tryParse(parts[parts.length - 2]);
      if (micros != null) {
        return DateTime.fromMicrosecondsSinceEpoch(micros, isUtc: true);
      }
    }
    return DateTime.now().toUtc();
  }

  Map<String, dynamic>? _decodeMeta(String color) {
    try {
      final decoded = jsonDecode(color);
      if (decoded is Map<String, dynamic>) return decoded;
      return null;
    } catch (_) {
      // Not our JSON encoding — likely a plain sample asset (e.g. seeded
      // "blue"/"red" demo data from asset-transfer-basic's InitLedger).
      return null;
    }
  }
}
