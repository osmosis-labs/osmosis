///
//  Generated code. Do not modify.
//  source: ibc/core/connection/v1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$connectionOpenInit = $grpc.ClientMethod<
          $1.MsgConnectionOpenInit, $1.MsgConnectionOpenInitResponse>(
      '/ibc.core.connection.v1.Msg/ConnectionOpenInit',
      ($1.MsgConnectionOpenInit value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgConnectionOpenInitResponse.fromBuffer(value));
  static final _$connectionOpenTry = $grpc.ClientMethod<$1.MsgConnectionOpenTry,
          $1.MsgConnectionOpenTryResponse>(
      '/ibc.core.connection.v1.Msg/ConnectionOpenTry',
      ($1.MsgConnectionOpenTry value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgConnectionOpenTryResponse.fromBuffer(value));
  static final _$connectionOpenAck = $grpc.ClientMethod<$1.MsgConnectionOpenAck,
          $1.MsgConnectionOpenAckResponse>(
      '/ibc.core.connection.v1.Msg/ConnectionOpenAck',
      ($1.MsgConnectionOpenAck value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgConnectionOpenAckResponse.fromBuffer(value));
  static final _$connectionOpenConfirm = $grpc.ClientMethod<
          $1.MsgConnectionOpenConfirm, $1.MsgConnectionOpenConfirmResponse>(
      '/ibc.core.connection.v1.Msg/ConnectionOpenConfirm',
      ($1.MsgConnectionOpenConfirm value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgConnectionOpenConfirmResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgConnectionOpenInitResponse> connectionOpenInit(
      $1.MsgConnectionOpenInit request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$connectionOpenInit, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgConnectionOpenTryResponse> connectionOpenTry(
      $1.MsgConnectionOpenTry request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$connectionOpenTry, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgConnectionOpenAckResponse> connectionOpenAck(
      $1.MsgConnectionOpenAck request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$connectionOpenAck, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgConnectionOpenConfirmResponse>
      connectionOpenConfirm($1.MsgConnectionOpenConfirm request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$connectionOpenConfirm, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'ibc.core.connection.v1.Msg';

  MsgServiceBase() {
    $addMethod($grpc.ServiceMethod<$1.MsgConnectionOpenInit,
            $1.MsgConnectionOpenInitResponse>(
        'ConnectionOpenInit',
        connectionOpenInit_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgConnectionOpenInit.fromBuffer(value),
        ($1.MsgConnectionOpenInitResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgConnectionOpenTry,
            $1.MsgConnectionOpenTryResponse>(
        'ConnectionOpenTry',
        connectionOpenTry_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgConnectionOpenTry.fromBuffer(value),
        ($1.MsgConnectionOpenTryResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgConnectionOpenAck,
            $1.MsgConnectionOpenAckResponse>(
        'ConnectionOpenAck',
        connectionOpenAck_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgConnectionOpenAck.fromBuffer(value),
        ($1.MsgConnectionOpenAckResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgConnectionOpenConfirm,
            $1.MsgConnectionOpenConfirmResponse>(
        'ConnectionOpenConfirm',
        connectionOpenConfirm_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgConnectionOpenConfirm.fromBuffer(value),
        ($1.MsgConnectionOpenConfirmResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgConnectionOpenInitResponse> connectionOpenInit_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgConnectionOpenInit> request) async {
    return connectionOpenInit(call, await request);
  }

  $async.Future<$1.MsgConnectionOpenTryResponse> connectionOpenTry_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgConnectionOpenTry> request) async {
    return connectionOpenTry(call, await request);
  }

  $async.Future<$1.MsgConnectionOpenAckResponse> connectionOpenAck_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgConnectionOpenAck> request) async {
    return connectionOpenAck(call, await request);
  }

  $async.Future<$1.MsgConnectionOpenConfirmResponse> connectionOpenConfirm_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgConnectionOpenConfirm> request) async {
    return connectionOpenConfirm(call, await request);
  }

  $async.Future<$1.MsgConnectionOpenInitResponse> connectionOpenInit(
      $grpc.ServiceCall call, $1.MsgConnectionOpenInit request);
  $async.Future<$1.MsgConnectionOpenTryResponse> connectionOpenTry(
      $grpc.ServiceCall call, $1.MsgConnectionOpenTry request);
  $async.Future<$1.MsgConnectionOpenAckResponse> connectionOpenAck(
      $grpc.ServiceCall call, $1.MsgConnectionOpenAck request);
  $async.Future<$1.MsgConnectionOpenConfirmResponse> connectionOpenConfirm(
      $grpc.ServiceCall call, $1.MsgConnectionOpenConfirm request);
}
