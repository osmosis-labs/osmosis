///
//  Generated code. Do not modify.
//  source: cosmos/evidence/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'query.pb.dart' as $0;
export 'query.pb.dart';

class QueryClient extends $grpc.Client {
  static final _$evidence =
      $grpc.ClientMethod<$0.QueryEvidenceRequest, $0.QueryEvidenceResponse>(
          '/cosmos.evidence.v1beta1.Query/Evidence',
          ($0.QueryEvidenceRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryEvidenceResponse.fromBuffer(value));
  static final _$allEvidence = $grpc.ClientMethod<$0.QueryAllEvidenceRequest,
          $0.QueryAllEvidenceResponse>(
      '/cosmos.evidence.v1beta1.Query/AllEvidence',
      ($0.QueryAllEvidenceRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryAllEvidenceResponse.fromBuffer(value));

  QueryClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.QueryEvidenceResponse> evidence(
      $0.QueryEvidenceRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$evidence, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryAllEvidenceResponse> allEvidence(
      $0.QueryAllEvidenceRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$allEvidence, request, options: options);
  }
}

abstract class QueryServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.evidence.v1beta1.Query';

  QueryServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$0.QueryEvidenceRequest, $0.QueryEvidenceResponse>(
            'Evidence',
            evidence_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.QueryEvidenceRequest.fromBuffer(value),
            ($0.QueryEvidenceResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryAllEvidenceRequest,
            $0.QueryAllEvidenceResponse>(
        'AllEvidence',
        allEvidence_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryAllEvidenceRequest.fromBuffer(value),
        ($0.QueryAllEvidenceResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.QueryEvidenceResponse> evidence_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryEvidenceRequest> request) async {
    return evidence(call, await request);
  }

  $async.Future<$0.QueryAllEvidenceResponse> allEvidence_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryAllEvidenceRequest> request) async {
    return allEvidence(call, await request);
  }

  $async.Future<$0.QueryEvidenceResponse> evidence(
      $grpc.ServiceCall call, $0.QueryEvidenceRequest request);
  $async.Future<$0.QueryAllEvidenceResponse> allEvidence(
      $grpc.ServiceCall call, $0.QueryAllEvidenceRequest request);
}
