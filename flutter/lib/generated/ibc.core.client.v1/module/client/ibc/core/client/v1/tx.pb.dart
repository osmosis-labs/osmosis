///
//  Generated code. Do not modify.
//  source: ibc/core/client/v1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../../../google/protobuf/any.pb.dart' as $2;

class MsgCreateClient extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgCreateClient', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..aOM<$2.Any>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientState', subBuilder: $2.Any.create)
    ..aOM<$2.Any>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'consensusState', subBuilder: $2.Any.create)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgCreateClient._() : super();
  factory MsgCreateClient() => create();
  factory MsgCreateClient.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgCreateClient.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgCreateClient clone() => MsgCreateClient()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgCreateClient copyWith(void Function(MsgCreateClient) updates) => super.copyWith((message) => updates(message as MsgCreateClient)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgCreateClient create() => MsgCreateClient._();
  MsgCreateClient createEmptyInstance() => create();
  static $pb.PbList<MsgCreateClient> createRepeated() => $pb.PbList<MsgCreateClient>();
  @$core.pragma('dart2js:noInline')
  static MsgCreateClient getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgCreateClient>(create);
  static MsgCreateClient _defaultInstance;

  @$pb.TagNumber(1)
  $2.Any get clientState => $_getN(0);
  @$pb.TagNumber(1)
  set clientState($2.Any v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientState() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientState() => clearField(1);
  @$pb.TagNumber(1)
  $2.Any ensureClientState() => $_ensure(0);

  @$pb.TagNumber(2)
  $2.Any get consensusState => $_getN(1);
  @$pb.TagNumber(2)
  set consensusState($2.Any v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasConsensusState() => $_has(1);
  @$pb.TagNumber(2)
  void clearConsensusState() => clearField(2);
  @$pb.TagNumber(2)
  $2.Any ensureConsensusState() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.String get signer => $_getSZ(2);
  @$pb.TagNumber(3)
  set signer($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSigner() => $_has(2);
  @$pb.TagNumber(3)
  void clearSigner() => clearField(3);
}

class MsgCreateClientResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgCreateClientResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgCreateClientResponse._() : super();
  factory MsgCreateClientResponse() => create();
  factory MsgCreateClientResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgCreateClientResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgCreateClientResponse clone() => MsgCreateClientResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgCreateClientResponse copyWith(void Function(MsgCreateClientResponse) updates) => super.copyWith((message) => updates(message as MsgCreateClientResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgCreateClientResponse create() => MsgCreateClientResponse._();
  MsgCreateClientResponse createEmptyInstance() => create();
  static $pb.PbList<MsgCreateClientResponse> createRepeated() => $pb.PbList<MsgCreateClientResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgCreateClientResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgCreateClientResponse>(create);
  static MsgCreateClientResponse _defaultInstance;
}

class MsgUpdateClient extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgUpdateClient', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..aOM<$2.Any>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'header', subBuilder: $2.Any.create)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgUpdateClient._() : super();
  factory MsgUpdateClient() => create();
  factory MsgUpdateClient.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgUpdateClient.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgUpdateClient clone() => MsgUpdateClient()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgUpdateClient copyWith(void Function(MsgUpdateClient) updates) => super.copyWith((message) => updates(message as MsgUpdateClient)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgUpdateClient create() => MsgUpdateClient._();
  MsgUpdateClient createEmptyInstance() => create();
  static $pb.PbList<MsgUpdateClient> createRepeated() => $pb.PbList<MsgUpdateClient>();
  @$core.pragma('dart2js:noInline')
  static MsgUpdateClient getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgUpdateClient>(create);
  static MsgUpdateClient _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $2.Any get header => $_getN(1);
  @$pb.TagNumber(2)
  set header($2.Any v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasHeader() => $_has(1);
  @$pb.TagNumber(2)
  void clearHeader() => clearField(2);
  @$pb.TagNumber(2)
  $2.Any ensureHeader() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.String get signer => $_getSZ(2);
  @$pb.TagNumber(3)
  set signer($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSigner() => $_has(2);
  @$pb.TagNumber(3)
  void clearSigner() => clearField(3);
}

class MsgUpdateClientResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgUpdateClientResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgUpdateClientResponse._() : super();
  factory MsgUpdateClientResponse() => create();
  factory MsgUpdateClientResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgUpdateClientResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgUpdateClientResponse clone() => MsgUpdateClientResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgUpdateClientResponse copyWith(void Function(MsgUpdateClientResponse) updates) => super.copyWith((message) => updates(message as MsgUpdateClientResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgUpdateClientResponse create() => MsgUpdateClientResponse._();
  MsgUpdateClientResponse createEmptyInstance() => create();
  static $pb.PbList<MsgUpdateClientResponse> createRepeated() => $pb.PbList<MsgUpdateClientResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgUpdateClientResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgUpdateClientResponse>(create);
  static MsgUpdateClientResponse _defaultInstance;
}

class MsgUpgradeClient extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgUpgradeClient', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..aOM<$2.Any>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientState', subBuilder: $2.Any.create)
    ..aOM<$2.Any>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'consensusState', subBuilder: $2.Any.create)
    ..a<$core.List<$core.int>>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofUpgradeClient', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofUpgradeConsensusState', $pb.PbFieldType.OY)
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgUpgradeClient._() : super();
  factory MsgUpgradeClient() => create();
  factory MsgUpgradeClient.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgUpgradeClient.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgUpgradeClient clone() => MsgUpgradeClient()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgUpgradeClient copyWith(void Function(MsgUpgradeClient) updates) => super.copyWith((message) => updates(message as MsgUpgradeClient)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgUpgradeClient create() => MsgUpgradeClient._();
  MsgUpgradeClient createEmptyInstance() => create();
  static $pb.PbList<MsgUpgradeClient> createRepeated() => $pb.PbList<MsgUpgradeClient>();
  @$core.pragma('dart2js:noInline')
  static MsgUpgradeClient getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgUpgradeClient>(create);
  static MsgUpgradeClient _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $2.Any get clientState => $_getN(1);
  @$pb.TagNumber(2)
  set clientState($2.Any v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasClientState() => $_has(1);
  @$pb.TagNumber(2)
  void clearClientState() => clearField(2);
  @$pb.TagNumber(2)
  $2.Any ensureClientState() => $_ensure(1);

  @$pb.TagNumber(3)
  $2.Any get consensusState => $_getN(2);
  @$pb.TagNumber(3)
  set consensusState($2.Any v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasConsensusState() => $_has(2);
  @$pb.TagNumber(3)
  void clearConsensusState() => clearField(3);
  @$pb.TagNumber(3)
  $2.Any ensureConsensusState() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.List<$core.int> get proofUpgradeClient => $_getN(3);
  @$pb.TagNumber(4)
  set proofUpgradeClient($core.List<$core.int> v) { $_setBytes(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasProofUpgradeClient() => $_has(3);
  @$pb.TagNumber(4)
  void clearProofUpgradeClient() => clearField(4);

  @$pb.TagNumber(5)
  $core.List<$core.int> get proofUpgradeConsensusState => $_getN(4);
  @$pb.TagNumber(5)
  set proofUpgradeConsensusState($core.List<$core.int> v) { $_setBytes(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasProofUpgradeConsensusState() => $_has(4);
  @$pb.TagNumber(5)
  void clearProofUpgradeConsensusState() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get signer => $_getSZ(5);
  @$pb.TagNumber(6)
  set signer($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasSigner() => $_has(5);
  @$pb.TagNumber(6)
  void clearSigner() => clearField(6);
}

class MsgUpgradeClientResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgUpgradeClientResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgUpgradeClientResponse._() : super();
  factory MsgUpgradeClientResponse() => create();
  factory MsgUpgradeClientResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgUpgradeClientResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgUpgradeClientResponse clone() => MsgUpgradeClientResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgUpgradeClientResponse copyWith(void Function(MsgUpgradeClientResponse) updates) => super.copyWith((message) => updates(message as MsgUpgradeClientResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgUpgradeClientResponse create() => MsgUpgradeClientResponse._();
  MsgUpgradeClientResponse createEmptyInstance() => create();
  static $pb.PbList<MsgUpgradeClientResponse> createRepeated() => $pb.PbList<MsgUpgradeClientResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgUpgradeClientResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgUpgradeClientResponse>(create);
  static MsgUpgradeClientResponse _defaultInstance;
}

class MsgSubmitMisbehaviour extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgSubmitMisbehaviour', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..aOM<$2.Any>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'misbehaviour', subBuilder: $2.Any.create)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgSubmitMisbehaviour._() : super();
  factory MsgSubmitMisbehaviour() => create();
  factory MsgSubmitMisbehaviour.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgSubmitMisbehaviour.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgSubmitMisbehaviour clone() => MsgSubmitMisbehaviour()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgSubmitMisbehaviour copyWith(void Function(MsgSubmitMisbehaviour) updates) => super.copyWith((message) => updates(message as MsgSubmitMisbehaviour)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgSubmitMisbehaviour create() => MsgSubmitMisbehaviour._();
  MsgSubmitMisbehaviour createEmptyInstance() => create();
  static $pb.PbList<MsgSubmitMisbehaviour> createRepeated() => $pb.PbList<MsgSubmitMisbehaviour>();
  @$core.pragma('dart2js:noInline')
  static MsgSubmitMisbehaviour getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgSubmitMisbehaviour>(create);
  static MsgSubmitMisbehaviour _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $2.Any get misbehaviour => $_getN(1);
  @$pb.TagNumber(2)
  set misbehaviour($2.Any v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasMisbehaviour() => $_has(1);
  @$pb.TagNumber(2)
  void clearMisbehaviour() => clearField(2);
  @$pb.TagNumber(2)
  $2.Any ensureMisbehaviour() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.String get signer => $_getSZ(2);
  @$pb.TagNumber(3)
  set signer($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSigner() => $_has(2);
  @$pb.TagNumber(3)
  void clearSigner() => clearField(3);
}

class MsgSubmitMisbehaviourResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgSubmitMisbehaviourResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgSubmitMisbehaviourResponse._() : super();
  factory MsgSubmitMisbehaviourResponse() => create();
  factory MsgSubmitMisbehaviourResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgSubmitMisbehaviourResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgSubmitMisbehaviourResponse clone() => MsgSubmitMisbehaviourResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgSubmitMisbehaviourResponse copyWith(void Function(MsgSubmitMisbehaviourResponse) updates) => super.copyWith((message) => updates(message as MsgSubmitMisbehaviourResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgSubmitMisbehaviourResponse create() => MsgSubmitMisbehaviourResponse._();
  MsgSubmitMisbehaviourResponse createEmptyInstance() => create();
  static $pb.PbList<MsgSubmitMisbehaviourResponse> createRepeated() => $pb.PbList<MsgSubmitMisbehaviourResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgSubmitMisbehaviourResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgSubmitMisbehaviourResponse>(create);
  static MsgSubmitMisbehaviourResponse _defaultInstance;
}

