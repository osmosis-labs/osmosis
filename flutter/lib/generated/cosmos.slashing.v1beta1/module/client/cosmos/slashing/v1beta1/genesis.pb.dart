///
//  Generated code. Do not modify.
//  source: cosmos/slashing/v1beta1/genesis.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'slashing.pb.dart' as $4;

class GenesisState extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GenesisState', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.slashing.v1beta1'), createEmptyInstance: create)
    ..aOM<$4.Params>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'params', subBuilder: $4.Params.create)
    ..pc<SigningInfo>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signingInfos', $pb.PbFieldType.PM, subBuilder: SigningInfo.create)
    ..pc<ValidatorMissedBlocks>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'missedBlocks', $pb.PbFieldType.PM, subBuilder: ValidatorMissedBlocks.create)
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
  $4.Params get params => $_getN(0);
  @$pb.TagNumber(1)
  set params($4.Params v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasParams() => $_has(0);
  @$pb.TagNumber(1)
  void clearParams() => clearField(1);
  @$pb.TagNumber(1)
  $4.Params ensureParams() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<SigningInfo> get signingInfos => $_getList(1);

  @$pb.TagNumber(3)
  $core.List<ValidatorMissedBlocks> get missedBlocks => $_getList(2);
}

class SigningInfo extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SigningInfo', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.slashing.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'address')
    ..aOM<$4.ValidatorSigningInfo>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorSigningInfo', subBuilder: $4.ValidatorSigningInfo.create)
    ..hasRequiredFields = false
  ;

  SigningInfo._() : super();
  factory SigningInfo() => create();
  factory SigningInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SigningInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SigningInfo clone() => SigningInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SigningInfo copyWith(void Function(SigningInfo) updates) => super.copyWith((message) => updates(message as SigningInfo)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static SigningInfo create() => SigningInfo._();
  SigningInfo createEmptyInstance() => create();
  static $pb.PbList<SigningInfo> createRepeated() => $pb.PbList<SigningInfo>();
  @$core.pragma('dart2js:noInline')
  static SigningInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SigningInfo>(create);
  static SigningInfo _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get address => $_getSZ(0);
  @$pb.TagNumber(1)
  set address($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => clearField(1);

  @$pb.TagNumber(2)
  $4.ValidatorSigningInfo get validatorSigningInfo => $_getN(1);
  @$pb.TagNumber(2)
  set validatorSigningInfo($4.ValidatorSigningInfo v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasValidatorSigningInfo() => $_has(1);
  @$pb.TagNumber(2)
  void clearValidatorSigningInfo() => clearField(2);
  @$pb.TagNumber(2)
  $4.ValidatorSigningInfo ensureValidatorSigningInfo() => $_ensure(1);
}

class ValidatorMissedBlocks extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorMissedBlocks', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.slashing.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'address')
    ..pc<MissedBlock>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'missedBlocks', $pb.PbFieldType.PM, subBuilder: MissedBlock.create)
    ..hasRequiredFields = false
  ;

  ValidatorMissedBlocks._() : super();
  factory ValidatorMissedBlocks() => create();
  factory ValidatorMissedBlocks.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorMissedBlocks.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorMissedBlocks clone() => ValidatorMissedBlocks()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorMissedBlocks copyWith(void Function(ValidatorMissedBlocks) updates) => super.copyWith((message) => updates(message as ValidatorMissedBlocks)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorMissedBlocks create() => ValidatorMissedBlocks._();
  ValidatorMissedBlocks createEmptyInstance() => create();
  static $pb.PbList<ValidatorMissedBlocks> createRepeated() => $pb.PbList<ValidatorMissedBlocks>();
  @$core.pragma('dart2js:noInline')
  static ValidatorMissedBlocks getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorMissedBlocks>(create);
  static ValidatorMissedBlocks _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get address => $_getSZ(0);
  @$pb.TagNumber(1)
  set address($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<MissedBlock> get missedBlocks => $_getList(1);
}

class MissedBlock extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MissedBlock', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.slashing.v1beta1'), createEmptyInstance: create)
    ..aInt64(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'index')
    ..aOB(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'missed')
    ..hasRequiredFields = false
  ;

  MissedBlock._() : super();
  factory MissedBlock() => create();
  factory MissedBlock.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MissedBlock.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MissedBlock clone() => MissedBlock()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MissedBlock copyWith(void Function(MissedBlock) updates) => super.copyWith((message) => updates(message as MissedBlock)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MissedBlock create() => MissedBlock._();
  MissedBlock createEmptyInstance() => create();
  static $pb.PbList<MissedBlock> createRepeated() => $pb.PbList<MissedBlock>();
  @$core.pragma('dart2js:noInline')
  static MissedBlock getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MissedBlock>(create);
  static MissedBlock _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get index => $_getI64(0);
  @$pb.TagNumber(1)
  set index($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasIndex() => $_has(0);
  @$pb.TagNumber(1)
  void clearIndex() => clearField(1);

  @$pb.TagNumber(2)
  $core.bool get missed => $_getBF(1);
  @$pb.TagNumber(2)
  set missed($core.bool v) { $_setBool(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasMissed() => $_has(1);
  @$pb.TagNumber(2)
  void clearMissed() => clearField(2);
}

