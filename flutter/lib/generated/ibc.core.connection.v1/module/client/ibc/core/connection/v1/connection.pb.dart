///
//  Generated code. Do not modify.
//  source: ibc/core/connection/v1/connection.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../commitment/v1/commitment.pb.dart' as $3;

import 'connection.pbenum.dart';

export 'connection.pbenum.dart';

class ConnectionEnd extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ConnectionEnd', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..pc<Version>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'versions', $pb.PbFieldType.PM, subBuilder: Version.create)
    ..e<State>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'state', $pb.PbFieldType.OE, defaultOrMaker: State.STATE_UNINITIALIZED_UNSPECIFIED, valueOf: State.valueOf, enumValues: State.values)
    ..aOM<Counterparty>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'counterparty', subBuilder: Counterparty.create)
    ..a<$fixnum.Int64>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delayPeriod', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  ConnectionEnd._() : super();
  factory ConnectionEnd() => create();
  factory ConnectionEnd.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConnectionEnd.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConnectionEnd clone() => ConnectionEnd()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConnectionEnd copyWith(void Function(ConnectionEnd) updates) => super.copyWith((message) => updates(message as ConnectionEnd)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ConnectionEnd create() => ConnectionEnd._();
  ConnectionEnd createEmptyInstance() => create();
  static $pb.PbList<ConnectionEnd> createRepeated() => $pb.PbList<ConnectionEnd>();
  @$core.pragma('dart2js:noInline')
  static ConnectionEnd getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConnectionEnd>(create);
  static ConnectionEnd _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<Version> get versions => $_getList(1);

  @$pb.TagNumber(3)
  State get state => $_getN(2);
  @$pb.TagNumber(3)
  set state(State v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasState() => $_has(2);
  @$pb.TagNumber(3)
  void clearState() => clearField(3);

  @$pb.TagNumber(4)
  Counterparty get counterparty => $_getN(3);
  @$pb.TagNumber(4)
  set counterparty(Counterparty v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasCounterparty() => $_has(3);
  @$pb.TagNumber(4)
  void clearCounterparty() => clearField(4);
  @$pb.TagNumber(4)
  Counterparty ensureCounterparty() => $_ensure(3);

  @$pb.TagNumber(5)
  $fixnum.Int64 get delayPeriod => $_getI64(4);
  @$pb.TagNumber(5)
  set delayPeriod($fixnum.Int64 v) { $_setInt64(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasDelayPeriod() => $_has(4);
  @$pb.TagNumber(5)
  void clearDelayPeriod() => clearField(5);
}

class IdentifiedConnection extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'IdentifiedConnection', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..pc<Version>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'versions', $pb.PbFieldType.PM, subBuilder: Version.create)
    ..e<State>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'state', $pb.PbFieldType.OE, defaultOrMaker: State.STATE_UNINITIALIZED_UNSPECIFIED, valueOf: State.valueOf, enumValues: State.values)
    ..aOM<Counterparty>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'counterparty', subBuilder: Counterparty.create)
    ..a<$fixnum.Int64>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delayPeriod', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  IdentifiedConnection._() : super();
  factory IdentifiedConnection() => create();
  factory IdentifiedConnection.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IdentifiedConnection.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IdentifiedConnection clone() => IdentifiedConnection()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IdentifiedConnection copyWith(void Function(IdentifiedConnection) updates) => super.copyWith((message) => updates(message as IdentifiedConnection)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static IdentifiedConnection create() => IdentifiedConnection._();
  IdentifiedConnection createEmptyInstance() => create();
  static $pb.PbList<IdentifiedConnection> createRepeated() => $pb.PbList<IdentifiedConnection>();
  @$core.pragma('dart2js:noInline')
  static IdentifiedConnection getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<IdentifiedConnection>(create);
  static IdentifiedConnection _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get clientId => $_getSZ(1);
  @$pb.TagNumber(2)
  set clientId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasClientId() => $_has(1);
  @$pb.TagNumber(2)
  void clearClientId() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<Version> get versions => $_getList(2);

  @$pb.TagNumber(4)
  State get state => $_getN(3);
  @$pb.TagNumber(4)
  set state(State v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasState() => $_has(3);
  @$pb.TagNumber(4)
  void clearState() => clearField(4);

  @$pb.TagNumber(5)
  Counterparty get counterparty => $_getN(4);
  @$pb.TagNumber(5)
  set counterparty(Counterparty v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasCounterparty() => $_has(4);
  @$pb.TagNumber(5)
  void clearCounterparty() => clearField(5);
  @$pb.TagNumber(5)
  Counterparty ensureCounterparty() => $_ensure(4);

  @$pb.TagNumber(6)
  $fixnum.Int64 get delayPeriod => $_getI64(5);
  @$pb.TagNumber(6)
  set delayPeriod($fixnum.Int64 v) { $_setInt64(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasDelayPeriod() => $_has(5);
  @$pb.TagNumber(6)
  void clearDelayPeriod() => clearField(6);
}

class Counterparty extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Counterparty', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'connectionId')
    ..aOM<$3.MerklePrefix>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'prefix', subBuilder: $3.MerklePrefix.create)
    ..hasRequiredFields = false
  ;

  Counterparty._() : super();
  factory Counterparty() => create();
  factory Counterparty.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Counterparty.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Counterparty clone() => Counterparty()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Counterparty copyWith(void Function(Counterparty) updates) => super.copyWith((message) => updates(message as Counterparty)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Counterparty create() => Counterparty._();
  Counterparty createEmptyInstance() => create();
  static $pb.PbList<Counterparty> createRepeated() => $pb.PbList<Counterparty>();
  @$core.pragma('dart2js:noInline')
  static Counterparty getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Counterparty>(create);
  static Counterparty _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get connectionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set connectionId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasConnectionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearConnectionId() => clearField(2);

  @$pb.TagNumber(3)
  $3.MerklePrefix get prefix => $_getN(2);
  @$pb.TagNumber(3)
  set prefix($3.MerklePrefix v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasPrefix() => $_has(2);
  @$pb.TagNumber(3)
  void clearPrefix() => clearField(3);
  @$pb.TagNumber(3)
  $3.MerklePrefix ensurePrefix() => $_ensure(2);
}

class ClientPaths extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ClientPaths', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..pPS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'paths')
    ..hasRequiredFields = false
  ;

  ClientPaths._() : super();
  factory ClientPaths() => create();
  factory ClientPaths.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ClientPaths.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ClientPaths clone() => ClientPaths()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ClientPaths copyWith(void Function(ClientPaths) updates) => super.copyWith((message) => updates(message as ClientPaths)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ClientPaths create() => ClientPaths._();
  ClientPaths createEmptyInstance() => create();
  static $pb.PbList<ClientPaths> createRepeated() => $pb.PbList<ClientPaths>();
  @$core.pragma('dart2js:noInline')
  static ClientPaths getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ClientPaths>(create);
  static ClientPaths _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.String> get paths => $_getList(0);
}

class ConnectionPaths extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ConnectionPaths', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..pPS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'paths')
    ..hasRequiredFields = false
  ;

  ConnectionPaths._() : super();
  factory ConnectionPaths() => create();
  factory ConnectionPaths.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConnectionPaths.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConnectionPaths clone() => ConnectionPaths()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConnectionPaths copyWith(void Function(ConnectionPaths) updates) => super.copyWith((message) => updates(message as ConnectionPaths)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ConnectionPaths create() => ConnectionPaths._();
  ConnectionPaths createEmptyInstance() => create();
  static $pb.PbList<ConnectionPaths> createRepeated() => $pb.PbList<ConnectionPaths>();
  @$core.pragma('dart2js:noInline')
  static ConnectionPaths getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConnectionPaths>(create);
  static ConnectionPaths _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.String> get paths => $_getList(1);
}

class Version extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Version', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.connection.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'identifier')
    ..pPS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'features')
    ..hasRequiredFields = false
  ;

  Version._() : super();
  factory Version() => create();
  factory Version.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Version.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Version clone() => Version()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Version copyWith(void Function(Version) updates) => super.copyWith((message) => updates(message as Version)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Version create() => Version._();
  Version createEmptyInstance() => create();
  static $pb.PbList<Version> createRepeated() => $pb.PbList<Version>();
  @$core.pragma('dart2js:noInline')
  static Version getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Version>(create);
  static Version _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get identifier => $_getSZ(0);
  @$pb.TagNumber(1)
  set identifier($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasIdentifier() => $_has(0);
  @$pb.TagNumber(1)
  void clearIdentifier() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.String> get features => $_getList(1);
}

