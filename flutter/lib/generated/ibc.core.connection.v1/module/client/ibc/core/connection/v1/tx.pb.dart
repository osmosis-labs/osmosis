///
//  Generated code. Do not modify.
//  source: ibc/core/connection/v1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'connection.pb.dart' as $4;
import '../../../../google/protobuf/any.pb.dart' as $5;
import '../../client/v1/client.pb.dart' as $7;

class MsgConnectionOpenInit extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgConnectionOpenInit', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..aOM<$4.Counterparty>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'counterparty', subBuilder: $4.Counterparty.create)
    ..aOM<$4.Version>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'version', subBuilder: $4.Version.create)
    ..a<$fixnum.Int64>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delayPeriod', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgConnectionOpenInit._() : super();
  factory MsgConnectionOpenInit() => create();
  factory MsgConnectionOpenInit.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgConnectionOpenInit.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenInit clone() => MsgConnectionOpenInit()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenInit copyWith(void Function(MsgConnectionOpenInit) updates) => super.copyWith((message) => updates(message as MsgConnectionOpenInit)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenInit create() => MsgConnectionOpenInit._();
  MsgConnectionOpenInit createEmptyInstance() => create();
  static $pb.PbList<MsgConnectionOpenInit> createRepeated() => $pb.PbList<MsgConnectionOpenInit>();
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenInit getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgConnectionOpenInit>(create);
  static MsgConnectionOpenInit _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $4.Counterparty get counterparty => $_getN(1);
  @$pb.TagNumber(2)
  set counterparty($4.Counterparty v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasCounterparty() => $_has(1);
  @$pb.TagNumber(2)
  void clearCounterparty() => clearField(2);
  @$pb.TagNumber(2)
  $4.Counterparty ensureCounterparty() => $_ensure(1);

  @$pb.TagNumber(3)
  $4.Version get version => $_getN(2);
  @$pb.TagNumber(3)
  set version($4.Version v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasVersion() => $_has(2);
  @$pb.TagNumber(3)
  void clearVersion() => clearField(3);
  @$pb.TagNumber(3)
  $4.Version ensureVersion() => $_ensure(2);

  @$pb.TagNumber(4)
  $fixnum.Int64 get delayPeriod => $_getI64(3);
  @$pb.TagNumber(4)
  set delayPeriod($fixnum.Int64 v) { $_setInt64(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasDelayPeriod() => $_has(3);
  @$pb.TagNumber(4)
  void clearDelayPeriod() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get signer => $_getSZ(4);
  @$pb.TagNumber(5)
  set signer($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasSigner() => $_has(4);
  @$pb.TagNumber(5)
  void clearSigner() => clearField(5);
}

class MsgConnectionOpenInitResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgConnectionOpenInitResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgConnectionOpenInitResponse._() : super();
  factory MsgConnectionOpenInitResponse() => create();
  factory MsgConnectionOpenInitResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgConnectionOpenInitResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenInitResponse clone() => MsgConnectionOpenInitResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenInitResponse copyWith(void Function(MsgConnectionOpenInitResponse) updates) => super.copyWith((message) => updates(message as MsgConnectionOpenInitResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenInitResponse create() => MsgConnectionOpenInitResponse._();
  MsgConnectionOpenInitResponse createEmptyInstance() => create();
  static $pb.PbList<MsgConnectionOpenInitResponse> createRepeated() => $pb.PbList<MsgConnectionOpenInitResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenInitResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgConnectionOpenInitResponse>(create);
  static MsgConnectionOpenInitResponse _defaultInstance;
}

class MsgConnectionOpenTry extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgConnectionOpenTry', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'previousConnectionId')
    ..aOM<$5.Any>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientState', subBuilder: $5.Any.create)
    ..aOM<$4.Counterparty>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'counterparty', subBuilder: $4.Counterparty.create)
    ..a<$fixnum.Int64>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delayPeriod', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..pc<$4.Version>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'counterpartyVersions', $pb.PbFieldType.PM, subBuilder: $4.Version.create)
    ..aOM<$7.Height>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $7.Height.create)
    ..a<$core.List<$core.int>>(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofInit', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofClient', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(10, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofConsensus', $pb.PbFieldType.OY)
    ..aOM<$7.Height>(11, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'consensusHeight', subBuilder: $7.Height.create)
    ..aOS(12, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgConnectionOpenTry._() : super();
  factory MsgConnectionOpenTry() => create();
  factory MsgConnectionOpenTry.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgConnectionOpenTry.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenTry clone() => MsgConnectionOpenTry()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenTry copyWith(void Function(MsgConnectionOpenTry) updates) => super.copyWith((message) => updates(message as MsgConnectionOpenTry)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenTry create() => MsgConnectionOpenTry._();
  MsgConnectionOpenTry createEmptyInstance() => create();
  static $pb.PbList<MsgConnectionOpenTry> createRepeated() => $pb.PbList<MsgConnectionOpenTry>();
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenTry getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgConnectionOpenTry>(create);
  static MsgConnectionOpenTry _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get previousConnectionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set previousConnectionId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPreviousConnectionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPreviousConnectionId() => clearField(2);

  @$pb.TagNumber(3)
  $5.Any get clientState => $_getN(2);
  @$pb.TagNumber(3)
  set clientState($5.Any v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasClientState() => $_has(2);
  @$pb.TagNumber(3)
  void clearClientState() => clearField(3);
  @$pb.TagNumber(3)
  $5.Any ensureClientState() => $_ensure(2);

  @$pb.TagNumber(4)
  $4.Counterparty get counterparty => $_getN(3);
  @$pb.TagNumber(4)
  set counterparty($4.Counterparty v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasCounterparty() => $_has(3);
  @$pb.TagNumber(4)
  void clearCounterparty() => clearField(4);
  @$pb.TagNumber(4)
  $4.Counterparty ensureCounterparty() => $_ensure(3);

  @$pb.TagNumber(5)
  $fixnum.Int64 get delayPeriod => $_getI64(4);
  @$pb.TagNumber(5)
  set delayPeriod($fixnum.Int64 v) { $_setInt64(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasDelayPeriod() => $_has(4);
  @$pb.TagNumber(5)
  void clearDelayPeriod() => clearField(5);

  @$pb.TagNumber(6)
  $core.List<$4.Version> get counterpartyVersions => $_getList(5);

  @$pb.TagNumber(7)
  $7.Height get proofHeight => $_getN(6);
  @$pb.TagNumber(7)
  set proofHeight($7.Height v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasProofHeight() => $_has(6);
  @$pb.TagNumber(7)
  void clearProofHeight() => clearField(7);
  @$pb.TagNumber(7)
  $7.Height ensureProofHeight() => $_ensure(6);

  @$pb.TagNumber(8)
  $core.List<$core.int> get proofInit => $_getN(7);
  @$pb.TagNumber(8)
  set proofInit($core.List<$core.int> v) { $_setBytes(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasProofInit() => $_has(7);
  @$pb.TagNumber(8)
  void clearProofInit() => clearField(8);

  @$pb.TagNumber(9)
  $core.List<$core.int> get proofClient => $_getN(8);
  @$pb.TagNumber(9)
  set proofClient($core.List<$core.int> v) { $_setBytes(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasProofClient() => $_has(8);
  @$pb.TagNumber(9)
  void clearProofClient() => clearField(9);

  @$pb.TagNumber(10)
  $core.List<$core.int> get proofConsensus => $_getN(9);
  @$pb.TagNumber(10)
  set proofConsensus($core.List<$core.int> v) { $_setBytes(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasProofConsensus() => $_has(9);
  @$pb.TagNumber(10)
  void clearProofConsensus() => clearField(10);

  @$pb.TagNumber(11)
  $7.Height get consensusHeight => $_getN(10);
  @$pb.TagNumber(11)
  set consensusHeight($7.Height v) { setField(11, v); }
  @$pb.TagNumber(11)
  $core.bool hasConsensusHeight() => $_has(10);
  @$pb.TagNumber(11)
  void clearConsensusHeight() => clearField(11);
  @$pb.TagNumber(11)
  $7.Height ensureConsensusHeight() => $_ensure(10);

  @$pb.TagNumber(12)
  $core.String get signer => $_getSZ(11);
  @$pb.TagNumber(12)
  set signer($core.String v) { $_setString(11, v); }
  @$pb.TagNumber(12)
  $core.bool hasSigner() => $_has(11);
  @$pb.TagNumber(12)
  void clearSigner() => clearField(12);
}

class MsgConnectionOpenTryResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgConnectionOpenTryResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgConnectionOpenTryResponse._() : super();
  factory MsgConnectionOpenTryResponse() => create();
  factory MsgConnectionOpenTryResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgConnectionOpenTryResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenTryResponse clone() => MsgConnectionOpenTryResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenTryResponse copyWith(void Function(MsgConnectionOpenTryResponse) updates) => super.copyWith((message) => updates(message as MsgConnectionOpenTryResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenTryResponse create() => MsgConnectionOpenTryResponse._();
  MsgConnectionOpenTryResponse createEmptyInstance() => create();
  static $pb.PbList<MsgConnectionOpenTryResponse> createRepeated() => $pb.PbList<MsgConnectionOpenTryResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenTryResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgConnectionOpenTryResponse>(create);
  static MsgConnectionOpenTryResponse _defaultInstance;
}

class MsgConnectionOpenAck extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgConnectionOpenAck', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'connectionId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'counterpartyConnectionId')
    ..aOM<$4.Version>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'version', subBuilder: $4.Version.create)
    ..aOM<$5.Any>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientState', subBuilder: $5.Any.create)
    ..aOM<$7.Height>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $7.Height.create)
    ..a<$core.List<$core.int>>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofTry', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofClient', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofConsensus', $pb.PbFieldType.OY)
    ..aOM<$7.Height>(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'consensusHeight', subBuilder: $7.Height.create)
    ..aOS(10, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgConnectionOpenAck._() : super();
  factory MsgConnectionOpenAck() => create();
  factory MsgConnectionOpenAck.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgConnectionOpenAck.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenAck clone() => MsgConnectionOpenAck()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenAck copyWith(void Function(MsgConnectionOpenAck) updates) => super.copyWith((message) => updates(message as MsgConnectionOpenAck)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenAck create() => MsgConnectionOpenAck._();
  MsgConnectionOpenAck createEmptyInstance() => create();
  static $pb.PbList<MsgConnectionOpenAck> createRepeated() => $pb.PbList<MsgConnectionOpenAck>();
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenAck getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgConnectionOpenAck>(create);
  static MsgConnectionOpenAck _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get connectionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set connectionId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasConnectionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearConnectionId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get counterpartyConnectionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set counterpartyConnectionId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCounterpartyConnectionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCounterpartyConnectionId() => clearField(2);

  @$pb.TagNumber(3)
  $4.Version get version => $_getN(2);
  @$pb.TagNumber(3)
  set version($4.Version v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasVersion() => $_has(2);
  @$pb.TagNumber(3)
  void clearVersion() => clearField(3);
  @$pb.TagNumber(3)
  $4.Version ensureVersion() => $_ensure(2);

  @$pb.TagNumber(4)
  $5.Any get clientState => $_getN(3);
  @$pb.TagNumber(4)
  set clientState($5.Any v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasClientState() => $_has(3);
  @$pb.TagNumber(4)
  void clearClientState() => clearField(4);
  @$pb.TagNumber(4)
  $5.Any ensureClientState() => $_ensure(3);

  @$pb.TagNumber(5)
  $7.Height get proofHeight => $_getN(4);
  @$pb.TagNumber(5)
  set proofHeight($7.Height v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasProofHeight() => $_has(4);
  @$pb.TagNumber(5)
  void clearProofHeight() => clearField(5);
  @$pb.TagNumber(5)
  $7.Height ensureProofHeight() => $_ensure(4);

  @$pb.TagNumber(6)
  $core.List<$core.int> get proofTry => $_getN(5);
  @$pb.TagNumber(6)
  set proofTry($core.List<$core.int> v) { $_setBytes(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasProofTry() => $_has(5);
  @$pb.TagNumber(6)
  void clearProofTry() => clearField(6);

  @$pb.TagNumber(7)
  $core.List<$core.int> get proofClient => $_getN(6);
  @$pb.TagNumber(7)
  set proofClient($core.List<$core.int> v) { $_setBytes(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasProofClient() => $_has(6);
  @$pb.TagNumber(7)
  void clearProofClient() => clearField(7);

  @$pb.TagNumber(8)
  $core.List<$core.int> get proofConsensus => $_getN(7);
  @$pb.TagNumber(8)
  set proofConsensus($core.List<$core.int> v) { $_setBytes(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasProofConsensus() => $_has(7);
  @$pb.TagNumber(8)
  void clearProofConsensus() => clearField(8);

  @$pb.TagNumber(9)
  $7.Height get consensusHeight => $_getN(8);
  @$pb.TagNumber(9)
  set consensusHeight($7.Height v) { setField(9, v); }
  @$pb.TagNumber(9)
  $core.bool hasConsensusHeight() => $_has(8);
  @$pb.TagNumber(9)
  void clearConsensusHeight() => clearField(9);
  @$pb.TagNumber(9)
  $7.Height ensureConsensusHeight() => $_ensure(8);

  @$pb.TagNumber(10)
  $core.String get signer => $_getSZ(9);
  @$pb.TagNumber(10)
  set signer($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasSigner() => $_has(9);
  @$pb.TagNumber(10)
  void clearSigner() => clearField(10);
}

class MsgConnectionOpenAckResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgConnectionOpenAckResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgConnectionOpenAckResponse._() : super();
  factory MsgConnectionOpenAckResponse() => create();
  factory MsgConnectionOpenAckResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgConnectionOpenAckResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenAckResponse clone() => MsgConnectionOpenAckResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenAckResponse copyWith(void Function(MsgConnectionOpenAckResponse) updates) => super.copyWith((message) => updates(message as MsgConnectionOpenAckResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenAckResponse create() => MsgConnectionOpenAckResponse._();
  MsgConnectionOpenAckResponse createEmptyInstance() => create();
  static $pb.PbList<MsgConnectionOpenAckResponse> createRepeated() => $pb.PbList<MsgConnectionOpenAckResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenAckResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgConnectionOpenAckResponse>(create);
  static MsgConnectionOpenAckResponse _defaultInstance;
}

class MsgConnectionOpenConfirm extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgConnectionOpenConfirm', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'connectionId')
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofAck', $pb.PbFieldType.OY)
    ..aOM<$7.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $7.Height.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signer')
    ..hasRequiredFields = false
  ;

  MsgConnectionOpenConfirm._() : super();
  factory MsgConnectionOpenConfirm() => create();
  factory MsgConnectionOpenConfirm.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgConnectionOpenConfirm.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenConfirm clone() => MsgConnectionOpenConfirm()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenConfirm copyWith(void Function(MsgConnectionOpenConfirm) updates) => super.copyWith((message) => updates(message as MsgConnectionOpenConfirm)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenConfirm create() => MsgConnectionOpenConfirm._();
  MsgConnectionOpenConfirm createEmptyInstance() => create();
  static $pb.PbList<MsgConnectionOpenConfirm> createRepeated() => $pb.PbList<MsgConnectionOpenConfirm>();
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenConfirm getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgConnectionOpenConfirm>(create);
  static MsgConnectionOpenConfirm _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get connectionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set connectionId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasConnectionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearConnectionId() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get proofAck => $_getN(1);
  @$pb.TagNumber(2)
  set proofAck($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasProofAck() => $_has(1);
  @$pb.TagNumber(2)
  void clearProofAck() => clearField(2);

  @$pb.TagNumber(3)
  $7.Height get proofHeight => $_getN(2);
  @$pb.TagNumber(3)
  set proofHeight($7.Height v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasProofHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearProofHeight() => clearField(3);
  @$pb.TagNumber(3)
  $7.Height ensureProofHeight() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.String get signer => $_getSZ(3);
  @$pb.TagNumber(4)
  set signer($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasSigner() => $_has(3);
  @$pb.TagNumber(4)
  void clearSigner() => clearField(4);
}

class MsgConnectionOpenConfirmResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MsgConnectionOpenConfirmResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  MsgConnectionOpenConfirmResponse._() : super();
  factory MsgConnectionOpenConfirmResponse() => create();
  factory MsgConnectionOpenConfirmResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MsgConnectionOpenConfirmResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenConfirmResponse clone() => MsgConnectionOpenConfirmResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MsgConnectionOpenConfirmResponse copyWith(void Function(MsgConnectionOpenConfirmResponse) updates) => super.copyWith((message) => updates(message as MsgConnectionOpenConfirmResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenConfirmResponse create() => MsgConnectionOpenConfirmResponse._();
  MsgConnectionOpenConfirmResponse createEmptyInstance() => create();
  static $pb.PbList<MsgConnectionOpenConfirmResponse> createRepeated() => $pb.PbList<MsgConnectionOpenConfirmResponse>();
  @$core.pragma('dart2js:noInline')
  static MsgConnectionOpenConfirmResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MsgConnectionOpenConfirmResponse>(create);
  static MsgConnectionOpenConfirmResponse _defaultInstance;
}

