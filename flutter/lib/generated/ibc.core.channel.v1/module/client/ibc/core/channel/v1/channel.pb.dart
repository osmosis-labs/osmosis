///
//  Generated code. Do not modify.
//  source: ibc/core/channel/v1/channel.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../client/v1/client.pb.dart' as $3;

import 'channel.pbenum.dart';

export 'channel.pbenum.dart';

class Channel extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Channel', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..e<State>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'state', $pb.PbFieldType.OE, defaultOrMaker: State.STATE_UNINITIALIZED_UNSPECIFIED, valueOf: State.valueOf, enumValues: State.values)
    ..e<Order>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'ordering', $pb.PbFieldType.OE, defaultOrMaker: Order.ORDER_NONE_UNSPECIFIED, valueOf: Order.valueOf, enumValues: Order.values)
    ..aOM<Counterparty>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'counterparty', subBuilder: Counterparty.create)
    ..pPS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'connectionHops')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'version')
    ..hasRequiredFields = false
  ;

  Channel._() : super();
  factory Channel() => create();
  factory Channel.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Channel.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Channel clone() => Channel()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Channel copyWith(void Function(Channel) updates) => super.copyWith((message) => updates(message as Channel)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Channel create() => Channel._();
  Channel createEmptyInstance() => create();
  static $pb.PbList<Channel> createRepeated() => $pb.PbList<Channel>();
  @$core.pragma('dart2js:noInline')
  static Channel getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Channel>(create);
  static Channel _defaultInstance;

  @$pb.TagNumber(1)
  State get state => $_getN(0);
  @$pb.TagNumber(1)
  set state(State v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasState() => $_has(0);
  @$pb.TagNumber(1)
  void clearState() => clearField(1);

  @$pb.TagNumber(2)
  Order get ordering => $_getN(1);
  @$pb.TagNumber(2)
  set ordering(Order v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasOrdering() => $_has(1);
  @$pb.TagNumber(2)
  void clearOrdering() => clearField(2);

  @$pb.TagNumber(3)
  Counterparty get counterparty => $_getN(2);
  @$pb.TagNumber(3)
  set counterparty(Counterparty v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasCounterparty() => $_has(2);
  @$pb.TagNumber(3)
  void clearCounterparty() => clearField(3);
  @$pb.TagNumber(3)
  Counterparty ensureCounterparty() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.List<$core.String> get connectionHops => $_getList(3);

  @$pb.TagNumber(5)
  $core.String get version => $_getSZ(4);
  @$pb.TagNumber(5)
  set version($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasVersion() => $_has(4);
  @$pb.TagNumber(5)
  void clearVersion() => clearField(5);
}

class IdentifiedChannel extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'IdentifiedChannel', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..e<State>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'state', $pb.PbFieldType.OE, defaultOrMaker: State.STATE_UNINITIALIZED_UNSPECIFIED, valueOf: State.valueOf, enumValues: State.values)
    ..e<Order>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'ordering', $pb.PbFieldType.OE, defaultOrMaker: Order.ORDER_NONE_UNSPECIFIED, valueOf: Order.valueOf, enumValues: Order.values)
    ..aOM<Counterparty>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'counterparty', subBuilder: Counterparty.create)
    ..pPS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'connectionHops')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'version')
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..hasRequiredFields = false
  ;

  IdentifiedChannel._() : super();
  factory IdentifiedChannel() => create();
  factory IdentifiedChannel.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IdentifiedChannel.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IdentifiedChannel clone() => IdentifiedChannel()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IdentifiedChannel copyWith(void Function(IdentifiedChannel) updates) => super.copyWith((message) => updates(message as IdentifiedChannel)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static IdentifiedChannel create() => IdentifiedChannel._();
  IdentifiedChannel createEmptyInstance() => create();
  static $pb.PbList<IdentifiedChannel> createRepeated() => $pb.PbList<IdentifiedChannel>();
  @$core.pragma('dart2js:noInline')
  static IdentifiedChannel getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<IdentifiedChannel>(create);
  static IdentifiedChannel _defaultInstance;

  @$pb.TagNumber(1)
  State get state => $_getN(0);
  @$pb.TagNumber(1)
  set state(State v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasState() => $_has(0);
  @$pb.TagNumber(1)
  void clearState() => clearField(1);

  @$pb.TagNumber(2)
  Order get ordering => $_getN(1);
  @$pb.TagNumber(2)
  set ordering(Order v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasOrdering() => $_has(1);
  @$pb.TagNumber(2)
  void clearOrdering() => clearField(2);

  @$pb.TagNumber(3)
  Counterparty get counterparty => $_getN(2);
  @$pb.TagNumber(3)
  set counterparty(Counterparty v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasCounterparty() => $_has(2);
  @$pb.TagNumber(3)
  void clearCounterparty() => clearField(3);
  @$pb.TagNumber(3)
  Counterparty ensureCounterparty() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.List<$core.String> get connectionHops => $_getList(3);

  @$pb.TagNumber(5)
  $core.String get version => $_getSZ(4);
  @$pb.TagNumber(5)
  set version($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasVersion() => $_has(4);
  @$pb.TagNumber(5)
  void clearVersion() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get portId => $_getSZ(5);
  @$pb.TagNumber(6)
  set portId($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasPortId() => $_has(5);
  @$pb.TagNumber(6)
  void clearPortId() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get channelId => $_getSZ(6);
  @$pb.TagNumber(7)
  set channelId($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasChannelId() => $_has(6);
  @$pb.TagNumber(7)
  void clearChannelId() => clearField(7);
}

class Counterparty extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Counterparty', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
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
}

class Packet extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Packet', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sequence', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sourcePort')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sourceChannel')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'destinationPort')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'destinationChannel')
    ..a<$core.List<$core.int>>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'data', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'timeoutHeight', subBuilder: $3.Height.create)
    ..a<$fixnum.Int64>(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'timeoutTimestamp', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  Packet._() : super();
  factory Packet() => create();
  factory Packet.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Packet.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Packet clone() => Packet()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Packet copyWith(void Function(Packet) updates) => super.copyWith((message) => updates(message as Packet)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Packet create() => Packet._();
  Packet createEmptyInstance() => create();
  static $pb.PbList<Packet> createRepeated() => $pb.PbList<Packet>();
  @$core.pragma('dart2js:noInline')
  static Packet getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Packet>(create);
  static Packet _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get sequence => $_getI64(0);
  @$pb.TagNumber(1)
  set sequence($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSequence() => $_has(0);
  @$pb.TagNumber(1)
  void clearSequence() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get sourcePort => $_getSZ(1);
  @$pb.TagNumber(2)
  set sourcePort($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasSourcePort() => $_has(1);
  @$pb.TagNumber(2)
  void clearSourcePort() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get sourceChannel => $_getSZ(2);
  @$pb.TagNumber(3)
  set sourceChannel($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSourceChannel() => $_has(2);
  @$pb.TagNumber(3)
  void clearSourceChannel() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get destinationPort => $_getSZ(3);
  @$pb.TagNumber(4)
  set destinationPort($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasDestinationPort() => $_has(3);
  @$pb.TagNumber(4)
  void clearDestinationPort() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get destinationChannel => $_getSZ(4);
  @$pb.TagNumber(5)
  set destinationChannel($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasDestinationChannel() => $_has(4);
  @$pb.TagNumber(5)
  void clearDestinationChannel() => clearField(5);

  @$pb.TagNumber(6)
  $core.List<$core.int> get data => $_getN(5);
  @$pb.TagNumber(6)
  set data($core.List<$core.int> v) { $_setBytes(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasData() => $_has(5);
  @$pb.TagNumber(6)
  void clearData() => clearField(6);

  @$pb.TagNumber(7)
  $3.Height get timeoutHeight => $_getN(6);
  @$pb.TagNumber(7)
  set timeoutHeight($3.Height v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasTimeoutHeight() => $_has(6);
  @$pb.TagNumber(7)
  void clearTimeoutHeight() => clearField(7);
  @$pb.TagNumber(7)
  $3.Height ensureTimeoutHeight() => $_ensure(6);

  @$pb.TagNumber(8)
  $fixnum.Int64 get timeoutTimestamp => $_getI64(7);
  @$pb.TagNumber(8)
  set timeoutTimestamp($fixnum.Int64 v) { $_setInt64(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasTimeoutTimestamp() => $_has(7);
  @$pb.TagNumber(8)
  void clearTimeoutTimestamp() => clearField(8);
}

class PacketState extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'PacketState', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..a<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sequence', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$core.List<$core.int>>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'data', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  PacketState._() : super();
  factory PacketState() => create();
  factory PacketState.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PacketState.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PacketState clone() => PacketState()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PacketState copyWith(void Function(PacketState) updates) => super.copyWith((message) => updates(message as PacketState)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static PacketState create() => PacketState._();
  PacketState createEmptyInstance() => create();
  static $pb.PbList<PacketState> createRepeated() => $pb.PbList<PacketState>();
  @$core.pragma('dart2js:noInline')
  static PacketState getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PacketState>(create);
  static PacketState _defaultInstance;

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
  $fixnum.Int64 get sequence => $_getI64(2);
  @$pb.TagNumber(3)
  set sequence($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSequence() => $_has(2);
  @$pb.TagNumber(3)
  void clearSequence() => clearField(3);

  @$pb.TagNumber(4)
  $core.List<$core.int> get data => $_getN(3);
  @$pb.TagNumber(4)
  set data($core.List<$core.int> v) { $_setBytes(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(4)
  void clearData() => clearField(4);
}

enum Acknowledgement_Response {
  result, 
  error, 
  notSet
}

class Acknowledgement extends $pb.GeneratedMessage {
  static const $core.Map<$core.int, Acknowledgement_Response> _Acknowledgement_ResponseByTag = {
    21 : Acknowledgement_Response.result,
    22 : Acknowledgement_Response.error,
    0 : Acknowledgement_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Acknowledgement', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..oo(0, [21, 22])
    ..a<$core.List<$core.int>>(21, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'result', $pb.PbFieldType.OY)
    ..aOS(22, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'error')
    ..hasRequiredFields = false
  ;

  Acknowledgement._() : super();
  factory Acknowledgement() => create();
  factory Acknowledgement.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Acknowledgement.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Acknowledgement clone() => Acknowledgement()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Acknowledgement copyWith(void Function(Acknowledgement) updates) => super.copyWith((message) => updates(message as Acknowledgement)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Acknowledgement create() => Acknowledgement._();
  Acknowledgement createEmptyInstance() => create();
  static $pb.PbList<Acknowledgement> createRepeated() => $pb.PbList<Acknowledgement>();
  @$core.pragma('dart2js:noInline')
  static Acknowledgement getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Acknowledgement>(create);
  static Acknowledgement _defaultInstance;

  Acknowledgement_Response whichResponse() => _Acknowledgement_ResponseByTag[$_whichOneof(0)];
  void clearResponse() => clearField($_whichOneof(0));

  @$pb.TagNumber(21)
  $core.List<$core.int> get result => $_getN(0);
  @$pb.TagNumber(21)
  set result($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(21)
  $core.bool hasResult() => $_has(0);
  @$pb.TagNumber(21)
  void clearResult() => clearField(21);

  @$pb.TagNumber(22)
  $core.String get error => $_getSZ(1);
  @$pb.TagNumber(22)
  set error($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(22)
  $core.bool hasError() => $_has(1);
  @$pb.TagNumber(22)
  void clearError() => clearField(22);
}

