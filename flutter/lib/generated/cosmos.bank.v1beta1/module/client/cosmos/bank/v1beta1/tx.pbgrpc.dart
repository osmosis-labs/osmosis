///
//  Generated code. Do not modify.
//  source: cosmos/bank/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$send = $grpc.ClientMethod<$1.MsgSend, $1.MsgSendResponse>(
      '/cosmos.bank.v1beta1.Msg/Send',
      ($1.MsgSend value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.MsgSendResponse.fromBuffer(value));
  static final _$multiSend =
      $grpc.ClientMethod<$1.MsgMultiSend, $1.MsgMultiSendResponse>(
          '/cosmos.bank.v1beta1.Msg/MultiSend',
          ($1.MsgMultiSend value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgMultiSendResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgSendResponse> send($1.MsgSend request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$send, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgMultiSendResponse> multiSend(
      $1.MsgMultiSend request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$multiSend, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.bank.v1beta1.Msg';

  MsgServiceBase() {
    $addMethod($grpc.ServiceMethod<$1.MsgSend, $1.MsgSendResponse>(
        'Send',
        send_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgSend.fromBuffer(value),
        ($1.MsgSendResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgMultiSend, $1.MsgMultiSendResponse>(
        'MultiSend',
        multiSend_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgMultiSend.fromBuffer(value),
        ($1.MsgMultiSendResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgSendResponse> send_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgSend> request) async {
    return send(call, await request);
  }

  $async.Future<$1.MsgMultiSendResponse> multiSend_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgMultiSend> request) async {
    return multiSend(call, await request);
  }

  $async.Future<$1.MsgSendResponse> send(
      $grpc.ServiceCall call, $1.MsgSend request);
  $async.Future<$1.MsgMultiSendResponse> multiSend(
      $grpc.ServiceCall call, $1.MsgMultiSend request);
}
