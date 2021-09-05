///
//  Generated code. Do not modify.
//  source: tendermint/abci/types.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../google/protobuf/timestamp.pb.dart' as $4;
import '../types/types.pb.dart' as $7;
import '../crypto/proof.pb.dart' as $2;
import '../types/params.pb.dart' as $8;
import '../crypto/keys.pb.dart' as $1;

import 'types.pbenum.dart';

export 'types.pbenum.dart';

enum Request_Value {
  echo, 
  flush, 
  info, 
  setOption, 
  initChain, 
  query, 
  beginBlock, 
  checkTx, 
  deliverTx, 
  endBlock, 
  commit, 
  listSnapshots, 
  offerSnapshot, 
  loadSnapshotChunk, 
  applySnapshotChunk, 
  notSet
}

class Request extends $pb.GeneratedMessage {
  static const $core.Map<$core.int, Request_Value> _Request_ValueByTag = {
    1 : Request_Value.echo,
    2 : Request_Value.flush,
    3 : Request_Value.info,
    4 : Request_Value.setOption,
    5 : Request_Value.initChain,
    6 : Request_Value.query,
    7 : Request_Value.beginBlock,
    8 : Request_Value.checkTx,
    9 : Request_Value.deliverTx,
    10 : Request_Value.endBlock,
    11 : Request_Value.commit,
    12 : Request_Value.listSnapshots,
    13 : Request_Value.offerSnapshot,
    14 : Request_Value.loadSnapshotChunk,
    15 : Request_Value.applySnapshotChunk,
    0 : Request_Value.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Request', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..oo(0, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15])
    ..aOM<RequestEcho>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'echo', subBuilder: RequestEcho.create)
    ..aOM<RequestFlush>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'flush', subBuilder: RequestFlush.create)
    ..aOM<RequestInfo>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'info', subBuilder: RequestInfo.create)
    ..aOM<RequestSetOption>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'setOption', subBuilder: RequestSetOption.create)
    ..aOM<RequestInitChain>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'initChain', subBuilder: RequestInitChain.create)
    ..aOM<RequestQuery>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'query', subBuilder: RequestQuery.create)
    ..aOM<RequestBeginBlock>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'beginBlock', subBuilder: RequestBeginBlock.create)
    ..aOM<RequestCheckTx>(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'checkTx', subBuilder: RequestCheckTx.create)
    ..aOM<RequestDeliverTx>(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'deliverTx', subBuilder: RequestDeliverTx.create)
    ..aOM<RequestEndBlock>(10, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'endBlock', subBuilder: RequestEndBlock.create)
    ..aOM<RequestCommit>(11, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'commit', subBuilder: RequestCommit.create)
    ..aOM<RequestListSnapshots>(12, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'listSnapshots', subBuilder: RequestListSnapshots.create)
    ..aOM<RequestOfferSnapshot>(13, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'offerSnapshot', subBuilder: RequestOfferSnapshot.create)
    ..aOM<RequestLoadSnapshotChunk>(14, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'loadSnapshotChunk', subBuilder: RequestLoadSnapshotChunk.create)
    ..aOM<RequestApplySnapshotChunk>(15, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'applySnapshotChunk', subBuilder: RequestApplySnapshotChunk.create)
    ..hasRequiredFields = false
  ;

  Request._() : super();
  factory Request() => create();
  factory Request.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Request.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Request clone() => Request()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Request copyWith(void Function(Request) updates) => super.copyWith((message) => updates(message as Request)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Request create() => Request._();
  Request createEmptyInstance() => create();
  static $pb.PbList<Request> createRepeated() => $pb.PbList<Request>();
  @$core.pragma('dart2js:noInline')
  static Request getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Request>(create);
  static Request _defaultInstance;

  Request_Value whichValue() => _Request_ValueByTag[$_whichOneof(0)];
  void clearValue() => clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  RequestEcho get echo => $_getN(0);
  @$pb.TagNumber(1)
  set echo(RequestEcho v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasEcho() => $_has(0);
  @$pb.TagNumber(1)
  void clearEcho() => clearField(1);
  @$pb.TagNumber(1)
  RequestEcho ensureEcho() => $_ensure(0);

  @$pb.TagNumber(2)
  RequestFlush get flush => $_getN(1);
  @$pb.TagNumber(2)
  set flush(RequestFlush v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasFlush() => $_has(1);
  @$pb.TagNumber(2)
  void clearFlush() => clearField(2);
  @$pb.TagNumber(2)
  RequestFlush ensureFlush() => $_ensure(1);

  @$pb.TagNumber(3)
  RequestInfo get info => $_getN(2);
  @$pb.TagNumber(3)
  set info(RequestInfo v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasInfo() => $_has(2);
  @$pb.TagNumber(3)
  void clearInfo() => clearField(3);
  @$pb.TagNumber(3)
  RequestInfo ensureInfo() => $_ensure(2);

  @$pb.TagNumber(4)
  RequestSetOption get setOption => $_getN(3);
  @$pb.TagNumber(4)
  set setOption(RequestSetOption v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasSetOption() => $_has(3);
  @$pb.TagNumber(4)
  void clearSetOption() => clearField(4);
  @$pb.TagNumber(4)
  RequestSetOption ensureSetOption() => $_ensure(3);

  @$pb.TagNumber(5)
  RequestInitChain get initChain => $_getN(4);
  @$pb.TagNumber(5)
  set initChain(RequestInitChain v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasInitChain() => $_has(4);
  @$pb.TagNumber(5)
  void clearInitChain() => clearField(5);
  @$pb.TagNumber(5)
  RequestInitChain ensureInitChain() => $_ensure(4);

  @$pb.TagNumber(6)
  RequestQuery get query => $_getN(5);
  @$pb.TagNumber(6)
  set query(RequestQuery v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasQuery() => $_has(5);
  @$pb.TagNumber(6)
  void clearQuery() => clearField(6);
  @$pb.TagNumber(6)
  RequestQuery ensureQuery() => $_ensure(5);

  @$pb.TagNumber(7)
  RequestBeginBlock get beginBlock => $_getN(6);
  @$pb.TagNumber(7)
  set beginBlock(RequestBeginBlock v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasBeginBlock() => $_has(6);
  @$pb.TagNumber(7)
  void clearBeginBlock() => clearField(7);
  @$pb.TagNumber(7)
  RequestBeginBlock ensureBeginBlock() => $_ensure(6);

  @$pb.TagNumber(8)
  RequestCheckTx get checkTx => $_getN(7);
  @$pb.TagNumber(8)
  set checkTx(RequestCheckTx v) { setField(8, v); }
  @$pb.TagNumber(8)
  $core.bool hasCheckTx() => $_has(7);
  @$pb.TagNumber(8)
  void clearCheckTx() => clearField(8);
  @$pb.TagNumber(8)
  RequestCheckTx ensureCheckTx() => $_ensure(7);

  @$pb.TagNumber(9)
  RequestDeliverTx get deliverTx => $_getN(8);
  @$pb.TagNumber(9)
  set deliverTx(RequestDeliverTx v) { setField(9, v); }
  @$pb.TagNumber(9)
  $core.bool hasDeliverTx() => $_has(8);
  @$pb.TagNumber(9)
  void clearDeliverTx() => clearField(9);
  @$pb.TagNumber(9)
  RequestDeliverTx ensureDeliverTx() => $_ensure(8);

  @$pb.TagNumber(10)
  RequestEndBlock get endBlock => $_getN(9);
  @$pb.TagNumber(10)
  set endBlock(RequestEndBlock v) { setField(10, v); }
  @$pb.TagNumber(10)
  $core.bool hasEndBlock() => $_has(9);
  @$pb.TagNumber(10)
  void clearEndBlock() => clearField(10);
  @$pb.TagNumber(10)
  RequestEndBlock ensureEndBlock() => $_ensure(9);

  @$pb.TagNumber(11)
  RequestCommit get commit => $_getN(10);
  @$pb.TagNumber(11)
  set commit(RequestCommit v) { setField(11, v); }
  @$pb.TagNumber(11)
  $core.bool hasCommit() => $_has(10);
  @$pb.TagNumber(11)
  void clearCommit() => clearField(11);
  @$pb.TagNumber(11)
  RequestCommit ensureCommit() => $_ensure(10);

  @$pb.TagNumber(12)
  RequestListSnapshots get listSnapshots => $_getN(11);
  @$pb.TagNumber(12)
  set listSnapshots(RequestListSnapshots v) { setField(12, v); }
  @$pb.TagNumber(12)
  $core.bool hasListSnapshots() => $_has(11);
  @$pb.TagNumber(12)
  void clearListSnapshots() => clearField(12);
  @$pb.TagNumber(12)
  RequestListSnapshots ensureListSnapshots() => $_ensure(11);

  @$pb.TagNumber(13)
  RequestOfferSnapshot get offerSnapshot => $_getN(12);
  @$pb.TagNumber(13)
  set offerSnapshot(RequestOfferSnapshot v) { setField(13, v); }
  @$pb.TagNumber(13)
  $core.bool hasOfferSnapshot() => $_has(12);
  @$pb.TagNumber(13)
  void clearOfferSnapshot() => clearField(13);
  @$pb.TagNumber(13)
  RequestOfferSnapshot ensureOfferSnapshot() => $_ensure(12);

  @$pb.TagNumber(14)
  RequestLoadSnapshotChunk get loadSnapshotChunk => $_getN(13);
  @$pb.TagNumber(14)
  set loadSnapshotChunk(RequestLoadSnapshotChunk v) { setField(14, v); }
  @$pb.TagNumber(14)
  $core.bool hasLoadSnapshotChunk() => $_has(13);
  @$pb.TagNumber(14)
  void clearLoadSnapshotChunk() => clearField(14);
  @$pb.TagNumber(14)
  RequestLoadSnapshotChunk ensureLoadSnapshotChunk() => $_ensure(13);

  @$pb.TagNumber(15)
  RequestApplySnapshotChunk get applySnapshotChunk => $_getN(14);
  @$pb.TagNumber(15)
  set applySnapshotChunk(RequestApplySnapshotChunk v) { setField(15, v); }
  @$pb.TagNumber(15)
  $core.bool hasApplySnapshotChunk() => $_has(14);
  @$pb.TagNumber(15)
  void clearApplySnapshotChunk() => clearField(15);
  @$pb.TagNumber(15)
  RequestApplySnapshotChunk ensureApplySnapshotChunk() => $_ensure(14);
}

class RequestEcho extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestEcho', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'message')
    ..hasRequiredFields = false
  ;

  RequestEcho._() : super();
  factory RequestEcho() => create();
  factory RequestEcho.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestEcho.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestEcho clone() => RequestEcho()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestEcho copyWith(void Function(RequestEcho) updates) => super.copyWith((message) => updates(message as RequestEcho)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestEcho create() => RequestEcho._();
  RequestEcho createEmptyInstance() => create();
  static $pb.PbList<RequestEcho> createRepeated() => $pb.PbList<RequestEcho>();
  @$core.pragma('dart2js:noInline')
  static RequestEcho getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestEcho>(create);
  static RequestEcho _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => clearField(1);
}

class RequestFlush extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestFlush', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  RequestFlush._() : super();
  factory RequestFlush() => create();
  factory RequestFlush.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestFlush.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestFlush clone() => RequestFlush()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestFlush copyWith(void Function(RequestFlush) updates) => super.copyWith((message) => updates(message as RequestFlush)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestFlush create() => RequestFlush._();
  RequestFlush createEmptyInstance() => create();
  static $pb.PbList<RequestFlush> createRepeated() => $pb.PbList<RequestFlush>();
  @$core.pragma('dart2js:noInline')
  static RequestFlush getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestFlush>(create);
  static RequestFlush _defaultInstance;
}

class RequestInfo extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestInfo', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'version')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'blockVersion', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'p2pVersion', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  RequestInfo._() : super();
  factory RequestInfo() => create();
  factory RequestInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestInfo clone() => RequestInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestInfo copyWith(void Function(RequestInfo) updates) => super.copyWith((message) => updates(message as RequestInfo)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestInfo create() => RequestInfo._();
  RequestInfo createEmptyInstance() => create();
  static $pb.PbList<RequestInfo> createRepeated() => $pb.PbList<RequestInfo>();
  @$core.pragma('dart2js:noInline')
  static RequestInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestInfo>(create);
  static RequestInfo _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get version => $_getSZ(0);
  @$pb.TagNumber(1)
  set version($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasVersion() => $_has(0);
  @$pb.TagNumber(1)
  void clearVersion() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get blockVersion => $_getI64(1);
  @$pb.TagNumber(2)
  set blockVersion($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasBlockVersion() => $_has(1);
  @$pb.TagNumber(2)
  void clearBlockVersion() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get p2pVersion => $_getI64(2);
  @$pb.TagNumber(3)
  set p2pVersion($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasP2pVersion() => $_has(2);
  @$pb.TagNumber(3)
  void clearP2pVersion() => clearField(3);
}

class RequestSetOption extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestSetOption', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'key')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'value')
    ..hasRequiredFields = false
  ;

  RequestSetOption._() : super();
  factory RequestSetOption() => create();
  factory RequestSetOption.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestSetOption.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestSetOption clone() => RequestSetOption()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestSetOption copyWith(void Function(RequestSetOption) updates) => super.copyWith((message) => updates(message as RequestSetOption)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestSetOption create() => RequestSetOption._();
  RequestSetOption createEmptyInstance() => create();
  static $pb.PbList<RequestSetOption> createRepeated() => $pb.PbList<RequestSetOption>();
  @$core.pragma('dart2js:noInline')
  static RequestSetOption getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestSetOption>(create);
  static RequestSetOption _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get key => $_getSZ(0);
  @$pb.TagNumber(1)
  set key($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasKey() => $_has(0);
  @$pb.TagNumber(1)
  void clearKey() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get value => $_getSZ(1);
  @$pb.TagNumber(2)
  set value($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasValue() => $_has(1);
  @$pb.TagNumber(2)
  void clearValue() => clearField(2);
}

class RequestInitChain extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestInitChain', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOM<$4.Timestamp>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'time', subBuilder: $4.Timestamp.create)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'chainId')
    ..aOM<ConsensusParams>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'consensusParams', subBuilder: ConsensusParams.create)
    ..pc<ValidatorUpdate>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validators', $pb.PbFieldType.PM, subBuilder: ValidatorUpdate.create)
    ..a<$core.List<$core.int>>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'appStateBytes', $pb.PbFieldType.OY)
    ..aInt64(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'initialHeight')
    ..hasRequiredFields = false
  ;

  RequestInitChain._() : super();
  factory RequestInitChain() => create();
  factory RequestInitChain.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestInitChain.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestInitChain clone() => RequestInitChain()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestInitChain copyWith(void Function(RequestInitChain) updates) => super.copyWith((message) => updates(message as RequestInitChain)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestInitChain create() => RequestInitChain._();
  RequestInitChain createEmptyInstance() => create();
  static $pb.PbList<RequestInitChain> createRepeated() => $pb.PbList<RequestInitChain>();
  @$core.pragma('dart2js:noInline')
  static RequestInitChain getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestInitChain>(create);
  static RequestInitChain _defaultInstance;

  @$pb.TagNumber(1)
  $4.Timestamp get time => $_getN(0);
  @$pb.TagNumber(1)
  set time($4.Timestamp v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasTime() => $_has(0);
  @$pb.TagNumber(1)
  void clearTime() => clearField(1);
  @$pb.TagNumber(1)
  $4.Timestamp ensureTime() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.String get chainId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chainId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChainId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChainId() => clearField(2);

  @$pb.TagNumber(3)
  ConsensusParams get consensusParams => $_getN(2);
  @$pb.TagNumber(3)
  set consensusParams(ConsensusParams v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasConsensusParams() => $_has(2);
  @$pb.TagNumber(3)
  void clearConsensusParams() => clearField(3);
  @$pb.TagNumber(3)
  ConsensusParams ensureConsensusParams() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.List<ValidatorUpdate> get validators => $_getList(3);

  @$pb.TagNumber(5)
  $core.List<$core.int> get appStateBytes => $_getN(4);
  @$pb.TagNumber(5)
  set appStateBytes($core.List<$core.int> v) { $_setBytes(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasAppStateBytes() => $_has(4);
  @$pb.TagNumber(5)
  void clearAppStateBytes() => clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get initialHeight => $_getI64(5);
  @$pb.TagNumber(6)
  set initialHeight($fixnum.Int64 v) { $_setInt64(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasInitialHeight() => $_has(5);
  @$pb.TagNumber(6)
  void clearInitialHeight() => clearField(6);
}

class RequestQuery extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestQuery', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'data', $pb.PbFieldType.OY)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'path')
    ..aInt64(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height')
    ..aOB(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'prove')
    ..hasRequiredFields = false
  ;

  RequestQuery._() : super();
  factory RequestQuery() => create();
  factory RequestQuery.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestQuery.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestQuery clone() => RequestQuery()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestQuery copyWith(void Function(RequestQuery) updates) => super.copyWith((message) => updates(message as RequestQuery)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestQuery create() => RequestQuery._();
  RequestQuery createEmptyInstance() => create();
  static $pb.PbList<RequestQuery> createRepeated() => $pb.PbList<RequestQuery>();
  @$core.pragma('dart2js:noInline')
  static RequestQuery getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestQuery>(create);
  static RequestQuery _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get data => $_getN(0);
  @$pb.TagNumber(1)
  set data($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get height => $_getI64(2);
  @$pb.TagNumber(3)
  set height($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearHeight() => clearField(3);

  @$pb.TagNumber(4)
  $core.bool get prove => $_getBF(3);
  @$pb.TagNumber(4)
  set prove($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasProve() => $_has(3);
  @$pb.TagNumber(4)
  void clearProve() => clearField(4);
}

class RequestBeginBlock extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestBeginBlock', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'hash', $pb.PbFieldType.OY)
    ..aOM<$7.Header>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'header', subBuilder: $7.Header.create)
    ..aOM<LastCommitInfo>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lastCommitInfo', subBuilder: LastCommitInfo.create)
    ..pc<Evidence>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'byzantineValidators', $pb.PbFieldType.PM, subBuilder: Evidence.create)
    ..hasRequiredFields = false
  ;

  RequestBeginBlock._() : super();
  factory RequestBeginBlock() => create();
  factory RequestBeginBlock.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestBeginBlock.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestBeginBlock clone() => RequestBeginBlock()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestBeginBlock copyWith(void Function(RequestBeginBlock) updates) => super.copyWith((message) => updates(message as RequestBeginBlock)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestBeginBlock create() => RequestBeginBlock._();
  RequestBeginBlock createEmptyInstance() => create();
  static $pb.PbList<RequestBeginBlock> createRepeated() => $pb.PbList<RequestBeginBlock>();
  @$core.pragma('dart2js:noInline')
  static RequestBeginBlock getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestBeginBlock>(create);
  static RequestBeginBlock _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get hash => $_getN(0);
  @$pb.TagNumber(1)
  set hash($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasHash() => $_has(0);
  @$pb.TagNumber(1)
  void clearHash() => clearField(1);

  @$pb.TagNumber(2)
  $7.Header get header => $_getN(1);
  @$pb.TagNumber(2)
  set header($7.Header v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasHeader() => $_has(1);
  @$pb.TagNumber(2)
  void clearHeader() => clearField(2);
  @$pb.TagNumber(2)
  $7.Header ensureHeader() => $_ensure(1);

  @$pb.TagNumber(3)
  LastCommitInfo get lastCommitInfo => $_getN(2);
  @$pb.TagNumber(3)
  set lastCommitInfo(LastCommitInfo v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasLastCommitInfo() => $_has(2);
  @$pb.TagNumber(3)
  void clearLastCommitInfo() => clearField(3);
  @$pb.TagNumber(3)
  LastCommitInfo ensureLastCommitInfo() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.List<Evidence> get byzantineValidators => $_getList(3);
}

class RequestCheckTx extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestCheckTx', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tx', $pb.PbFieldType.OY)
    ..e<CheckTxType>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'type', $pb.PbFieldType.OE, defaultOrMaker: CheckTxType.NEW, valueOf: CheckTxType.valueOf, enumValues: CheckTxType.values)
    ..hasRequiredFields = false
  ;

  RequestCheckTx._() : super();
  factory RequestCheckTx() => create();
  factory RequestCheckTx.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestCheckTx.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestCheckTx clone() => RequestCheckTx()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestCheckTx copyWith(void Function(RequestCheckTx) updates) => super.copyWith((message) => updates(message as RequestCheckTx)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestCheckTx create() => RequestCheckTx._();
  RequestCheckTx createEmptyInstance() => create();
  static $pb.PbList<RequestCheckTx> createRepeated() => $pb.PbList<RequestCheckTx>();
  @$core.pragma('dart2js:noInline')
  static RequestCheckTx getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestCheckTx>(create);
  static RequestCheckTx _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get tx => $_getN(0);
  @$pb.TagNumber(1)
  set tx($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTx() => $_has(0);
  @$pb.TagNumber(1)
  void clearTx() => clearField(1);

  @$pb.TagNumber(2)
  CheckTxType get type => $_getN(1);
  @$pb.TagNumber(2)
  set type(CheckTxType v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => clearField(2);
}

class RequestDeliverTx extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestDeliverTx', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tx', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  RequestDeliverTx._() : super();
  factory RequestDeliverTx() => create();
  factory RequestDeliverTx.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestDeliverTx.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestDeliverTx clone() => RequestDeliverTx()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestDeliverTx copyWith(void Function(RequestDeliverTx) updates) => super.copyWith((message) => updates(message as RequestDeliverTx)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestDeliverTx create() => RequestDeliverTx._();
  RequestDeliverTx createEmptyInstance() => create();
  static $pb.PbList<RequestDeliverTx> createRepeated() => $pb.PbList<RequestDeliverTx>();
  @$core.pragma('dart2js:noInline')
  static RequestDeliverTx getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestDeliverTx>(create);
  static RequestDeliverTx _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get tx => $_getN(0);
  @$pb.TagNumber(1)
  set tx($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTx() => $_has(0);
  @$pb.TagNumber(1)
  void clearTx() => clearField(1);
}

class RequestEndBlock extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestEndBlock', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aInt64(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height')
    ..hasRequiredFields = false
  ;

  RequestEndBlock._() : super();
  factory RequestEndBlock() => create();
  factory RequestEndBlock.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestEndBlock.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestEndBlock clone() => RequestEndBlock()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestEndBlock copyWith(void Function(RequestEndBlock) updates) => super.copyWith((message) => updates(message as RequestEndBlock)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestEndBlock create() => RequestEndBlock._();
  RequestEndBlock createEmptyInstance() => create();
  static $pb.PbList<RequestEndBlock> createRepeated() => $pb.PbList<RequestEndBlock>();
  @$core.pragma('dart2js:noInline')
  static RequestEndBlock getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestEndBlock>(create);
  static RequestEndBlock _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get height => $_getI64(0);
  @$pb.TagNumber(1)
  set height($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasHeight() => $_has(0);
  @$pb.TagNumber(1)
  void clearHeight() => clearField(1);
}

class RequestCommit extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestCommit', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  RequestCommit._() : super();
  factory RequestCommit() => create();
  factory RequestCommit.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestCommit.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestCommit clone() => RequestCommit()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestCommit copyWith(void Function(RequestCommit) updates) => super.copyWith((message) => updates(message as RequestCommit)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestCommit create() => RequestCommit._();
  RequestCommit createEmptyInstance() => create();
  static $pb.PbList<RequestCommit> createRepeated() => $pb.PbList<RequestCommit>();
  @$core.pragma('dart2js:noInline')
  static RequestCommit getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestCommit>(create);
  static RequestCommit _defaultInstance;
}

class RequestListSnapshots extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestListSnapshots', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  RequestListSnapshots._() : super();
  factory RequestListSnapshots() => create();
  factory RequestListSnapshots.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestListSnapshots.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestListSnapshots clone() => RequestListSnapshots()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestListSnapshots copyWith(void Function(RequestListSnapshots) updates) => super.copyWith((message) => updates(message as RequestListSnapshots)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestListSnapshots create() => RequestListSnapshots._();
  RequestListSnapshots createEmptyInstance() => create();
  static $pb.PbList<RequestListSnapshots> createRepeated() => $pb.PbList<RequestListSnapshots>();
  @$core.pragma('dart2js:noInline')
  static RequestListSnapshots getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestListSnapshots>(create);
  static RequestListSnapshots _defaultInstance;
}

class RequestOfferSnapshot extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestOfferSnapshot', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOM<Snapshot>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'snapshot', subBuilder: Snapshot.create)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'appHash', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  RequestOfferSnapshot._() : super();
  factory RequestOfferSnapshot() => create();
  factory RequestOfferSnapshot.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestOfferSnapshot.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestOfferSnapshot clone() => RequestOfferSnapshot()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestOfferSnapshot copyWith(void Function(RequestOfferSnapshot) updates) => super.copyWith((message) => updates(message as RequestOfferSnapshot)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestOfferSnapshot create() => RequestOfferSnapshot._();
  RequestOfferSnapshot createEmptyInstance() => create();
  static $pb.PbList<RequestOfferSnapshot> createRepeated() => $pb.PbList<RequestOfferSnapshot>();
  @$core.pragma('dart2js:noInline')
  static RequestOfferSnapshot getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestOfferSnapshot>(create);
  static RequestOfferSnapshot _defaultInstance;

  @$pb.TagNumber(1)
  Snapshot get snapshot => $_getN(0);
  @$pb.TagNumber(1)
  set snapshot(Snapshot v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasSnapshot() => $_has(0);
  @$pb.TagNumber(1)
  void clearSnapshot() => clearField(1);
  @$pb.TagNumber(1)
  Snapshot ensureSnapshot() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<$core.int> get appHash => $_getN(1);
  @$pb.TagNumber(2)
  set appHash($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasAppHash() => $_has(1);
  @$pb.TagNumber(2)
  void clearAppHash() => clearField(2);
}

class RequestLoadSnapshotChunk extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestLoadSnapshotChunk', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$core.int>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'format', $pb.PbFieldType.OU3)
    ..a<$core.int>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'chunk', $pb.PbFieldType.OU3)
    ..hasRequiredFields = false
  ;

  RequestLoadSnapshotChunk._() : super();
  factory RequestLoadSnapshotChunk() => create();
  factory RequestLoadSnapshotChunk.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestLoadSnapshotChunk.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestLoadSnapshotChunk clone() => RequestLoadSnapshotChunk()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestLoadSnapshotChunk copyWith(void Function(RequestLoadSnapshotChunk) updates) => super.copyWith((message) => updates(message as RequestLoadSnapshotChunk)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestLoadSnapshotChunk create() => RequestLoadSnapshotChunk._();
  RequestLoadSnapshotChunk createEmptyInstance() => create();
  static $pb.PbList<RequestLoadSnapshotChunk> createRepeated() => $pb.PbList<RequestLoadSnapshotChunk>();
  @$core.pragma('dart2js:noInline')
  static RequestLoadSnapshotChunk getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestLoadSnapshotChunk>(create);
  static RequestLoadSnapshotChunk _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get height => $_getI64(0);
  @$pb.TagNumber(1)
  set height($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasHeight() => $_has(0);
  @$pb.TagNumber(1)
  void clearHeight() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get format => $_getIZ(1);
  @$pb.TagNumber(2)
  set format($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasFormat() => $_has(1);
  @$pb.TagNumber(2)
  void clearFormat() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get chunk => $_getIZ(2);
  @$pb.TagNumber(3)
  set chunk($core.int v) { $_setUnsignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasChunk() => $_has(2);
  @$pb.TagNumber(3)
  void clearChunk() => clearField(3);
}

class RequestApplySnapshotChunk extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RequestApplySnapshotChunk', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.int>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'index', $pb.PbFieldType.OU3)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'chunk', $pb.PbFieldType.OY)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..hasRequiredFields = false
  ;

  RequestApplySnapshotChunk._() : super();
  factory RequestApplySnapshotChunk() => create();
  factory RequestApplySnapshotChunk.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestApplySnapshotChunk.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestApplySnapshotChunk clone() => RequestApplySnapshotChunk()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestApplySnapshotChunk copyWith(void Function(RequestApplySnapshotChunk) updates) => super.copyWith((message) => updates(message as RequestApplySnapshotChunk)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RequestApplySnapshotChunk create() => RequestApplySnapshotChunk._();
  RequestApplySnapshotChunk createEmptyInstance() => create();
  static $pb.PbList<RequestApplySnapshotChunk> createRepeated() => $pb.PbList<RequestApplySnapshotChunk>();
  @$core.pragma('dart2js:noInline')
  static RequestApplySnapshotChunk getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestApplySnapshotChunk>(create);
  static RequestApplySnapshotChunk _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get index => $_getIZ(0);
  @$pb.TagNumber(1)
  set index($core.int v) { $_setUnsignedInt32(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasIndex() => $_has(0);
  @$pb.TagNumber(1)
  void clearIndex() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get chunk => $_getN(1);
  @$pb.TagNumber(2)
  set chunk($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChunk() => $_has(1);
  @$pb.TagNumber(2)
  void clearChunk() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get sender => $_getSZ(2);
  @$pb.TagNumber(3)
  set sender($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSender() => $_has(2);
  @$pb.TagNumber(3)
  void clearSender() => clearField(3);
}

enum Response_Value {
  exception, 
  echo, 
  flush, 
  info, 
  setOption, 
  initChain, 
  query, 
  beginBlock, 
  checkTx, 
  deliverTx, 
  endBlock, 
  commit, 
  listSnapshots, 
  offerSnapshot, 
  loadSnapshotChunk, 
  applySnapshotChunk, 
  notSet
}

class Response extends $pb.GeneratedMessage {
  static const $core.Map<$core.int, Response_Value> _Response_ValueByTag = {
    1 : Response_Value.exception,
    2 : Response_Value.echo,
    3 : Response_Value.flush,
    4 : Response_Value.info,
    5 : Response_Value.setOption,
    6 : Response_Value.initChain,
    7 : Response_Value.query,
    8 : Response_Value.beginBlock,
    9 : Response_Value.checkTx,
    10 : Response_Value.deliverTx,
    11 : Response_Value.endBlock,
    12 : Response_Value.commit,
    13 : Response_Value.listSnapshots,
    14 : Response_Value.offerSnapshot,
    15 : Response_Value.loadSnapshotChunk,
    16 : Response_Value.applySnapshotChunk,
    0 : Response_Value.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Response', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..oo(0, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16])
    ..aOM<ResponseException>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'exception', subBuilder: ResponseException.create)
    ..aOM<ResponseEcho>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'echo', subBuilder: ResponseEcho.create)
    ..aOM<ResponseFlush>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'flush', subBuilder: ResponseFlush.create)
    ..aOM<ResponseInfo>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'info', subBuilder: ResponseInfo.create)
    ..aOM<ResponseSetOption>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'setOption', subBuilder: ResponseSetOption.create)
    ..aOM<ResponseInitChain>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'initChain', subBuilder: ResponseInitChain.create)
    ..aOM<ResponseQuery>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'query', subBuilder: ResponseQuery.create)
    ..aOM<ResponseBeginBlock>(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'beginBlock', subBuilder: ResponseBeginBlock.create)
    ..aOM<ResponseCheckTx>(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'checkTx', subBuilder: ResponseCheckTx.create)
    ..aOM<ResponseDeliverTx>(10, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'deliverTx', subBuilder: ResponseDeliverTx.create)
    ..aOM<ResponseEndBlock>(11, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'endBlock', subBuilder: ResponseEndBlock.create)
    ..aOM<ResponseCommit>(12, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'commit', subBuilder: ResponseCommit.create)
    ..aOM<ResponseListSnapshots>(13, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'listSnapshots', subBuilder: ResponseListSnapshots.create)
    ..aOM<ResponseOfferSnapshot>(14, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'offerSnapshot', subBuilder: ResponseOfferSnapshot.create)
    ..aOM<ResponseLoadSnapshotChunk>(15, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'loadSnapshotChunk', subBuilder: ResponseLoadSnapshotChunk.create)
    ..aOM<ResponseApplySnapshotChunk>(16, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'applySnapshotChunk', subBuilder: ResponseApplySnapshotChunk.create)
    ..hasRequiredFields = false
  ;

  Response._() : super();
  factory Response() => create();
  factory Response.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Response.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Response clone() => Response()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Response copyWith(void Function(Response) updates) => super.copyWith((message) => updates(message as Response)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Response create() => Response._();
  Response createEmptyInstance() => create();
  static $pb.PbList<Response> createRepeated() => $pb.PbList<Response>();
  @$core.pragma('dart2js:noInline')
  static Response getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Response>(create);
  static Response _defaultInstance;

  Response_Value whichValue() => _Response_ValueByTag[$_whichOneof(0)];
  void clearValue() => clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  ResponseException get exception => $_getN(0);
  @$pb.TagNumber(1)
  set exception(ResponseException v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasException() => $_has(0);
  @$pb.TagNumber(1)
  void clearException() => clearField(1);
  @$pb.TagNumber(1)
  ResponseException ensureException() => $_ensure(0);

  @$pb.TagNumber(2)
  ResponseEcho get echo => $_getN(1);
  @$pb.TagNumber(2)
  set echo(ResponseEcho v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasEcho() => $_has(1);
  @$pb.TagNumber(2)
  void clearEcho() => clearField(2);
  @$pb.TagNumber(2)
  ResponseEcho ensureEcho() => $_ensure(1);

  @$pb.TagNumber(3)
  ResponseFlush get flush => $_getN(2);
  @$pb.TagNumber(3)
  set flush(ResponseFlush v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasFlush() => $_has(2);
  @$pb.TagNumber(3)
  void clearFlush() => clearField(3);
  @$pb.TagNumber(3)
  ResponseFlush ensureFlush() => $_ensure(2);

  @$pb.TagNumber(4)
  ResponseInfo get info => $_getN(3);
  @$pb.TagNumber(4)
  set info(ResponseInfo v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasInfo() => $_has(3);
  @$pb.TagNumber(4)
  void clearInfo() => clearField(4);
  @$pb.TagNumber(4)
  ResponseInfo ensureInfo() => $_ensure(3);

  @$pb.TagNumber(5)
  ResponseSetOption get setOption => $_getN(4);
  @$pb.TagNumber(5)
  set setOption(ResponseSetOption v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasSetOption() => $_has(4);
  @$pb.TagNumber(5)
  void clearSetOption() => clearField(5);
  @$pb.TagNumber(5)
  ResponseSetOption ensureSetOption() => $_ensure(4);

  @$pb.TagNumber(6)
  ResponseInitChain get initChain => $_getN(5);
  @$pb.TagNumber(6)
  set initChain(ResponseInitChain v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasInitChain() => $_has(5);
  @$pb.TagNumber(6)
  void clearInitChain() => clearField(6);
  @$pb.TagNumber(6)
  ResponseInitChain ensureInitChain() => $_ensure(5);

  @$pb.TagNumber(7)
  ResponseQuery get query => $_getN(6);
  @$pb.TagNumber(7)
  set query(ResponseQuery v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasQuery() => $_has(6);
  @$pb.TagNumber(7)
  void clearQuery() => clearField(7);
  @$pb.TagNumber(7)
  ResponseQuery ensureQuery() => $_ensure(6);

  @$pb.TagNumber(8)
  ResponseBeginBlock get beginBlock => $_getN(7);
  @$pb.TagNumber(8)
  set beginBlock(ResponseBeginBlock v) { setField(8, v); }
  @$pb.TagNumber(8)
  $core.bool hasBeginBlock() => $_has(7);
  @$pb.TagNumber(8)
  void clearBeginBlock() => clearField(8);
  @$pb.TagNumber(8)
  ResponseBeginBlock ensureBeginBlock() => $_ensure(7);

  @$pb.TagNumber(9)
  ResponseCheckTx get checkTx => $_getN(8);
  @$pb.TagNumber(9)
  set checkTx(ResponseCheckTx v) { setField(9, v); }
  @$pb.TagNumber(9)
  $core.bool hasCheckTx() => $_has(8);
  @$pb.TagNumber(9)
  void clearCheckTx() => clearField(9);
  @$pb.TagNumber(9)
  ResponseCheckTx ensureCheckTx() => $_ensure(8);

  @$pb.TagNumber(10)
  ResponseDeliverTx get deliverTx => $_getN(9);
  @$pb.TagNumber(10)
  set deliverTx(ResponseDeliverTx v) { setField(10, v); }
  @$pb.TagNumber(10)
  $core.bool hasDeliverTx() => $_has(9);
  @$pb.TagNumber(10)
  void clearDeliverTx() => clearField(10);
  @$pb.TagNumber(10)
  ResponseDeliverTx ensureDeliverTx() => $_ensure(9);

  @$pb.TagNumber(11)
  ResponseEndBlock get endBlock => $_getN(10);
  @$pb.TagNumber(11)
  set endBlock(ResponseEndBlock v) { setField(11, v); }
  @$pb.TagNumber(11)
  $core.bool hasEndBlock() => $_has(10);
  @$pb.TagNumber(11)
  void clearEndBlock() => clearField(11);
  @$pb.TagNumber(11)
  ResponseEndBlock ensureEndBlock() => $_ensure(10);

  @$pb.TagNumber(12)
  ResponseCommit get commit => $_getN(11);
  @$pb.TagNumber(12)
  set commit(ResponseCommit v) { setField(12, v); }
  @$pb.TagNumber(12)
  $core.bool hasCommit() => $_has(11);
  @$pb.TagNumber(12)
  void clearCommit() => clearField(12);
  @$pb.TagNumber(12)
  ResponseCommit ensureCommit() => $_ensure(11);

  @$pb.TagNumber(13)
  ResponseListSnapshots get listSnapshots => $_getN(12);
  @$pb.TagNumber(13)
  set listSnapshots(ResponseListSnapshots v) { setField(13, v); }
  @$pb.TagNumber(13)
  $core.bool hasListSnapshots() => $_has(12);
  @$pb.TagNumber(13)
  void clearListSnapshots() => clearField(13);
  @$pb.TagNumber(13)
  ResponseListSnapshots ensureListSnapshots() => $_ensure(12);

  @$pb.TagNumber(14)
  ResponseOfferSnapshot get offerSnapshot => $_getN(13);
  @$pb.TagNumber(14)
  set offerSnapshot(ResponseOfferSnapshot v) { setField(14, v); }
  @$pb.TagNumber(14)
  $core.bool hasOfferSnapshot() => $_has(13);
  @$pb.TagNumber(14)
  void clearOfferSnapshot() => clearField(14);
  @$pb.TagNumber(14)
  ResponseOfferSnapshot ensureOfferSnapshot() => $_ensure(13);

  @$pb.TagNumber(15)
  ResponseLoadSnapshotChunk get loadSnapshotChunk => $_getN(14);
  @$pb.TagNumber(15)
  set loadSnapshotChunk(ResponseLoadSnapshotChunk v) { setField(15, v); }
  @$pb.TagNumber(15)
  $core.bool hasLoadSnapshotChunk() => $_has(14);
  @$pb.TagNumber(15)
  void clearLoadSnapshotChunk() => clearField(15);
  @$pb.TagNumber(15)
  ResponseLoadSnapshotChunk ensureLoadSnapshotChunk() => $_ensure(14);

  @$pb.TagNumber(16)
  ResponseApplySnapshotChunk get applySnapshotChunk => $_getN(15);
  @$pb.TagNumber(16)
  set applySnapshotChunk(ResponseApplySnapshotChunk v) { setField(16, v); }
  @$pb.TagNumber(16)
  $core.bool hasApplySnapshotChunk() => $_has(15);
  @$pb.TagNumber(16)
  void clearApplySnapshotChunk() => clearField(16);
  @$pb.TagNumber(16)
  ResponseApplySnapshotChunk ensureApplySnapshotChunk() => $_ensure(15);
}

class ResponseException extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseException', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'error')
    ..hasRequiredFields = false
  ;

  ResponseException._() : super();
  factory ResponseException() => create();
  factory ResponseException.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseException.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseException clone() => ResponseException()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseException copyWith(void Function(ResponseException) updates) => super.copyWith((message) => updates(message as ResponseException)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseException create() => ResponseException._();
  ResponseException createEmptyInstance() => create();
  static $pb.PbList<ResponseException> createRepeated() => $pb.PbList<ResponseException>();
  @$core.pragma('dart2js:noInline')
  static ResponseException getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseException>(create);
  static ResponseException _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get error => $_getSZ(0);
  @$pb.TagNumber(1)
  set error($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasError() => $_has(0);
  @$pb.TagNumber(1)
  void clearError() => clearField(1);
}

class ResponseEcho extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseEcho', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'message')
    ..hasRequiredFields = false
  ;

  ResponseEcho._() : super();
  factory ResponseEcho() => create();
  factory ResponseEcho.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseEcho.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseEcho clone() => ResponseEcho()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseEcho copyWith(void Function(ResponseEcho) updates) => super.copyWith((message) => updates(message as ResponseEcho)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseEcho create() => ResponseEcho._();
  ResponseEcho createEmptyInstance() => create();
  static $pb.PbList<ResponseEcho> createRepeated() => $pb.PbList<ResponseEcho>();
  @$core.pragma('dart2js:noInline')
  static ResponseEcho getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseEcho>(create);
  static ResponseEcho _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => clearField(1);
}

class ResponseFlush extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseFlush', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  ResponseFlush._() : super();
  factory ResponseFlush() => create();
  factory ResponseFlush.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseFlush.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseFlush clone() => ResponseFlush()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseFlush copyWith(void Function(ResponseFlush) updates) => super.copyWith((message) => updates(message as ResponseFlush)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseFlush create() => ResponseFlush._();
  ResponseFlush createEmptyInstance() => create();
  static $pb.PbList<ResponseFlush> createRepeated() => $pb.PbList<ResponseFlush>();
  @$core.pragma('dart2js:noInline')
  static ResponseFlush getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseFlush>(create);
  static ResponseFlush _defaultInstance;
}

class ResponseInfo extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseInfo', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'data')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'version')
    ..a<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'appVersion', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aInt64(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lastBlockHeight')
    ..a<$core.List<$core.int>>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lastBlockAppHash', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  ResponseInfo._() : super();
  factory ResponseInfo() => create();
  factory ResponseInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseInfo clone() => ResponseInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseInfo copyWith(void Function(ResponseInfo) updates) => super.copyWith((message) => updates(message as ResponseInfo)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseInfo create() => ResponseInfo._();
  ResponseInfo createEmptyInstance() => create();
  static $pb.PbList<ResponseInfo> createRepeated() => $pb.PbList<ResponseInfo>();
  @$core.pragma('dart2js:noInline')
  static ResponseInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseInfo>(create);
  static ResponseInfo _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get data => $_getSZ(0);
  @$pb.TagNumber(1)
  set data($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get version => $_getSZ(1);
  @$pb.TagNumber(2)
  set version($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasVersion() => $_has(1);
  @$pb.TagNumber(2)
  void clearVersion() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get appVersion => $_getI64(2);
  @$pb.TagNumber(3)
  set appVersion($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasAppVersion() => $_has(2);
  @$pb.TagNumber(3)
  void clearAppVersion() => clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get lastBlockHeight => $_getI64(3);
  @$pb.TagNumber(4)
  set lastBlockHeight($fixnum.Int64 v) { $_setInt64(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasLastBlockHeight() => $_has(3);
  @$pb.TagNumber(4)
  void clearLastBlockHeight() => clearField(4);

  @$pb.TagNumber(5)
  $core.List<$core.int> get lastBlockAppHash => $_getN(4);
  @$pb.TagNumber(5)
  set lastBlockAppHash($core.List<$core.int> v) { $_setBytes(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasLastBlockAppHash() => $_has(4);
  @$pb.TagNumber(5)
  void clearLastBlockAppHash() => clearField(5);
}

class ResponseSetOption extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseSetOption', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.int>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'code', $pb.PbFieldType.OU3)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'log')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'info')
    ..hasRequiredFields = false
  ;

  ResponseSetOption._() : super();
  factory ResponseSetOption() => create();
  factory ResponseSetOption.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseSetOption.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseSetOption clone() => ResponseSetOption()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseSetOption copyWith(void Function(ResponseSetOption) updates) => super.copyWith((message) => updates(message as ResponseSetOption)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseSetOption create() => ResponseSetOption._();
  ResponseSetOption createEmptyInstance() => create();
  static $pb.PbList<ResponseSetOption> createRepeated() => $pb.PbList<ResponseSetOption>();
  @$core.pragma('dart2js:noInline')
  static ResponseSetOption getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseSetOption>(create);
  static ResponseSetOption _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get code => $_getIZ(0);
  @$pb.TagNumber(1)
  set code($core.int v) { $_setUnsignedInt32(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCode() => clearField(1);

  @$pb.TagNumber(3)
  $core.String get log => $_getSZ(1);
  @$pb.TagNumber(3)
  set log($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(3)
  $core.bool hasLog() => $_has(1);
  @$pb.TagNumber(3)
  void clearLog() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get info => $_getSZ(2);
  @$pb.TagNumber(4)
  set info($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(4)
  $core.bool hasInfo() => $_has(2);
  @$pb.TagNumber(4)
  void clearInfo() => clearField(4);
}

class ResponseInitChain extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseInitChain', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOM<ConsensusParams>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'consensusParams', subBuilder: ConsensusParams.create)
    ..pc<ValidatorUpdate>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validators', $pb.PbFieldType.PM, subBuilder: ValidatorUpdate.create)
    ..a<$core.List<$core.int>>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'appHash', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  ResponseInitChain._() : super();
  factory ResponseInitChain() => create();
  factory ResponseInitChain.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseInitChain.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseInitChain clone() => ResponseInitChain()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseInitChain copyWith(void Function(ResponseInitChain) updates) => super.copyWith((message) => updates(message as ResponseInitChain)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseInitChain create() => ResponseInitChain._();
  ResponseInitChain createEmptyInstance() => create();
  static $pb.PbList<ResponseInitChain> createRepeated() => $pb.PbList<ResponseInitChain>();
  @$core.pragma('dart2js:noInline')
  static ResponseInitChain getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseInitChain>(create);
  static ResponseInitChain _defaultInstance;

  @$pb.TagNumber(1)
  ConsensusParams get consensusParams => $_getN(0);
  @$pb.TagNumber(1)
  set consensusParams(ConsensusParams v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasConsensusParams() => $_has(0);
  @$pb.TagNumber(1)
  void clearConsensusParams() => clearField(1);
  @$pb.TagNumber(1)
  ConsensusParams ensureConsensusParams() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<ValidatorUpdate> get validators => $_getList(1);

  @$pb.TagNumber(3)
  $core.List<$core.int> get appHash => $_getN(2);
  @$pb.TagNumber(3)
  set appHash($core.List<$core.int> v) { $_setBytes(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasAppHash() => $_has(2);
  @$pb.TagNumber(3)
  void clearAppHash() => clearField(3);
}

class ResponseQuery extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseQuery', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.int>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'code', $pb.PbFieldType.OU3)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'log')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'info')
    ..aInt64(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'index')
    ..a<$core.List<$core.int>>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'key', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'value', $pb.PbFieldType.OY)
    ..aOM<$2.ProofOps>(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofOps', subBuilder: $2.ProofOps.create)
    ..aInt64(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height')
    ..aOS(10, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'codespace')
    ..hasRequiredFields = false
  ;

  ResponseQuery._() : super();
  factory ResponseQuery() => create();
  factory ResponseQuery.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseQuery.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseQuery clone() => ResponseQuery()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseQuery copyWith(void Function(ResponseQuery) updates) => super.copyWith((message) => updates(message as ResponseQuery)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseQuery create() => ResponseQuery._();
  ResponseQuery createEmptyInstance() => create();
  static $pb.PbList<ResponseQuery> createRepeated() => $pb.PbList<ResponseQuery>();
  @$core.pragma('dart2js:noInline')
  static ResponseQuery getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseQuery>(create);
  static ResponseQuery _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get code => $_getIZ(0);
  @$pb.TagNumber(1)
  set code($core.int v) { $_setUnsignedInt32(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCode() => clearField(1);

  @$pb.TagNumber(3)
  $core.String get log => $_getSZ(1);
  @$pb.TagNumber(3)
  set log($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(3)
  $core.bool hasLog() => $_has(1);
  @$pb.TagNumber(3)
  void clearLog() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get info => $_getSZ(2);
  @$pb.TagNumber(4)
  set info($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(4)
  $core.bool hasInfo() => $_has(2);
  @$pb.TagNumber(4)
  void clearInfo() => clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get index => $_getI64(3);
  @$pb.TagNumber(5)
  set index($fixnum.Int64 v) { $_setInt64(3, v); }
  @$pb.TagNumber(5)
  $core.bool hasIndex() => $_has(3);
  @$pb.TagNumber(5)
  void clearIndex() => clearField(5);

  @$pb.TagNumber(6)
  $core.List<$core.int> get key => $_getN(4);
  @$pb.TagNumber(6)
  set key($core.List<$core.int> v) { $_setBytes(4, v); }
  @$pb.TagNumber(6)
  $core.bool hasKey() => $_has(4);
  @$pb.TagNumber(6)
  void clearKey() => clearField(6);

  @$pb.TagNumber(7)
  $core.List<$core.int> get value => $_getN(5);
  @$pb.TagNumber(7)
  set value($core.List<$core.int> v) { $_setBytes(5, v); }
  @$pb.TagNumber(7)
  $core.bool hasValue() => $_has(5);
  @$pb.TagNumber(7)
  void clearValue() => clearField(7);

  @$pb.TagNumber(8)
  $2.ProofOps get proofOps => $_getN(6);
  @$pb.TagNumber(8)
  set proofOps($2.ProofOps v) { setField(8, v); }
  @$pb.TagNumber(8)
  $core.bool hasProofOps() => $_has(6);
  @$pb.TagNumber(8)
  void clearProofOps() => clearField(8);
  @$pb.TagNumber(8)
  $2.ProofOps ensureProofOps() => $_ensure(6);

  @$pb.TagNumber(9)
  $fixnum.Int64 get height => $_getI64(7);
  @$pb.TagNumber(9)
  set height($fixnum.Int64 v) { $_setInt64(7, v); }
  @$pb.TagNumber(9)
  $core.bool hasHeight() => $_has(7);
  @$pb.TagNumber(9)
  void clearHeight() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get codespace => $_getSZ(8);
  @$pb.TagNumber(10)
  set codespace($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(10)
  $core.bool hasCodespace() => $_has(8);
  @$pb.TagNumber(10)
  void clearCodespace() => clearField(10);
}

class ResponseBeginBlock extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseBeginBlock', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..pc<Event>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'events', $pb.PbFieldType.PM, subBuilder: Event.create)
    ..hasRequiredFields = false
  ;

  ResponseBeginBlock._() : super();
  factory ResponseBeginBlock() => create();
  factory ResponseBeginBlock.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseBeginBlock.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseBeginBlock clone() => ResponseBeginBlock()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseBeginBlock copyWith(void Function(ResponseBeginBlock) updates) => super.copyWith((message) => updates(message as ResponseBeginBlock)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseBeginBlock create() => ResponseBeginBlock._();
  ResponseBeginBlock createEmptyInstance() => create();
  static $pb.PbList<ResponseBeginBlock> createRepeated() => $pb.PbList<ResponseBeginBlock>();
  @$core.pragma('dart2js:noInline')
  static ResponseBeginBlock getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseBeginBlock>(create);
  static ResponseBeginBlock _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Event> get events => $_getList(0);
}

class ResponseCheckTx extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseCheckTx', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.int>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'code', $pb.PbFieldType.OU3)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'data', $pb.PbFieldType.OY)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'log')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'info')
    ..aInt64(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'gas_wanted')
    ..aInt64(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'gas_used')
    ..pc<Event>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'events', $pb.PbFieldType.PM, subBuilder: Event.create)
    ..aOS(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'codespace')
    ..hasRequiredFields = false
  ;

  ResponseCheckTx._() : super();
  factory ResponseCheckTx() => create();
  factory ResponseCheckTx.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseCheckTx.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseCheckTx clone() => ResponseCheckTx()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseCheckTx copyWith(void Function(ResponseCheckTx) updates) => super.copyWith((message) => updates(message as ResponseCheckTx)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseCheckTx create() => ResponseCheckTx._();
  ResponseCheckTx createEmptyInstance() => create();
  static $pb.PbList<ResponseCheckTx> createRepeated() => $pb.PbList<ResponseCheckTx>();
  @$core.pragma('dart2js:noInline')
  static ResponseCheckTx getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseCheckTx>(create);
  static ResponseCheckTx _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get code => $_getIZ(0);
  @$pb.TagNumber(1)
  set code($core.int v) { $_setUnsignedInt32(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCode() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get data => $_getN(1);
  @$pb.TagNumber(2)
  set data($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasData() => $_has(1);
  @$pb.TagNumber(2)
  void clearData() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get log => $_getSZ(2);
  @$pb.TagNumber(3)
  set log($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasLog() => $_has(2);
  @$pb.TagNumber(3)
  void clearLog() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get info => $_getSZ(3);
  @$pb.TagNumber(4)
  set info($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasInfo() => $_has(3);
  @$pb.TagNumber(4)
  void clearInfo() => clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get gasWanted => $_getI64(4);
  @$pb.TagNumber(5)
  set gasWanted($fixnum.Int64 v) { $_setInt64(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasGasWanted() => $_has(4);
  @$pb.TagNumber(5)
  void clearGasWanted() => clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get gasUsed => $_getI64(5);
  @$pb.TagNumber(6)
  set gasUsed($fixnum.Int64 v) { $_setInt64(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasGasUsed() => $_has(5);
  @$pb.TagNumber(6)
  void clearGasUsed() => clearField(6);

  @$pb.TagNumber(7)
  $core.List<Event> get events => $_getList(6);

  @$pb.TagNumber(8)
  $core.String get codespace => $_getSZ(7);
  @$pb.TagNumber(8)
  set codespace($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasCodespace() => $_has(7);
  @$pb.TagNumber(8)
  void clearCodespace() => clearField(8);
}

class ResponseDeliverTx extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseDeliverTx', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.int>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'code', $pb.PbFieldType.OU3)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'data', $pb.PbFieldType.OY)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'log')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'info')
    ..aInt64(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'gas_wanted')
    ..aInt64(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'gas_used')
    ..pc<Event>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'events', $pb.PbFieldType.PM, subBuilder: Event.create)
    ..aOS(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'codespace')
    ..hasRequiredFields = false
  ;

  ResponseDeliverTx._() : super();
  factory ResponseDeliverTx() => create();
  factory ResponseDeliverTx.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseDeliverTx.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseDeliverTx clone() => ResponseDeliverTx()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseDeliverTx copyWith(void Function(ResponseDeliverTx) updates) => super.copyWith((message) => updates(message as ResponseDeliverTx)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseDeliverTx create() => ResponseDeliverTx._();
  ResponseDeliverTx createEmptyInstance() => create();
  static $pb.PbList<ResponseDeliverTx> createRepeated() => $pb.PbList<ResponseDeliverTx>();
  @$core.pragma('dart2js:noInline')
  static ResponseDeliverTx getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseDeliverTx>(create);
  static ResponseDeliverTx _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get code => $_getIZ(0);
  @$pb.TagNumber(1)
  set code($core.int v) { $_setUnsignedInt32(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCode() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get data => $_getN(1);
  @$pb.TagNumber(2)
  set data($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasData() => $_has(1);
  @$pb.TagNumber(2)
  void clearData() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get log => $_getSZ(2);
  @$pb.TagNumber(3)
  set log($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasLog() => $_has(2);
  @$pb.TagNumber(3)
  void clearLog() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get info => $_getSZ(3);
  @$pb.TagNumber(4)
  set info($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasInfo() => $_has(3);
  @$pb.TagNumber(4)
  void clearInfo() => clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get gasWanted => $_getI64(4);
  @$pb.TagNumber(5)
  set gasWanted($fixnum.Int64 v) { $_setInt64(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasGasWanted() => $_has(4);
  @$pb.TagNumber(5)
  void clearGasWanted() => clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get gasUsed => $_getI64(5);
  @$pb.TagNumber(6)
  set gasUsed($fixnum.Int64 v) { $_setInt64(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasGasUsed() => $_has(5);
  @$pb.TagNumber(6)
  void clearGasUsed() => clearField(6);

  @$pb.TagNumber(7)
  $core.List<Event> get events => $_getList(6);

  @$pb.TagNumber(8)
  $core.String get codespace => $_getSZ(7);
  @$pb.TagNumber(8)
  set codespace($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasCodespace() => $_has(7);
  @$pb.TagNumber(8)
  void clearCodespace() => clearField(8);
}

class ResponseEndBlock extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseEndBlock', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..pc<ValidatorUpdate>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorUpdates', $pb.PbFieldType.PM, subBuilder: ValidatorUpdate.create)
    ..aOM<ConsensusParams>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'consensusParamUpdates', subBuilder: ConsensusParams.create)
    ..pc<Event>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'events', $pb.PbFieldType.PM, subBuilder: Event.create)
    ..hasRequiredFields = false
  ;

  ResponseEndBlock._() : super();
  factory ResponseEndBlock() => create();
  factory ResponseEndBlock.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseEndBlock.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseEndBlock clone() => ResponseEndBlock()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseEndBlock copyWith(void Function(ResponseEndBlock) updates) => super.copyWith((message) => updates(message as ResponseEndBlock)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseEndBlock create() => ResponseEndBlock._();
  ResponseEndBlock createEmptyInstance() => create();
  static $pb.PbList<ResponseEndBlock> createRepeated() => $pb.PbList<ResponseEndBlock>();
  @$core.pragma('dart2js:noInline')
  static ResponseEndBlock getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseEndBlock>(create);
  static ResponseEndBlock _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<ValidatorUpdate> get validatorUpdates => $_getList(0);

  @$pb.TagNumber(2)
  ConsensusParams get consensusParamUpdates => $_getN(1);
  @$pb.TagNumber(2)
  set consensusParamUpdates(ConsensusParams v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasConsensusParamUpdates() => $_has(1);
  @$pb.TagNumber(2)
  void clearConsensusParamUpdates() => clearField(2);
  @$pb.TagNumber(2)
  ConsensusParams ensureConsensusParamUpdates() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.List<Event> get events => $_getList(2);
}

class ResponseCommit extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseCommit', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'data', $pb.PbFieldType.OY)
    ..aInt64(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'retainHeight')
    ..hasRequiredFields = false
  ;

  ResponseCommit._() : super();
  factory ResponseCommit() => create();
  factory ResponseCommit.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseCommit.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseCommit clone() => ResponseCommit()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseCommit copyWith(void Function(ResponseCommit) updates) => super.copyWith((message) => updates(message as ResponseCommit)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseCommit create() => ResponseCommit._();
  ResponseCommit createEmptyInstance() => create();
  static $pb.PbList<ResponseCommit> createRepeated() => $pb.PbList<ResponseCommit>();
  @$core.pragma('dart2js:noInline')
  static ResponseCommit getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseCommit>(create);
  static ResponseCommit _defaultInstance;

  @$pb.TagNumber(2)
  $core.List<$core.int> get data => $_getN(0);
  @$pb.TagNumber(2)
  set data($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(2)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(2)
  void clearData() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get retainHeight => $_getI64(1);
  @$pb.TagNumber(3)
  set retainHeight($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(3)
  $core.bool hasRetainHeight() => $_has(1);
  @$pb.TagNumber(3)
  void clearRetainHeight() => clearField(3);
}

class ResponseListSnapshots extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseListSnapshots', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..pc<Snapshot>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'snapshots', $pb.PbFieldType.PM, subBuilder: Snapshot.create)
    ..hasRequiredFields = false
  ;

  ResponseListSnapshots._() : super();
  factory ResponseListSnapshots() => create();
  factory ResponseListSnapshots.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseListSnapshots.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseListSnapshots clone() => ResponseListSnapshots()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseListSnapshots copyWith(void Function(ResponseListSnapshots) updates) => super.copyWith((message) => updates(message as ResponseListSnapshots)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseListSnapshots create() => ResponseListSnapshots._();
  ResponseListSnapshots createEmptyInstance() => create();
  static $pb.PbList<ResponseListSnapshots> createRepeated() => $pb.PbList<ResponseListSnapshots>();
  @$core.pragma('dart2js:noInline')
  static ResponseListSnapshots getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseListSnapshots>(create);
  static ResponseListSnapshots _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Snapshot> get snapshots => $_getList(0);
}

class ResponseOfferSnapshot extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseOfferSnapshot', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..e<ResponseOfferSnapshot_Result>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'result', $pb.PbFieldType.OE, defaultOrMaker: ResponseOfferSnapshot_Result.UNKNOWN, valueOf: ResponseOfferSnapshot_Result.valueOf, enumValues: ResponseOfferSnapshot_Result.values)
    ..hasRequiredFields = false
  ;

  ResponseOfferSnapshot._() : super();
  factory ResponseOfferSnapshot() => create();
  factory ResponseOfferSnapshot.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseOfferSnapshot.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseOfferSnapshot clone() => ResponseOfferSnapshot()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseOfferSnapshot copyWith(void Function(ResponseOfferSnapshot) updates) => super.copyWith((message) => updates(message as ResponseOfferSnapshot)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseOfferSnapshot create() => ResponseOfferSnapshot._();
  ResponseOfferSnapshot createEmptyInstance() => create();
  static $pb.PbList<ResponseOfferSnapshot> createRepeated() => $pb.PbList<ResponseOfferSnapshot>();
  @$core.pragma('dart2js:noInline')
  static ResponseOfferSnapshot getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseOfferSnapshot>(create);
  static ResponseOfferSnapshot _defaultInstance;

  @$pb.TagNumber(1)
  ResponseOfferSnapshot_Result get result => $_getN(0);
  @$pb.TagNumber(1)
  set result(ResponseOfferSnapshot_Result v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasResult() => $_has(0);
  @$pb.TagNumber(1)
  void clearResult() => clearField(1);
}

class ResponseLoadSnapshotChunk extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseLoadSnapshotChunk', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'chunk', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  ResponseLoadSnapshotChunk._() : super();
  factory ResponseLoadSnapshotChunk() => create();
  factory ResponseLoadSnapshotChunk.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseLoadSnapshotChunk.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseLoadSnapshotChunk clone() => ResponseLoadSnapshotChunk()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseLoadSnapshotChunk copyWith(void Function(ResponseLoadSnapshotChunk) updates) => super.copyWith((message) => updates(message as ResponseLoadSnapshotChunk)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseLoadSnapshotChunk create() => ResponseLoadSnapshotChunk._();
  ResponseLoadSnapshotChunk createEmptyInstance() => create();
  static $pb.PbList<ResponseLoadSnapshotChunk> createRepeated() => $pb.PbList<ResponseLoadSnapshotChunk>();
  @$core.pragma('dart2js:noInline')
  static ResponseLoadSnapshotChunk getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseLoadSnapshotChunk>(create);
  static ResponseLoadSnapshotChunk _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get chunk => $_getN(0);
  @$pb.TagNumber(1)
  set chunk($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasChunk() => $_has(0);
  @$pb.TagNumber(1)
  void clearChunk() => clearField(1);
}

class ResponseApplySnapshotChunk extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ResponseApplySnapshotChunk', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..e<ResponseApplySnapshotChunk_Result>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'result', $pb.PbFieldType.OE, defaultOrMaker: ResponseApplySnapshotChunk_Result.UNKNOWN, valueOf: ResponseApplySnapshotChunk_Result.valueOf, enumValues: ResponseApplySnapshotChunk_Result.values)
    ..p<$core.int>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'refetchChunks', $pb.PbFieldType.PU3)
    ..pPS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'rejectSenders')
    ..hasRequiredFields = false
  ;

  ResponseApplySnapshotChunk._() : super();
  factory ResponseApplySnapshotChunk() => create();
  factory ResponseApplySnapshotChunk.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ResponseApplySnapshotChunk.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ResponseApplySnapshotChunk clone() => ResponseApplySnapshotChunk()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ResponseApplySnapshotChunk copyWith(void Function(ResponseApplySnapshotChunk) updates) => super.copyWith((message) => updates(message as ResponseApplySnapshotChunk)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ResponseApplySnapshotChunk create() => ResponseApplySnapshotChunk._();
  ResponseApplySnapshotChunk createEmptyInstance() => create();
  static $pb.PbList<ResponseApplySnapshotChunk> createRepeated() => $pb.PbList<ResponseApplySnapshotChunk>();
  @$core.pragma('dart2js:noInline')
  static ResponseApplySnapshotChunk getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResponseApplySnapshotChunk>(create);
  static ResponseApplySnapshotChunk _defaultInstance;

  @$pb.TagNumber(1)
  ResponseApplySnapshotChunk_Result get result => $_getN(0);
  @$pb.TagNumber(1)
  set result(ResponseApplySnapshotChunk_Result v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasResult() => $_has(0);
  @$pb.TagNumber(1)
  void clearResult() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get refetchChunks => $_getList(1);

  @$pb.TagNumber(3)
  $core.List<$core.String> get rejectSenders => $_getList(2);
}

class ConsensusParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ConsensusParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOM<BlockParams>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'block', subBuilder: BlockParams.create)
    ..aOM<$8.EvidenceParams>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'evidence', subBuilder: $8.EvidenceParams.create)
    ..aOM<$8.ValidatorParams>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validator', subBuilder: $8.ValidatorParams.create)
    ..aOM<$8.VersionParams>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'version', subBuilder: $8.VersionParams.create)
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
  $8.EvidenceParams get evidence => $_getN(1);
  @$pb.TagNumber(2)
  set evidence($8.EvidenceParams v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasEvidence() => $_has(1);
  @$pb.TagNumber(2)
  void clearEvidence() => clearField(2);
  @$pb.TagNumber(2)
  $8.EvidenceParams ensureEvidence() => $_ensure(1);

  @$pb.TagNumber(3)
  $8.ValidatorParams get validator => $_getN(2);
  @$pb.TagNumber(3)
  set validator($8.ValidatorParams v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasValidator() => $_has(2);
  @$pb.TagNumber(3)
  void clearValidator() => clearField(3);
  @$pb.TagNumber(3)
  $8.ValidatorParams ensureValidator() => $_ensure(2);

  @$pb.TagNumber(4)
  $8.VersionParams get version => $_getN(3);
  @$pb.TagNumber(4)
  set version($8.VersionParams v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasVersion() => $_has(3);
  @$pb.TagNumber(4)
  void clearVersion() => clearField(4);
  @$pb.TagNumber(4)
  $8.VersionParams ensureVersion() => $_ensure(3);
}

class BlockParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'BlockParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aInt64(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxBytes')
    ..aInt64(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxGas')
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
}

class LastCommitInfo extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LastCommitInfo', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.int>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'round', $pb.PbFieldType.O3)
    ..pc<VoteInfo>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'votes', $pb.PbFieldType.PM, subBuilder: VoteInfo.create)
    ..hasRequiredFields = false
  ;

  LastCommitInfo._() : super();
  factory LastCommitInfo() => create();
  factory LastCommitInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LastCommitInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LastCommitInfo clone() => LastCommitInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LastCommitInfo copyWith(void Function(LastCommitInfo) updates) => super.copyWith((message) => updates(message as LastCommitInfo)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static LastCommitInfo create() => LastCommitInfo._();
  LastCommitInfo createEmptyInstance() => create();
  static $pb.PbList<LastCommitInfo> createRepeated() => $pb.PbList<LastCommitInfo>();
  @$core.pragma('dart2js:noInline')
  static LastCommitInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LastCommitInfo>(create);
  static LastCommitInfo _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get round => $_getIZ(0);
  @$pb.TagNumber(1)
  set round($core.int v) { $_setSignedInt32(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasRound() => $_has(0);
  @$pb.TagNumber(1)
  void clearRound() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<VoteInfo> get votes => $_getList(1);
}

class Event extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Event', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'type')
    ..pc<EventAttribute>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'attributes', $pb.PbFieldType.PM, subBuilder: EventAttribute.create)
    ..hasRequiredFields = false
  ;

  Event._() : super();
  factory Event() => create();
  factory Event.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Event.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Event clone() => Event()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Event copyWith(void Function(Event) updates) => super.copyWith((message) => updates(message as Event)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Event create() => Event._();
  Event createEmptyInstance() => create();
  static $pb.PbList<Event> createRepeated() => $pb.PbList<Event>();
  @$core.pragma('dart2js:noInline')
  static Event getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Event>(create);
  static Event _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get type => $_getSZ(0);
  @$pb.TagNumber(1)
  set type($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<EventAttribute> get attributes => $_getList(1);
}

class EventAttribute extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'EventAttribute', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'key', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'value', $pb.PbFieldType.OY)
    ..aOB(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'index')
    ..hasRequiredFields = false
  ;

  EventAttribute._() : super();
  factory EventAttribute() => create();
  factory EventAttribute.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory EventAttribute.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  EventAttribute clone() => EventAttribute()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  EventAttribute copyWith(void Function(EventAttribute) updates) => super.copyWith((message) => updates(message as EventAttribute)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static EventAttribute create() => EventAttribute._();
  EventAttribute createEmptyInstance() => create();
  static $pb.PbList<EventAttribute> createRepeated() => $pb.PbList<EventAttribute>();
  @$core.pragma('dart2js:noInline')
  static EventAttribute getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<EventAttribute>(create);
  static EventAttribute _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get key => $_getN(0);
  @$pb.TagNumber(1)
  set key($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasKey() => $_has(0);
  @$pb.TagNumber(1)
  void clearKey() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get value => $_getN(1);
  @$pb.TagNumber(2)
  set value($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasValue() => $_has(1);
  @$pb.TagNumber(2)
  void clearValue() => clearField(2);

  @$pb.TagNumber(3)
  $core.bool get index => $_getBF(2);
  @$pb.TagNumber(3)
  set index($core.bool v) { $_setBool(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasIndex() => $_has(2);
  @$pb.TagNumber(3)
  void clearIndex() => clearField(3);
}

class TxResult extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'TxResult', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aInt64(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height')
    ..a<$core.int>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'index', $pb.PbFieldType.OU3)
    ..a<$core.List<$core.int>>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tx', $pb.PbFieldType.OY)
    ..aOM<ResponseDeliverTx>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'result', subBuilder: ResponseDeliverTx.create)
    ..hasRequiredFields = false
  ;

  TxResult._() : super();
  factory TxResult() => create();
  factory TxResult.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory TxResult.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  TxResult clone() => TxResult()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  TxResult copyWith(void Function(TxResult) updates) => super.copyWith((message) => updates(message as TxResult)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static TxResult create() => TxResult._();
  TxResult createEmptyInstance() => create();
  static $pb.PbList<TxResult> createRepeated() => $pb.PbList<TxResult>();
  @$core.pragma('dart2js:noInline')
  static TxResult getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TxResult>(create);
  static TxResult _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get height => $_getI64(0);
  @$pb.TagNumber(1)
  set height($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasHeight() => $_has(0);
  @$pb.TagNumber(1)
  void clearHeight() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get index => $_getIZ(1);
  @$pb.TagNumber(2)
  set index($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasIndex() => $_has(1);
  @$pb.TagNumber(2)
  void clearIndex() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get tx => $_getN(2);
  @$pb.TagNumber(3)
  set tx($core.List<$core.int> v) { $_setBytes(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTx() => $_has(2);
  @$pb.TagNumber(3)
  void clearTx() => clearField(3);

  @$pb.TagNumber(4)
  ResponseDeliverTx get result => $_getN(3);
  @$pb.TagNumber(4)
  set result(ResponseDeliverTx v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasResult() => $_has(3);
  @$pb.TagNumber(4)
  void clearResult() => clearField(4);
  @$pb.TagNumber(4)
  ResponseDeliverTx ensureResult() => $_ensure(3);
}

class Validator extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Validator', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'address', $pb.PbFieldType.OY)
    ..aInt64(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'power')
    ..hasRequiredFields = false
  ;

  Validator._() : super();
  factory Validator() => create();
  factory Validator.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Validator.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Validator clone() => Validator()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Validator copyWith(void Function(Validator) updates) => super.copyWith((message) => updates(message as Validator)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Validator create() => Validator._();
  Validator createEmptyInstance() => create();
  static $pb.PbList<Validator> createRepeated() => $pb.PbList<Validator>();
  @$core.pragma('dart2js:noInline')
  static Validator getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Validator>(create);
  static Validator _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get address => $_getN(0);
  @$pb.TagNumber(1)
  set address($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => clearField(1);

  @$pb.TagNumber(3)
  $fixnum.Int64 get power => $_getI64(1);
  @$pb.TagNumber(3)
  set power($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(3)
  $core.bool hasPower() => $_has(1);
  @$pb.TagNumber(3)
  void clearPower() => clearField(3);
}

class ValidatorUpdate extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValidatorUpdate', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOM<$1.PublicKey>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pubKey', subBuilder: $1.PublicKey.create)
    ..aInt64(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'power')
    ..hasRequiredFields = false
  ;

  ValidatorUpdate._() : super();
  factory ValidatorUpdate() => create();
  factory ValidatorUpdate.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValidatorUpdate.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValidatorUpdate clone() => ValidatorUpdate()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValidatorUpdate copyWith(void Function(ValidatorUpdate) updates) => super.copyWith((message) => updates(message as ValidatorUpdate)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValidatorUpdate create() => ValidatorUpdate._();
  ValidatorUpdate createEmptyInstance() => create();
  static $pb.PbList<ValidatorUpdate> createRepeated() => $pb.PbList<ValidatorUpdate>();
  @$core.pragma('dart2js:noInline')
  static ValidatorUpdate getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValidatorUpdate>(create);
  static ValidatorUpdate _defaultInstance;

  @$pb.TagNumber(1)
  $1.PublicKey get pubKey => $_getN(0);
  @$pb.TagNumber(1)
  set pubKey($1.PublicKey v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasPubKey() => $_has(0);
  @$pb.TagNumber(1)
  void clearPubKey() => clearField(1);
  @$pb.TagNumber(1)
  $1.PublicKey ensurePubKey() => $_ensure(0);

  @$pb.TagNumber(2)
  $fixnum.Int64 get power => $_getI64(1);
  @$pb.TagNumber(2)
  set power($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPower() => $_has(1);
  @$pb.TagNumber(2)
  void clearPower() => clearField(2);
}

class VoteInfo extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'VoteInfo', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..aOM<Validator>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validator', subBuilder: Validator.create)
    ..aOB(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signedLastBlock')
    ..hasRequiredFields = false
  ;

  VoteInfo._() : super();
  factory VoteInfo() => create();
  factory VoteInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory VoteInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  VoteInfo clone() => VoteInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  VoteInfo copyWith(void Function(VoteInfo) updates) => super.copyWith((message) => updates(message as VoteInfo)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static VoteInfo create() => VoteInfo._();
  VoteInfo createEmptyInstance() => create();
  static $pb.PbList<VoteInfo> createRepeated() => $pb.PbList<VoteInfo>();
  @$core.pragma('dart2js:noInline')
  static VoteInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VoteInfo>(create);
  static VoteInfo _defaultInstance;

  @$pb.TagNumber(1)
  Validator get validator => $_getN(0);
  @$pb.TagNumber(1)
  set validator(Validator v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidator() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidator() => clearField(1);
  @$pb.TagNumber(1)
  Validator ensureValidator() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.bool get signedLastBlock => $_getBF(1);
  @$pb.TagNumber(2)
  set signedLastBlock($core.bool v) { $_setBool(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasSignedLastBlock() => $_has(1);
  @$pb.TagNumber(2)
  void clearSignedLastBlock() => clearField(2);
}

class Evidence extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Evidence', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..e<EvidenceType>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'type', $pb.PbFieldType.OE, defaultOrMaker: EvidenceType.UNKNOWN, valueOf: EvidenceType.valueOf, enumValues: EvidenceType.values)
    ..aOM<Validator>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validator', subBuilder: Validator.create)
    ..aInt64(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height')
    ..aOM<$4.Timestamp>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'time', subBuilder: $4.Timestamp.create)
    ..aInt64(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'totalVotingPower')
    ..hasRequiredFields = false
  ;

  Evidence._() : super();
  factory Evidence() => create();
  factory Evidence.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Evidence.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Evidence clone() => Evidence()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Evidence copyWith(void Function(Evidence) updates) => super.copyWith((message) => updates(message as Evidence)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Evidence create() => Evidence._();
  Evidence createEmptyInstance() => create();
  static $pb.PbList<Evidence> createRepeated() => $pb.PbList<Evidence>();
  @$core.pragma('dart2js:noInline')
  static Evidence getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Evidence>(create);
  static Evidence _defaultInstance;

  @$pb.TagNumber(1)
  EvidenceType get type => $_getN(0);
  @$pb.TagNumber(1)
  set type(EvidenceType v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => clearField(1);

  @$pb.TagNumber(2)
  Validator get validator => $_getN(1);
  @$pb.TagNumber(2)
  set validator(Validator v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasValidator() => $_has(1);
  @$pb.TagNumber(2)
  void clearValidator() => clearField(2);
  @$pb.TagNumber(2)
  Validator ensureValidator() => $_ensure(1);

  @$pb.TagNumber(3)
  $fixnum.Int64 get height => $_getI64(2);
  @$pb.TagNumber(3)
  set height($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearHeight() => clearField(3);

  @$pb.TagNumber(4)
  $4.Timestamp get time => $_getN(3);
  @$pb.TagNumber(4)
  set time($4.Timestamp v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasTime() => $_has(3);
  @$pb.TagNumber(4)
  void clearTime() => clearField(4);
  @$pb.TagNumber(4)
  $4.Timestamp ensureTime() => $_ensure(3);

  @$pb.TagNumber(5)
  $fixnum.Int64 get totalVotingPower => $_getI64(4);
  @$pb.TagNumber(5)
  set totalVotingPower($fixnum.Int64 v) { $_setInt64(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasTotalVotingPower() => $_has(4);
  @$pb.TagNumber(5)
  void clearTotalVotingPower() => clearField(5);
}

class Snapshot extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Snapshot', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'tendermint.abci'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$core.int>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'format', $pb.PbFieldType.OU3)
    ..a<$core.int>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'chunks', $pb.PbFieldType.OU3)
    ..a<$core.List<$core.int>>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'hash', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'metadata', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  Snapshot._() : super();
  factory Snapshot() => create();
  factory Snapshot.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Snapshot.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Snapshot clone() => Snapshot()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Snapshot copyWith(void Function(Snapshot) updates) => super.copyWith((message) => updates(message as Snapshot)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Snapshot create() => Snapshot._();
  Snapshot createEmptyInstance() => create();
  static $pb.PbList<Snapshot> createRepeated() => $pb.PbList<Snapshot>();
  @$core.pragma('dart2js:noInline')
  static Snapshot getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Snapshot>(create);
  static Snapshot _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get height => $_getI64(0);
  @$pb.TagNumber(1)
  set height($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasHeight() => $_has(0);
  @$pb.TagNumber(1)
  void clearHeight() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get format => $_getIZ(1);
  @$pb.TagNumber(2)
  set format($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasFormat() => $_has(1);
  @$pb.TagNumber(2)
  void clearFormat() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get chunks => $_getIZ(2);
  @$pb.TagNumber(3)
  set chunks($core.int v) { $_setUnsignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasChunks() => $_has(2);
  @$pb.TagNumber(3)
  void clearChunks() => clearField(3);

  @$pb.TagNumber(4)
  $core.List<$core.int> get hash => $_getN(3);
  @$pb.TagNumber(4)
  set hash($core.List<$core.int> v) { $_setBytes(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasHash() => $_has(3);
  @$pb.TagNumber(4)
  void clearHash() => clearField(4);

  @$pb.TagNumber(5)
  $core.List<$core.int> get metadata => $_getN(4);
  @$pb.TagNumber(5)
  set metadata($core.List<$core.int> v) { $_setBytes(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasMetadata() => $_has(4);
  @$pb.TagNumber(5)
  void clearMetadata() => clearField(5);
}

