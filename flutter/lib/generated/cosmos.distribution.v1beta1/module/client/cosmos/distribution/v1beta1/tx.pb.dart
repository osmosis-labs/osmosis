///
//  Generated code. Do not modify.
//  source: cosmos/distribution/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../base/v1beta1/coin.pb.dart' as $2;

class MsgSetWithdrawAddress extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgSetWithdrawAddress', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'withdrawAddress')
    ..hasRequiredFields = false
  ;

  MsgSetWithdrawAddress._() : super();
  factory MsgSetWithdrawAddress() => create();
  factory MsgSetWithdrawAddress.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgSetWithdrawAddress.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgSetWithdrawAddress clone() => MsgSetWithdrawAddress()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgSetWithdrawAddress copyWith(void Function(MsgSetWithdrawAddress) updates) => super.copyWith((message) => updates(message as MsgSetWithdrawAddress)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgSetWithdrawAddress create() => MsgSetWithdrawAddress._();
  MsgSetWithdrawAddress createEmptyInstance() => create();
  static $pb.PbList<MsgSetWithdrawAddress> createRepeated() => $pb.PbList<MsgSetWithdrawAddress>();
  @$core.pragma('dart2js:noInline')
  static MsgSetWithdrawAddress getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgSetWithdrawAddress>(create);
  static MsgSetWithdrawAddress _defaultInstance;

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

class MsgSetWithdrawAddressResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgSetWithdrawAddressResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgSetWithdrawAddressResponse._() : super();
  factory MsgSetWithdrawAddressResponse() => create();
  factory MsgSetWithdrawAddressResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgSetWithdrawAddressResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgSetWithdrawAddressResponse clone() => MsgSetWithdrawAddressResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgSetWithdrawAddressResponse copyWith(void Function(MsgSetWithdrawAddressResponse) updates) => super.copyWith((message) => updates(message as MsgSetWithdrawAddressResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgSetWithdrawAddressResponse create() => MsgSetWithdrawAddressResponse._();
  MsgSetWithdrawAddressResponse createEmptyInstance() => create();
  static $pb.PbList<MsgSetWithdrawAddressResponse> createRepeated() => $pb.PbList<MsgSetWithdrawAddressResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgSetWithdrawAddressResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgSetWithdrawAddressResponse>(create);
  static MsgSetWithdrawAddressResponse _defaultInstance;
}

class MsgWithdrawDelegatorReward extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgWithdrawDelegatorReward', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..hasRequiredFields = false
  ;

  MsgWithdrawDelegatorReward._() : super();
  factory MsgWithdrawDelegatorReward() => create();
  factory MsgWithdrawDelegatorReward.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgWithdrawDelegatorReward.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgWithdrawDelegatorReward clone() => MsgWithdrawDelegatorReward()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgWithdrawDelegatorReward copyWith(void Function(MsgWithdrawDelegatorReward) updates) => super.copyWith((message) => updates(message as MsgWithdrawDelegatorReward)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgWithdrawDelegatorReward create() => MsgWithdrawDelegatorReward._();
  MsgWithdrawDelegatorReward createEmptyInstance() => create();
  static $pb.PbList<MsgWithdrawDelegatorReward> createRepeated() => $pb.PbList<MsgWithdrawDelegatorReward>();
  @$core.pragma('dart2js:noInline')
  static MsgWithdrawDelegatorReward getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgWithdrawDelegatorReward>(create);
  static MsgWithdrawDelegatorReward _defaultInstance;

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
}

class MsgWithdrawDelegatorRewardResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgWithdrawDelegatorRewardResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgWithdrawDelegatorRewardResponse._() : super();
  factory MsgWithdrawDelegatorRewardResponse() => create();
  factory MsgWithdrawDelegatorRewardResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgWithdrawDelegatorRewardResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgWithdrawDelegatorRewardResponse clone() => MsgWithdrawDelegatorRewardResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgWithdrawDelegatorRewardResponse copyWith(void Function(MsgWithdrawDelegatorRewardResponse) updates) => super.copyWith((message) => updates(message as MsgWithdrawDelegatorRewardResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgWithdrawDelegatorRewardResponse create() => MsgWithdrawDelegatorRewardResponse._();
  MsgWithdrawDelegatorRewardResponse createEmptyInstance() => create();
  static $pb.PbList<MsgWithdrawDelegatorRewardResponse> createRepeated() => $pb.PbList<MsgWithdrawDelegatorRewardResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgWithdrawDelegatorRewardResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgWithdrawDelegatorRewardResponse>(create);
  static MsgWithdrawDelegatorRewardResponse _defaultInstance;
}

class MsgWithdrawValidatorCommission extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgWithdrawValidatorCommission', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..hasRequiredFields = false
  ;

  MsgWithdrawValidatorCommission._() : super();
  factory MsgWithdrawValidatorCommission() => create();
  factory MsgWithdrawValidatorCommission.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgWithdrawValidatorCommission.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgWithdrawValidatorCommission clone() => MsgWithdrawValidatorCommission()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgWithdrawValidatorCommission copyWith(void Function(MsgWithdrawValidatorCommission) updates) => super.copyWith((message) => updates(message as MsgWithdrawValidatorCommission)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgWithdrawValidatorCommission create() => MsgWithdrawValidatorCommission._();
  MsgWithdrawValidatorCommission createEmptyInstance() => create();
  static $pb.PbList<MsgWithdrawValidatorCommission> createRepeated() => $pb.PbList<MsgWithdrawValidatorCommission>();
  @$core.pragma('dart2js:noInline')
  static MsgWithdrawValidatorCommission getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgWithdrawValidatorCommission>(create);
  static MsgWithdrawValidatorCommission _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get validatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set validatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorAddress() => clearField(1);
}

class MsgWithdrawValidatorCommissionResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgWithdrawValidatorCommissionResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgWithdrawValidatorCommissionResponse._() : super();
  factory MsgWithdrawValidatorCommissionResponse() => create();
  factory MsgWithdrawValidatorCommissionResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgWithdrawValidatorCommissionResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgWithdrawValidatorCommissionResponse clone() => MsgWithdrawValidatorCommissionResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgWithdrawValidatorCommissionResponse copyWith(void Function(MsgWithdrawValidatorCommissionResponse) updates) => super.copyWith((message) => updates(message as MsgWithdrawValidatorCommissionResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgWithdrawValidatorCommissionResponse create() => MsgWithdrawValidatorCommissionResponse._();
  MsgWithdrawValidatorCommissionResponse createEmptyInstance() => create();
  static $pb.PbList<MsgWithdrawValidatorCommissionResponse> createRepeated() => $pb.PbList<MsgWithdrawValidatorCommissionResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgWithdrawValidatorCommissionResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgWithdrawValidatorCommissionResponse>(create);
  static MsgWithdrawValidatorCommissionResponse _defaultInstance;
}

class MsgFundCommunityPool extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgFundCommunityPool', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pc<$2.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', $pb.PbFieldType.PM, subBuilder: $2.Coin.create)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'depositor')
    ..hasRequiredFields = false
  ;

  MsgFundCommunityPool._() : super();
  factory MsgFundCommunityPool() => create();
  factory MsgFundCommunityPool.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgFundCommunityPool.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgFundCommunityPool clone() => MsgFundCommunityPool()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgFundCommunityPool copyWith(void Function(MsgFundCommunityPool) updates) => super.copyWith((message) => updates(message as MsgFundCommunityPool)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgFundCommunityPool create() => MsgFundCommunityPool._();
  MsgFundCommunityPool createEmptyInstance() => create();
  static $pb.PbList<MsgFundCommunityPool> createRepeated() => $pb.PbList<MsgFundCommunityPool>();
  @$core.pragma('dart2js:noInline')
  static MsgFundCommunityPool getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgFundCommunityPool>(create);
  static MsgFundCommunityPool _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.Coin> get amount => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get depositor => $_getSZ(1);
  @$pb.TagNumber(2)
  set depositor($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasDepositor() => $_has(1);
  @$pb.TagNumber(2)
  void clearDepositor() => clearField(2);
}

class MsgFundCommunityPoolResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgFundCommunityPoolResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgFundCommunityPoolResponse._() : super();
  factory MsgFundCommunityPoolResponse() => create();
  factory MsgFundCommunityPoolResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgFundCommunityPoolResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgFundCommunityPoolResponse clone() => MsgFundCommunityPoolResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgFundCommunityPoolResponse copyWith(void Function(MsgFundCommunityPoolResponse) updates) => super.copyWith((message) => updates(message as MsgFundCommunityPoolResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgFundCommunityPoolResponse create() => MsgFundCommunityPoolResponse._();
  MsgFundCommunityPoolResponse createEmptyInstance() => create();
  static $pb.PbList<MsgFundCommunityPoolResponse> createRepeated() => $pb.PbList<MsgFundCommunityPoolResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgFundCommunityPoolResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgFundCommunityPoolResponse>(create);
  static MsgFundCommunityPoolResponse _defaultInstance;
}

