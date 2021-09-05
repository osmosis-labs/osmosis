///
//  Generated code. Do not modify.
//  source: osmosis/lockup/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$lockTokens =
      $grpc.ClientMethod<$1.MsgLockTokens, $1.MsgLockTokensResponse>(
          '/osmosis.lockup.Msg/LockTokens',
          ($1.MsgLockTokens value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgLockTokensResponse.fromBuffer(value));
  static final _$beginUnlockingAll = $grpc.ClientMethod<$1.MsgBeginUnlockingAll,
          $1.MsgBeginUnlockingAllResponse>(
      '/osmosis.lockup.Msg/BeginUnlockingAll',
      ($1.MsgBeginUnlockingAll value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgBeginUnlockingAllResponse.fromBuffer(value));
  static final _$beginUnlocking =
      $grpc.ClientMethod<$1.MsgBeginUnlocking, $1.MsgBeginUnlockingResponse>(
          '/osmosis.lockup.Msg/BeginUnlocking',
          ($1.MsgBeginUnlocking value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgBeginUnlockingResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgLockTokensResponse> lockTokens(
      $1.MsgLockTokens request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$lockTokens, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgBeginUnlockingAllResponse> beginUnlockingAll(
      $1.MsgBeginUnlockingAll request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$beginUnlockingAll, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgBeginUnlockingResponse> beginUnlocking(
      $1.MsgBeginUnlocking request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$beginUnlocking, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'osmosis.lockup.Msg';

  MsgServiceBase() {
    $addMethod($grpc.ServiceMethod<$1.MsgLockTokens, $1.MsgLockTokensResponse>(
        'LockTokens',
        lockTokens_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgLockTokens.fromBuffer(value),
        ($1.MsgLockTokensResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgBeginUnlockingAll,
            $1.MsgBeginUnlockingAllResponse>(
        'BeginUnlockingAll',
        beginUnlockingAll_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgBeginUnlockingAll.fromBuffer(value),
        ($1.MsgBeginUnlockingAllResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.MsgBeginUnlocking, $1.MsgBeginUnlockingResponse>(
            'BeginUnlocking',
            beginUnlocking_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.MsgBeginUnlocking.fromBuffer(value),
            ($1.MsgBeginUnlockingResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgLockTokensResponse> lockTokens_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgLockTokens> request) async {
    return lockTokens(call, await request);
  }

  $async.Future<$1.MsgBeginUnlockingAllResponse> beginUnlockingAll_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgBeginUnlockingAll> request) async {
    return beginUnlockingAll(call, await request);
  }

  $async.Future<$1.MsgBeginUnlockingResponse> beginUnlocking_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgBeginUnlocking> request) async {
    return beginUnlocking(call, await request);
  }

  $async.Future<$1.MsgLockTokensResponse> lockTokens(
      $grpc.ServiceCall call, $1.MsgLockTokens request);
  $async.Future<$1.MsgBeginUnlockingAllResponse> beginUnlockingAll(
      $grpc.ServiceCall call, $1.MsgBeginUnlockingAll request);
  $async.Future<$1.MsgBeginUnlockingResponse> beginUnlocking(
      $grpc.ServiceCall call, $1.MsgBeginUnlocking request);
}
