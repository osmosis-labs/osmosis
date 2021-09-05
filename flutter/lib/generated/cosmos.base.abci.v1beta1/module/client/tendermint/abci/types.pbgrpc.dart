///
//  Generated code. Do not modify.
//  source: tendermint/abci/types.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'types.pb.dart' as $0;
export 'types.pb.dart';

class ABCIApplicationClient extends $grpc.Client {
  static final _$echo = $grpc.ClientMethod<$0.RequestEcho, $0.ResponseEcho>(
      '/tendermint.abci.ABCIApplication/Echo',
      ($0.RequestEcho value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.ResponseEcho.fromBuffer(value));
  static final _$flush = $grpc.ClientMethod<$0.RequestFlush, $0.ResponseFlush>(
      '/tendermint.abci.ABCIApplication/Flush',
      ($0.RequestFlush value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.ResponseFlush.fromBuffer(value));
  static final _$info = $grpc.ClientMethod<$0.RequestInfo, $0.ResponseInfo>(
      '/tendermint.abci.ABCIApplication/Info',
      ($0.RequestInfo value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.ResponseInfo.fromBuffer(value));
  static final _$setOption =
      $grpc.ClientMethod<$0.RequestSetOption, $0.ResponseSetOption>(
          '/tendermint.abci.ABCIApplication/SetOption',
          ($0.RequestSetOption value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.ResponseSetOption.fromBuffer(value));
  static final _$deliverTx =
      $grpc.ClientMethod<$0.RequestDeliverTx, $0.ResponseDeliverTx>(
          '/tendermint.abci.ABCIApplication/DeliverTx',
          ($0.RequestDeliverTx value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.ResponseDeliverTx.fromBuffer(value));
  static final _$checkTx =
      $grpc.ClientMethod<$0.RequestCheckTx, $0.ResponseCheckTx>(
          '/tendermint.abci.ABCIApplication/CheckTx',
          ($0.RequestCheckTx value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.ResponseCheckTx.fromBuffer(value));
  static final _$query = $grpc.ClientMethod<$0.RequestQuery, $0.ResponseQuery>(
      '/tendermint.abci.ABCIApplication/Query',
      ($0.RequestQuery value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.ResponseQuery.fromBuffer(value));
  static final _$commit =
      $grpc.ClientMethod<$0.RequestCommit, $0.ResponseCommit>(
          '/tendermint.abci.ABCIApplication/Commit',
          ($0.RequestCommit value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $0.ResponseCommit.fromBuffer(value));
  static final _$initChain =
      $grpc.ClientMethod<$0.RequestInitChain, $0.ResponseInitChain>(
          '/tendermint.abci.ABCIApplication/InitChain',
          ($0.RequestInitChain value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.ResponseInitChain.fromBuffer(value));
  static final _$beginBlock =
      $grpc.ClientMethod<$0.RequestBeginBlock, $0.ResponseBeginBlock>(
          '/tendermint.abci.ABCIApplication/BeginBlock',
          ($0.RequestBeginBlock value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.ResponseBeginBlock.fromBuffer(value));
  static final _$endBlock =
      $grpc.ClientMethod<$0.RequestEndBlock, $0.ResponseEndBlock>(
          '/tendermint.abci.ABCIApplication/EndBlock',
          ($0.RequestEndBlock value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.ResponseEndBlock.fromBuffer(value));
  static final _$listSnapshots =
      $grpc.ClientMethod<$0.RequestListSnapshots, $0.ResponseListSnapshots>(
          '/tendermint.abci.ABCIApplication/ListSnapshots',
          ($0.RequestListSnapshots value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.ResponseListSnapshots.fromBuffer(value));
  static final _$offerSnapshot =
      $grpc.ClientMethod<$0.RequestOfferSnapshot, $0.ResponseOfferSnapshot>(
          '/tendermint.abci.ABCIApplication/OfferSnapshot',
          ($0.RequestOfferSnapshot value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.ResponseOfferSnapshot.fromBuffer(value));
  static final _$loadSnapshotChunk = $grpc.ClientMethod<
          $0.RequestLoadSnapshotChunk, $0.ResponseLoadSnapshotChunk>(
      '/tendermint.abci.ABCIApplication/LoadSnapshotChunk',
      ($0.RequestLoadSnapshotChunk value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.ResponseLoadSnapshotChunk.fromBuffer(value));
  static final _$applySnapshotChunk = $grpc.ClientMethod<
          $0.RequestApplySnapshotChunk, $0.ResponseApplySnapshotChunk>(
      '/tendermint.abci.ABCIApplication/ApplySnapshotChunk',
      ($0.RequestApplySnapshotChunk value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.ResponseApplySnapshotChunk.fromBuffer(value));

  ABCIApplicationClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.ResponseEcho> echo($0.RequestEcho request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$echo, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseFlush> flush($0.RequestFlush request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$flush, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseInfo> info($0.RequestInfo request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$info, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseSetOption> setOption(
      $0.RequestSetOption request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$setOption, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseDeliverTx> deliverTx(
      $0.RequestDeliverTx request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$deliverTx, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseCheckTx> checkTx($0.RequestCheckTx request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$checkTx, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseQuery> query($0.RequestQuery request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$query, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseCommit> commit($0.RequestCommit request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$commit, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseInitChain> initChain(
      $0.RequestInitChain request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$initChain, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseBeginBlock> beginBlock(
      $0.RequestBeginBlock request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$beginBlock, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseEndBlock> endBlock($0.RequestEndBlock request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$endBlock, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseListSnapshots> listSnapshots(
      $0.RequestListSnapshots request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$listSnapshots, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseOfferSnapshot> offerSnapshot(
      $0.RequestOfferSnapshot request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$offerSnapshot, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseLoadSnapshotChunk> loadSnapshotChunk(
      $0.RequestLoadSnapshotChunk request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$loadSnapshotChunk, request, options: options);
  }

  $grpc.ResponseFuture<$0.ResponseApplySnapshotChunk> applySnapshotChunk(
      $0.RequestApplySnapshotChunk request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$applySnapshotChunk, request, options: options);
  }
}

abstract class ABCIApplicationServiceBase extends $grpc.Service {
  $core.String get $name => 'tendermint.abci.ABCIApplication';

  ABCIApplicationServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.RequestEcho, $0.ResponseEcho>(
        'Echo',
        echo_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RequestEcho.fromBuffer(value),
        ($0.ResponseEcho value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestFlush, $0.ResponseFlush>(
        'Flush',
        flush_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RequestFlush.fromBuffer(value),
        ($0.ResponseFlush value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestInfo, $0.ResponseInfo>(
        'Info',
        info_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RequestInfo.fromBuffer(value),
        ($0.ResponseInfo value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestSetOption, $0.ResponseSetOption>(
        'SetOption',
        setOption_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RequestSetOption.fromBuffer(value),
        ($0.ResponseSetOption value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestDeliverTx, $0.ResponseDeliverTx>(
        'DeliverTx',
        deliverTx_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RequestDeliverTx.fromBuffer(value),
        ($0.ResponseDeliverTx value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestCheckTx, $0.ResponseCheckTx>(
        'CheckTx',
        checkTx_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RequestCheckTx.fromBuffer(value),
        ($0.ResponseCheckTx value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestQuery, $0.ResponseQuery>(
        'Query',
        query_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RequestQuery.fromBuffer(value),
        ($0.ResponseQuery value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestCommit, $0.ResponseCommit>(
        'Commit',
        commit_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RequestCommit.fromBuffer(value),
        ($0.ResponseCommit value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestInitChain, $0.ResponseInitChain>(
        'InitChain',
        initChain_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RequestInitChain.fromBuffer(value),
        ($0.ResponseInitChain value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestBeginBlock, $0.ResponseBeginBlock>(
        'BeginBlock',
        beginBlock_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RequestBeginBlock.fromBuffer(value),
        ($0.ResponseBeginBlock value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestEndBlock, $0.ResponseEndBlock>(
        'EndBlock',
        endBlock_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RequestEndBlock.fromBuffer(value),
        ($0.ResponseEndBlock value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.RequestListSnapshots, $0.ResponseListSnapshots>(
            'ListSnapshots',
            listSnapshots_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.RequestListSnapshots.fromBuffer(value),
            ($0.ResponseListSnapshots value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.RequestOfferSnapshot, $0.ResponseOfferSnapshot>(
            'OfferSnapshot',
            offerSnapshot_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.RequestOfferSnapshot.fromBuffer(value),
            ($0.ResponseOfferSnapshot value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestLoadSnapshotChunk,
            $0.ResponseLoadSnapshotChunk>(
        'LoadSnapshotChunk',
        loadSnapshotChunk_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.RequestLoadSnapshotChunk.fromBuffer(value),
        ($0.ResponseLoadSnapshotChunk value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RequestApplySnapshotChunk,
            $0.ResponseApplySnapshotChunk>(
        'ApplySnapshotChunk',
        applySnapshotChunk_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.RequestApplySnapshotChunk.fromBuffer(value),
        ($0.ResponseApplySnapshotChunk value) => value.writeToBuffer()));
  }

  $async.Future<$0.ResponseEcho> echo_Pre(
      $grpc.ServiceCall call, $async.Future<$0.RequestEcho> request) async {
    return echo(call, await request);
  }

  $async.Future<$0.ResponseFlush> flush_Pre(
      $grpc.ServiceCall call, $async.Future<$0.RequestFlush> request) async {
    return flush(call, await request);
  }

  $async.Future<$0.ResponseInfo> info_Pre(
      $grpc.ServiceCall call, $async.Future<$0.RequestInfo> request) async {
    return info(call, await request);
  }

  $async.Future<$0.ResponseSetOption> setOption_Pre($grpc.ServiceCall call,
      $async.Future<$0.RequestSetOption> request) async {
    return setOption(call, await request);
  }

  $async.Future<$0.ResponseDeliverTx> deliverTx_Pre($grpc.ServiceCall call,
      $async.Future<$0.RequestDeliverTx> request) async {
    return deliverTx(call, await request);
  }

  $async.Future<$0.ResponseCheckTx> checkTx_Pre(
      $grpc.ServiceCall call, $async.Future<$0.RequestCheckTx> request) async {
    return checkTx(call, await request);
  }

  $async.Future<$0.ResponseQuery> query_Pre(
      $grpc.ServiceCall call, $async.Future<$0.RequestQuery> request) async {
    return query(call, await request);
  }

  $async.Future<$0.ResponseCommit> commit_Pre(
      $grpc.ServiceCall call, $async.Future<$0.RequestCommit> request) async {
    return commit(call, await request);
  }

  $async.Future<$0.ResponseInitChain> initChain_Pre($grpc.ServiceCall call,
      $async.Future<$0.RequestInitChain> request) async {
    return initChain(call, await request);
  }

  $async.Future<$0.ResponseBeginBlock> beginBlock_Pre($grpc.ServiceCall call,
      $async.Future<$0.RequestBeginBlock> request) async {
    return beginBlock(call, await request);
  }

  $async.Future<$0.ResponseEndBlock> endBlock_Pre(
      $grpc.ServiceCall call, $async.Future<$0.RequestEndBlock> request) async {
    return endBlock(call, await request);
  }

  $async.Future<$0.ResponseListSnapshots> listSnapshots_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.RequestListSnapshots> request) async {
    return listSnapshots(call, await request);
  }

  $async.Future<$0.ResponseOfferSnapshot> offerSnapshot_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.RequestOfferSnapshot> request) async {
    return offerSnapshot(call, await request);
  }

  $async.Future<$0.ResponseLoadSnapshotChunk> loadSnapshotChunk_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.RequestLoadSnapshotChunk> request) async {
    return loadSnapshotChunk(call, await request);
  }

  $async.Future<$0.ResponseApplySnapshotChunk> applySnapshotChunk_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.RequestApplySnapshotChunk> request) async {
    return applySnapshotChunk(call, await request);
  }

  $async.Future<$0.ResponseEcho> echo(
      $grpc.ServiceCall call, $0.RequestEcho request);
  $async.Future<$0.ResponseFlush> flush(
      $grpc.ServiceCall call, $0.RequestFlush request);
  $async.Future<$0.ResponseInfo> info(
      $grpc.ServiceCall call, $0.RequestInfo request);
  $async.Future<$0.ResponseSetOption> setOption(
      $grpc.ServiceCall call, $0.RequestSetOption request);
  $async.Future<$0.ResponseDeliverTx> deliverTx(
      $grpc.ServiceCall call, $0.RequestDeliverTx request);
  $async.Future<$0.ResponseCheckTx> checkTx(
      $grpc.ServiceCall call, $0.RequestCheckTx request);
  $async.Future<$0.ResponseQuery> query(
      $grpc.ServiceCall call, $0.RequestQuery request);
  $async.Future<$0.ResponseCommit> commit(
      $grpc.ServiceCall call, $0.RequestCommit request);
  $async.Future<$0.ResponseInitChain> initChain(
      $grpc.ServiceCall call, $0.RequestInitChain request);
  $async.Future<$0.ResponseBeginBlock> beginBlock(
      $grpc.ServiceCall call, $0.RequestBeginBlock request);
  $async.Future<$0.ResponseEndBlock> endBlock(
      $grpc.ServiceCall call, $0.RequestEndBlock request);
  $async.Future<$0.ResponseListSnapshots> listSnapshots(
      $grpc.ServiceCall call, $0.RequestListSnapshots request);
  $async.Future<$0.ResponseOfferSnapshot> offerSnapshot(
      $grpc.ServiceCall call, $0.RequestOfferSnapshot request);
  $async.Future<$0.ResponseLoadSnapshotChunk> loadSnapshotChunk(
      $grpc.ServiceCall call, $0.RequestLoadSnapshotChunk request);
  $async.Future<$0.ResponseApplySnapshotChunk> applySnapshotChunk(
      $grpc.ServiceCall call, $0.RequestApplySnapshotChunk request);
}
