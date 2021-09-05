///
//  Generated code. Do not modify.
//  source: cosmos/crisis/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $0;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$verifyInvariant =
      $grpc.ClientMethod<$0.MsgVerifyInvariant, $0.MsgVerifyInvariantResponse>(
          '/cosmos.crisis.v1beta1.Msg/VerifyInvariant',
          ($0.MsgVerifyInvariant value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.MsgVerifyInvariantResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.MsgVerifyInvariantResponse> verifyInvariant(
      $0.MsgVerifyInvariant request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$verifyInvariant, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.crisis.v1beta1.Msg';

  MsgServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.MsgVerifyInvariant,
            $0.MsgVerifyInvariantResponse>(
        'VerifyInvariant',
        verifyInvariant_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.MsgVerifyInvariant.fromBuffer(value),
        ($0.MsgVerifyInvariantResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.MsgVerifyInvariantResponse> verifyInvariant_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.MsgVerifyInvariant> request) async {
    return verifyInvariant(call, await request);
  }

  $async.Future<$0.MsgVerifyInvariantResponse> verifyInvariant(
      $grpc.ServiceCall call, $0.MsgVerifyInvariant request);
}
