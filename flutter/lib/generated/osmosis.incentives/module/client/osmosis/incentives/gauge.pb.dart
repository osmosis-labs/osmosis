///
//  Generated code. Do not modify.
//  source: osmosis/incentives/gauge.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../lockup/lock.pb.dart' as $5;
import '../../cosmos/base/v1beta1/coin.pb.dart' as $4;
import '../../google/protobuf/timestamp.pb.dart' as $3;
import '../../google/protobuf/duration.pb.dart' as $2;

class Gauge extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Gauge', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOB(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'isPerpetual')
    ..aOM<$5.QueryCondition>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'distributeTo', subBuilder: $5.QueryCondition.create)
    ..pc<$4.Coin>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'coins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..aOM<$3.Timestamp>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'startTime', subBuilder: $3.Timestamp.create)
    ..a<$fixnum.Int64>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'numEpochsPaidOver', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'filledEpochs', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..pc<$4.Coin>(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'distributedCoins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..hasRequiredFields = false
  ;

  Gauge._() : super();
  factory Gauge() => create();
  factory Gauge.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Gauge.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Gauge clone() => Gauge()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Gauge copyWith(void Function(Gauge) updates) => super.copyWith((message) => updates(message as Gauge)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Gauge create() => Gauge._();
  Gauge createEmptyInstance() => create();
  static $pb.PbList<Gauge> createRepeated() => $pb.PbList<Gauge>();
  @$core.pragma('dart2js:noInline')
  static Gauge getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Gauge>(create);
  static Gauge _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get id => $_getI64(0);
  @$pb.TagNumber(1)
  set id($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.bool get isPerpetual => $_getBF(1);
  @$pb.TagNumber(2)
  set isPerpetual($core.bool v) { $_setBool(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasIsPerpetual() => $_has(1);
  @$pb.TagNumber(2)
  void clearIsPerpetual() => clearField(2);

  @$pb.TagNumber(3)
  $5.QueryCondition get distributeTo => $_getN(2);
  @$pb.TagNumber(3)
  set distributeTo($5.QueryCondition v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasDistributeTo() => $_has(2);
  @$pb.TagNumber(3)
  void clearDistributeTo() => clearField(3);
  @$pb.TagNumber(3)
  $5.QueryCondition ensureDistributeTo() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.List<$4.Coin> get coins => $_getList(3);

  @$pb.TagNumber(5)
  $3.Timestamp get startTime => $_getN(4);
  @$pb.TagNumber(5)
  set startTime($3.Timestamp v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasStartTime() => $_has(4);
  @$pb.TagNumber(5)
  void clearStartTime() => clearField(5);
  @$pb.TagNumber(5)
  $3.Timestamp ensureStartTime() => $_ensure(4);

  @$pb.TagNumber(6)
  $fixnum.Int64 get numEpochsPaidOver => $_getI64(5);
  @$pb.TagNumber(6)
  set numEpochsPaidOver($fixnum.Int64 v) { $_setInt64(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasNumEpochsPaidOver() => $_has(5);
  @$pb.TagNumber(6)
  void clearNumEpochsPaidOver() => clearField(6);

  @$pb.TagNumber(7)
  $fixnum.Int64 get filledEpochs => $_getI64(6);
  @$pb.TagNumber(7)
  set filledEpochs($fixnum.Int64 v) { $_setInt64(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasFilledEpochs() => $_has(6);
  @$pb.TagNumber(7)
  void clearFilledEpochs() => clearField(7);

  @$pb.TagNumber(8)
  $core.List<$4.Coin> get distributedCoins => $_getList(7);
}

class LockableDurationsInfo extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LockableDurationsInfo', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..pc<$2.Duration>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lockableDurations', $pb.PbFieldType.PM, subBuilder: $2.Duration.create)
    ..hasRequiredFields = false
  ;

  LockableDurationsInfo._() : super();
  factory LockableDurationsInfo() => create();
  factory LockableDurationsInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LockableDurationsInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LockableDurationsInfo clone() => LockableDurationsInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LockableDurationsInfo copyWith(void Function(LockableDurationsInfo) updates) => super.copyWith((message) => updates(message as LockableDurationsInfo)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static LockableDurationsInfo create() => LockableDurationsInfo._();
  LockableDurationsInfo createEmptyInstance() => create();
  static $pb.PbList<LockableDurationsInfo> createRepeated() => $pb.PbList<LockableDurationsInfo>();
  @$core.pragma('dart2js:noInline')
  static LockableDurationsInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LockableDurationsInfo>(create);
  static LockableDurationsInfo _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.Duration> get lockableDurations => $_getList(0);
}

