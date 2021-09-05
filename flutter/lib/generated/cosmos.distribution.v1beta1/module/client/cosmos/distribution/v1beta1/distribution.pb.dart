///
//  Generated code. Do not modify.
//  source: cosmos/distribution/v1beta1/distribution.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../base/v1beta1/coin.pb.dart' as $2;

class Params extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Params', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'communityTax')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'baseProposerReward')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'bonusProposerReward')
    ..aOB(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'withdrawAddrEnabled')
    ..hasRequiredFields = false
  ;

  Params._() : super();
  factory Params() => create();
  factory Params.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Params.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Params clone() => Params()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Params copyWith(void Function(Params) updates) => super.copyWith((message) => updates(message as Params)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Params create() => Params._();
  Params createEmptyInstance() => create();
  static $pb.PbList<Params> createRepeated() => $pb.PbList<Params>();
  @$core.pragma('dart2js:noInline')
  static Params getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Params>(create);
  static Params _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get communityTax => $_getSZ(0);
  @$pb.TagNumber(1)
  set communityTax($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCommunityTax() => $_has(0);
  @$pb.TagNumber(1)
  void clearCommunityTax() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get baseProposerReward => $_getSZ(1);
  @$pb.TagNumber(2)
  set baseProposerReward($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasBaseProposerReward() => $_has(1);
  @$pb.TagNumber(2)
  void clearBaseProposerReward() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get bonusProposerReward => $_getSZ(2);
  @$pb.TagNumber(3)
  set bonusProposerReward($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasBonusProposerReward() => $_has(2);
  @$pb.TagNumber(3)
  void clearBonusProposerReward() => clearField(3);

  @$pb.TagNumber(4)
  $core.bool get withdrawAddrEnabled => $_getBF(3);
  @$pb.TagNumber(4)
  set withdrawAddrEnabled($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasWithdrawAddrEnabled() => $_has(3);
  @$pb.TagNumber(4)
  void clearWithdrawAddrEnabled() => clearField(4);
}

class ValidatorHistoricalRewards extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorHistoricalRewards', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pc<$2.DecCoin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'cumulativeRewardRatio', $pb.PbFieldType.PM, subBuilder: $2.DecCoin.create)
    ..a<$core.int>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'referenceCount', $pb.PbFieldType.OU3)
    ..hasRequiredFields = false
  ;

  ValidatorHistoricalRewards._() : super();
  factory ValidatorHistoricalRewards() => create();
  factory ValidatorHistoricalRewards.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorHistoricalRewards.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorHistoricalRewards clone() => ValidatorHistoricalRewards()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorHistoricalRewards copyWith(void Function(ValidatorHistoricalRewards) updates) => super.copyWith((message) => updates(message as ValidatorHistoricalRewards)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorHistoricalRewards create() => ValidatorHistoricalRewards._();
  ValidatorHistoricalRewards createEmptyInstance() => create();
  static $pb.PbList<ValidatorHistoricalRewards> createRepeated() => $pb.PbList<ValidatorHistoricalRewards>();
  @$core.pragma('dart2js:noInline')
  static ValidatorHistoricalRewards getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorHistoricalRewards>(create);
  static ValidatorHistoricalRewards _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.DecCoin> get cumulativeRewardRatio => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get referenceCount => $_getIZ(1);
  @$pb.TagNumber(2)
  set referenceCount($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasReferenceCount() => $_has(1);
  @$pb.TagNumber(2)
  void clearReferenceCount() => clearField(2);
}

class ValidatorCurrentRewards extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorCurrentRewards', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pc<$2.DecCoin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'rewards', $pb.PbFieldType.PM, subBuilder: $2.DecCoin.create)
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'period', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  ValidatorCurrentRewards._() : super();
  factory ValidatorCurrentRewards() => create();
  factory ValidatorCurrentRewards.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorCurrentRewards.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorCurrentRewards clone() => ValidatorCurrentRewards()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorCurrentRewards copyWith(void Function(ValidatorCurrentRewards) updates) => super.copyWith((message) => updates(message as ValidatorCurrentRewards)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorCurrentRewards create() => ValidatorCurrentRewards._();
  ValidatorCurrentRewards createEmptyInstance() => create();
  static $pb.PbList<ValidatorCurrentRewards> createRepeated() => $pb.PbList<ValidatorCurrentRewards>();
  @$core.pragma('dart2js:noInline')
  static ValidatorCurrentRewards getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorCurrentRewards>(create);
  static ValidatorCurrentRewards _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.DecCoin> get rewards => $_getList(0);

  @$pb.TagNumber(2)
  $fixnum.Int64 get period => $_getI64(1);
  @$pb.TagNumber(2)
  set period($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPeriod() => $_has(1);
  @$pb.TagNumber(2)
  void clearPeriod() => clearField(2);
}

class ValidatorAccumulatedCommission extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorAccumulatedCommission', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pc<$2.DecCoin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'commission', $pb.PbFieldType.PM, subBuilder: $2.DecCoin.create)
    ..hasRequiredFields = false
  ;

  ValidatorAccumulatedCommission._() : super();
  factory ValidatorAccumulatedCommission() => create();
  factory ValidatorAccumulatedCommission.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorAccumulatedCommission.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorAccumulatedCommission clone() => ValidatorAccumulatedCommission()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorAccumulatedCommission copyWith(void Function(ValidatorAccumulatedCommission) updates) => super.copyWith((message) => updates(message as ValidatorAccumulatedCommission)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorAccumulatedCommission create() => ValidatorAccumulatedCommission._();
  ValidatorAccumulatedCommission createEmptyInstance() => create();
  static $pb.PbList<ValidatorAccumulatedCommission> createRepeated() => $pb.PbList<ValidatorAccumulatedCommission>();
  @$core.pragma('dart2js:noInline')
  static ValidatorAccumulatedCommission getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorAccumulatedCommission>(create);
  static ValidatorAccumulatedCommission _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.DecCoin> get commission => $_getList(0);
}

class ValidatorOutstandingRewards extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorOutstandingRewards', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pc<$2.DecCoin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'rewards', $pb.PbFieldType.PM, subBuilder: $2.DecCoin.create)
    ..hasRequiredFields = false
  ;

  ValidatorOutstandingRewards._() : super();
  factory ValidatorOutstandingRewards() => create();
  factory ValidatorOutstandingRewards.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorOutstandingRewards.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorOutstandingRewards clone() => ValidatorOutstandingRewards()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorOutstandingRewards copyWith(void Function(ValidatorOutstandingRewards) updates) => super.copyWith((message) => updates(message as ValidatorOutstandingRewards)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorOutstandingRewards create() => ValidatorOutstandingRewards._();
  ValidatorOutstandingRewards createEmptyInstance() => create();
  static $pb.PbList<ValidatorOutstandingRewards> createRepeated() => $pb.PbList<ValidatorOutstandingRewards>();
  @$core.pragma('dart2js:noInline')
  static ValidatorOutstandingRewards getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorOutstandingRewards>(create);
  static ValidatorOutstandingRewards _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.DecCoin> get rewards => $_getList(0);
}

class ValidatorSlashEvent extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorSlashEvent', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorPeriod', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'fraction')
    ..hasRequiredFields = false
  ;

  ValidatorSlashEvent._() : super();
  factory ValidatorSlashEvent() => create();
  factory ValidatorSlashEvent.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorSlashEvent.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorSlashEvent clone() => ValidatorSlashEvent()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorSlashEvent copyWith(void Function(ValidatorSlashEvent) updates) => super.copyWith((message) => updates(message as ValidatorSlashEvent)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorSlashEvent create() => ValidatorSlashEvent._();
  ValidatorSlashEvent createEmptyInstance() => create();
  static $pb.PbList<ValidatorSlashEvent> createRepeated() => $pb.PbList<ValidatorSlashEvent>();
  @$core.pragma('dart2js:noInline')
  static ValidatorSlashEvent getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorSlashEvent>(create);
  static ValidatorSlashEvent _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get validatorPeriod => $_getI64(0);
  @$pb.TagNumber(1)
  set validatorPeriod($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorPeriod() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorPeriod() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get fraction => $_getSZ(1);
  @$pb.TagNumber(2)
  set fraction($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasFraction() => $_has(1);
  @$pb.TagNumber(2)
  void clearFraction() => clearField(2);
}

class ValidatorSlashEvents extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorSlashEvents', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pc<ValidatorSlashEvent>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorSlashEvents', $pb.PbFieldType.PM, subBuilder: ValidatorSlashEvent.create)
    ..hasRequiredFields = false
  ;

  ValidatorSlashEvents._() : super();
  factory ValidatorSlashEvents() => create();
  factory ValidatorSlashEvents.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorSlashEvents.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorSlashEvents clone() => ValidatorSlashEvents()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorSlashEvents copyWith(void Function(ValidatorSlashEvents) updates) => super.copyWith((message) => updates(message as ValidatorSlashEvents)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorSlashEvents create() => ValidatorSlashEvents._();
  ValidatorSlashEvents createEmptyInstance() => create();
  static $pb.PbList<ValidatorSlashEvents> createRepeated() => $pb.PbList<ValidatorSlashEvents>();
  @$core.pragma('dart2js:noInline')
  static ValidatorSlashEvents getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorSlashEvents>(create);
  static ValidatorSlashEvents _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<ValidatorSlashEvent> get validatorSlashEvents => $_getList(0);
}

class FeePool extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'FeePool', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pc<$2.DecCoin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'communityPool', $pb.PbFieldType.PM, subBuilder: $2.DecCoin.create)
    ..hasRequiredFields = false
  ;

  FeePool._() : super();
  factory FeePool() => create();
  factory FeePool.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory FeePool.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  FeePool clone() => FeePool()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  FeePool copyWith(void Function(FeePool) updates) => super.copyWith((message) => updates(message as FeePool)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static FeePool create() => FeePool._();
  FeePool createEmptyInstance() => create();
  static $pb.PbList<FeePool> createRepeated() => $pb.PbList<FeePool>();
  @$core.pragma('dart2js:noInline')
  static FeePool getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<FeePool>(create);
  static FeePool _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.DecCoin> get communityPool => $_getList(0);
}

class CommunityPoolSpendProposal extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CommunityPoolSpendProposal', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'title')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'description')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'recipient')
    ..pc<$2.Coin>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', $pb.PbFieldType.PM, subBuilder: $2.Coin.create)
    ..hasRequiredFields = false
  ;

  CommunityPoolSpendProposal._() : super();
  factory CommunityPoolSpendProposal() => create();
  factory CommunityPoolSpendProposal.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CommunityPoolSpendProposal.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CommunityPoolSpendProposal clone() => CommunityPoolSpendProposal()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CommunityPoolSpendProposal copyWith(void Function(CommunityPoolSpendProposal) updates) => super.copyWith((message) => updates(message as CommunityPoolSpendProposal)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CommunityPoolSpendProposal create() => CommunityPoolSpendProposal._();
  CommunityPoolSpendProposal createEmptyInstance() => create();
  static $pb.PbList<CommunityPoolSpendProposal> createRepeated() => $pb.PbList<CommunityPoolSpendProposal>();
  @$core.pragma('dart2js:noInline')
  static CommunityPoolSpendProposal getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CommunityPoolSpendProposal>(create);
  static CommunityPoolSpendProposal _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get title => $_getSZ(0);
  @$pb.TagNumber(1)
  set title($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTitle() => $_has(0);
  @$pb.TagNumber(1)
  void clearTitle() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get recipient => $_getSZ(2);
  @$pb.TagNumber(3)
  set recipient($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasRecipient() => $_has(2);
  @$pb.TagNumber(3)
  void clearRecipient() => clearField(3);

  @$pb.TagNumber(4)
  $core.List<$2.Coin> get amount => $_getList(3);
}

class DelegatorStartingInfo extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DelegatorStartingInfo', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'previousPeriod', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'stake')
    ..a<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  DelegatorStartingInfo._() : super();
  factory DelegatorStartingInfo() => create();
  factory DelegatorStartingInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DelegatorStartingInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DelegatorStartingInfo clone() => DelegatorStartingInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DelegatorStartingInfo copyWith(void Function(DelegatorStartingInfo) updates) => super.copyWith((message) => updates(message as DelegatorStartingInfo)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static DelegatorStartingInfo create() => DelegatorStartingInfo._();
  DelegatorStartingInfo createEmptyInstance() => create();
  static $pb.PbList<DelegatorStartingInfo> createRepeated() => $pb.PbList<DelegatorStartingInfo>();
  @$core.pragma('dart2js:noInline')
  static DelegatorStartingInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DelegatorStartingInfo>(create);
  static DelegatorStartingInfo _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get previousPeriod => $_getI64(0);
  @$pb.TagNumber(1)
  set previousPeriod($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPreviousPeriod() => $_has(0);
  @$pb.TagNumber(1)
  void clearPreviousPeriod() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get stake => $_getSZ(1);
  @$pb.TagNumber(2)
  set stake($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasStake() => $_has(1);
  @$pb.TagNumber(2)
  void clearStake() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get height => $_getI64(2);
  @$pb.TagNumber(3)
  set height($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearHeight() => clearField(3);
}

class DelegationDelegatorReward extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DelegationDelegatorReward', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..pc<$2.DecCoin>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'reward', $pb.PbFieldType.PM, subBuilder: $2.DecCoin.create)
    ..hasRequiredFields = false
  ;

  DelegationDelegatorReward._() : super();
  factory DelegationDelegatorReward() => create();
  factory DelegationDelegatorReward.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DelegationDelegatorReward.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DelegationDelegatorReward clone() => DelegationDelegatorReward()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DelegationDelegatorReward copyWith(void Function(DelegationDelegatorReward) updates) => super.copyWith((message) => updates(message as DelegationDelegatorReward)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static DelegationDelegatorReward create() => DelegationDelegatorReward._();
  DelegationDelegatorReward createEmptyInstance() => create();
  static $pb.PbList<DelegationDelegatorReward> createRepeated() => $pb.PbList<DelegationDelegatorReward>();
  @$core.pragma('dart2js:noInline')
  static DelegationDelegatorReward getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DelegationDelegatorReward>(create);
  static DelegationDelegatorReward _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get validatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set validatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$2.DecCoin> get reward => $_getList(1);
}

class CommunityPoolSpendProposalWithDeposit extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CommunityPoolSpendProposalWithDeposit', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'title')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'description')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'recipient')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'deposit')
    ..hasRequiredFields = false
  ;

  CommunityPoolSpendProposalWithDeposit._() : super();
  factory CommunityPoolSpendProposalWithDeposit() => create();
  factory CommunityPoolSpendProposalWithDeposit.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CommunityPoolSpendProposalWithDeposit.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CommunityPoolSpendProposalWithDeposit clone() => CommunityPoolSpendProposalWithDeposit()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CommunityPoolSpendProposalWithDeposit copyWith(void Function(CommunityPoolSpendProposalWithDeposit) updates) => super.copyWith((message) => updates(message as CommunityPoolSpendProposalWithDeposit)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CommunityPoolSpendProposalWithDeposit create() => CommunityPoolSpendProposalWithDeposit._();
  CommunityPoolSpendProposalWithDeposit createEmptyInstance() => create();
  static $pb.PbList<CommunityPoolSpendProposalWithDeposit> createRepeated() => $pb.PbList<CommunityPoolSpendProposalWithDeposit>();
  @$core.pragma('dart2js:noInline')
  static CommunityPoolSpendProposalWithDeposit getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CommunityPoolSpendProposalWithDeposit>(create);
  static CommunityPoolSpendProposalWithDeposit _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get title => $_getSZ(0);
  @$pb.TagNumber(1)
  set title($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTitle() => $_has(0);
  @$pb.TagNumber(1)
  void clearTitle() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get recipient => $_getSZ(2);
  @$pb.TagNumber(3)
  set recipient($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasRecipient() => $_has(2);
  @$pb.TagNumber(3)
  void clearRecipient() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get amount => $_getSZ(3);
  @$pb.TagNumber(4)
  set amount($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearAmount() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get deposit => $_getSZ(4);
  @$pb.TagNumber(5)
  set deposit($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasDeposit() => $_has(4);
  @$pb.TagNumber(5)
  void clearDeposit() => clearField(5);
}

