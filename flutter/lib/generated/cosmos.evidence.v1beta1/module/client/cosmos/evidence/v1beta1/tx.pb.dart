///
//  Generated code. Do not modify.
//  source: cosmos/evidence/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../../google/protobuf/any.pb.dart' as $3;

class MsgSubmitEvidence extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgSubmitEvidence', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.evidence.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'submitter')
    ..aOM<$3.Any>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'evidence', subBuilder: $3.Any.create)
    ..hasRequiredFields = false
  ;

  MsgSubmitEvidence._() : super();
  factory MsgSubmitEvidence() => create();
  factory MsgSubmitEvidence.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgSubmitEvidence.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgSubmitEvidence clone() => MsgSubmitEvidence()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgSubmitEvidence copyWith(void Function(MsgSubmitEvidence) updates) => super.copyWith((message) => updates(message as MsgSubmitEvidence)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgSubmitEvidence create() => MsgSubmitEvidence._();
  MsgSubmitEvidence createEmptyInstance() => create();
  static $pb.PbList<MsgSubmitEvidence> createRepeated() => $pb.PbList<MsgSubmitEvidence>();
  @$core.pragma('dart2js:noInline')
  static MsgSubmitEvidence getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgSubmitEvidence>(create);
  static MsgSubmitEvidence _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get submitter => $_getSZ(0);
  @$pb.TagNumber(1)
  set submitter($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSubmitter() => $_has(0);
  @$pb.TagNumber(1)
  void clearSubmitter() => clearField(1);

  @$pb.TagNumber(2)
  $3.Any get evidence => $_getN(1);
  @$pb.TagNumber(2)
  set evidence($3.Any v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasEvidence() => $_has(1);
  @$pb.TagNumber(2)
  void clearEvidence() => clearField(2);
  @$pb.TagNumber(2)
  $3.Any ensureEvidence() => $_ensure(1);
}

class MsgSubmitEvidenceResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgSubmitEvidenceResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.evidence.v1beta1'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'hash', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  MsgSubmitEvidenceResponse._() : super();
  factory MsgSubmitEvidenceResponse() => create();
  factory MsgSubmitEvidenceResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgSubmitEvidenceResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgSubmitEvidenceResponse clone() => MsgSubmitEvidenceResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgSubmitEvidenceResponse copyWith(void Function(MsgSubmitEvidenceResponse) updates) => super.copyWith((message) => updates(message as MsgSubmitEvidenceResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgSubmitEvidenceResponse create() => MsgSubmitEvidenceResponse._();
  MsgSubmitEvidenceResponse createEmptyInstance() => create();
  static $pb.PbList<MsgSubmitEvidenceResponse> createRepeated() => $pb.PbList<MsgSubmitEvidenceResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgSubmitEvidenceResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgSubmitEvidenceResponse>(create);
  static MsgSubmitEvidenceResponse _defaultInstance;

  @$pb.TagNumber(4)
  $core.List<$core.int> get hash => $_getN(0);
  @$pb.TagNumber(4)
  set hash($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(4)
  $core.bool hasHash() => $_has(0);
  @$pb.TagNumber(4)
  void clearHash() => clearField(4);
}

