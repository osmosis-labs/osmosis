///
//  Generated code. Do not modify.
//  source: cosmos/slashing/v1beta1/slashing.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../../google/protobuf/timestamp.pb.dart' as $2;
import '../../../google/protobuf/duration.pb.dart' as $3;

class ValidatorSigningInfo extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorSigningInfo', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.slashing.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'address')
    ..aInt64(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'startHeight')
    ..aInt64(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'indexOffset')
    ..aOM<$2.Timestamp>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'jailedUntil', subBuilder: $2.Timestamp.create)
    ..aOB(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tombstoned')
    ..aInt64(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'missedBlocksCounter')
    ..hasRequiredFields = false
  ;

  ValidatorSigningInfo._() : super();
  factory ValidatorSigningInfo() => create();
  factory ValidatorSigningInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorSigningInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorSigningInfo clone() => ValidatorSigningInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorSigningInfo copyWith(void Function(ValidatorSigningInfo) updates) => super.copyWith((message) => updates(message as ValidatorSigningInfo)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorSigningInfo create() => ValidatorSigningInfo._();
  ValidatorSigningInfo createEmptyInstance() => create();
  static $pb.PbList<ValidatorSigningInfo> createRepeated() => $pb.PbList<ValidatorSigningInfo>();
  @$core.pragma('dart2js:noInline')
  static ValidatorSigningInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorSigningInfo>(create);
  static ValidatorSigningInfo _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get address => $_getSZ(0);
  @$pb.TagNumber(1)
  set address($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get startHeight => $_getI64(1);
  @$pb.TagNumber(2)
  set startHeight($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasStartHeight() => $_has(1);
  @$pb.TagNumber(2)
  void clearStartHeight() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get indexOffset => $_getI64(2);
  @$pb.TagNumber(3)
  set indexOffset($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasIndexOffset() => $_has(2);
  @$pb.TagNumber(3)
  void clearIndexOffset() => clearField(3);

  @$pb.TagNumber(4)
  $2.Timestamp get jailedUntil => $_getN(3);
  @$pb.TagNumber(4)
  set jailedUntil($2.Timestamp v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasJailedUntil() => $_has(3);
  @$pb.TagNumber(4)
  void clearJailedUntil() => clearField(4);
  @$pb.TagNumber(4)
  $2.Timestamp ensureJailedUntil() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.bool get tombstoned => $_getBF(4);
  @$pb.TagNumber(5)
  set tombstoned($core.bool v) { $_setBool(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasTombstoned() => $_has(4);
  @$pb.TagNumber(5)
  void clearTombstoned() => clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get missedBlocksCounter => $_getI64(5);
  @$pb.TagNumber(6)
  set missedBlocksCounter($fixnum.Int64 v) { $_setInt64(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasMissedBlocksCounter() => $_has(5);
  @$pb.TagNumber(6)
  void clearMissedBlocksCounter() => clearField(6);
}

class Params extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Params', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.slashing.v1beta1'), createEmptyInstance: create)
    ..aInt64(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signedBlocksWindow')
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'minSignedPerWindow', $pb.PbFieldType.OY)
    ..aOM<$3.Duration>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'downtimeJailDuration', subBuilder: $3.Duration.create)
    ..a<$core.List<$core.int>>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'slashFractionDoubleSign', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'slashFractionDowntime', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  Params._() : super();
  factory Params() => create();
  factory Params.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Params.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Params clone() => Params()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Params copyWith(void Function(Params) updates) => super.copyWith((message) => updates(message as Params)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Params create() => Params._();
  Params createEmptyInstance() => create();
  static $pb.PbList<Params> createRepeated() => $pb.PbList<Params>();
  @$core.pragma('dart2js:noInline')
  static Params getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Params>(create);
  static Params _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get signedBlocksWindow => $_getI64(0);
  @$pb.TagNumber(1)
  set signedBlocksWindow($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSignedBlocksWindow() => $_has(0);
  @$pb.TagNumber(1)
  void clearSignedBlocksWindow() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get minSignedPerWindow => $_getN(1);
  @$pb.TagNumber(2)
  set minSignedPerWindow($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasMinSignedPerWindow() => $_has(1);
  @$pb.TagNumber(2)
  void clearMinSignedPerWindow() => clearField(2);

  @$pb.TagNumber(3)
  $3.Duration get downtimeJailDuration => $_getN(2);
  @$pb.TagNumber(3)
  set downtimeJailDuration($3.Duration v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasDowntimeJailDuration() => $_has(2);
  @$pb.TagNumber(3)
  void clearDowntimeJailDuration() => clearField(3);
  @$pb.TagNumber(3)
  $3.Duration ensureDowntimeJailDuration() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.List<$core.int> get slashFractionDoubleSign => $_getN(3);
  @$pb.TagNumber(4)
  set slashFractionDoubleSign($core.List<$core.int> v) { $_setBytes(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasSlashFractionDoubleSign() => $_has(3);
  @$pb.TagNumber(4)
  void clearSlashFractionDoubleSign() => clearField(4);

  @$pb.TagNumber(5)
  $core.List<$core.int> get slashFractionDowntime => $_getN(4);
  @$pb.TagNumber(5)
  set slashFractionDowntime($core.List<$core.int> v) { $_setBytes(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasSlashFractionDowntime() => $_has(4);
  @$pb.TagNumber(5)
  void clearSlashFractionDowntime() => clearField(5);
}

