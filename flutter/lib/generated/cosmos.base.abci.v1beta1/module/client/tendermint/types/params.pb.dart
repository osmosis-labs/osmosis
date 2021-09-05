///
//  Generated code. Do not modify.
//  source: tendermint/types/params.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../google/protobuf/duration.pb.dart' as $6;

class ConsensusParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ConsensusParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.types'), createEmptyInstance: create)
    ..aOM<BlockParams>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'block', subBuilder: BlockParams.create)
    ..aOM<EvidenceParams>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'evidence', subBuilder: EvidenceParams.create)
    ..aOM<ValidatorParams>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validator', subBuilder: ValidatorParams.create)
    ..aOM<VersionParams>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'version', subBuilder: VersionParams.create)
    ..hasRequiredFields = false
  ;

  ConsensusParams._() : super();
  factory ConsensusParams() => create();
  factory ConsensusParams.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConsensusParams.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConsensusParams clone() => ConsensusParams()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConsensusParams copyWith(void Function(ConsensusParams) updates) => super.copyWith((message) => updates(message as ConsensusParams)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ConsensusParams create() => ConsensusParams._();
  ConsensusParams createEmptyInstance() => create();
  static $pb.PbList<ConsensusParams> createRepeated() => $pb.PbList<ConsensusParams>();
  @$core.pragma('dart2js:noInline')
  static ConsensusParams getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConsensusParams>(create);
  static ConsensusParams _defaultInstance;

  @$pb.TagNumber(1)
  BlockParams get block => $_getN(0);
  @$pb.TagNumber(1)
  set block(BlockParams v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasBlock() => $_has(0);
  @$pb.TagNumber(1)
  void clearBlock() => clearField(1);
  @$pb.TagNumber(1)
  BlockParams ensureBlock() => $_ensure(0);

  @$pb.TagNumber(2)
  EvidenceParams get evidence => $_getN(1);
  @$pb.TagNumber(2)
  set evidence(EvidenceParams v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasEvidence() => $_has(1);
  @$pb.TagNumber(2)
  void clearEvidence() => clearField(2);
  @$pb.TagNumber(2)
  EvidenceParams ensureEvidence() => $_ensure(1);

  @$pb.TagNumber(3)
  ValidatorParams get validator => $_getN(2);
  @$pb.TagNumber(3)
  set validator(ValidatorParams v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasValidator() => $_has(2);
  @$pb.TagNumber(3)
  void clearValidator() => clearField(3);
  @$pb.TagNumber(3)
  ValidatorParams ensureValidator() => $_ensure(2);

  @$pb.TagNumber(4)
  VersionParams get version => $_getN(3);
  @$pb.TagNumber(4)
  set version(VersionParams v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasVersion() => $_has(3);
  @$pb.TagNumber(4)
  void clearVersion() => clearField(4);
  @$pb.TagNumber(4)
  VersionParams ensureVersion() => $_ensure(3);
}

class BlockParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'BlockParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.types'), createEmptyInstance: create)
    ..aInt64(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxBytes')
    ..aInt64(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxGas')
    ..aInt64(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'timeIotaMs')
    ..hasRequiredFields = false
  ;

  BlockParams._() : super();
  factory BlockParams() => create();
  factory BlockParams.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory BlockParams.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  BlockParams clone() => BlockParams()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  BlockParams copyWith(void Function(BlockParams) updates) => super.copyWith((message) => updates(message as BlockParams)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static BlockParams create() => BlockParams._();
  BlockParams createEmptyInstance() => create();
  static $pb.PbList<BlockParams> createRepeated() => $pb.PbList<BlockParams>();
  @$core.pragma('dart2js:noInline')
  static BlockParams getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<BlockParams>(create);
  static BlockParams _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get maxBytes => $_getI64(0);
  @$pb.TagNumber(1)
  set maxBytes($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMaxBytes() => $_has(0);
  @$pb.TagNumber(1)
  void clearMaxBytes() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get maxGas => $_getI64(1);
  @$pb.TagNumber(2)
  set maxGas($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasMaxGas() => $_has(1);
  @$pb.TagNumber(2)
  void clearMaxGas() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get timeIotaMs => $_getI64(2);
  @$pb.TagNumber(3)
  set timeIotaMs($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTimeIotaMs() => $_has(2);
  @$pb.TagNumber(3)
  void clearTimeIotaMs() => clearField(3);
}

class EvidenceParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'EvidenceParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.types'), createEmptyInstance: create)
    ..aInt64(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxAgeNumBlocks')
    ..aOM<$6.Duration>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxAgeDuration', subBuilder: $6.Duration.create)
    ..aInt64(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxBytes')
    ..hasRequiredFields = false
  ;

  EvidenceParams._() : super();
  factory EvidenceParams() => create();
  factory EvidenceParams.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory EvidenceParams.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  EvidenceParams clone() => EvidenceParams()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  EvidenceParams copyWith(void Function(EvidenceParams) updates) => super.copyWith((message) => updates(message as EvidenceParams)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static EvidenceParams create() => EvidenceParams._();
  EvidenceParams createEmptyInstance() => create();
  static $pb.PbList<EvidenceParams> createRepeated() => $pb.PbList<EvidenceParams>();
  @$core.pragma('dart2js:noInline')
  static EvidenceParams getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<EvidenceParams>(create);
  static EvidenceParams _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get maxAgeNumBlocks => $_getI64(0);
  @$pb.TagNumber(1)
  set maxAgeNumBlocks($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMaxAgeNumBlocks() => $_has(0);
  @$pb.TagNumber(1)
  void clearMaxAgeNumBlocks() => clearField(1);

  @$pb.TagNumber(2)
  $6.Duration get maxAgeDuration => $_getN(1);
  @$pb.TagNumber(2)
  set maxAgeDuration($6.Duration v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasMaxAgeDuration() => $_has(1);
  @$pb.TagNumber(2)
  void clearMaxAgeDuration() => clearField(2);
  @$pb.TagNumber(2)
  $6.Duration ensureMaxAgeDuration() => $_ensure(1);

  @$pb.TagNumber(3)
  $fixnum.Int64 get maxBytes => $_getI64(2);
  @$pb.TagNumber(3)
  set maxBytes($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasMaxBytes() => $_has(2);
  @$pb.TagNumber(3)
  void clearMaxBytes() => clearField(3);
}

class ValidatorParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.types'), createEmptyInstance: create)
    ..pPS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pubKeyTypes')
    ..hasRequiredFields = false
  ;

  ValidatorParams._() : super();
  factory ValidatorParams() => create();
  factory ValidatorParams.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorParams.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorParams clone() => ValidatorParams()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorParams copyWith(void Function(ValidatorParams) updates) => super.copyWith((message) => updates(message as ValidatorParams)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorParams create() => ValidatorParams._();
  ValidatorParams createEmptyInstance() => create();
  static $pb.PbList<ValidatorParams> createRepeated() => $pb.PbList<ValidatorParams>();
  @$core.pragma('dart2js:noInline')
  static ValidatorParams getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorParams>(create);
  static ValidatorParams _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.String> get pubKeyTypes => $_getList(0);
}

class VersionParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'VersionParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.types'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'appVersion', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  VersionParams._() : super();
  factory VersionParams() => create();
  factory VersionParams.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory VersionParams.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  VersionParams clone() => VersionParams()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  VersionParams copyWith(void Function(VersionParams) updates) => super.copyWith((message) => updates(message as VersionParams)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static VersionParams create() => VersionParams._();
  VersionParams createEmptyInstance() => create();
  static $pb.PbList<VersionParams> createRepeated() => $pb.PbList<VersionParams>();
  @$core.pragma('dart2js:noInline')
  static VersionParams getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VersionParams>(create);
  static VersionParams _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get appVersion => $_getI64(0);
  @$pb.TagNumber(1)
  set appVersion($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAppVersion() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppVersion() => clearField(1);
}

class HashedParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'HashedParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.types'), createEmptyInstance: create)
    ..aInt64(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'blockMaxBytes')
    ..aInt64(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'blockMaxGas')
    ..hasRequiredFields = false
  ;

  HashedParams._() : super();
  factory HashedParams() => create();
  factory HashedParams.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory HashedParams.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  HashedParams clone() => HashedParams()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  HashedParams copyWith(void Function(HashedParams) updates) => super.copyWith((message) => updates(message as HashedParams)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static HashedParams create() => HashedParams._();
  HashedParams createEmptyInstance() => create();
  static $pb.PbList<HashedParams> createRepeated() => $pb.PbList<HashedParams>();
  @$core.pragma('dart2js:noInline')
  static HashedParams getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<HashedParams>(create);
  static HashedParams _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get blockMaxBytes => $_getI64(0);
  @$pb.TagNumber(1)
  set blockMaxBytes($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasBlockMaxBytes() => $_has(0);
  @$pb.TagNumber(1)
  void clearBlockMaxBytes() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get blockMaxGas => $_getI64(1);
  @$pb.TagNumber(2)
  set blockMaxGas($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasBlockMaxGas() => $_has(1);
  @$pb.TagNumber(2)
  void clearBlockMaxGas() => clearField(2);
}

