///
//  Generated code. Do not modify.
//  source: ibc/core/channel/v1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'channel.pb.dart' as $4;
import '../../client/v1/client.pb.dart' as $3;

class MsgChannelOpenInit extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelOpenInit', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOM<$4.Channel>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channel', subBuilder: $4.Channel.create)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgChannelOpenInit._() : super();
  factory MsgChannelOpenInit() => create();
  factory MsgChannelOpenInit.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelOpenInit.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelOpenInit clone() => MsgChannelOpenInit()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelOpenInit copyWith(void Function(MsgChannelOpenInit) updates) => super.copyWith((message) => updates(message as MsgChannelOpenInit)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenInit create() => MsgChannelOpenInit._();
  MsgChannelOpenInit createEmptyInstance() => create();
  static $pb.PbList<MsgChannelOpenInit> createRepeated() => $pb.PbList<MsgChannelOpenInit>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenInit getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelOpenInit>(create);
  static MsgChannelOpenInit _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $4.Channel get channel => $_getN(1);
  @$pb.TagNumber(2)
  set channel($4.Channel v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannel() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannel() => clearField(2);
  @$pb.TagNumber(2)
  $4.Channel ensureChannel() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.String get signer => $_getSZ(2);
  @$pb.TagNumber(3)
  set signer($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSigner() => $_has(2);
  @$pb.TagNumber(3)
  void clearSigner() => clearField(3);
}

class MsgChannelOpenInitResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelOpenInitResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgChannelOpenInitResponse._() : super();
  factory MsgChannelOpenInitResponse() => create();
  factory MsgChannelOpenInitResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelOpenInitResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelOpenInitResponse clone() => MsgChannelOpenInitResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelOpenInitResponse copyWith(void Function(MsgChannelOpenInitResponse) updates) => super.copyWith((message) => updates(message as MsgChannelOpenInitResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenInitResponse create() => MsgChannelOpenInitResponse._();
  MsgChannelOpenInitResponse createEmptyInstance() => create();
  static $pb.PbList<MsgChannelOpenInitResponse> createRepeated() => $pb.PbList<MsgChannelOpenInitResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenInitResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelOpenInitResponse>(create);
  static MsgChannelOpenInitResponse _defaultInstance;
}

class MsgChannelOpenTry extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelOpenTry', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'previousChannelId')
    ..aOM<$4.Channel>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channel', subBuilder: $4.Channel.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'counterpartyVersion')
    ..a<$core.List<$core.int>>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofInit', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..aOS(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgChannelOpenTry._() : super();
  factory MsgChannelOpenTry() => create();
  factory MsgChannelOpenTry.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelOpenTry.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelOpenTry clone() => MsgChannelOpenTry()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelOpenTry copyWith(void Function(MsgChannelOpenTry) updates) => super.copyWith((message) => updates(message as MsgChannelOpenTry)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenTry create() => MsgChannelOpenTry._();
  MsgChannelOpenTry createEmptyInstance() => create();
  static $pb.PbList<MsgChannelOpenTry> createRepeated() => $pb.PbList<MsgChannelOpenTry>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenTry getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelOpenTry>(create);
  static MsgChannelOpenTry _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get previousChannelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set previousChannelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPreviousChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPreviousChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $4.Channel get channel => $_getN(2);
  @$pb.TagNumber(3)
  set channel($4.Channel v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasChannel() => $_has(2);
  @$pb.TagNumber(3)
  void clearChannel() => clearField(3);
  @$pb.TagNumber(3)
  $4.Channel ensureChannel() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.String get counterpartyVersion => $_getSZ(3);
  @$pb.TagNumber(4)
  set counterpartyVersion($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasCounterpartyVersion() => $_has(3);
  @$pb.TagNumber(4)
  void clearCounterpartyVersion() => clearField(4);

  @$pb.TagNumber(5)
  $core.List<$core.int> get proofInit => $_getN(4);
  @$pb.TagNumber(5)
  set proofInit($core.List<$core.int> v) { $_setBytes(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasProofInit() => $_has(4);
  @$pb.TagNumber(5)
  void clearProofInit() => clearField(5);

  @$pb.TagNumber(6)
  $3.Height get proofHeight => $_getN(5);
  @$pb.TagNumber(6)
  set proofHeight($3.Height v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasProofHeight() => $_has(5);
  @$pb.TagNumber(6)
  void clearProofHeight() => clearField(6);
  @$pb.TagNumber(6)
  $3.Height ensureProofHeight() => $_ensure(5);

  @$pb.TagNumber(7)
  $core.String get signer => $_getSZ(6);
  @$pb.TagNumber(7)
  set signer($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasSigner() => $_has(6);
  @$pb.TagNumber(7)
  void clearSigner() => clearField(7);
}

class MsgChannelOpenTryResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelOpenTryResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgChannelOpenTryResponse._() : super();
  factory MsgChannelOpenTryResponse() => create();
  factory MsgChannelOpenTryResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelOpenTryResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelOpenTryResponse clone() => MsgChannelOpenTryResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelOpenTryResponse copyWith(void Function(MsgChannelOpenTryResponse) updates) => super.copyWith((message) => updates(message as MsgChannelOpenTryResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenTryResponse create() => MsgChannelOpenTryResponse._();
  MsgChannelOpenTryResponse createEmptyInstance() => create();
  static $pb.PbList<MsgChannelOpenTryResponse> createRepeated() => $pb.PbList<MsgChannelOpenTryResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenTryResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelOpenTryResponse>(create);
  static MsgChannelOpenTryResponse _defaultInstance;
}

class MsgChannelOpenAck extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelOpenAck', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'counterpartyChannelId')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'counterpartyVersion')
    ..a<$core.List<$core.int>>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofTry', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..aOS(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgChannelOpenAck._() : super();
  factory MsgChannelOpenAck() => create();
  factory MsgChannelOpenAck.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelOpenAck.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelOpenAck clone() => MsgChannelOpenAck()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelOpenAck copyWith(void Function(MsgChannelOpenAck) updates) => super.copyWith((message) => updates(message as MsgChannelOpenAck)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenAck create() => MsgChannelOpenAck._();
  MsgChannelOpenAck createEmptyInstance() => create();
  static $pb.PbList<MsgChannelOpenAck> createRepeated() => $pb.PbList<MsgChannelOpenAck>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenAck getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelOpenAck>(create);
  static MsgChannelOpenAck _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get counterpartyChannelId => $_getSZ(2);
  @$pb.TagNumber(3)
  set counterpartyChannelId($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasCounterpartyChannelId() => $_has(2);
  @$pb.TagNumber(3)
  void clearCounterpartyChannelId() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get counterpartyVersion => $_getSZ(3);
  @$pb.TagNumber(4)
  set counterpartyVersion($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasCounterpartyVersion() => $_has(3);
  @$pb.TagNumber(4)
  void clearCounterpartyVersion() => clearField(4);

  @$pb.TagNumber(5)
  $core.List<$core.int> get proofTry => $_getN(4);
  @$pb.TagNumber(5)
  set proofTry($core.List<$core.int> v) { $_setBytes(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasProofTry() => $_has(4);
  @$pb.TagNumber(5)
  void clearProofTry() => clearField(5);

  @$pb.TagNumber(6)
  $3.Height get proofHeight => $_getN(5);
  @$pb.TagNumber(6)
  set proofHeight($3.Height v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasProofHeight() => $_has(5);
  @$pb.TagNumber(6)
  void clearProofHeight() => clearField(6);
  @$pb.TagNumber(6)
  $3.Height ensureProofHeight() => $_ensure(5);

  @$pb.TagNumber(7)
  $core.String get signer => $_getSZ(6);
  @$pb.TagNumber(7)
  set signer($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasSigner() => $_has(6);
  @$pb.TagNumber(7)
  void clearSigner() => clearField(7);
}

class MsgChannelOpenAckResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelOpenAckResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgChannelOpenAckResponse._() : super();
  factory MsgChannelOpenAckResponse() => create();
  factory MsgChannelOpenAckResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelOpenAckResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelOpenAckResponse clone() => MsgChannelOpenAckResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelOpenAckResponse copyWith(void Function(MsgChannelOpenAckResponse) updates) => super.copyWith((message) => updates(message as MsgChannelOpenAckResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenAckResponse create() => MsgChannelOpenAckResponse._();
  MsgChannelOpenAckResponse createEmptyInstance() => create();
  static $pb.PbList<MsgChannelOpenAckResponse> createRepeated() => $pb.PbList<MsgChannelOpenAckResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenAckResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelOpenAckResponse>(create);
  static MsgChannelOpenAckResponse _defaultInstance;
}

class MsgChannelOpenConfirm extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelOpenConfirm', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..a<$core.List<$core.int>>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofAck', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgChannelOpenConfirm._() : super();
  factory MsgChannelOpenConfirm() => create();
  factory MsgChannelOpenConfirm.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelOpenConfirm.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelOpenConfirm clone() => MsgChannelOpenConfirm()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelOpenConfirm copyWith(void Function(MsgChannelOpenConfirm) updates) => super.copyWith((message) => updates(message as MsgChannelOpenConfirm)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenConfirm create() => MsgChannelOpenConfirm._();
  MsgChannelOpenConfirm createEmptyInstance() => create();
  static $pb.PbList<MsgChannelOpenConfirm> createRepeated() => $pb.PbList<MsgChannelOpenConfirm>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenConfirm getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelOpenConfirm>(create);
  static MsgChannelOpenConfirm _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get proofAck => $_getN(2);
  @$pb.TagNumber(3)
  set proofAck($core.List<$core.int> v) { $_setBytes(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasProofAck() => $_has(2);
  @$pb.TagNumber(3)
  void clearProofAck() => clearField(3);

  @$pb.TagNumber(4)
  $3.Height get proofHeight => $_getN(3);
  @$pb.TagNumber(4)
  set proofHeight($3.Height v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasProofHeight() => $_has(3);
  @$pb.TagNumber(4)
  void clearProofHeight() => clearField(4);
  @$pb.TagNumber(4)
  $3.Height ensureProofHeight() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.String get signer => $_getSZ(4);
  @$pb.TagNumber(5)
  set signer($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasSigner() => $_has(4);
  @$pb.TagNumber(5)
  void clearSigner() => clearField(5);
}

class MsgChannelOpenConfirmResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelOpenConfirmResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgChannelOpenConfirmResponse._() : super();
  factory MsgChannelOpenConfirmResponse() => create();
  factory MsgChannelOpenConfirmResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelOpenConfirmResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelOpenConfirmResponse clone() => MsgChannelOpenConfirmResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelOpenConfirmResponse copyWith(void Function(MsgChannelOpenConfirmResponse) updates) => super.copyWith((message) => updates(message as MsgChannelOpenConfirmResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenConfirmResponse create() => MsgChannelOpenConfirmResponse._();
  MsgChannelOpenConfirmResponse createEmptyInstance() => create();
  static $pb.PbList<MsgChannelOpenConfirmResponse> createRepeated() => $pb.PbList<MsgChannelOpenConfirmResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelOpenConfirmResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelOpenConfirmResponse>(create);
  static MsgChannelOpenConfirmResponse _defaultInstance;
}

class MsgChannelCloseInit extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelCloseInit', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgChannelCloseInit._() : super();
  factory MsgChannelCloseInit() => create();
  factory MsgChannelCloseInit.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelCloseInit.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelCloseInit clone() => MsgChannelCloseInit()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelCloseInit copyWith(void Function(MsgChannelCloseInit) updates) => super.copyWith((message) => updates(message as MsgChannelCloseInit)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelCloseInit create() => MsgChannelCloseInit._();
  MsgChannelCloseInit createEmptyInstance() => create();
  static $pb.PbList<MsgChannelCloseInit> createRepeated() => $pb.PbList<MsgChannelCloseInit>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelCloseInit getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelCloseInit>(create);
  static MsgChannelCloseInit _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get signer => $_getSZ(2);
  @$pb.TagNumber(3)
  set signer($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSigner() => $_has(2);
  @$pb.TagNumber(3)
  void clearSigner() => clearField(3);
}

class MsgChannelCloseInitResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelCloseInitResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgChannelCloseInitResponse._() : super();
  factory MsgChannelCloseInitResponse() => create();
  factory MsgChannelCloseInitResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelCloseInitResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelCloseInitResponse clone() => MsgChannelCloseInitResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelCloseInitResponse copyWith(void Function(MsgChannelCloseInitResponse) updates) => super.copyWith((message) => updates(message as MsgChannelCloseInitResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelCloseInitResponse create() => MsgChannelCloseInitResponse._();
  MsgChannelCloseInitResponse createEmptyInstance() => create();
  static $pb.PbList<MsgChannelCloseInitResponse> createRepeated() => $pb.PbList<MsgChannelCloseInitResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelCloseInitResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelCloseInitResponse>(create);
  static MsgChannelCloseInitResponse _defaultInstance;
}

class MsgChannelCloseConfirm extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelCloseConfirm', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..a<$core.List<$core.int>>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofInit', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgChannelCloseConfirm._() : super();
  factory MsgChannelCloseConfirm() => create();
  factory MsgChannelCloseConfirm.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelCloseConfirm.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelCloseConfirm clone() => MsgChannelCloseConfirm()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelCloseConfirm copyWith(void Function(MsgChannelCloseConfirm) updates) => super.copyWith((message) => updates(message as MsgChannelCloseConfirm)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelCloseConfirm create() => MsgChannelCloseConfirm._();
  MsgChannelCloseConfirm createEmptyInstance() => create();
  static $pb.PbList<MsgChannelCloseConfirm> createRepeated() => $pb.PbList<MsgChannelCloseConfirm>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelCloseConfirm getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelCloseConfirm>(create);
  static MsgChannelCloseConfirm _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get proofInit => $_getN(2);
  @$pb.TagNumber(3)
  set proofInit($core.List<$core.int> v) { $_setBytes(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasProofInit() => $_has(2);
  @$pb.TagNumber(3)
  void clearProofInit() => clearField(3);

  @$pb.TagNumber(4)
  $3.Height get proofHeight => $_getN(3);
  @$pb.TagNumber(4)
  set proofHeight($3.Height v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasProofHeight() => $_has(3);
  @$pb.TagNumber(4)
  void clearProofHeight() => clearField(4);
  @$pb.TagNumber(4)
  $3.Height ensureProofHeight() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.String get signer => $_getSZ(4);
  @$pb.TagNumber(5)
  set signer($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasSigner() => $_has(4);
  @$pb.TagNumber(5)
  void clearSigner() => clearField(5);
}

class MsgChannelCloseConfirmResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgChannelCloseConfirmResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgChannelCloseConfirmResponse._() : super();
  factory MsgChannelCloseConfirmResponse() => create();
  factory MsgChannelCloseConfirmResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgChannelCloseConfirmResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgChannelCloseConfirmResponse clone() => MsgChannelCloseConfirmResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgChannelCloseConfirmResponse copyWith(void Function(MsgChannelCloseConfirmResponse) updates) => super.copyWith((message) => updates(message as MsgChannelCloseConfirmResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgChannelCloseConfirmResponse create() => MsgChannelCloseConfirmResponse._();
  MsgChannelCloseConfirmResponse createEmptyInstance() => create();
  static $pb.PbList<MsgChannelCloseConfirmResponse> createRepeated() => $pb.PbList<MsgChannelCloseConfirmResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgChannelCloseConfirmResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgChannelCloseConfirmResponse>(create);
  static MsgChannelCloseConfirmResponse _defaultInstance;
}

class MsgRecvPacket extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgRecvPacket', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOM<$4.Packet>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'packet', subBuilder: $4.Packet.create)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofCommitment', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgRecvPacket._() : super();
  factory MsgRecvPacket() => create();
  factory MsgRecvPacket.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgRecvPacket.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgRecvPacket clone() => MsgRecvPacket()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgRecvPacket copyWith(void Function(MsgRecvPacket) updates) => super.copyWith((message) => updates(message as MsgRecvPacket)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgRecvPacket create() => MsgRecvPacket._();
  MsgRecvPacket createEmptyInstance() => create();
  static $pb.PbList<MsgRecvPacket> createRepeated() => $pb.PbList<MsgRecvPacket>();
  @$core.pragma('dart2js:noInline')
  static MsgRecvPacket getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgRecvPacket>(create);
  static MsgRecvPacket _defaultInstance;

  @$pb.TagNumber(1)
  $4.Packet get packet => $_getN(0);
  @$pb.TagNumber(1)
  set packet($4.Packet v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasPacket() => $_has(0);
  @$pb.TagNumber(1)
  void clearPacket() => clearField(1);
  @$pb.TagNumber(1)
  $4.Packet ensurePacket() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<$core.int> get proofCommitment => $_getN(1);
  @$pb.TagNumber(2)
  set proofCommitment($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasProofCommitment() => $_has(1);
  @$pb.TagNumber(2)
  void clearProofCommitment() => clearField(2);

  @$pb.TagNumber(3)
  $3.Height get proofHeight => $_getN(2);
  @$pb.TagNumber(3)
  set proofHeight($3.Height v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasProofHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearProofHeight() => clearField(3);
  @$pb.TagNumber(3)
  $3.Height ensureProofHeight() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.String get signer => $_getSZ(3);
  @$pb.TagNumber(4)
  set signer($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasSigner() => $_has(3);
  @$pb.TagNumber(4)
  void clearSigner() => clearField(4);
}

class MsgRecvPacketResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgRecvPacketResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgRecvPacketResponse._() : super();
  factory MsgRecvPacketResponse() => create();
  factory MsgRecvPacketResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgRecvPacketResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgRecvPacketResponse clone() => MsgRecvPacketResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgRecvPacketResponse copyWith(void Function(MsgRecvPacketResponse) updates) => super.copyWith((message) => updates(message as MsgRecvPacketResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgRecvPacketResponse create() => MsgRecvPacketResponse._();
  MsgRecvPacketResponse createEmptyInstance() => create();
  static $pb.PbList<MsgRecvPacketResponse> createRepeated() => $pb.PbList<MsgRecvPacketResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgRecvPacketResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgRecvPacketResponse>(create);
  static MsgRecvPacketResponse _defaultInstance;
}

class MsgTimeout extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgTimeout', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOM<$4.Packet>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'packet', subBuilder: $4.Packet.create)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofUnreceived', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..a<$fixnum.Int64>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nextSequenceRecv', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgTimeout._() : super();
  factory MsgTimeout() => create();
  factory MsgTimeout.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgTimeout.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgTimeout clone() => MsgTimeout()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgTimeout copyWith(void Function(MsgTimeout) updates) => super.copyWith((message) => updates(message as MsgTimeout)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgTimeout create() => MsgTimeout._();
  MsgTimeout createEmptyInstance() => create();
  static $pb.PbList<MsgTimeout> createRepeated() => $pb.PbList<MsgTimeout>();
  @$core.pragma('dart2js:noInline')
  static MsgTimeout getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgTimeout>(create);
  static MsgTimeout _defaultInstance;

  @$pb.TagNumber(1)
  $4.Packet get packet => $_getN(0);
  @$pb.TagNumber(1)
  set packet($4.Packet v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasPacket() => $_has(0);
  @$pb.TagNumber(1)
  void clearPacket() => clearField(1);
  @$pb.TagNumber(1)
  $4.Packet ensurePacket() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<$core.int> get proofUnreceived => $_getN(1);
  @$pb.TagNumber(2)
  set proofUnreceived($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasProofUnreceived() => $_has(1);
  @$pb.TagNumber(2)
  void clearProofUnreceived() => clearField(2);

  @$pb.TagNumber(3)
  $3.Height get proofHeight => $_getN(2);
  @$pb.TagNumber(3)
  set proofHeight($3.Height v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasProofHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearProofHeight() => clearField(3);
  @$pb.TagNumber(3)
  $3.Height ensureProofHeight() => $_ensure(2);

  @$pb.TagNumber(4)
  $fixnum.Int64 get nextSequenceRecv => $_getI64(3);
  @$pb.TagNumber(4)
  set nextSequenceRecv($fixnum.Int64 v) { $_setInt64(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasNextSequenceRecv() => $_has(3);
  @$pb.TagNumber(4)
  void clearNextSequenceRecv() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get signer => $_getSZ(4);
  @$pb.TagNumber(5)
  set signer($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasSigner() => $_has(4);
  @$pb.TagNumber(5)
  void clearSigner() => clearField(5);
}

class MsgTimeoutResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgTimeoutResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgTimeoutResponse._() : super();
  factory MsgTimeoutResponse() => create();
  factory MsgTimeoutResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgTimeoutResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgTimeoutResponse clone() => MsgTimeoutResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgTimeoutResponse copyWith(void Function(MsgTimeoutResponse) updates) => super.copyWith((message) => updates(message as MsgTimeoutResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgTimeoutResponse create() => MsgTimeoutResponse._();
  MsgTimeoutResponse createEmptyInstance() => create();
  static $pb.PbList<MsgTimeoutResponse> createRepeated() => $pb.PbList<MsgTimeoutResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgTimeoutResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgTimeoutResponse>(create);
  static MsgTimeoutResponse _defaultInstance;
}

class MsgTimeoutOnClose extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgTimeoutOnClose', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOM<$4.Packet>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'packet', subBuilder: $4.Packet.create)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofUnreceived', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofClose', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..a<$fixnum.Int64>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nextSequenceRecv', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgTimeoutOnClose._() : super();
  factory MsgTimeoutOnClose() => create();
  factory MsgTimeoutOnClose.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgTimeoutOnClose.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgTimeoutOnClose clone() => MsgTimeoutOnClose()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgTimeoutOnClose copyWith(void Function(MsgTimeoutOnClose) updates) => super.copyWith((message) => updates(message as MsgTimeoutOnClose)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgTimeoutOnClose create() => MsgTimeoutOnClose._();
  MsgTimeoutOnClose createEmptyInstance() => create();
  static $pb.PbList<MsgTimeoutOnClose> createRepeated() => $pb.PbList<MsgTimeoutOnClose>();
  @$core.pragma('dart2js:noInline')
  static MsgTimeoutOnClose getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgTimeoutOnClose>(create);
  static MsgTimeoutOnClose _defaultInstance;

  @$pb.TagNumber(1)
  $4.Packet get packet => $_getN(0);
  @$pb.TagNumber(1)
  set packet($4.Packet v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasPacket() => $_has(0);
  @$pb.TagNumber(1)
  void clearPacket() => clearField(1);
  @$pb.TagNumber(1)
  $4.Packet ensurePacket() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<$core.int> get proofUnreceived => $_getN(1);
  @$pb.TagNumber(2)
  set proofUnreceived($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasProofUnreceived() => $_has(1);
  @$pb.TagNumber(2)
  void clearProofUnreceived() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get proofClose => $_getN(2);
  @$pb.TagNumber(3)
  set proofClose($core.List<$core.int> v) { $_setBytes(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasProofClose() => $_has(2);
  @$pb.TagNumber(3)
  void clearProofClose() => clearField(3);

  @$pb.TagNumber(4)
  $3.Height get proofHeight => $_getN(3);
  @$pb.TagNumber(4)
  set proofHeight($3.Height v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasProofHeight() => $_has(3);
  @$pb.TagNumber(4)
  void clearProofHeight() => clearField(4);
  @$pb.TagNumber(4)
  $3.Height ensureProofHeight() => $_ensure(3);

  @$pb.TagNumber(5)
  $fixnum.Int64 get nextSequenceRecv => $_getI64(4);
  @$pb.TagNumber(5)
  set nextSequenceRecv($fixnum.Int64 v) { $_setInt64(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasNextSequenceRecv() => $_has(4);
  @$pb.TagNumber(5)
  void clearNextSequenceRecv() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get signer => $_getSZ(5);
  @$pb.TagNumber(6)
  set signer($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasSigner() => $_has(5);
  @$pb.TagNumber(6)
  void clearSigner() => clearField(6);
}

class MsgTimeoutOnCloseResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgTimeoutOnCloseResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgTimeoutOnCloseResponse._() : super();
  factory MsgTimeoutOnCloseResponse() => create();
  factory MsgTimeoutOnCloseResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgTimeoutOnCloseResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgTimeoutOnCloseResponse clone() => MsgTimeoutOnCloseResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgTimeoutOnCloseResponse copyWith(void Function(MsgTimeoutOnCloseResponse) updates) => super.copyWith((message) => updates(message as MsgTimeoutOnCloseResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgTimeoutOnCloseResponse create() => MsgTimeoutOnCloseResponse._();
  MsgTimeoutOnCloseResponse createEmptyInstance() => create();
  static $pb.PbList<MsgTimeoutOnCloseResponse> createRepeated() => $pb.PbList<MsgTimeoutOnCloseResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgTimeoutOnCloseResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgTimeoutOnCloseResponse>(create);
  static MsgTimeoutOnCloseResponse _defaultInstance;
}

class MsgAcknowledgement extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgAcknowledgement', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOM<$4.Packet>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'packet', subBuilder: $4.Packet.create)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'acknowledgement', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofAcked', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgAcknowledgement._() : super();
  factory MsgAcknowledgement() => create();
  factory MsgAcknowledgement.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgAcknowledgement.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgAcknowledgement clone() => MsgAcknowledgement()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgAcknowledgement copyWith(void Function(MsgAcknowledgement) updates) => super.copyWith((message) => updates(message as MsgAcknowledgement)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgAcknowledgement create() => MsgAcknowledgement._();
  MsgAcknowledgement createEmptyInstance() => create();
  static $pb.PbList<MsgAcknowledgement> createRepeated() => $pb.PbList<MsgAcknowledgement>();
  @$core.pragma('dart2js:noInline')
  static MsgAcknowledgement getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgAcknowledgement>(create);
  static MsgAcknowledgement _defaultInstance;

  @$pb.TagNumber(1)
  $4.Packet get packet => $_getN(0);
  @$pb.TagNumber(1)
  set packet($4.Packet v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasPacket() => $_has(0);
  @$pb.TagNumber(1)
  void clearPacket() => clearField(1);
  @$pb.TagNumber(1)
  $4.Packet ensurePacket() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<$core.int> get acknowledgement => $_getN(1);
  @$pb.TagNumber(2)
  set acknowledgement($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasAcknowledgement() => $_has(1);
  @$pb.TagNumber(2)
  void clearAcknowledgement() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get proofAcked => $_getN(2);
  @$pb.TagNumber(3)
  set proofAcked($core.List<$core.int> v) { $_setBytes(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasProofAcked() => $_has(2);
  @$pb.TagNumber(3)
  void clearProofAcked() => clearField(3);

  @$pb.TagNumber(4)
  $3.Height get proofHeight => $_getN(3);
  @$pb.TagNumber(4)
  set proofHeight($3.Height v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasProofHeight() => $_has(3);
  @$pb.TagNumber(4)
  void clearProofHeight() => clearField(4);
  @$pb.TagNumber(4)
  $3.Height ensureProofHeight() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.String get signer => $_getSZ(4);
  @$pb.TagNumber(5)
  set signer($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasSigner() => $_has(4);
  @$pb.TagNumber(5)
  void clearSigner() => clearField(5);
}

class MsgAcknowledgementResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgAcknowledgementResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgAcknowledgementResponse._() : super();
  factory MsgAcknowledgementResponse() => create();
  factory MsgAcknowledgementResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgAcknowledgementResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgAcknowledgementResponse clone() => MsgAcknowledgementResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgAcknowledgementResponse copyWith(void Function(MsgAcknowledgementResponse) updates) => super.copyWith((message) => updates(message as MsgAcknowledgementResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgAcknowledgementResponse create() => MsgAcknowledgementResponse._();
  MsgAcknowledgementResponse createEmptyInstance() => create();
  static $pb.PbList<MsgAcknowledgementResponse> createRepeated() => $pb.PbList<MsgAcknowledgementResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgAcknowledgementResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgAcknowledgementResponse>(create);
  static MsgAcknowledgementResponse _defaultInstance;
}

