///
//  Generated code. Do not modify.
//  source: osmosis/incentives/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../lockup/lock.pb.dart' as $5;
import '../../cosmos/base/v1beta1/coin.pb.dart' as $4;
import '../../google/protobuf/timestamp.pb.dart' as $3;

class MsgCreateGauge extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgCreateGauge', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'isPerpetual')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..aOM<$5.QueryCondition>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'distributeTo', subBuilder: $5.QueryCondition.create)
    ..pc<$4.Coin>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'coins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..aOM<$3.Timestamp>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'startTime', subBuilder: $3.Timestamp.create)
    ..a<$fixnum.Int64>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'numEpochsPaidOver', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  MsgCreateGauge._() : super();
  factory MsgCreateGauge() => create();
  factory MsgCreateGauge.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgCreateGauge.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgCreateGauge clone() => MsgCreateGauge()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgCreateGauge copyWith(void Function(MsgCreateGauge) updates) => super.copyWith((message) => updates(message as MsgCreateGauge)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgCreateGauge create() => MsgCreateGauge._();
  MsgCreateGauge createEmptyInstance() => create();
  static $pb.PbList<MsgCreateGauge> createRepeated() => $pb.PbList<MsgCreateGauge>();
  @$core.pragma('dart2js:noInline')
  static MsgCreateGauge getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgCreateGauge>(create);
  static MsgCreateGauge _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get isPerpetual => $_getBF(0);
  @$pb.TagNumber(1)
  set isPerpetual($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasIsPerpetual() => $_has(0);
  @$pb.TagNumber(1)
  void clearIsPerpetual() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get owner => $_getSZ(1);
  @$pb.TagNumber(2)
  set owner($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasOwner() => $_has(1);
  @$pb.TagNumber(2)
  void clearOwner() => clearField(2);

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
}

class MsgCreateGaugeResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgCreateGaugeResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgCreateGaugeResponse._() : super();
  factory MsgCreateGaugeResponse() => create();
  factory MsgCreateGaugeResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgCreateGaugeResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgCreateGaugeResponse clone() => MsgCreateGaugeResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgCreateGaugeResponse copyWith(void Function(MsgCreateGaugeResponse) updates) => super.copyWith((message) => updates(message as MsgCreateGaugeResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgCreateGaugeResponse create() => MsgCreateGaugeResponse._();
  MsgCreateGaugeResponse createEmptyInstance() => create();
  static $pb.PbList<MsgCreateGaugeResponse> createRepeated() => $pb.PbList<MsgCreateGaugeResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgCreateGaugeResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgCreateGaugeResponse>(create);
  static MsgCreateGaugeResponse _defaultInstance;
}

class MsgAddToGauge extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgAddToGauge', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'gaugeId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..pc<$4.Coin>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'rewards', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..hasRequiredFields = false
  ;

  MsgAddToGauge._() : super();
  factory MsgAddToGauge() => create();
  factory MsgAddToGauge.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgAddToGauge.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgAddToGauge clone() => MsgAddToGauge()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgAddToGauge copyWith(void Function(MsgAddToGauge) updates) => super.copyWith((message) => updates(message as MsgAddToGauge)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgAddToGauge create() => MsgAddToGauge._();
  MsgAddToGauge createEmptyInstance() => create();
  static $pb.PbList<MsgAddToGauge> createRepeated() => $pb.PbList<MsgAddToGauge>();
  @$core.pragma('dart2js:noInline')
  static MsgAddToGauge getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgAddToGauge>(create);
  static MsgAddToGauge _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get gaugeId => $_getI64(1);
  @$pb.TagNumber(2)
  set gaugeId($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasGaugeId() => $_has(1);
  @$pb.TagNumber(2)
  void clearGaugeId() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$4.Coin> get rewards => $_getList(2);
}

class MsgAddToGaugeResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgAddToGaugeResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgAddToGaugeResponse._() : super();
  factory MsgAddToGaugeResponse() => create();
  factory MsgAddToGaugeResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgAddToGaugeResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgAddToGaugeResponse clone() => MsgAddToGaugeResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgAddToGaugeResponse copyWith(void Function(MsgAddToGaugeResponse) updates) => super.copyWith((message) => updates(message as MsgAddToGaugeResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgAddToGaugeResponse create() => MsgAddToGaugeResponse._();
  MsgAddToGaugeResponse createEmptyInstance() => create();
  static $pb.PbList<MsgAddToGaugeResponse> createRepeated() => $pb.PbList<MsgAddToGaugeResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgAddToGaugeResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgAddToGaugeResponse>(create);
  static MsgAddToGaugeResponse _defaultInstance;
}

