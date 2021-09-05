///
//  Generated code. Do not modify.
//  source: osmosis/lockup/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../google/protobuf/duration.pb.dart' as $2;
import '../../cosmos/base/v1beta1/coin.pb.dart' as $4;
import 'lock.pb.dart' as $5;

class MsgLockTokens extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgLockTokens', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..aOM<$2.Duration>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'duration', subBuilder: $2.Duration.create)
    ..pc<$4.Coin>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'coins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..hasRequiredFields = false
  ;

  MsgLockTokens._() : super();
  factory MsgLockTokens() => create();
  factory MsgLockTokens.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgLockTokens.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgLockTokens clone() => MsgLockTokens()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgLockTokens copyWith(void Function(MsgLockTokens) updates) => super.copyWith((message) => updates(message as MsgLockTokens)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgLockTokens create() => MsgLockTokens._();
  MsgLockTokens createEmptyInstance() => create();
  static $pb.PbList<MsgLockTokens> createRepeated() => $pb.PbList<MsgLockTokens>();
  @$core.pragma('dart2js:noInline')
  static MsgLockTokens getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgLockTokens>(create);
  static MsgLockTokens _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);

  @$pb.TagNumber(2)
  $2.Duration get duration => $_getN(1);
  @$pb.TagNumber(2)
  set duration($2.Duration v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasDuration() => $_has(1);
  @$pb.TagNumber(2)
  void clearDuration() => clearField(2);
  @$pb.TagNumber(2)
  $2.Duration ensureDuration() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.List<$4.Coin> get coins => $_getList(2);
}

class MsgLockTokensResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgLockTokensResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'ID', $pb.PbFieldType.OU6, protoName: 'ID', defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  MsgLockTokensResponse._() : super();
  factory MsgLockTokensResponse() => create();
  factory MsgLockTokensResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgLockTokensResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgLockTokensResponse clone() => MsgLockTokensResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgLockTokensResponse copyWith(void Function(MsgLockTokensResponse) updates) => super.copyWith((message) => updates(message as MsgLockTokensResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgLockTokensResponse create() => MsgLockTokensResponse._();
  MsgLockTokensResponse createEmptyInstance() => create();
  static $pb.PbList<MsgLockTokensResponse> createRepeated() => $pb.PbList<MsgLockTokensResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgLockTokensResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgLockTokensResponse>(create);
  static MsgLockTokensResponse _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get iD => $_getI64(0);
  @$pb.TagNumber(1)
  set iD($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasID() => $_has(0);
  @$pb.TagNumber(1)
  void clearID() => clearField(1);
}

class MsgBeginUnlockingAll extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgBeginUnlockingAll', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..hasRequiredFields = false
  ;

  MsgBeginUnlockingAll._() : super();
  factory MsgBeginUnlockingAll() => create();
  factory MsgBeginUnlockingAll.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgBeginUnlockingAll.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgBeginUnlockingAll clone() => MsgBeginUnlockingAll()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgBeginUnlockingAll copyWith(void Function(MsgBeginUnlockingAll) updates) => super.copyWith((message) => updates(message as MsgBeginUnlockingAll)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgBeginUnlockingAll create() => MsgBeginUnlockingAll._();
  MsgBeginUnlockingAll createEmptyInstance() => create();
  static $pb.PbList<MsgBeginUnlockingAll> createRepeated() => $pb.PbList<MsgBeginUnlockingAll>();
  @$core.pragma('dart2js:noInline')
  static MsgBeginUnlockingAll getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgBeginUnlockingAll>(create);
  static MsgBeginUnlockingAll _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);
}

class MsgBeginUnlockingAllResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgBeginUnlockingAllResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$5.PeriodLock>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'unlocks', $pb.PbFieldType.PM, subBuilder: $5.PeriodLock.create)
    ..hasRequiredFields = false
  ;

  MsgBeginUnlockingAllResponse._() : super();
  factory MsgBeginUnlockingAllResponse() => create();
  factory MsgBeginUnlockingAllResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgBeginUnlockingAllResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgBeginUnlockingAllResponse clone() => MsgBeginUnlockingAllResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgBeginUnlockingAllResponse copyWith(void Function(MsgBeginUnlockingAllResponse) updates) => super.copyWith((message) => updates(message as MsgBeginUnlockingAllResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgBeginUnlockingAllResponse create() => MsgBeginUnlockingAllResponse._();
  MsgBeginUnlockingAllResponse createEmptyInstance() => create();
  static $pb.PbList<MsgBeginUnlockingAllResponse> createRepeated() => $pb.PbList<MsgBeginUnlockingAllResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgBeginUnlockingAllResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgBeginUnlockingAllResponse>(create);
  static MsgBeginUnlockingAllResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$5.PeriodLock> get unlocks => $_getList(0);
}

class MsgBeginUnlocking extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgBeginUnlocking', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'ID', $pb.PbFieldType.OU6, protoName: 'ID', defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  MsgBeginUnlocking._() : super();
  factory MsgBeginUnlocking() => create();
  factory MsgBeginUnlocking.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgBeginUnlocking.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgBeginUnlocking clone() => MsgBeginUnlocking()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgBeginUnlocking copyWith(void Function(MsgBeginUnlocking) updates) => super.copyWith((message) => updates(message as MsgBeginUnlocking)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgBeginUnlocking create() => MsgBeginUnlocking._();
  MsgBeginUnlocking createEmptyInstance() => create();
  static $pb.PbList<MsgBeginUnlocking> createRepeated() => $pb.PbList<MsgBeginUnlocking>();
  @$core.pragma('dart2js:noInline')
  static MsgBeginUnlocking getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgBeginUnlocking>(create);
  static MsgBeginUnlocking _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get iD => $_getI64(1);
  @$pb.TagNumber(2)
  set iD($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasID() => $_has(1);
  @$pb.TagNumber(2)
  void clearID() => clearField(2);
}

class MsgBeginUnlockingResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgBeginUnlockingResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'success')
    ..hasRequiredFields = false
  ;

  MsgBeginUnlockingResponse._() : super();
  factory MsgBeginUnlockingResponse() => create();
  factory MsgBeginUnlockingResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgBeginUnlockingResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgBeginUnlockingResponse clone() => MsgBeginUnlockingResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgBeginUnlockingResponse copyWith(void Function(MsgBeginUnlockingResponse) updates) => super.copyWith((message) => updates(message as MsgBeginUnlockingResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgBeginUnlockingResponse create() => MsgBeginUnlockingResponse._();
  MsgBeginUnlockingResponse createEmptyInstance() => create();
  static $pb.PbList<MsgBeginUnlockingResponse> createRepeated() => $pb.PbList<MsgBeginUnlockingResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgBeginUnlockingResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgBeginUnlockingResponse>(create);
  static MsgBeginUnlockingResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get success => $_getBF(0);
  @$pb.TagNumber(1)
  set success($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSuccess() => $_has(0);
  @$pb.TagNumber(1)
  void clearSuccess() => clearField(1);
}

