///
//  Generated code. Do not modify.
//  source: cosmos/slashing/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class MsgUnjail extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgUnjail', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.slashing.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddr')
    ..hasRequiredFields = false
  ;

  MsgUnjail._() : super();
  factory MsgUnjail() => create();
  factory MsgUnjail.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgUnjail.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgUnjail clone() => MsgUnjail()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgUnjail copyWith(void Function(MsgUnjail) updates) => super.copyWith((message) => updates(message as MsgUnjail)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgUnjail create() => MsgUnjail._();
  MsgUnjail createEmptyInstance() => create();
  static $pb.PbList<MsgUnjail> createRepeated() => $pb.PbList<MsgUnjail>();
  @$core.pragma('dart2js:noInline')
  static MsgUnjail getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgUnjail>(create);
  static MsgUnjail _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get validatorAddr => $_getSZ(0);
  @$pb.TagNumber(1)
  set validatorAddr($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorAddr() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorAddr() => clearField(1);
}

class MsgUnjailResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgUnjailResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.slashing.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgUnjailResponse._() : super();
  factory MsgUnjailResponse() => create();
  factory MsgUnjailResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgUnjailResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgUnjailResponse clone() => MsgUnjailResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgUnjailResponse copyWith(void Function(MsgUnjailResponse) updates) => super.copyWith((message) => updates(message as MsgUnjailResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgUnjailResponse create() => MsgUnjailResponse._();
  MsgUnjailResponse createEmptyInstance() => create();
  static $pb.PbList<MsgUnjailResponse> createRepeated() => $pb.PbList<MsgUnjailResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgUnjailResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgUnjailResponse>(create);
  static MsgUnjailResponse _defaultInstance;
}

