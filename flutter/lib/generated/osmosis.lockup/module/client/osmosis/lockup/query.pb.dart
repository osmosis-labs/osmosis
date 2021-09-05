///
//  Generated code. Do not modify.
//  source: osmosis/lockup/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../cosmos/base/v1beta1/coin.pb.dart' as $4;
import '../../google/protobuf/timestamp.pb.dart' as $3;
import 'lock.pb.dart' as $5;
import '../../google/protobuf/duration.pb.dart' as $2;

class ModuleBalanceRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ModuleBalanceRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  ModuleBalanceRequest._() : super();
  factory ModuleBalanceRequest() => create();
  factory ModuleBalanceRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ModuleBalanceRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ModuleBalanceRequest clone() => ModuleBalanceRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ModuleBalanceRequest copyWith(void Function(ModuleBalanceRequest) updates) => super.copyWith((message) => updates(message as ModuleBalanceRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ModuleBalanceRequest create() => ModuleBalanceRequest._();
  ModuleBalanceRequest createEmptyInstance() => create();
  static $pb.PbList<ModuleBalanceRequest> createRepeated() => $pb.PbList<ModuleBalanceRequest>();
  @$core.pragma('dart2js:noInline')
  static ModuleBalanceRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ModuleBalanceRequest>(create);
  static ModuleBalanceRequest _defaultInstance;
}

class ModuleBalanceResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ModuleBalanceResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$4.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'coins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..hasRequiredFields = false
  ;

  ModuleBalanceResponse._() : super();
  factory ModuleBalanceResponse() => create();
  factory ModuleBalanceResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ModuleBalanceResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ModuleBalanceResponse clone() => ModuleBalanceResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ModuleBalanceResponse copyWith(void Function(ModuleBalanceResponse) updates) => super.copyWith((message) => updates(message as ModuleBalanceResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ModuleBalanceResponse create() => ModuleBalanceResponse._();
  ModuleBalanceResponse createEmptyInstance() => create();
  static $pb.PbList<ModuleBalanceResponse> createRepeated() => $pb.PbList<ModuleBalanceResponse>();
  @$core.pragma('dart2js:noInline')
  static ModuleBalanceResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ModuleBalanceResponse>(create);
  static ModuleBalanceResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.Coin> get coins => $_getList(0);
}

class ModuleLockedAmountRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ModuleLockedAmountRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  ModuleLockedAmountRequest._() : super();
  factory ModuleLockedAmountRequest() => create();
  factory ModuleLockedAmountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ModuleLockedAmountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ModuleLockedAmountRequest clone() => ModuleLockedAmountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ModuleLockedAmountRequest copyWith(void Function(ModuleLockedAmountRequest) updates) => super.copyWith((message) => updates(message as ModuleLockedAmountRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ModuleLockedAmountRequest create() => ModuleLockedAmountRequest._();
  ModuleLockedAmountRequest createEmptyInstance() => create();
  static $pb.PbList<ModuleLockedAmountRequest> createRepeated() => $pb.PbList<ModuleLockedAmountRequest>();
  @$core.pragma('dart2js:noInline')
  static ModuleLockedAmountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ModuleLockedAmountRequest>(create);
  static ModuleLockedAmountRequest _defaultInstance;
}

class ModuleLockedAmountResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ModuleLockedAmountResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$4.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'coins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..hasRequiredFields = false
  ;

  ModuleLockedAmountResponse._() : super();
  factory ModuleLockedAmountResponse() => create();
  factory ModuleLockedAmountResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ModuleLockedAmountResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ModuleLockedAmountResponse clone() => ModuleLockedAmountResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ModuleLockedAmountResponse copyWith(void Function(ModuleLockedAmountResponse) updates) => super.copyWith((message) => updates(message as ModuleLockedAmountResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ModuleLockedAmountResponse create() => ModuleLockedAmountResponse._();
  ModuleLockedAmountResponse createEmptyInstance() => create();
  static $pb.PbList<ModuleLockedAmountResponse> createRepeated() => $pb.PbList<ModuleLockedAmountResponse>();
  @$core.pragma('dart2js:noInline')
  static ModuleLockedAmountResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ModuleLockedAmountResponse>(create);
  static ModuleLockedAmountResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.Coin> get coins => $_getList(0);
}

class AccountUnlockableCoinsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountUnlockableCoinsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..hasRequiredFields = false
  ;

  AccountUnlockableCoinsRequest._() : super();
  factory AccountUnlockableCoinsRequest() => create();
  factory AccountUnlockableCoinsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountUnlockableCoinsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountUnlockableCoinsRequest clone() => AccountUnlockableCoinsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountUnlockableCoinsRequest copyWith(void Function(AccountUnlockableCoinsRequest) updates) => super.copyWith((message) => updates(message as AccountUnlockableCoinsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountUnlockableCoinsRequest create() => AccountUnlockableCoinsRequest._();
  AccountUnlockableCoinsRequest createEmptyInstance() => create();
  static $pb.PbList<AccountUnlockableCoinsRequest> createRepeated() => $pb.PbList<AccountUnlockableCoinsRequest>();
  @$core.pragma('dart2js:noInline')
  static AccountUnlockableCoinsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountUnlockableCoinsRequest>(create);
  static AccountUnlockableCoinsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);
}

class AccountUnlockableCoinsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountUnlockableCoinsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$4.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'coins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..hasRequiredFields = false
  ;

  AccountUnlockableCoinsResponse._() : super();
  factory AccountUnlockableCoinsResponse() => create();
  factory AccountUnlockableCoinsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountUnlockableCoinsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountUnlockableCoinsResponse clone() => AccountUnlockableCoinsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountUnlockableCoinsResponse copyWith(void Function(AccountUnlockableCoinsResponse) updates) => super.copyWith((message) => updates(message as AccountUnlockableCoinsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountUnlockableCoinsResponse create() => AccountUnlockableCoinsResponse._();
  AccountUnlockableCoinsResponse createEmptyInstance() => create();
  static $pb.PbList<AccountUnlockableCoinsResponse> createRepeated() => $pb.PbList<AccountUnlockableCoinsResponse>();
  @$core.pragma('dart2js:noInline')
  static AccountUnlockableCoinsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountUnlockableCoinsResponse>(create);
  static AccountUnlockableCoinsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.Coin> get coins => $_getList(0);
}

class AccountUnlockingCoinsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountUnlockingCoinsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..hasRequiredFields = false
  ;

  AccountUnlockingCoinsRequest._() : super();
  factory AccountUnlockingCoinsRequest() => create();
  factory AccountUnlockingCoinsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountUnlockingCoinsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountUnlockingCoinsRequest clone() => AccountUnlockingCoinsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountUnlockingCoinsRequest copyWith(void Function(AccountUnlockingCoinsRequest) updates) => super.copyWith((message) => updates(message as AccountUnlockingCoinsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountUnlockingCoinsRequest create() => AccountUnlockingCoinsRequest._();
  AccountUnlockingCoinsRequest createEmptyInstance() => create();
  static $pb.PbList<AccountUnlockingCoinsRequest> createRepeated() => $pb.PbList<AccountUnlockingCoinsRequest>();
  @$core.pragma('dart2js:noInline')
  static AccountUnlockingCoinsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountUnlockingCoinsRequest>(create);
  static AccountUnlockingCoinsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);
}

class AccountUnlockingCoinsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountUnlockingCoinsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$4.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'coins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..hasRequiredFields = false
  ;

  AccountUnlockingCoinsResponse._() : super();
  factory AccountUnlockingCoinsResponse() => create();
  factory AccountUnlockingCoinsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountUnlockingCoinsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountUnlockingCoinsResponse clone() => AccountUnlockingCoinsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountUnlockingCoinsResponse copyWith(void Function(AccountUnlockingCoinsResponse) updates) => super.copyWith((message) => updates(message as AccountUnlockingCoinsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountUnlockingCoinsResponse create() => AccountUnlockingCoinsResponse._();
  AccountUnlockingCoinsResponse createEmptyInstance() => create();
  static $pb.PbList<AccountUnlockingCoinsResponse> createRepeated() => $pb.PbList<AccountUnlockingCoinsResponse>();
  @$core.pragma('dart2js:noInline')
  static AccountUnlockingCoinsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountUnlockingCoinsResponse>(create);
  static AccountUnlockingCoinsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.Coin> get coins => $_getList(0);
}

class AccountLockedCoinsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedCoinsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..hasRequiredFields = false
  ;

  AccountLockedCoinsRequest._() : super();
  factory AccountLockedCoinsRequest() => create();
  factory AccountLockedCoinsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedCoinsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedCoinsRequest clone() => AccountLockedCoinsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedCoinsRequest copyWith(void Function(AccountLockedCoinsRequest) updates) => super.copyWith((message) => updates(message as AccountLockedCoinsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedCoinsRequest create() => AccountLockedCoinsRequest._();
  AccountLockedCoinsRequest createEmptyInstance() => create();
  static $pb.PbList<AccountLockedCoinsRequest> createRepeated() => $pb.PbList<AccountLockedCoinsRequest>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedCoinsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedCoinsRequest>(create);
  static AccountLockedCoinsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);
}

class AccountLockedCoinsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedCoinsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$4.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'coins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..hasRequiredFields = false
  ;

  AccountLockedCoinsResponse._() : super();
  factory AccountLockedCoinsResponse() => create();
  factory AccountLockedCoinsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedCoinsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedCoinsResponse clone() => AccountLockedCoinsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedCoinsResponse copyWith(void Function(AccountLockedCoinsResponse) updates) => super.copyWith((message) => updates(message as AccountLockedCoinsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedCoinsResponse create() => AccountLockedCoinsResponse._();
  AccountLockedCoinsResponse createEmptyInstance() => create();
  static $pb.PbList<AccountLockedCoinsResponse> createRepeated() => $pb.PbList<AccountLockedCoinsResponse>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedCoinsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedCoinsResponse>(create);
  static AccountLockedCoinsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.Coin> get coins => $_getList(0);
}

class AccountLockedPastTimeRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedPastTimeRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..aOM<$3.Timestamp>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'timestamp', subBuilder: $3.Timestamp.create)
    ..hasRequiredFields = false
  ;

  AccountLockedPastTimeRequest._() : super();
  factory AccountLockedPastTimeRequest() => create();
  factory AccountLockedPastTimeRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedPastTimeRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeRequest clone() => AccountLockedPastTimeRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeRequest copyWith(void Function(AccountLockedPastTimeRequest) updates) => super.copyWith((message) => updates(message as AccountLockedPastTimeRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeRequest create() => AccountLockedPastTimeRequest._();
  AccountLockedPastTimeRequest createEmptyInstance() => create();
  static $pb.PbList<AccountLockedPastTimeRequest> createRepeated() => $pb.PbList<AccountLockedPastTimeRequest>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedPastTimeRequest>(create);
  static AccountLockedPastTimeRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);

  @$pb.TagNumber(2)
  $3.Timestamp get timestamp => $_getN(1);
  @$pb.TagNumber(2)
  set timestamp($3.Timestamp v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasTimestamp() => $_has(1);
  @$pb.TagNumber(2)
  void clearTimestamp() => clearField(2);
  @$pb.TagNumber(2)
  $3.Timestamp ensureTimestamp() => $_ensure(1);
}

class AccountLockedPastTimeResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedPastTimeResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$5.PeriodLock>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'locks', $pb.PbFieldType.PM, subBuilder: $5.PeriodLock.create)
    ..hasRequiredFields = false
  ;

  AccountLockedPastTimeResponse._() : super();
  factory AccountLockedPastTimeResponse() => create();
  factory AccountLockedPastTimeResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedPastTimeResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeResponse clone() => AccountLockedPastTimeResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeResponse copyWith(void Function(AccountLockedPastTimeResponse) updates) => super.copyWith((message) => updates(message as AccountLockedPastTimeResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeResponse create() => AccountLockedPastTimeResponse._();
  AccountLockedPastTimeResponse createEmptyInstance() => create();
  static $pb.PbList<AccountLockedPastTimeResponse> createRepeated() => $pb.PbList<AccountLockedPastTimeResponse>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedPastTimeResponse>(create);
  static AccountLockedPastTimeResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$5.PeriodLock> get locks => $_getList(0);
}

class AccountLockedPastTimeNotUnlockingOnlyRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedPastTimeNotUnlockingOnlyRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..aOM<$3.Timestamp>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'timestamp', subBuilder: $3.Timestamp.create)
    ..hasRequiredFields = false
  ;

  AccountLockedPastTimeNotUnlockingOnlyRequest._() : super();
  factory AccountLockedPastTimeNotUnlockingOnlyRequest() => create();
  factory AccountLockedPastTimeNotUnlockingOnlyRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedPastTimeNotUnlockingOnlyRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeNotUnlockingOnlyRequest clone() => AccountLockedPastTimeNotUnlockingOnlyRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeNotUnlockingOnlyRequest copyWith(void Function(AccountLockedPastTimeNotUnlockingOnlyRequest) updates) => super.copyWith((message) => updates(message as AccountLockedPastTimeNotUnlockingOnlyRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeNotUnlockingOnlyRequest create() => AccountLockedPastTimeNotUnlockingOnlyRequest._();
  AccountLockedPastTimeNotUnlockingOnlyRequest createEmptyInstance() => create();
  static $pb.PbList<AccountLockedPastTimeNotUnlockingOnlyRequest> createRepeated() => $pb.PbList<AccountLockedPastTimeNotUnlockingOnlyRequest>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeNotUnlockingOnlyRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedPastTimeNotUnlockingOnlyRequest>(create);
  static AccountLockedPastTimeNotUnlockingOnlyRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);

  @$pb.TagNumber(2)
  $3.Timestamp get timestamp => $_getN(1);
  @$pb.TagNumber(2)
  set timestamp($3.Timestamp v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasTimestamp() => $_has(1);
  @$pb.TagNumber(2)
  void clearTimestamp() => clearField(2);
  @$pb.TagNumber(2)
  $3.Timestamp ensureTimestamp() => $_ensure(1);
}

class AccountLockedPastTimeNotUnlockingOnlyResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedPastTimeNotUnlockingOnlyResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$5.PeriodLock>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'locks', $pb.PbFieldType.PM, subBuilder: $5.PeriodLock.create)
    ..hasRequiredFields = false
  ;

  AccountLockedPastTimeNotUnlockingOnlyResponse._() : super();
  factory AccountLockedPastTimeNotUnlockingOnlyResponse() => create();
  factory AccountLockedPastTimeNotUnlockingOnlyResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedPastTimeNotUnlockingOnlyResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeNotUnlockingOnlyResponse clone() => AccountLockedPastTimeNotUnlockingOnlyResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeNotUnlockingOnlyResponse copyWith(void Function(AccountLockedPastTimeNotUnlockingOnlyResponse) updates) => super.copyWith((message) => updates(message as AccountLockedPastTimeNotUnlockingOnlyResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeNotUnlockingOnlyResponse create() => AccountLockedPastTimeNotUnlockingOnlyResponse._();
  AccountLockedPastTimeNotUnlockingOnlyResponse createEmptyInstance() => create();
  static $pb.PbList<AccountLockedPastTimeNotUnlockingOnlyResponse> createRepeated() => $pb.PbList<AccountLockedPastTimeNotUnlockingOnlyResponse>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeNotUnlockingOnlyResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedPastTimeNotUnlockingOnlyResponse>(create);
  static AccountLockedPastTimeNotUnlockingOnlyResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$5.PeriodLock> get locks => $_getList(0);
}

class AccountUnlockedBeforeTimeRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountUnlockedBeforeTimeRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..aOM<$3.Timestamp>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'timestamp', subBuilder: $3.Timestamp.create)
    ..hasRequiredFields = false
  ;

  AccountUnlockedBeforeTimeRequest._() : super();
  factory AccountUnlockedBeforeTimeRequest() => create();
  factory AccountUnlockedBeforeTimeRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountUnlockedBeforeTimeRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountUnlockedBeforeTimeRequest clone() => AccountUnlockedBeforeTimeRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountUnlockedBeforeTimeRequest copyWith(void Function(AccountUnlockedBeforeTimeRequest) updates) => super.copyWith((message) => updates(message as AccountUnlockedBeforeTimeRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountUnlockedBeforeTimeRequest create() => AccountUnlockedBeforeTimeRequest._();
  AccountUnlockedBeforeTimeRequest createEmptyInstance() => create();
  static $pb.PbList<AccountUnlockedBeforeTimeRequest> createRepeated() => $pb.PbList<AccountUnlockedBeforeTimeRequest>();
  @$core.pragma('dart2js:noInline')
  static AccountUnlockedBeforeTimeRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountUnlockedBeforeTimeRequest>(create);
  static AccountUnlockedBeforeTimeRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);

  @$pb.TagNumber(2)
  $3.Timestamp get timestamp => $_getN(1);
  @$pb.TagNumber(2)
  set timestamp($3.Timestamp v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasTimestamp() => $_has(1);
  @$pb.TagNumber(2)
  void clearTimestamp() => clearField(2);
  @$pb.TagNumber(2)
  $3.Timestamp ensureTimestamp() => $_ensure(1);
}

class AccountUnlockedBeforeTimeResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountUnlockedBeforeTimeResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$5.PeriodLock>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'locks', $pb.PbFieldType.PM, subBuilder: $5.PeriodLock.create)
    ..hasRequiredFields = false
  ;

  AccountUnlockedBeforeTimeResponse._() : super();
  factory AccountUnlockedBeforeTimeResponse() => create();
  factory AccountUnlockedBeforeTimeResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountUnlockedBeforeTimeResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountUnlockedBeforeTimeResponse clone() => AccountUnlockedBeforeTimeResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountUnlockedBeforeTimeResponse copyWith(void Function(AccountUnlockedBeforeTimeResponse) updates) => super.copyWith((message) => updates(message as AccountUnlockedBeforeTimeResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountUnlockedBeforeTimeResponse create() => AccountUnlockedBeforeTimeResponse._();
  AccountUnlockedBeforeTimeResponse createEmptyInstance() => create();
  static $pb.PbList<AccountUnlockedBeforeTimeResponse> createRepeated() => $pb.PbList<AccountUnlockedBeforeTimeResponse>();
  @$core.pragma('dart2js:noInline')
  static AccountUnlockedBeforeTimeResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountUnlockedBeforeTimeResponse>(create);
  static AccountUnlockedBeforeTimeResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$5.PeriodLock> get locks => $_getList(0);
}

class AccountLockedPastTimeDenomRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedPastTimeDenomRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..aOM<$3.Timestamp>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'timestamp', subBuilder: $3.Timestamp.create)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'denom')
    ..hasRequiredFields = false
  ;

  AccountLockedPastTimeDenomRequest._() : super();
  factory AccountLockedPastTimeDenomRequest() => create();
  factory AccountLockedPastTimeDenomRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedPastTimeDenomRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeDenomRequest clone() => AccountLockedPastTimeDenomRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeDenomRequest copyWith(void Function(AccountLockedPastTimeDenomRequest) updates) => super.copyWith((message) => updates(message as AccountLockedPastTimeDenomRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeDenomRequest create() => AccountLockedPastTimeDenomRequest._();
  AccountLockedPastTimeDenomRequest createEmptyInstance() => create();
  static $pb.PbList<AccountLockedPastTimeDenomRequest> createRepeated() => $pb.PbList<AccountLockedPastTimeDenomRequest>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeDenomRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedPastTimeDenomRequest>(create);
  static AccountLockedPastTimeDenomRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);

  @$pb.TagNumber(2)
  $3.Timestamp get timestamp => $_getN(1);
  @$pb.TagNumber(2)
  set timestamp($3.Timestamp v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasTimestamp() => $_has(1);
  @$pb.TagNumber(2)
  void clearTimestamp() => clearField(2);
  @$pb.TagNumber(2)
  $3.Timestamp ensureTimestamp() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.String get denom => $_getSZ(2);
  @$pb.TagNumber(3)
  set denom($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasDenom() => $_has(2);
  @$pb.TagNumber(3)
  void clearDenom() => clearField(3);
}

class AccountLockedPastTimeDenomResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedPastTimeDenomResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$5.PeriodLock>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'locks', $pb.PbFieldType.PM, subBuilder: $5.PeriodLock.create)
    ..hasRequiredFields = false
  ;

  AccountLockedPastTimeDenomResponse._() : super();
  factory AccountLockedPastTimeDenomResponse() => create();
  factory AccountLockedPastTimeDenomResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedPastTimeDenomResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeDenomResponse clone() => AccountLockedPastTimeDenomResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedPastTimeDenomResponse copyWith(void Function(AccountLockedPastTimeDenomResponse) updates) => super.copyWith((message) => updates(message as AccountLockedPastTimeDenomResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeDenomResponse create() => AccountLockedPastTimeDenomResponse._();
  AccountLockedPastTimeDenomResponse createEmptyInstance() => create();
  static $pb.PbList<AccountLockedPastTimeDenomResponse> createRepeated() => $pb.PbList<AccountLockedPastTimeDenomResponse>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedPastTimeDenomResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedPastTimeDenomResponse>(create);
  static AccountLockedPastTimeDenomResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$5.PeriodLock> get locks => $_getList(0);
}

class LockedDenomRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LockedDenomRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'denom')
    ..aOM<$2.Duration>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'duration', subBuilder: $2.Duration.create)
    ..hasRequiredFields = false
  ;

  LockedDenomRequest._() : super();
  factory LockedDenomRequest() => create();
  factory LockedDenomRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LockedDenomRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LockedDenomRequest clone() => LockedDenomRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LockedDenomRequest copyWith(void Function(LockedDenomRequest) updates) => super.copyWith((message) => updates(message as LockedDenomRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static LockedDenomRequest create() => LockedDenomRequest._();
  LockedDenomRequest createEmptyInstance() => create();
  static $pb.PbList<LockedDenomRequest> createRepeated() => $pb.PbList<LockedDenomRequest>();
  @$core.pragma('dart2js:noInline')
  static LockedDenomRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LockedDenomRequest>(create);
  static LockedDenomRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get denom => $_getSZ(0);
  @$pb.TagNumber(1)
  set denom($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDenom() => $_has(0);
  @$pb.TagNumber(1)
  void clearDenom() => clearField(1);

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
}

class LockedDenomResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LockedDenomResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount')
    ..hasRequiredFields = false
  ;

  LockedDenomResponse._() : super();
  factory LockedDenomResponse() => create();
  factory LockedDenomResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LockedDenomResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LockedDenomResponse clone() => LockedDenomResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LockedDenomResponse copyWith(void Function(LockedDenomResponse) updates) => super.copyWith((message) => updates(message as LockedDenomResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static LockedDenomResponse create() => LockedDenomResponse._();
  LockedDenomResponse createEmptyInstance() => create();
  static $pb.PbList<LockedDenomResponse> createRepeated() => $pb.PbList<LockedDenomResponse>();
  @$core.pragma('dart2js:noInline')
  static LockedDenomResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LockedDenomResponse>(create);
  static LockedDenomResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get amount => $_getSZ(0);
  @$pb.TagNumber(1)
  set amount($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAmount() => $_has(0);
  @$pb.TagNumber(1)
  void clearAmount() => clearField(1);
}

class LockedRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LockedRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lockId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  LockedRequest._() : super();
  factory LockedRequest() => create();
  factory LockedRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LockedRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LockedRequest clone() => LockedRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LockedRequest copyWith(void Function(LockedRequest) updates) => super.copyWith((message) => updates(message as LockedRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static LockedRequest create() => LockedRequest._();
  LockedRequest createEmptyInstance() => create();
  static $pb.PbList<LockedRequest> createRepeated() => $pb.PbList<LockedRequest>();
  @$core.pragma('dart2js:noInline')
  static LockedRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LockedRequest>(create);
  static LockedRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get lockId => $_getI64(0);
  @$pb.TagNumber(1)
  set lockId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasLockId() => $_has(0);
  @$pb.TagNumber(1)
  void clearLockId() => clearField(1);
}

class LockedResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LockedResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOM<$5.PeriodLock>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lock', subBuilder: $5.PeriodLock.create)
    ..hasRequiredFields = false
  ;

  LockedResponse._() : super();
  factory LockedResponse() => create();
  factory LockedResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LockedResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LockedResponse clone() => LockedResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LockedResponse copyWith(void Function(LockedResponse) updates) => super.copyWith((message) => updates(message as LockedResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static LockedResponse create() => LockedResponse._();
  LockedResponse createEmptyInstance() => create();
  static $pb.PbList<LockedResponse> createRepeated() => $pb.PbList<LockedResponse>();
  @$core.pragma('dart2js:noInline')
  static LockedResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LockedResponse>(create);
  static LockedResponse _defaultInstance;

  @$pb.TagNumber(1)
  $5.PeriodLock get lock => $_getN(0);
  @$pb.TagNumber(1)
  set lock($5.PeriodLock v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasLock() => $_has(0);
  @$pb.TagNumber(1)
  void clearLock() => clearField(1);
  @$pb.TagNumber(1)
  $5.PeriodLock ensureLock() => $_ensure(0);
}

class AccountLockedLongerDurationRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedLongerDurationRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..aOM<$2.Duration>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'duration', subBuilder: $2.Duration.create)
    ..hasRequiredFields = false
  ;

  AccountLockedLongerDurationRequest._() : super();
  factory AccountLockedLongerDurationRequest() => create();
  factory AccountLockedLongerDurationRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedLongerDurationRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationRequest clone() => AccountLockedLongerDurationRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationRequest copyWith(void Function(AccountLockedLongerDurationRequest) updates) => super.copyWith((message) => updates(message as AccountLockedLongerDurationRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationRequest create() => AccountLockedLongerDurationRequest._();
  AccountLockedLongerDurationRequest createEmptyInstance() => create();
  static $pb.PbList<AccountLockedLongerDurationRequest> createRepeated() => $pb.PbList<AccountLockedLongerDurationRequest>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedLongerDurationRequest>(create);
  static AccountLockedLongerDurationRequest _defaultInstance;

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
}

class AccountLockedLongerDurationResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedLongerDurationResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$5.PeriodLock>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'locks', $pb.PbFieldType.PM, subBuilder: $5.PeriodLock.create)
    ..hasRequiredFields = false
  ;

  AccountLockedLongerDurationResponse._() : super();
  factory AccountLockedLongerDurationResponse() => create();
  factory AccountLockedLongerDurationResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedLongerDurationResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationResponse clone() => AccountLockedLongerDurationResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationResponse copyWith(void Function(AccountLockedLongerDurationResponse) updates) => super.copyWith((message) => updates(message as AccountLockedLongerDurationResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationResponse create() => AccountLockedLongerDurationResponse._();
  AccountLockedLongerDurationResponse createEmptyInstance() => create();
  static $pb.PbList<AccountLockedLongerDurationResponse> createRepeated() => $pb.PbList<AccountLockedLongerDurationResponse>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedLongerDurationResponse>(create);
  static AccountLockedLongerDurationResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$5.PeriodLock> get locks => $_getList(0);
}

class AccountLockedLongerDurationNotUnlockingOnlyRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedLongerDurationNotUnlockingOnlyRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..aOM<$2.Duration>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'duration', subBuilder: $2.Duration.create)
    ..hasRequiredFields = false
  ;

  AccountLockedLongerDurationNotUnlockingOnlyRequest._() : super();
  factory AccountLockedLongerDurationNotUnlockingOnlyRequest() => create();
  factory AccountLockedLongerDurationNotUnlockingOnlyRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedLongerDurationNotUnlockingOnlyRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationNotUnlockingOnlyRequest clone() => AccountLockedLongerDurationNotUnlockingOnlyRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationNotUnlockingOnlyRequest copyWith(void Function(AccountLockedLongerDurationNotUnlockingOnlyRequest) updates) => super.copyWith((message) => updates(message as AccountLockedLongerDurationNotUnlockingOnlyRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationNotUnlockingOnlyRequest create() => AccountLockedLongerDurationNotUnlockingOnlyRequest._();
  AccountLockedLongerDurationNotUnlockingOnlyRequest createEmptyInstance() => create();
  static $pb.PbList<AccountLockedLongerDurationNotUnlockingOnlyRequest> createRepeated() => $pb.PbList<AccountLockedLongerDurationNotUnlockingOnlyRequest>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationNotUnlockingOnlyRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedLongerDurationNotUnlockingOnlyRequest>(create);
  static AccountLockedLongerDurationNotUnlockingOnlyRequest _defaultInstance;

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
}

class AccountLockedLongerDurationNotUnlockingOnlyResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedLongerDurationNotUnlockingOnlyResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$5.PeriodLock>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'locks', $pb.PbFieldType.PM, subBuilder: $5.PeriodLock.create)
    ..hasRequiredFields = false
  ;

  AccountLockedLongerDurationNotUnlockingOnlyResponse._() : super();
  factory AccountLockedLongerDurationNotUnlockingOnlyResponse() => create();
  factory AccountLockedLongerDurationNotUnlockingOnlyResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedLongerDurationNotUnlockingOnlyResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationNotUnlockingOnlyResponse clone() => AccountLockedLongerDurationNotUnlockingOnlyResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationNotUnlockingOnlyResponse copyWith(void Function(AccountLockedLongerDurationNotUnlockingOnlyResponse) updates) => super.copyWith((message) => updates(message as AccountLockedLongerDurationNotUnlockingOnlyResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationNotUnlockingOnlyResponse create() => AccountLockedLongerDurationNotUnlockingOnlyResponse._();
  AccountLockedLongerDurationNotUnlockingOnlyResponse createEmptyInstance() => create();
  static $pb.PbList<AccountLockedLongerDurationNotUnlockingOnlyResponse> createRepeated() => $pb.PbList<AccountLockedLongerDurationNotUnlockingOnlyResponse>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationNotUnlockingOnlyResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedLongerDurationNotUnlockingOnlyResponse>(create);
  static AccountLockedLongerDurationNotUnlockingOnlyResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$5.PeriodLock> get locks => $_getList(0);
}

class AccountLockedLongerDurationDenomRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedLongerDurationDenomRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..aOM<$2.Duration>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'duration', subBuilder: $2.Duration.create)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'denom')
    ..hasRequiredFields = false
  ;

  AccountLockedLongerDurationDenomRequest._() : super();
  factory AccountLockedLongerDurationDenomRequest() => create();
  factory AccountLockedLongerDurationDenomRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedLongerDurationDenomRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationDenomRequest clone() => AccountLockedLongerDurationDenomRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationDenomRequest copyWith(void Function(AccountLockedLongerDurationDenomRequest) updates) => super.copyWith((message) => updates(message as AccountLockedLongerDurationDenomRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationDenomRequest create() => AccountLockedLongerDurationDenomRequest._();
  AccountLockedLongerDurationDenomRequest createEmptyInstance() => create();
  static $pb.PbList<AccountLockedLongerDurationDenomRequest> createRepeated() => $pb.PbList<AccountLockedLongerDurationDenomRequest>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationDenomRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedLongerDurationDenomRequest>(create);
  static AccountLockedLongerDurationDenomRequest _defaultInstance;

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
  $core.String get denom => $_getSZ(2);
  @$pb.TagNumber(3)
  set denom($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasDenom() => $_has(2);
  @$pb.TagNumber(3)
  void clearDenom() => clearField(3);
}

class AccountLockedLongerDurationDenomResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AccountLockedLongerDurationDenomResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.lockup'), createEmptyInstance: create)
    ..pc<$5.PeriodLock>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'locks', $pb.PbFieldType.PM, subBuilder: $5.PeriodLock.create)
    ..hasRequiredFields = false
  ;

  AccountLockedLongerDurationDenomResponse._() : super();
  factory AccountLockedLongerDurationDenomResponse() => create();
  factory AccountLockedLongerDurationDenomResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AccountLockedLongerDurationDenomResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationDenomResponse clone() => AccountLockedLongerDurationDenomResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AccountLockedLongerDurationDenomResponse copyWith(void Function(AccountLockedLongerDurationDenomResponse) updates) => super.copyWith((message) => updates(message as AccountLockedLongerDurationDenomResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationDenomResponse create() => AccountLockedLongerDurationDenomResponse._();
  AccountLockedLongerDurationDenomResponse createEmptyInstance() => create();
  static $pb.PbList<AccountLockedLongerDurationDenomResponse> createRepeated() => $pb.PbList<AccountLockedLongerDurationDenomResponse>();
  @$core.pragma('dart2js:noInline')
  static AccountLockedLongerDurationDenomResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountLockedLongerDurationDenomResponse>(create);
  static AccountLockedLongerDurationDenomResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$5.PeriodLock> get locks => $_getList(0);
}

