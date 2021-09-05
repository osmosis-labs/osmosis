///
//  Generated code. Do not modify.
//  source: osmosis/gamm/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $0;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$createPool =
      $grpc.ClientMethod<$0.MsgCreatePool, $0.MsgCreatePoolResponse>(
          '/osmosis.gamm.v1beta1.Msg/CreatePool',
          ($0.MsgCreatePool value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.MsgCreatePoolResponse.fromBuffer(value));
  static final _$joinPool =
      $grpc.ClientMethod<$0.MsgJoinPool, $0.MsgJoinPoolResponse>(
          '/osmosis.gamm.v1beta1.Msg/JoinPool',
          ($0.MsgJoinPool value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.MsgJoinPoolResponse.fromBuffer(value));
  static final _$exitPool =
      $grpc.ClientMethod<$0.MsgExitPool, $0.MsgExitPoolResponse>(
          '/osmosis.gamm.v1beta1.Msg/ExitPool',
          ($0.MsgExitPool value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.MsgExitPoolResponse.fromBuffer(value));
  static final _$swapExactAmountIn = $grpc.ClientMethod<$0.MsgSwapExactAmountIn,
          $0.MsgSwapExactAmountInResponse>(
      '/osmosis.gamm.v1beta1.Msg/SwapExactAmountIn',
      ($0.MsgSwapExactAmountIn value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.MsgSwapExactAmountInResponse.fromBuffer(value));
  static final _$swapExactAmountOut = $grpc.ClientMethod<
          $0.MsgSwapExactAmountOut, $0.MsgSwapExactAmountOutResponse>(
      '/osmosis.gamm.v1beta1.Msg/SwapExactAmountOut',
      ($0.MsgSwapExactAmountOut value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.MsgSwapExactAmountOutResponse.fromBuffer(value));
  static final _$joinSwapExternAmountIn = $grpc.ClientMethod<
          $0.MsgJoinSwapExternAmountIn, $0.MsgJoinSwapExternAmountInResponse>(
      '/osmosis.gamm.v1beta1.Msg/JoinSwapExternAmountIn',
      ($0.MsgJoinSwapExternAmountIn value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.MsgJoinSwapExternAmountInResponse.fromBuffer(value));
  static final _$joinSwapShareAmountOut = $grpc.ClientMethod<
          $0.MsgJoinSwapShareAmountOut, $0.MsgJoinSwapShareAmountOutResponse>(
      '/osmosis.gamm.v1beta1.Msg/JoinSwapShareAmountOut',
      ($0.MsgJoinSwapShareAmountOut value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.MsgJoinSwapShareAmountOutResponse.fromBuffer(value));
  static final _$exitSwapExternAmountOut = $grpc.ClientMethod<
          $0.MsgExitSwapExternAmountOut, $0.MsgExitSwapExternAmountOutResponse>(
      '/osmosis.gamm.v1beta1.Msg/ExitSwapExternAmountOut',
      ($0.MsgExitSwapExternAmountOut value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.MsgExitSwapExternAmountOutResponse.fromBuffer(value));
  static final _$exitSwapShareAmountIn = $grpc.ClientMethod<
          $0.MsgExitSwapShareAmountIn, $0.MsgExitSwapShareAmountInResponse>(
      '/osmosis.gamm.v1beta1.Msg/ExitSwapShareAmountIn',
      ($0.MsgExitSwapShareAmountIn value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.MsgExitSwapShareAmountInResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.MsgCreatePoolResponse> createPool(
      $0.MsgCreatePool request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$createPool, request, options: options);
  }

  $grpc.ResponseFuture<$0.MsgJoinPoolResponse> joinPool($0.MsgJoinPool request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$joinPool, request, options: options);
  }

  $grpc.ResponseFuture<$0.MsgExitPoolResponse> exitPool($0.MsgExitPool request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$exitPool, request, options: options);
  }

  $grpc.ResponseFuture<$0.MsgSwapExactAmountInResponse> swapExactAmountIn(
      $0.MsgSwapExactAmountIn request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$swapExactAmountIn, request, options: options);
  }

  $grpc.ResponseFuture<$0.MsgSwapExactAmountOutResponse> swapExactAmountOut(
      $0.MsgSwapExactAmountOut request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$swapExactAmountOut, request, options: options);
  }

  $grpc.ResponseFuture<$0.MsgJoinSwapExternAmountInResponse>
      joinSwapExternAmountIn($0.MsgJoinSwapExternAmountIn request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$joinSwapExternAmountIn, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.MsgJoinSwapShareAmountOutResponse>
      joinSwapShareAmountOut($0.MsgJoinSwapShareAmountOut request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$joinSwapShareAmountOut, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.MsgExitSwapExternAmountOutResponse>
      exitSwapExternAmountOut($0.MsgExitSwapExternAmountOut request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$exitSwapExternAmountOut, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.MsgExitSwapShareAmountInResponse>
      exitSwapShareAmountIn($0.MsgExitSwapShareAmountIn request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$exitSwapShareAmountIn, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'osmosis.gamm.v1beta1.Msg';

  MsgServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.MsgCreatePool, $0.MsgCreatePoolResponse>(
        'CreatePool',
        createPool_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.MsgCreatePool.fromBuffer(value),
        ($0.MsgCreatePoolResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MsgJoinPool, $0.MsgJoinPoolResponse>(
        'JoinPool',
        joinPool_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.MsgJoinPool.fromBuffer(value),
        ($0.MsgJoinPoolResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MsgExitPool, $0.MsgExitPoolResponse>(
        'ExitPool',
        exitPool_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.MsgExitPool.fromBuffer(value),
        ($0.MsgExitPoolResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MsgSwapExactAmountIn,
            $0.MsgSwapExactAmountInResponse>(
        'SwapExactAmountIn',
        swapExactAmountIn_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.MsgSwapExactAmountIn.fromBuffer(value),
        ($0.MsgSwapExactAmountInResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MsgSwapExactAmountOut,
            $0.MsgSwapExactAmountOutResponse>(
        'SwapExactAmountOut',
        swapExactAmountOut_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.MsgSwapExactAmountOut.fromBuffer(value),
        ($0.MsgSwapExactAmountOutResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MsgJoinSwapExternAmountIn,
            $0.MsgJoinSwapExternAmountInResponse>(
        'JoinSwapExternAmountIn',
        joinSwapExternAmountIn_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.MsgJoinSwapExternAmountIn.fromBuffer(value),
        ($0.MsgJoinSwapExternAmountInResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MsgJoinSwapShareAmountOut,
            $0.MsgJoinSwapShareAmountOutResponse>(
        'JoinSwapShareAmountOut',
        joinSwapShareAmountOut_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.MsgJoinSwapShareAmountOut.fromBuffer(value),
        ($0.MsgJoinSwapShareAmountOutResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MsgExitSwapExternAmountOut,
            $0.MsgExitSwapExternAmountOutResponse>(
        'ExitSwapExternAmountOut',
        exitSwapExternAmountOut_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.MsgExitSwapExternAmountOut.fromBuffer(value),
        ($0.MsgExitSwapExternAmountOutResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MsgExitSwapShareAmountIn,
            $0.MsgExitSwapShareAmountInResponse>(
        'ExitSwapShareAmountIn',
        exitSwapShareAmountIn_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.MsgExitSwapShareAmountIn.fromBuffer(value),
        ($0.MsgExitSwapShareAmountInResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.MsgCreatePoolResponse> createPool_Pre(
      $grpc.ServiceCall call, $async.Future<$0.MsgCreatePool> request) async {
    return createPool(call, await request);
  }

  $async.Future<$0.MsgJoinPoolResponse> joinPool_Pre(
      $grpc.ServiceCall call, $async.Future<$0.MsgJoinPool> request) async {
    return joinPool(call, await request);
  }

  $async.Future<$0.MsgExitPoolResponse> exitPool_Pre(
      $grpc.ServiceCall call, $async.Future<$0.MsgExitPool> request) async {
    return exitPool(call, await request);
  }

  $async.Future<$0.MsgSwapExactAmountInResponse> swapExactAmountIn_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.MsgSwapExactAmountIn> request) async {
    return swapExactAmountIn(call, await request);
  }

  $async.Future<$0.MsgSwapExactAmountOutResponse> swapExactAmountOut_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.MsgSwapExactAmountOut> request) async {
    return swapExactAmountOut(call, await request);
  }

  $async.Future<$0.MsgJoinSwapExternAmountInResponse>
      joinSwapExternAmountIn_Pre($grpc.ServiceCall call,
          $async.Future<$0.MsgJoinSwapExternAmountIn> request) async {
    return joinSwapExternAmountIn(call, await request);
  }

  $async.Future<$0.MsgJoinSwapShareAmountOutResponse>
      joinSwapShareAmountOut_Pre($grpc.ServiceCall call,
          $async.Future<$0.MsgJoinSwapShareAmountOut> request) async {
    return joinSwapShareAmountOut(call, await request);
  }

  $async.Future<$0.MsgExitSwapExternAmountOutResponse>
      exitSwapExternAmountOut_Pre($grpc.ServiceCall call,
          $async.Future<$0.MsgExitSwapExternAmountOut> request) async {
    return exitSwapExternAmountOut(call, await request);
  }

  $async.Future<$0.MsgExitSwapShareAmountInResponse> exitSwapShareAmountIn_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.MsgExitSwapShareAmountIn> request) async {
    return exitSwapShareAmountIn(call, await request);
  }

  $async.Future<$0.MsgCreatePoolResponse> createPool(
      $grpc.ServiceCall call, $0.MsgCreatePool request);
  $async.Future<$0.MsgJoinPoolResponse> joinPool(
      $grpc.ServiceCall call, $0.MsgJoinPool request);
  $async.Future<$0.MsgExitPoolResponse> exitPool(
      $grpc.ServiceCall call, $0.MsgExitPool request);
  $async.Future<$0.MsgSwapExactAmountInResponse> swapExactAmountIn(
      $grpc.ServiceCall call, $0.MsgSwapExactAmountIn request);
  $async.Future<$0.MsgSwapExactAmountOutResponse> swapExactAmountOut(
      $grpc.ServiceCall call, $0.MsgSwapExactAmountOut request);
  $async.Future<$0.MsgJoinSwapExternAmountInResponse> joinSwapExternAmountIn(
      $grpc.ServiceCall call, $0.MsgJoinSwapExternAmountIn request);
  $async.Future<$0.MsgJoinSwapShareAmountOutResponse> joinSwapShareAmountOut(
      $grpc.ServiceCall call, $0.MsgJoinSwapShareAmountOut request);
  $async.Future<$0.MsgExitSwapExternAmountOutResponse> exitSwapExternAmountOut(
      $grpc.ServiceCall call, $0.MsgExitSwapExternAmountOut request);
  $async.Future<$0.MsgExitSwapShareAmountInResponse> exitSwapShareAmountIn(
      $grpc.ServiceCall call, $0.MsgExitSwapShareAmountIn request);
}
