///
//  Generated code. Do not modify.
//  source: osmosis/gamm/v1beta1/pool.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../../cosmos/base/v1beta1/coin.pb.dart' as $2;
import '../../../google/protobuf/timestamp.pb.dart' as $4;
import '../../../google/protobuf/duration.pb.dart' as $5;

class PoolAsset extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'PoolAsset', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOM<$2.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'token', subBuilder: $2.Coin.create)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'weight')
    ..hasRequiredFields = false
  ;

  PoolAsset._() : super();
  factory PoolAsset() => create();
  factory PoolAsset.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PoolAsset.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PoolAsset clone() => PoolAsset()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PoolAsset copyWith(void Function(PoolAsset) updates) => super.copyWith((message) => updates(message as PoolAsset)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static PoolAsset create() => PoolAsset._();
  PoolAsset createEmptyInstance() => create();
  static $pb.PbList<PoolAsset> createRepeated() => $pb.PbList<PoolAsset>();
  @$core.pragma('dart2js:noInline')
  static PoolAsset getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PoolAsset>(create);
  static PoolAsset _defaultInstance;

  @$pb.TagNumber(1)
  $2.Coin get token => $_getN(0);
  @$pb.TagNumber(1)
  set token($2.Coin v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => clearField(1);
  @$pb.TagNumber(1)
  $2.Coin ensureToken() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.String get weight => $_getSZ(1);
  @$pb.TagNumber(2)
  set weight($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasWeight() => $_has(1);
  @$pb.TagNumber(2)
  void clearWeight() => clearField(2);
}

class SmoothWeightChangeParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SmoothWeightChangeParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOM<$4.Timestamp>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'startTime', subBuilder: $4.Timestamp.create)
    ..aOM<$5.Duration>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'duration', subBuilder: $5.Duration.create)
    ..pc<PoolAsset>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'initialPoolWeights', $pb.PbFieldType.PM, protoName: 'initialPoolWeights', subBuilder: PoolAsset.create)
    ..pc<PoolAsset>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'targetPoolWeights', $pb.PbFieldType.PM, protoName: 'targetPoolWeights', subBuilder: PoolAsset.create)
    ..hasRequiredFields = false
  ;

  SmoothWeightChangeParams._() : super();
  factory SmoothWeightChangeParams() => create();
  factory SmoothWeightChangeParams.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SmoothWeightChangeParams.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SmoothWeightChangeParams clone() => SmoothWeightChangeParams()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SmoothWeightChangeParams copyWith(void Function(SmoothWeightChangeParams) updates) => super.copyWith((message) => updates(message as SmoothWeightChangeParams)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static SmoothWeightChangeParams create() => SmoothWeightChangeParams._();
  SmoothWeightChangeParams createEmptyInstance() => create();
  static $pb.PbList<SmoothWeightChangeParams> createRepeated() => $pb.PbList<SmoothWeightChangeParams>();
  @$core.pragma('dart2js:noInline')
  static SmoothWeightChangeParams getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SmoothWeightChangeParams>(create);
  static SmoothWeightChangeParams _defaultInstance;

  @$pb.TagNumber(1)
  $4.Timestamp get startTime => $_getN(0);
  @$pb.TagNumber(1)
  set startTime($4.Timestamp v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasStartTime() => $_has(0);
  @$pb.TagNumber(1)
  void clearStartTime() => clearField(1);
  @$pb.TagNumber(1)
  $4.Timestamp ensureStartTime() => $_ensure(0);

  @$pb.TagNumber(2)
  $5.Duration get duration => $_getN(1);
  @$pb.TagNumber(2)
  set duration($5.Duration v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasDuration() => $_has(1);
  @$pb.TagNumber(2)
  void clearDuration() => clearField(2);
  @$pb.TagNumber(2)
  $5.Duration ensureDuration() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.List<PoolAsset> get initialPoolWeights => $_getList(2);

  @$pb.TagNumber(4)
  $core.List<PoolAsset> get targetPoolWeights => $_getList(3);
}

class PoolParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'PoolParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'swapFee', protoName: 'swapFee')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'exitFee', protoName: 'exitFee')
    ..aOM<SmoothWeightChangeParams>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'smoothWeightChangeParams', protoName: 'smoothWeightChangeParams', subBuilder: SmoothWeightChangeParams.create)
    ..hasRequiredFields = false
  ;

  PoolParams._() : super();
  factory PoolParams() => create();
  factory PoolParams.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PoolParams.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PoolParams clone() => PoolParams()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PoolParams copyWith(void Function(PoolParams) updates) => super.copyWith((message) => updates(message as PoolParams)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static PoolParams create() => PoolParams._();
  PoolParams createEmptyInstance() => create();
  static $pb.PbList<PoolParams> createRepeated() => $pb.PbList<PoolParams>();
  @$core.pragma('dart2js:noInline')
  static PoolParams getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PoolParams>(create);
  static PoolParams _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get swapFee => $_getSZ(0);
  @$pb.TagNumber(1)
  set swapFee($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSwapFee() => $_has(0);
  @$pb.TagNumber(1)
  void clearSwapFee() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get exitFee => $_getSZ(1);
  @$pb.TagNumber(2)
  set exitFee($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasExitFee() => $_has(1);
  @$pb.TagNumber(2)
  void clearExitFee() => clearField(2);

  @$pb.TagNumber(3)
  SmoothWeightChangeParams get smoothWeightChangeParams => $_getN(2);
  @$pb.TagNumber(3)
  set smoothWeightChangeParams(SmoothWeightChangeParams v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasSmoothWeightChangeParams() => $_has(2);
  @$pb.TagNumber(3)
  void clearSmoothWeightChangeParams() => clearField(3);
  @$pb.TagNumber(3)
  SmoothWeightChangeParams ensureSmoothWeightChangeParams() => $_ensure(2);
}

class Pool extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Pool', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'address')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOM<PoolParams>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolParams', protoName: 'poolParams', subBuilder: PoolParams.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'futurePoolGovernor')
    ..aOM<$2.Coin>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'totalShares', protoName: 'totalShares', subBuilder: $2.Coin.create)
    ..pc<PoolAsset>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolAssets', $pb.PbFieldType.PM, protoName: 'poolAssets', subBuilder: PoolAsset.create)
    ..aOS(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'totalWeight', protoName: 'totalWeight')
    ..hasRequiredFields = false
  ;

  Pool._() : super();
  factory Pool() => create();
  factory Pool.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Pool.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Pool clone() => Pool()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Pool copyWith(void Function(Pool) updates) => super.copyWith((message) => updates(message as Pool)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Pool create() => Pool._();
  Pool createEmptyInstance() => create();
  static $pb.PbList<Pool> createRepeated() => $pb.PbList<Pool>();
  @$core.pragma('dart2js:noInline')
  static Pool getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Pool>(create);
  static Pool _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get address => $_getSZ(0);
  @$pb.TagNumber(1)
  set address($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get id => $_getI64(1);
  @$pb.TagNumber(2)
  set id($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasId() => $_has(1);
  @$pb.TagNumber(2)
  void clearId() => clearField(2);

  @$pb.TagNumber(3)
  PoolParams get poolParams => $_getN(2);
  @$pb.TagNumber(3)
  set poolParams(PoolParams v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasPoolParams() => $_has(2);
  @$pb.TagNumber(3)
  void clearPoolParams() => clearField(3);
  @$pb.TagNumber(3)
  PoolParams ensurePoolParams() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.String get futurePoolGovernor => $_getSZ(3);
  @$pb.TagNumber(4)
  set futurePoolGovernor($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasFuturePoolGovernor() => $_has(3);
  @$pb.TagNumber(4)
  void clearFuturePoolGovernor() => clearField(4);

  @$pb.TagNumber(5)
  $2.Coin get totalShares => $_getN(4);
  @$pb.TagNumber(5)
  set totalShares($2.Coin v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasTotalShares() => $_has(4);
  @$pb.TagNumber(5)
  void clearTotalShares() => clearField(5);
  @$pb.TagNumber(5)
  $2.Coin ensureTotalShares() => $_ensure(4);

  @$pb.TagNumber(6)
  $core.List<PoolAsset> get poolAssets => $_getList(5);

  @$pb.TagNumber(7)
  $core.String get totalWeight => $_getSZ(6);
  @$pb.TagNumber(7)
  set totalWeight($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasTotalWeight() => $_has(6);
  @$pb.TagNumber(7)
  void clearTotalWeight() => clearField(7);
}

