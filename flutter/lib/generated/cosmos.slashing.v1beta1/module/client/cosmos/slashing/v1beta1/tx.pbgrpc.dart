///
//  Generated code. Do not modify.
//  source: cosmos/slashing/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$unjail =
      $grpc.ClientMethod<$1.MsgUnjail, $1.MsgUnjailResponse>(
          '/cosmos.slashing.v1beta1.Msg/Unjail',
          ($1.MsgUnjail value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgUnjailResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgUnjailResponse> unjail($1.MsgUnjail request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$unjail, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.slashing.v1beta1.Msg';

  MsgServiceBase() {
    $addMethod($grpc.ServiceMethod<$1.MsgUnjail, $1.MsgUnjailResponse>(
        'Unjail',
        unjail_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgUnjail.fromBuffer(value),
        ($1.MsgUnjailResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgUnjailResponse> unjail_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgUnjail> request) async {
    return unjail(call, await request);
  }

  $async.Future<$1.MsgUnjailResponse> unjail(
      $grpc.ServiceCall call, $1.MsgUnjail request);
}
