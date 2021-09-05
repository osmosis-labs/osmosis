///
//  Generated code. Do not modify.
//  source: osmosis/gamm/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'pool.pb.dart' as $6;
import '../../../cosmos/base/v1beta1/coin.pb.dart' as $2;

class MsgCreatePool extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgCreatePool', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..aOM<$6.PoolParams>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolParams', protoName: 'poolParams', subBuilder: $6.PoolParams.create)
    ..pc<$6.PoolAsset>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolAssets', $pb.PbFieldType.PM, protoName: 'poolAssets', subBuilder: $6.PoolAsset.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'futurePoolGovernor')
    ..hasRequiredFields = false
  ;

  MsgCreatePool._() : super();
  factory MsgCreatePool() => create();
  factory MsgCreatePool.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgCreatePool.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgCreatePool clone() => MsgCreatePool()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgCreatePool copyWith(void Function(MsgCreatePool) updates) => super.copyWith((message) => updates(message as MsgCreatePool)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgCreatePool create() => MsgCreatePool._();
  MsgCreatePool createEmptyInstance() => create();
  static $pb.PbList<MsgCreatePool> createRepeated() => $pb.PbList<MsgCreatePool>();
  @$core.pragma('dart2js:noInline')
  static MsgCreatePool getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgCreatePool>(create);
  static MsgCreatePool _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sender => $_getSZ(0);
  @$pb.TagNumber(1)
  set sender($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSender() => $_has(0);
  @$pb.TagNumber(1)
  void clearSender() => clearField(1);

  @$pb.TagNumber(2)
  $6.PoolParams get poolParams => $_getN(1);
  @$pb.TagNumber(2)
  set poolParams($6.PoolParams v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPoolParams() => $_has(1);
  @$pb.TagNumber(2)
  void clearPoolParams() => clearField(2);
  @$pb.TagNumber(2)
  $6.PoolParams ensurePoolParams() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.List<$6.PoolAsset> get poolAssets => $_getList(2);

  @$pb.TagNumber(4)
  $core.String get futurePoolGovernor => $_getSZ(3);
  @$pb.TagNumber(4)
  set futurePoolGovernor($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasFuturePoolGovernor() => $_has(3);
  @$pb.TagNumber(4)
  void clearFuturePoolGovernor() => clearField(4);
}

class MsgCreatePoolResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgCreatePoolResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgCreatePoolResponse._() : super();
  factory MsgCreatePoolResponse() => create();
  factory MsgCreatePoolResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgCreatePoolResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgCreatePoolResponse clone() => MsgCreatePoolResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgCreatePoolResponse copyWith(void Function(MsgCreatePoolResponse) updates) => super.copyWith((message) => updates(message as MsgCreatePoolResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgCreatePoolResponse create() => MsgCreatePoolResponse._();
  MsgCreatePoolResponse createEmptyInstance() => create();
  static $pb.PbList<MsgCreatePoolResponse> createRepeated() => $pb.PbList<MsgCreatePoolResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgCreatePoolResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgCreatePoolResponse>(create);
  static MsgCreatePoolResponse _defaultInstance;
}

class MsgJoinPool extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgJoinPool', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'shareOutAmount', protoName: 'shareOutAmount')
    ..pc<$2.Coin>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenInMaxs', $pb.PbFieldType.PM, protoName: 'tokenInMaxs', subBuilder: $2.Coin.create)
    ..hasRequiredFields = false
  ;

  MsgJoinPool._() : super();
  factory MsgJoinPool() => create();
  factory MsgJoinPool.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgJoinPool.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgJoinPool clone() => MsgJoinPool()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgJoinPool copyWith(void Function(MsgJoinPool) updates) => super.copyWith((message) => updates(message as MsgJoinPool)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgJoinPool create() => MsgJoinPool._();
  MsgJoinPool createEmptyInstance() => create();
  static $pb.PbList<MsgJoinPool> createRepeated() => $pb.PbList<MsgJoinPool>();
  @$core.pragma('dart2js:noInline')
  static MsgJoinPool getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgJoinPool>(create);
  static MsgJoinPool _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sender => $_getSZ(0);
  @$pb.TagNumber(1)
  set sender($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSender() => $_has(0);
  @$pb.TagNumber(1)
  void clearSender() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get poolId => $_getI64(1);
  @$pb.TagNumber(2)
  set poolId($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPoolId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPoolId() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get shareOutAmount => $_getSZ(2);
  @$pb.TagNumber(3)
  set shareOutAmount($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasShareOutAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearShareOutAmount() => clearField(3);

  @$pb.TagNumber(4)
  $core.List<$2.Coin> get tokenInMaxs => $_getList(3);
}

class MsgJoinPoolResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgJoinPoolResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgJoinPoolResponse._() : super();
  factory MsgJoinPoolResponse() => create();
  factory MsgJoinPoolResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgJoinPoolResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgJoinPoolResponse clone() => MsgJoinPoolResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgJoinPoolResponse copyWith(void Function(MsgJoinPoolResponse) updates) => super.copyWith((message) => updates(message as MsgJoinPoolResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgJoinPoolResponse create() => MsgJoinPoolResponse._();
  MsgJoinPoolResponse createEmptyInstance() => create();
  static $pb.PbList<MsgJoinPoolResponse> createRepeated() => $pb.PbList<MsgJoinPoolResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgJoinPoolResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgJoinPoolResponse>(create);
  static MsgJoinPoolResponse _defaultInstance;
}

class MsgExitPool extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgExitPool', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'shareInAmount', protoName: 'shareInAmount')
    ..pc<$2.Coin>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenOutMins', $pb.PbFieldType.PM, protoName: 'tokenOutMins', subBuilder: $2.Coin.create)
    ..hasRequiredFields = false
  ;

  MsgExitPool._() : super();
  factory MsgExitPool() => create();
  factory MsgExitPool.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgExitPool.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgExitPool clone() => MsgExitPool()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgExitPool copyWith(void Function(MsgExitPool) updates) => super.copyWith((message) => updates(message as MsgExitPool)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgExitPool create() => MsgExitPool._();
  MsgExitPool createEmptyInstance() => create();
  static $pb.PbList<MsgExitPool> createRepeated() => $pb.PbList<MsgExitPool>();
  @$core.pragma('dart2js:noInline')
  static MsgExitPool getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgExitPool>(create);
  static MsgExitPool _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sender => $_getSZ(0);
  @$pb.TagNumber(1)
  set sender($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSender() => $_has(0);
  @$pb.TagNumber(1)
  void clearSender() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get poolId => $_getI64(1);
  @$pb.TagNumber(2)
  set poolId($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPoolId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPoolId() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get shareInAmount => $_getSZ(2);
  @$pb.TagNumber(3)
  set shareInAmount($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasShareInAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearShareInAmount() => clearField(3);

  @$pb.TagNumber(4)
  $core.List<$2.Coin> get tokenOutMins => $_getList(3);
}

class MsgExitPoolResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgExitPoolResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgExitPoolResponse._() : super();
  factory MsgExitPoolResponse() => create();
  factory MsgExitPoolResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgExitPoolResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgExitPoolResponse clone() => MsgExitPoolResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgExitPoolResponse copyWith(void Function(MsgExitPoolResponse) updates) => super.copyWith((message) => updates(message as MsgExitPoolResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgExitPoolResponse create() => MsgExitPoolResponse._();
  MsgExitPoolResponse createEmptyInstance() => create();
  static $pb.PbList<MsgExitPoolResponse> createRepeated() => $pb.PbList<MsgExitPoolResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgExitPoolResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgExitPoolResponse>(create);
  static MsgExitPoolResponse _defaultInstance;
}

class SwapAmountInRoute extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SwapAmountInRoute', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenOutDenom', protoName: 'tokenOutDenom')
    ..hasRequiredFields = false
  ;

  SwapAmountInRoute._() : super();
  factory SwapAmountInRoute() => create();
  factory SwapAmountInRoute.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SwapAmountInRoute.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SwapAmountInRoute clone() => SwapAmountInRoute()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SwapAmountInRoute copyWith(void Function(SwapAmountInRoute) updates) => super.copyWith((message) => updates(message as SwapAmountInRoute)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static SwapAmountInRoute create() => SwapAmountInRoute._();
  SwapAmountInRoute createEmptyInstance() => create();
  static $pb.PbList<SwapAmountInRoute> createRepeated() => $pb.PbList<SwapAmountInRoute>();
  @$core.pragma('dart2js:noInline')
  static SwapAmountInRoute getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SwapAmountInRoute>(create);
  static SwapAmountInRoute _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get poolId => $_getI64(0);
  @$pb.TagNumber(1)
  set poolId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPoolId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPoolId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get tokenOutDenom => $_getSZ(1);
  @$pb.TagNumber(2)
  set tokenOutDenom($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasTokenOutDenom() => $_has(1);
  @$pb.TagNumber(2)
  void clearTokenOutDenom() => clearField(2);
}

class MsgSwapExactAmountIn extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgSwapExactAmountIn', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..pc<SwapAmountInRoute>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'routes', $pb.PbFieldType.PM, subBuilder: SwapAmountInRoute.create)
    ..aOM<$2.Coin>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenIn', protoName: 'tokenIn', subBuilder: $2.Coin.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenOutMinAmount', protoName: 'tokenOutMinAmount')
    ..hasRequiredFields = false
  ;

  MsgSwapExactAmountIn._() : super();
  factory MsgSwapExactAmountIn() => create();
  factory MsgSwapExactAmountIn.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgSwapExactAmountIn.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgSwapExactAmountIn clone() => MsgSwapExactAmountIn()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgSwapExactAmountIn copyWith(void Function(MsgSwapExactAmountIn) updates) => super.copyWith((message) => updates(message as MsgSwapExactAmountIn)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgSwapExactAmountIn create() => MsgSwapExactAmountIn._();
  MsgSwapExactAmountIn createEmptyInstance() => create();
  static $pb.PbList<MsgSwapExactAmountIn> createRepeated() => $pb.PbList<MsgSwapExactAmountIn>();
  @$core.pragma('dart2js:noInline')
  static MsgSwapExactAmountIn getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgSwapExactAmountIn>(create);
  static MsgSwapExactAmountIn _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sender => $_getSZ(0);
  @$pb.TagNumber(1)
  set sender($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSender() => $_has(0);
  @$pb.TagNumber(1)
  void clearSender() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<SwapAmountInRoute> get routes => $_getList(1);

  @$pb.TagNumber(3)
  $2.Coin get tokenIn => $_getN(2);
  @$pb.TagNumber(3)
  set tokenIn($2.Coin v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasTokenIn() => $_has(2);
  @$pb.TagNumber(3)
  void clearTokenIn() => clearField(3);
  @$pb.TagNumber(3)
  $2.Coin ensureTokenIn() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.String get tokenOutMinAmount => $_getSZ(3);
  @$pb.TagNumber(4)
  set tokenOutMinAmount($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasTokenOutMinAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearTokenOutMinAmount() => clearField(4);
}

class MsgSwapExactAmountInResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgSwapExactAmountInResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgSwapExactAmountInResponse._() : super();
  factory MsgSwapExactAmountInResponse() => create();
  factory MsgSwapExactAmountInResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgSwapExactAmountInResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgSwapExactAmountInResponse clone() => MsgSwapExactAmountInResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgSwapExactAmountInResponse copyWith(void Function(MsgSwapExactAmountInResponse) updates) => super.copyWith((message) => updates(message as MsgSwapExactAmountInResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgSwapExactAmountInResponse create() => MsgSwapExactAmountInResponse._();
  MsgSwapExactAmountInResponse createEmptyInstance() => create();
  static $pb.PbList<MsgSwapExactAmountInResponse> createRepeated() => $pb.PbList<MsgSwapExactAmountInResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgSwapExactAmountInResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgSwapExactAmountInResponse>(create);
  static MsgSwapExactAmountInResponse _defaultInstance;
}

class SwapAmountOutRoute extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SwapAmountOutRoute', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenInDenom', protoName: 'tokenInDenom')
    ..hasRequiredFields = false
  ;

  SwapAmountOutRoute._() : super();
  factory SwapAmountOutRoute() => create();
  factory SwapAmountOutRoute.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SwapAmountOutRoute.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SwapAmountOutRoute clone() => SwapAmountOutRoute()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SwapAmountOutRoute copyWith(void Function(SwapAmountOutRoute) updates) => super.copyWith((message) => updates(message as SwapAmountOutRoute)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static SwapAmountOutRoute create() => SwapAmountOutRoute._();
  SwapAmountOutRoute createEmptyInstance() => create();
  static $pb.PbList<SwapAmountOutRoute> createRepeated() => $pb.PbList<SwapAmountOutRoute>();
  @$core.pragma('dart2js:noInline')
  static SwapAmountOutRoute getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SwapAmountOutRoute>(create);
  static SwapAmountOutRoute _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get poolId => $_getI64(0);
  @$pb.TagNumber(1)
  set poolId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPoolId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPoolId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get tokenInDenom => $_getSZ(1);
  @$pb.TagNumber(2)
  set tokenInDenom($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasTokenInDenom() => $_has(1);
  @$pb.TagNumber(2)
  void clearTokenInDenom() => clearField(2);
}

class MsgSwapExactAmountOut extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgSwapExactAmountOut', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..pc<SwapAmountOutRoute>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'routes', $pb.PbFieldType.PM, subBuilder: SwapAmountOutRoute.create)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenInMaxAmount', protoName: 'tokenInMaxAmount')
    ..aOM<$2.Coin>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenOut', protoName: 'tokenOut', subBuilder: $2.Coin.create)
    ..hasRequiredFields = false
  ;

  MsgSwapExactAmountOut._() : super();
  factory MsgSwapExactAmountOut() => create();
  factory MsgSwapExactAmountOut.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgSwapExactAmountOut.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgSwapExactAmountOut clone() => MsgSwapExactAmountOut()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgSwapExactAmountOut copyWith(void Function(MsgSwapExactAmountOut) updates) => super.copyWith((message) => updates(message as MsgSwapExactAmountOut)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgSwapExactAmountOut create() => MsgSwapExactAmountOut._();
  MsgSwapExactAmountOut createEmptyInstance() => create();
  static $pb.PbList<MsgSwapExactAmountOut> createRepeated() => $pb.PbList<MsgSwapExactAmountOut>();
  @$core.pragma('dart2js:noInline')
  static MsgSwapExactAmountOut getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgSwapExactAmountOut>(create);
  static MsgSwapExactAmountOut _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sender => $_getSZ(0);
  @$pb.TagNumber(1)
  set sender($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSender() => $_has(0);
  @$pb.TagNumber(1)
  void clearSender() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<SwapAmountOutRoute> get routes => $_getList(1);

  @$pb.TagNumber(3)
  $core.String get tokenInMaxAmount => $_getSZ(2);
  @$pb.TagNumber(3)
  set tokenInMaxAmount($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTokenInMaxAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearTokenInMaxAmount() => clearField(3);

  @$pb.TagNumber(4)
  $2.Coin get tokenOut => $_getN(3);
  @$pb.TagNumber(4)
  set tokenOut($2.Coin v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasTokenOut() => $_has(3);
  @$pb.TagNumber(4)
  void clearTokenOut() => clearField(4);
  @$pb.TagNumber(4)
  $2.Coin ensureTokenOut() => $_ensure(3);
}

class MsgSwapExactAmountOutResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgSwapExactAmountOutResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgSwapExactAmountOutResponse._() : super();
  factory MsgSwapExactAmountOutResponse() => create();
  factory MsgSwapExactAmountOutResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgSwapExactAmountOutResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgSwapExactAmountOutResponse clone() => MsgSwapExactAmountOutResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgSwapExactAmountOutResponse copyWith(void Function(MsgSwapExactAmountOutResponse) updates) => super.copyWith((message) => updates(message as MsgSwapExactAmountOutResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgSwapExactAmountOutResponse create() => MsgSwapExactAmountOutResponse._();
  MsgSwapExactAmountOutResponse createEmptyInstance() => create();
  static $pb.PbList<MsgSwapExactAmountOutResponse> createRepeated() => $pb.PbList<MsgSwapExactAmountOutResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgSwapExactAmountOutResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgSwapExactAmountOutResponse>(create);
  static MsgSwapExactAmountOutResponse _defaultInstance;
}

class MsgJoinSwapExternAmountIn extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgJoinSwapExternAmountIn', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOM<$2.Coin>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenIn', protoName: 'tokenIn', subBuilder: $2.Coin.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'shareOutMinAmount', protoName: 'shareOutMinAmount')
    ..hasRequiredFields = false
  ;

  MsgJoinSwapExternAmountIn._() : super();
  factory MsgJoinSwapExternAmountIn() => create();
  factory MsgJoinSwapExternAmountIn.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgJoinSwapExternAmountIn.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgJoinSwapExternAmountIn clone() => MsgJoinSwapExternAmountIn()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgJoinSwapExternAmountIn copyWith(void Function(MsgJoinSwapExternAmountIn) updates) => super.copyWith((message) => updates(message as MsgJoinSwapExternAmountIn)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgJoinSwapExternAmountIn create() => MsgJoinSwapExternAmountIn._();
  MsgJoinSwapExternAmountIn createEmptyInstance() => create();
  static $pb.PbList<MsgJoinSwapExternAmountIn> createRepeated() => $pb.PbList<MsgJoinSwapExternAmountIn>();
  @$core.pragma('dart2js:noInline')
  static MsgJoinSwapExternAmountIn getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgJoinSwapExternAmountIn>(create);
  static MsgJoinSwapExternAmountIn _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sender => $_getSZ(0);
  @$pb.TagNumber(1)
  set sender($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSender() => $_has(0);
  @$pb.TagNumber(1)
  void clearSender() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get poolId => $_getI64(1);
  @$pb.TagNumber(2)
  set poolId($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPoolId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPoolId() => clearField(2);

  @$pb.TagNumber(3)
  $2.Coin get tokenIn => $_getN(2);
  @$pb.TagNumber(3)
  set tokenIn($2.Coin v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasTokenIn() => $_has(2);
  @$pb.TagNumber(3)
  void clearTokenIn() => clearField(3);
  @$pb.TagNumber(3)
  $2.Coin ensureTokenIn() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.String get shareOutMinAmount => $_getSZ(3);
  @$pb.TagNumber(4)
  set shareOutMinAmount($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasShareOutMinAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearShareOutMinAmount() => clearField(4);
}

class MsgJoinSwapExternAmountInResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgJoinSwapExternAmountInResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgJoinSwapExternAmountInResponse._() : super();
  factory MsgJoinSwapExternAmountInResponse() => create();
  factory MsgJoinSwapExternAmountInResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgJoinSwapExternAmountInResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgJoinSwapExternAmountInResponse clone() => MsgJoinSwapExternAmountInResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgJoinSwapExternAmountInResponse copyWith(void Function(MsgJoinSwapExternAmountInResponse) updates) => super.copyWith((message) => updates(message as MsgJoinSwapExternAmountInResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgJoinSwapExternAmountInResponse create() => MsgJoinSwapExternAmountInResponse._();
  MsgJoinSwapExternAmountInResponse createEmptyInstance() => create();
  static $pb.PbList<MsgJoinSwapExternAmountInResponse> createRepeated() => $pb.PbList<MsgJoinSwapExternAmountInResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgJoinSwapExternAmountInResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgJoinSwapExternAmountInResponse>(create);
  static MsgJoinSwapExternAmountInResponse _defaultInstance;
}

class MsgJoinSwapShareAmountOut extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgJoinSwapShareAmountOut', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenInDenom', protoName: 'tokenInDenom')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'shareOutAmount', protoName: 'shareOutAmount')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenInMaxAmount', protoName: 'tokenInMaxAmount')
    ..hasRequiredFields = false
  ;

  MsgJoinSwapShareAmountOut._() : super();
  factory MsgJoinSwapShareAmountOut() => create();
  factory MsgJoinSwapShareAmountOut.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgJoinSwapShareAmountOut.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgJoinSwapShareAmountOut clone() => MsgJoinSwapShareAmountOut()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgJoinSwapShareAmountOut copyWith(void Function(MsgJoinSwapShareAmountOut) updates) => super.copyWith((message) => updates(message as MsgJoinSwapShareAmountOut)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgJoinSwapShareAmountOut create() => MsgJoinSwapShareAmountOut._();
  MsgJoinSwapShareAmountOut createEmptyInstance() => create();
  static $pb.PbList<MsgJoinSwapShareAmountOut> createRepeated() => $pb.PbList<MsgJoinSwapShareAmountOut>();
  @$core.pragma('dart2js:noInline')
  static MsgJoinSwapShareAmountOut getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgJoinSwapShareAmountOut>(create);
  static MsgJoinSwapShareAmountOut _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sender => $_getSZ(0);
  @$pb.TagNumber(1)
  set sender($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSender() => $_has(0);
  @$pb.TagNumber(1)
  void clearSender() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get poolId => $_getI64(1);
  @$pb.TagNumber(2)
  set poolId($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPoolId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPoolId() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get tokenInDenom => $_getSZ(2);
  @$pb.TagNumber(3)
  set tokenInDenom($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTokenInDenom() => $_has(2);
  @$pb.TagNumber(3)
  void clearTokenInDenom() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get shareOutAmount => $_getSZ(3);
  @$pb.TagNumber(4)
  set shareOutAmount($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasShareOutAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearShareOutAmount() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get tokenInMaxAmount => $_getSZ(4);
  @$pb.TagNumber(5)
  set tokenInMaxAmount($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasTokenInMaxAmount() => $_has(4);
  @$pb.TagNumber(5)
  void clearTokenInMaxAmount() => clearField(5);
}

class MsgJoinSwapShareAmountOutResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgJoinSwapShareAmountOutResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgJoinSwapShareAmountOutResponse._() : super();
  factory MsgJoinSwapShareAmountOutResponse() => create();
  factory MsgJoinSwapShareAmountOutResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgJoinSwapShareAmountOutResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgJoinSwapShareAmountOutResponse clone() => MsgJoinSwapShareAmountOutResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgJoinSwapShareAmountOutResponse copyWith(void Function(MsgJoinSwapShareAmountOutResponse) updates) => super.copyWith((message) => updates(message as MsgJoinSwapShareAmountOutResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgJoinSwapShareAmountOutResponse create() => MsgJoinSwapShareAmountOutResponse._();
  MsgJoinSwapShareAmountOutResponse createEmptyInstance() => create();
  static $pb.PbList<MsgJoinSwapShareAmountOutResponse> createRepeated() => $pb.PbList<MsgJoinSwapShareAmountOutResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgJoinSwapShareAmountOutResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgJoinSwapShareAmountOutResponse>(create);
  static MsgJoinSwapShareAmountOutResponse _defaultInstance;
}

class MsgExitSwapShareAmountIn extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgExitSwapShareAmountIn', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenOutDenom', protoName: 'tokenOutDenom')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'shareInAmount', protoName: 'shareInAmount')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenOutMinAmount', protoName: 'tokenOutMinAmount')
    ..hasRequiredFields = false
  ;

  MsgExitSwapShareAmountIn._() : super();
  factory MsgExitSwapShareAmountIn() => create();
  factory MsgExitSwapShareAmountIn.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgExitSwapShareAmountIn.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgExitSwapShareAmountIn clone() => MsgExitSwapShareAmountIn()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgExitSwapShareAmountIn copyWith(void Function(MsgExitSwapShareAmountIn) updates) => super.copyWith((message) => updates(message as MsgExitSwapShareAmountIn)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgExitSwapShareAmountIn create() => MsgExitSwapShareAmountIn._();
  MsgExitSwapShareAmountIn createEmptyInstance() => create();
  static $pb.PbList<MsgExitSwapShareAmountIn> createRepeated() => $pb.PbList<MsgExitSwapShareAmountIn>();
  @$core.pragma('dart2js:noInline')
  static MsgExitSwapShareAmountIn getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgExitSwapShareAmountIn>(create);
  static MsgExitSwapShareAmountIn _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sender => $_getSZ(0);
  @$pb.TagNumber(1)
  set sender($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSender() => $_has(0);
  @$pb.TagNumber(1)
  void clearSender() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get poolId => $_getI64(1);
  @$pb.TagNumber(2)
  set poolId($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPoolId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPoolId() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get tokenOutDenom => $_getSZ(2);
  @$pb.TagNumber(3)
  set tokenOutDenom($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTokenOutDenom() => $_has(2);
  @$pb.TagNumber(3)
  void clearTokenOutDenom() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get shareInAmount => $_getSZ(3);
  @$pb.TagNumber(4)
  set shareInAmount($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasShareInAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearShareInAmount() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get tokenOutMinAmount => $_getSZ(4);
  @$pb.TagNumber(5)
  set tokenOutMinAmount($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasTokenOutMinAmount() => $_has(4);
  @$pb.TagNumber(5)
  void clearTokenOutMinAmount() => clearField(5);
}

class MsgExitSwapShareAmountInResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgExitSwapShareAmountInResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgExitSwapShareAmountInResponse._() : super();
  factory MsgExitSwapShareAmountInResponse() => create();
  factory MsgExitSwapShareAmountInResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgExitSwapShareAmountInResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgExitSwapShareAmountInResponse clone() => MsgExitSwapShareAmountInResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgExitSwapShareAmountInResponse copyWith(void Function(MsgExitSwapShareAmountInResponse) updates) => super.copyWith((message) => updates(message as MsgExitSwapShareAmountInResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgExitSwapShareAmountInResponse create() => MsgExitSwapShareAmountInResponse._();
  MsgExitSwapShareAmountInResponse createEmptyInstance() => create();
  static $pb.PbList<MsgExitSwapShareAmountInResponse> createRepeated() => $pb.PbList<MsgExitSwapShareAmountInResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgExitSwapShareAmountInResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgExitSwapShareAmountInResponse>(create);
  static MsgExitSwapShareAmountInResponse _defaultInstance;
}

class MsgExitSwapExternAmountOut extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgExitSwapExternAmountOut', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOM<$2.Coin>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenOut', protoName: 'tokenOut', subBuilder: $2.Coin.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'shareInMaxAmount', protoName: 'shareInMaxAmount')
    ..hasRequiredFields = false
  ;

  MsgExitSwapExternAmountOut._() : super();
  factory MsgExitSwapExternAmountOut() => create();
  factory MsgExitSwapExternAmountOut.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgExitSwapExternAmountOut.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgExitSwapExternAmountOut clone() => MsgExitSwapExternAmountOut()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgExitSwapExternAmountOut copyWith(void Function(MsgExitSwapExternAmountOut) updates) => super.copyWith((message) => updates(message as MsgExitSwapExternAmountOut)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgExitSwapExternAmountOut create() => MsgExitSwapExternAmountOut._();
  MsgExitSwapExternAmountOut createEmptyInstance() => create();
  static $pb.PbList<MsgExitSwapExternAmountOut> createRepeated() => $pb.PbList<MsgExitSwapExternAmountOut>();
  @$core.pragma('dart2js:noInline')
  static MsgExitSwapExternAmountOut getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgExitSwapExternAmountOut>(create);
  static MsgExitSwapExternAmountOut _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sender => $_getSZ(0);
  @$pb.TagNumber(1)
  set sender($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSender() => $_has(0);
  @$pb.TagNumber(1)
  void clearSender() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get poolId => $_getI64(1);
  @$pb.TagNumber(2)
  set poolId($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPoolId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPoolId() => clearField(2);

  @$pb.TagNumber(3)
  $2.Coin get tokenOut => $_getN(2);
  @$pb.TagNumber(3)
  set tokenOut($2.Coin v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasTokenOut() => $_has(2);
  @$pb.TagNumber(3)
  void clearTokenOut() => clearField(3);
  @$pb.TagNumber(3)
  $2.Coin ensureTokenOut() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.String get shareInMaxAmount => $_getSZ(3);
  @$pb.TagNumber(4)
  set shareInMaxAmount($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasShareInMaxAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearShareInMaxAmount() => clearField(4);
}

class MsgExitSwapExternAmountOutResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgExitSwapExternAmountOutResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgExitSwapExternAmountOutResponse._() : super();
  factory MsgExitSwapExternAmountOutResponse() => create();
  factory MsgExitSwapExternAmountOutResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgExitSwapExternAmountOutResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgExitSwapExternAmountOutResponse clone() => MsgExitSwapExternAmountOutResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgExitSwapExternAmountOutResponse copyWith(void Function(MsgExitSwapExternAmountOutResponse) updates) => super.copyWith((message) => updates(message as MsgExitSwapExternAmountOutResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgExitSwapExternAmountOutResponse create() => MsgExitSwapExternAmountOutResponse._();
  MsgExitSwapExternAmountOutResponse createEmptyInstance() => create();
  static $pb.PbList<MsgExitSwapExternAmountOutResponse> createRepeated() => $pb.PbList<MsgExitSwapExternAmountOutResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgExitSwapExternAmountOutResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgExitSwapExternAmountOutResponse>(create);
  static MsgExitSwapExternAmountOutResponse _defaultInstance;
}

