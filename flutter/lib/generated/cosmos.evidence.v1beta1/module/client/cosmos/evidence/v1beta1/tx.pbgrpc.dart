///
//  Generated code. Do not modify.
//  source: cosmos/evidence/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$submitEvidence =
      $grpc.ClientMethod<$1.MsgSubmitEvidence, $1.MsgSubmitEvidenceResponse>(
          '/cosmos.evidence.v1beta1.Msg/SubmitEvidence',
          ($1.MsgSubmitEvidence value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgSubmitEvidenceResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgSubmitEvidenceResponse> submitEvidence(
      $1.MsgSubmitEvidence request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$submitEvidence, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.evidence.v1beta1.Msg';

  MsgServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$1.MsgSubmitEvidence, $1.MsgSubmitEvidenceResponse>(
            'SubmitEvidence',
            submitEvidence_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.MsgSubmitEvidence.fromBuffer(value),
            ($1.MsgSubmitEvidenceResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgSubmitEvidenceResponse> submitEvidence_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgSubmitEvidence> request) async {
    return submitEvidence(call, await request);
  }

  $async.Future<$1.MsgSubmitEvidenceResponse> submitEvidence(
      $grpc.ServiceCall call, $1.MsgSubmitEvidence request);
}
