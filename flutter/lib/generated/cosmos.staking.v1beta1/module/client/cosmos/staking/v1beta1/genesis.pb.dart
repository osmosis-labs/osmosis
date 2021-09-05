///
//  Generated code. Do not modify.
//  source: cosmos/staking/v1beta1/genesis.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'staking.pb.dart' as $11;

class GenesisState extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GenesisState', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOM<$11.Params>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'params', subBuilder: $11.Params.create)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lastTotalPower', $pb.PbFieldType.OY)
    ..pc<LastValidatorPower>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lastValidatorPowers', $pb.PbFieldType.PM, subBuilder: LastValidatorPower.create)
    ..pc<$11.Validator>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validators', $pb.PbFieldType.PM, subBuilder: $11.Validator.create)
    ..pc<$11.Delegation>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegations', $pb.PbFieldType.PM, subBuilder: $11.Delegation.create)
    ..pc<$11.UnbondingDelegation>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'unbondingDelegations', $pb.PbFieldType.PM, subBuilder: $11.UnbondingDelegation.create)
    ..pc<$11.Redelegation>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'redelegations', $pb.PbFieldType.PM, subBuilder: $11.Redelegation.create)
    ..aOB(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'exported')
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
  $11.Params get params => $_getN(0);
  @$pb.TagNumber(1)
  set params($11.Params v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasParams() => $_has(0);
  @$pb.TagNumber(1)
  void clearParams() => clearField(1);
  @$pb.TagNumber(1)
  $11.Params ensureParams() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<$core.int> get lastTotalPower => $_getN(1);
  @$pb.TagNumber(2)
  set lastTotalPower($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasLastTotalPower() => $_has(1);
  @$pb.TagNumber(2)
  void clearLastTotalPower() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<LastValidatorPower> get lastValidatorPowers => $_getList(2);

  @$pb.TagNumber(4)
  $core.List<$11.Validator> get validators => $_getList(3);

  @$pb.TagNumber(5)
  $core.List<$11.Delegation> get delegations => $_getList(4);

  @$pb.TagNumber(6)
  $core.List<$11.UnbondingDelegation> get unbondingDelegations => $_getList(5);

  @$pb.TagNumber(7)
  $core.List<$11.Redelegation> get redelegations => $_getList(6);

  @$pb.TagNumber(8)
  $core.bool get exported => $_getBF(7);
  @$pb.TagNumber(8)
  set exported($core.bool v) { $_setBool(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasExported() => $_has(7);
  @$pb.TagNumber(8)
  void clearExported() => clearField(8);
}

class LastValidatorPower extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LastValidatorPower', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'address')
    ..aInt64(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'power')
    ..hasRequiredFields = false
  ;

  LastValidatorPower._() : super();
  factory LastValidatorPower() => create();
  factory LastValidatorPower.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LastValidatorPower.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LastValidatorPower clone() => LastValidatorPower()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LastValidatorPower copyWith(void Function(LastValidatorPower) updates) => super.copyWith((message) => updates(message as LastValidatorPower)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static LastValidatorPower create() => LastValidatorPower._();
  LastValidatorPower createEmptyInstance() => create();
  static $pb.PbList<LastValidatorPower> createRepeated() => $pb.PbList<LastValidatorPower>();
  @$core.pragma('dart2js:noInline')
  static LastValidatorPower getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LastValidatorPower>(create);
  static LastValidatorPower _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get address => $_getSZ(0);
  @$pb.TagNumber(1)
  set address($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get power => $_getI64(1);
  @$pb.TagNumber(2)
  set power($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPower() => $_has(1);
  @$pb.TagNumber(2)
  void clearPower() => clearField(2);
}

