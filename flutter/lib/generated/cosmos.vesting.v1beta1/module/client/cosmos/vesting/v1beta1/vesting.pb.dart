///
//  Generated code. Do not modify.
//  source: cosmos/vesting/v1beta1/vesting.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../auth/v1beta1/auth.pb.dart' as $3;
import '../../base/v1beta1/coin.pb.dart' as $1;

class BaseVestingAccount extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'BaseVestingAccount', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.vesting.v1beta1'), createEmptyInstance: create)
    ..aOM<$3.BaseAccount>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'baseAccount', subBuilder: $3.BaseAccount.create)
    ..pc<$1.Coin>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'originalVesting', $pb.PbFieldType.PM, subBuilder: $1.Coin.create)
    ..pc<$1.Coin>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatedFree', $pb.PbFieldType.PM, subBuilder: $1.Coin.create)
    ..pc<$1.Coin>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatedVesting', $pb.PbFieldType.PM, subBuilder: $1.Coin.create)
    ..aInt64(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'endTime')
    ..hasRequiredFields = false
  ;

  BaseVestingAccount._() : super();
  factory BaseVestingAccount() => create();
  factory BaseVestingAccount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory BaseVestingAccount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  BaseVestingAccount clone() => BaseVestingAccount()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  BaseVestingAccount copyWith(void Function(BaseVestingAccount) updates) => super.copyWith((message) => updates(message as BaseVestingAccount)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static BaseVestingAccount create() => BaseVestingAccount._();
  BaseVestingAccount createEmptyInstance() => create();
  static $pb.PbList<BaseVestingAccount> createRepeated() => $pb.PbList<BaseVestingAccount>();
  @$core.pragma('dart2js:noInline')
  static BaseVestingAccount getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<BaseVestingAccount>(create);
  static BaseVestingAccount _defaultInstance;

  @$pb.TagNumber(1)
  $3.BaseAccount get baseAccount => $_getN(0);
  @$pb.TagNumber(1)
  set baseAccount($3.BaseAccount v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasBaseAccount() => $_has(0);
  @$pb.TagNumber(1)
  void clearBaseAccount() => clearField(1);
  @$pb.TagNumber(1)
  $3.BaseAccount ensureBaseAccount() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<$1.Coin> get originalVesting => $_getList(1);

  @$pb.TagNumber(3)
  $core.List<$1.Coin> get delegatedFree => $_getList(2);

  @$pb.TagNumber(4)
  $core.List<$1.Coin> get delegatedVesting => $_getList(3);

  @$pb.TagNumber(5)
  $fixnum.Int64 get endTime => $_getI64(4);
  @$pb.TagNumber(5)
  set endTime($fixnum.Int64 v) { $_setInt64(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasEndTime() => $_has(4);
  @$pb.TagNumber(5)
  void clearEndTime() => clearField(5);
}

class ContinuousVestingAccount extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ContinuousVestingAccount', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.vesting.v1beta1'), createEmptyInstance: create)
    ..aOM<BaseVestingAccount>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'baseVestingAccount', subBuilder: BaseVestingAccount.create)
    ..aInt64(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'startTime')
    ..hasRequiredFields = false
  ;

  ContinuousVestingAccount._() : super();
  factory ContinuousVestingAccount() => create();
  factory ContinuousVestingAccount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ContinuousVestingAccount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ContinuousVestingAccount clone() => ContinuousVestingAccount()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ContinuousVestingAccount copyWith(void Function(ContinuousVestingAccount) updates) => super.copyWith((message) => updates(message as ContinuousVestingAccount)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ContinuousVestingAccount create() => ContinuousVestingAccount._();
  ContinuousVestingAccount createEmptyInstance() => create();
  static $pb.PbList<ContinuousVestingAccount> createRepeated() => $pb.PbList<ContinuousVestingAccount>();
  @$core.pragma('dart2js:noInline')
  static ContinuousVestingAccount getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ContinuousVestingAccount>(create);
  static ContinuousVestingAccount _defaultInstance;

  @$pb.TagNumber(1)
  BaseVestingAccount get baseVestingAccount => $_getN(0);
  @$pb.TagNumber(1)
  set baseVestingAccount(BaseVestingAccount v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasBaseVestingAccount() => $_has(0);
  @$pb.TagNumber(1)
  void clearBaseVestingAccount() => clearField(1);
  @$pb.TagNumber(1)
  BaseVestingAccount ensureBaseVestingAccount() => $_ensure(0);

  @$pb.TagNumber(2)
  $fixnum.Int64 get startTime => $_getI64(1);
  @$pb.TagNumber(2)
  set startTime($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasStartTime() => $_has(1);
  @$pb.TagNumber(2)
  void clearStartTime() => clearField(2);
}

class DelayedVestingAccount extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DelayedVestingAccount', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.vesting.v1beta1'), createEmptyInstance: create)
    ..aOM<BaseVestingAccount>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'baseVestingAccount', subBuilder: BaseVestingAccount.create)
    ..hasRequiredFields = false
  ;

  DelayedVestingAccount._() : super();
  factory DelayedVestingAccount() => create();
  factory DelayedVestingAccount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DelayedVestingAccount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DelayedVestingAccount clone() => DelayedVestingAccount()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DelayedVestingAccount copyWith(void Function(DelayedVestingAccount) updates) => super.copyWith((message) => updates(message as DelayedVestingAccount)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static DelayedVestingAccount create() => DelayedVestingAccount._();
  DelayedVestingAccount createEmptyInstance() => create();
  static $pb.PbList<DelayedVestingAccount> createRepeated() => $pb.PbList<DelayedVestingAccount>();
  @$core.pragma('dart2js:noInline')
  static DelayedVestingAccount getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DelayedVestingAccount>(create);
  static DelayedVestingAccount _defaultInstance;

  @$pb.TagNumber(1)
  BaseVestingAccount get baseVestingAccount => $_getN(0);
  @$pb.TagNumber(1)
  set baseVestingAccount(BaseVestingAccount v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasBaseVestingAccount() => $_has(0);
  @$pb.TagNumber(1)
  void clearBaseVestingAccount() => clearField(1);
  @$pb.TagNumber(1)
  BaseVestingAccount ensureBaseVestingAccount() => $_ensure(0);
}

class Period extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Period', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.vesting.v1beta1'), createEmptyInstance: create)
    ..aInt64(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'length')
    ..pc<$1.Coin>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', $pb.PbFieldType.PM, subBuilder: $1.Coin.create)
    ..hasRequiredFields = false
  ;

  Period._() : super();
  factory Period() => create();
  factory Period.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Period.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Period clone() => Period()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Period copyWith(void Function(Period) updates) => super.copyWith((message) => updates(message as Period)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Period create() => Period._();
  Period createEmptyInstance() => create();
  static $pb.PbList<Period> createRepeated() => $pb.PbList<Period>();
  @$core.pragma('dart2js:noInline')
  static Period getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Period>(create);
  static Period _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get length => $_getI64(0);
  @$pb.TagNumber(1)
  set length($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasLength() => $_has(0);
  @$pb.TagNumber(1)
  void clearLength() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$1.Coin> get amount => $_getList(1);
}

class PeriodicVestingAccount extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'PeriodicVestingAccount', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.vesting.v1beta1'), createEmptyInstance: create)
    ..aOM<BaseVestingAccount>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'baseVestingAccount', subBuilder: BaseVestingAccount.create)
    ..aInt64(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'startTime')
    ..pc<Period>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'vestingPeriods', $pb.PbFieldType.PM, subBuilder: Period.create)
    ..hasRequiredFields = false
  ;

  PeriodicVestingAccount._() : super();
  factory PeriodicVestingAccount() => create();
  factory PeriodicVestingAccount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PeriodicVestingAccount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PeriodicVestingAccount clone() => PeriodicVestingAccount()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PeriodicVestingAccount copyWith(void Function(PeriodicVestingAccount) updates) => super.copyWith((message) => updates(message as PeriodicVestingAccount)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static PeriodicVestingAccount create() => PeriodicVestingAccount._();
  PeriodicVestingAccount createEmptyInstance() => create();
  static $pb.PbList<PeriodicVestingAccount> createRepeated() => $pb.PbList<PeriodicVestingAccount>();
  @$core.pragma('dart2js:noInline')
  static PeriodicVestingAccount getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PeriodicVestingAccount>(create);
  static PeriodicVestingAccount _defaultInstance;

  @$pb.TagNumber(1)
  BaseVestingAccount get baseVestingAccount => $_getN(0);
  @$pb.TagNumber(1)
  set baseVestingAccount(BaseVestingAccount v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasBaseVestingAccount() => $_has(0);
  @$pb.TagNumber(1)
  void clearBaseVestingAccount() => clearField(1);
  @$pb.TagNumber(1)
  BaseVestingAccount ensureBaseVestingAccount() => $_ensure(0);

  @$pb.TagNumber(2)
  $fixnum.Int64 get startTime => $_getI64(1);
  @$pb.TagNumber(2)
  set startTime($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasStartTime() => $_has(1);
  @$pb.TagNumber(2)
  void clearStartTime() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<Period> get vestingPeriods => $_getList(2);
}

