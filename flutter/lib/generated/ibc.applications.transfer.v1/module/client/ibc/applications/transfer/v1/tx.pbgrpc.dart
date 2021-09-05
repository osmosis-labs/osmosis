///
//  Generated code. Do not modify.
//  source: ibc/applications/transfer/v1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$transfer =
      $grpc.ClientMethod<$1.MsgTransfer, $1.MsgTransferResponse>(
          '/ibc.applications.transfer.v1.Msg/Transfer',
          ($1.MsgTransfer value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgTransferResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgTransferResponse> transfer($1.MsgTransfer request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$transfer, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'ibc.applications.transfer.v1.Msg';

  MsgServiceBase() {
    $addMethod($grpc.ServiceMethod<$1.MsgTransfer, $1.MsgTransferResponse>(
        'Transfer',
        transfer_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgTransfer.fromBuffer(value),
        ($1.MsgTransferResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgTransferResponse> transfer_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgTransfer> request) async {
    return transfer(call, await request);
  }

  $async.Future<$1.MsgTransferResponse> transfer(
      $grpc.ServiceCall call, $1.MsgTransfer request);
}
