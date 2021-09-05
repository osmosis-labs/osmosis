///
//  Generated code. Do not modify.
//  source: cosmos/distribution/v1beta1/genesis.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../base/v1beta1/coin.pb.dart' as $2;
import 'distribution.pb.dart' as $3;

class DelegatorWithdrawInfo extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DelegatorWithdrawInfo', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'withdrawAddress')
    ..hasRequiredFields = false
  ;

  DelegatorWithdrawInfo._() : super();
  factory DelegatorWithdrawInfo() => create();
  factory DelegatorWithdrawInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DelegatorWithdrawInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DelegatorWithdrawInfo clone() => DelegatorWithdrawInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DelegatorWithdrawInfo copyWith(void Function(DelegatorWithdrawInfo) updates) => super.copyWith((message) => updates(message as DelegatorWithdrawInfo)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static DelegatorWithdrawInfo create() => DelegatorWithdrawInfo._();
  DelegatorWithdrawInfo createEmptyInstance() => create();
  static $pb.PbList<DelegatorWithdrawInfo> createRepeated() => $pb.PbList<DelegatorWithdrawInfo>();
  @$core.pragma('dart2js:noInline')
  static DelegatorWithdrawInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DelegatorWithdrawInfo>(create);
  static DelegatorWithdrawInfo _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get delegatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set delegatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get withdrawAddress => $_getSZ(1);
  @$pb.TagNumber(2)
  set withdrawAddress($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasWithdrawAddress() => $_has(1);
  @$pb.TagNumber(2)
  void clearWithdrawAddress() => clearField(2);
}

class ValidatorOutstandingRewardsRecord extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorOutstandingRewardsRecord', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..pc<$2.DecCoin>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'outstandingRewards', $pb.PbFieldType.PM, subBuilder: $2.DecCoin.create)
    ..hasRequiredFields = false
  ;

  ValidatorOutstandingRewardsRecord._() : super();
  factory ValidatorOutstandingRewardsRecord() => create();
  factory ValidatorOutstandingRewardsRecord.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorOutstandingRewardsRecord.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorOutstandingRewardsRecord clone() => ValidatorOutstandingRewardsRecord()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorOutstandingRewardsRecord copyWith(void Function(ValidatorOutstandingRewardsRecord) updates) => super.copyWith((message) => updates(message as ValidatorOutstandingRewardsRecord)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorOutstandingRewardsRecord create() => ValidatorOutstandingRewardsRecord._();
  ValidatorOutstandingRewardsRecord createEmptyInstance() => create();
  static $pb.PbList<ValidatorOutstandingRewardsRecord> createRepeated() => $pb.PbList<ValidatorOutstandingRewardsRecord>();
  @$core.pragma('dart2js:noInline')
  static ValidatorOutstandingRewardsRecord getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorOutstandingRewardsRecord>(create);
  static ValidatorOutstandingRewardsRecord _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get validatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set validatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$2.DecCoin> get outstandingRewards => $_getList(1);
}

class ValidatorAccumulatedCommissionRecord extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorAccumulatedCommissionRecord', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..aOM<$3.ValidatorAccumulatedCommission>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'accumulated', subBuilder: $3.ValidatorAccumulatedCommission.create)
    ..hasRequiredFields = false
  ;

  ValidatorAccumulatedCommissionRecord._() : super();
  factory ValidatorAccumulatedCommissionRecord() => create();
  factory ValidatorAccumulatedCommissionRecord.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorAccumulatedCommissionRecord.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorAccumulatedCommissionRecord clone() => ValidatorAccumulatedCommissionRecord()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorAccumulatedCommissionRecord copyWith(void Function(ValidatorAccumulatedCommissionRecord) updates) => super.copyWith((message) => updates(message as ValidatorAccumulatedCommissionRecord)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorAccumulatedCommissionRecord create() => ValidatorAccumulatedCommissionRecord._();
  ValidatorAccumulatedCommissionRecord createEmptyInstance() => create();
  static $pb.PbList<ValidatorAccumulatedCommissionRecord> createRepeated() => $pb.PbList<ValidatorAccumulatedCommissionRecord>();
  @$core.pragma('dart2js:noInline')
  static ValidatorAccumulatedCommissionRecord getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorAccumulatedCommissionRecord>(create);
  static ValidatorAccumulatedCommissionRecord _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get validatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set validatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $3.ValidatorAccumulatedCommission get accumulated => $_getN(1);
  @$pb.TagNumber(2)
  set accumulated($3.ValidatorAccumulatedCommission v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasAccumulated() => $_has(1);
  @$pb.TagNumber(2)
  void clearAccumulated() => clearField(2);
  @$pb.TagNumber(2)
  $3.ValidatorAccumulatedCommission ensureAccumulated() => $_ensure(1);
}

class ValidatorHistoricalRewardsRecord extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorHistoricalRewardsRecord', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'period', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOM<$3.ValidatorHistoricalRewards>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'rewards', subBuilder: $3.ValidatorHistoricalRewards.create)
    ..hasRequiredFields = false
  ;

  ValidatorHistoricalRewardsRecord._() : super();
  factory ValidatorHistoricalRewardsRecord() => create();
  factory ValidatorHistoricalRewardsRecord.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorHistoricalRewardsRecord.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorHistoricalRewardsRecord clone() => ValidatorHistoricalRewardsRecord()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorHistoricalRewardsRecord copyWith(void Function(ValidatorHistoricalRewardsRecord) updates) => super.copyWith((message) => updates(message as ValidatorHistoricalRewardsRecord)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorHistoricalRewardsRecord create() => ValidatorHistoricalRewardsRecord._();
  ValidatorHistoricalRewardsRecord createEmptyInstance() => create();
  static $pb.PbList<ValidatorHistoricalRewardsRecord> createRepeated() => $pb.PbList<ValidatorHistoricalRewardsRecord>();
  @$core.pragma('dart2js:noInline')
  static ValidatorHistoricalRewardsRecord getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorHistoricalRewardsRecord>(create);
  static ValidatorHistoricalRewardsRecord _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get validatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set validatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get period => $_getI64(1);
  @$pb.TagNumber(2)
  set period($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPeriod() => $_has(1);
  @$pb.TagNumber(2)
  void clearPeriod() => clearField(2);

  @$pb.TagNumber(3)
  $3.ValidatorHistoricalRewards get rewards => $_getN(2);
  @$pb.TagNumber(3)
  set rewards($3.ValidatorHistoricalRewards v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasRewards() => $_has(2);
  @$pb.TagNumber(3)
  void clearRewards() => clearField(3);
  @$pb.TagNumber(3)
  $3.ValidatorHistoricalRewards ensureRewards() => $_ensure(2);
}

class ValidatorCurrentRewardsRecord extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorCurrentRewardsRecord', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..aOM<$3.ValidatorCurrentRewards>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'rewards', subBuilder: $3.ValidatorCurrentRewards.create)
    ..hasRequiredFields = false
  ;

  ValidatorCurrentRewardsRecord._() : super();
  factory ValidatorCurrentRewardsRecord() => create();
  factory ValidatorCurrentRewardsRecord.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorCurrentRewardsRecord.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorCurrentRewardsRecord clone() => ValidatorCurrentRewardsRecord()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorCurrentRewardsRecord copyWith(void Function(ValidatorCurrentRewardsRecord) updates) => super.copyWith((message) => updates(message as ValidatorCurrentRewardsRecord)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorCurrentRewardsRecord create() => ValidatorCurrentRewardsRecord._();
  ValidatorCurrentRewardsRecord createEmptyInstance() => create();
  static $pb.PbList<ValidatorCurrentRewardsRecord> createRepeated() => $pb.PbList<ValidatorCurrentRewardsRecord>();
  @$core.pragma('dart2js:noInline')
  static ValidatorCurrentRewardsRecord getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorCurrentRewardsRecord>(create);
  static ValidatorCurrentRewardsRecord _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get validatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set validatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $3.ValidatorCurrentRewards get rewards => $_getN(1);
  @$pb.TagNumber(2)
  set rewards($3.ValidatorCurrentRewards v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasRewards() => $_has(1);
  @$pb.TagNumber(2)
  void clearRewards() => clearField(2);
  @$pb.TagNumber(2)
  $3.ValidatorCurrentRewards ensureRewards() => $_ensure(1);
}

class DelegatorStartingInfoRecord extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DelegatorStartingInfoRecord', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..aOM<$3.DelegatorStartingInfo>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'startingInfo', subBuilder: $3.DelegatorStartingInfo.create)
    ..hasRequiredFields = false
  ;

  DelegatorStartingInfoRecord._() : super();
  factory DelegatorStartingInfoRecord() => create();
  factory DelegatorStartingInfoRecord.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DelegatorStartingInfoRecord.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DelegatorStartingInfoRecord clone() => DelegatorStartingInfoRecord()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DelegatorStartingInfoRecord copyWith(void Function(DelegatorStartingInfoRecord) updates) => super.copyWith((message) => updates(message as DelegatorStartingInfoRecord)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static DelegatorStartingInfoRecord create() => DelegatorStartingInfoRecord._();
  DelegatorStartingInfoRecord createEmptyInstance() => create();
  static $pb.PbList<DelegatorStartingInfoRecord> createRepeated() => $pb.PbList<DelegatorStartingInfoRecord>();
  @$core.pragma('dart2js:noInline')
  static DelegatorStartingInfoRecord getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DelegatorStartingInfoRecord>(create);
  static DelegatorStartingInfoRecord _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get delegatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set delegatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get validatorAddress => $_getSZ(1);
  @$pb.TagNumber(2)
  set validatorAddress($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasValidatorAddress() => $_has(1);
  @$pb.TagNumber(2)
  void clearValidatorAddress() => clearField(2);

  @$pb.TagNumber(3)
  $3.DelegatorStartingInfo get startingInfo => $_getN(2);
  @$pb.TagNumber(3)
  set startingInfo($3.DelegatorStartingInfo v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasStartingInfo() => $_has(2);
  @$pb.TagNumber(3)
  void clearStartingInfo() => clearField(3);
  @$pb.TagNumber(3)
  $3.DelegatorStartingInfo ensureStartingInfo() => $_ensure(2);
}

class ValidatorSlashEventRecord extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorSlashEventRecord', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'period', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOM<$3.ValidatorSlashEvent>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorSlashEvent', subBuilder: $3.ValidatorSlashEvent.create)
    ..hasRequiredFields = false
  ;

  ValidatorSlashEventRecord._() : super();
  factory ValidatorSlashEventRecord() => create();
  factory ValidatorSlashEventRecord.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorSlashEventRecord.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorSlashEventRecord clone() => ValidatorSlashEventRecord()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorSlashEventRecord copyWith(void Function(ValidatorSlashEventRecord) updates) => super.copyWith((message) => updates(message as ValidatorSlashEventRecord)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorSlashEventRecord create() => ValidatorSlashEventRecord._();
  ValidatorSlashEventRecord createEmptyInstance() => create();
  static $pb.PbList<ValidatorSlashEventRecord> createRepeated() => $pb.PbList<ValidatorSlashEventRecord>();
  @$core.pragma('dart2js:noInline')
  static ValidatorSlashEventRecord getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorSlashEventRecord>(create);
  static ValidatorSlashEventRecord _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get validatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set validatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get height => $_getI64(1);
  @$pb.TagNumber(2)
  set height($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasHeight() => $_has(1);
  @$pb.TagNumber(2)
  void clearHeight() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get period => $_getI64(2);
  @$pb.TagNumber(3)
  set period($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasPeriod() => $_has(2);
  @$pb.TagNumber(3)
  void clearPeriod() => clearField(3);

  @$pb.TagNumber(4)
  $3.ValidatorSlashEvent get validatorSlashEvent => $_getN(3);
  @$pb.TagNumber(4)
  set validatorSlashEvent($3.ValidatorSlashEvent v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasValidatorSlashEvent() => $_has(3);
  @$pb.TagNumber(4)
  void clearValidatorSlashEvent() => clearField(4);
  @$pb.TagNumber(4)
  $3.ValidatorSlashEvent ensureValidatorSlashEvent() => $_ensure(3);
}

class GenesisState extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GenesisState', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOM<$3.Params>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'params', subBuilder: $3.Params.create)
    ..aOM<$3.FeePool>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'feePool', subBuilder: $3.FeePool.create)
    ..pc<DelegatorWithdrawInfo>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorWithdrawInfos', $pb.PbFieldType.PM, subBuilder: DelegatorWithdrawInfo.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'previousProposer')
    ..pc<ValidatorOutstandingRewardsRecord>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'outstandingRewards', $pb.PbFieldType.PM, subBuilder: ValidatorOutstandingRewardsRecord.create)
    ..pc<ValidatorAccumulatedCommissionRecord>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAccumulatedCommissions', $pb.PbFieldType.PM, subBuilder: ValidatorAccumulatedCommissionRecord.create)
    ..pc<ValidatorHistoricalRewardsRecord>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorHistoricalRewards', $pb.PbFieldType.PM, subBuilder: ValidatorHistoricalRewardsRecord.create)
    ..pc<ValidatorCurrentRewardsRecord>(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorCurrentRewards', $pb.PbFieldType.PM, subBuilder: ValidatorCurrentRewardsRecord.create)
    ..pc<DelegatorStartingInfoRecord>(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorStartingInfos', $pb.PbFieldType.PM, subBuilder: DelegatorStartingInfoRecord.create)
    ..pc<ValidatorSlashEventRecord>(10, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorSlashEvents', $pb.PbFieldType.PM, subBuilder: ValidatorSlashEventRecord.create)
    ..hasRequiredFields = false
  ;

  GenesisState._() : super();
  factory GenesisState() => create();
  factory GenesisState.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GenesisState.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GenesisState clone() => GenesisState()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GenesisState copyWith(void Function(GenesisState) updates) => super.copyWith((message) => updates(message as GenesisState)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static GenesisState create() => GenesisState._();
  GenesisState createEmptyInstance() => create();
  static $pb.PbList<GenesisState> createRepeated() => $pb.PbList<GenesisState>();
  @$core.pragma('dart2js:noInline')
  static GenesisState getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GenesisState>(create);
  static GenesisState _defaultInstance;

  @$pb.TagNumber(1)
  $3.Params get params => $_getN(0);
  @$pb.TagNumber(1)
  set params($3.Params v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasParams() => $_has(0);
  @$pb.TagNumber(1)
  void clearParams() => clearField(1);
  @$pb.TagNumber(1)
  $3.Params ensureParams() => $_ensure(0);

  @$pb.TagNumber(2)
  $3.FeePool get feePool => $_getN(1);
  @$pb.TagNumber(2)
  set feePool($3.FeePool v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasFeePool() => $_has(1);
  @$pb.TagNumber(2)
  void clearFeePool() => clearField(2);
  @$pb.TagNumber(2)
  $3.FeePool ensureFeePool() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.List<DelegatorWithdrawInfo> get delegatorWithdrawInfos => $_getList(2);

  @$pb.TagNumber(4)
  $core.String get previousProposer => $_getSZ(3);
  @$pb.TagNumber(4)
  set previousProposer($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasPreviousProposer() => $_has(3);
  @$pb.TagNumber(4)
  void clearPreviousProposer() => clearField(4);

  @$pb.TagNumber(5)
  $core.List<ValidatorOutstandingRewardsRecord> get outstandingRewards => $_getList(4);

  @$pb.TagNumber(6)
  $core.List<ValidatorAccumulatedCommissionRecord> get validatorAccumulatedCommissions => $_getList(5);

  @$pb.TagNumber(7)
  $core.List<ValidatorHistoricalRewardsRecord> get validatorHistoricalRewards => $_getList(6);

  @$pb.TagNumber(8)
  $core.List<ValidatorCurrentRewardsRecord> get validatorCurrentRewards => $_getList(7);

  @$pb.TagNumber(9)
  $core.List<DelegatorStartingInfoRecord> get delegatorStartingInfos => $_getList(8);

  @$pb.TagNumber(10)
  $core.List<ValidatorSlashEventRecord> get validatorSlashEvents => $_getList(9);
}

