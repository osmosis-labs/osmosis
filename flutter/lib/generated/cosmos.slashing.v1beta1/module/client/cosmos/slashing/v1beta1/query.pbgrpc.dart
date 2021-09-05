///
//  Generated code. Do not modify.
//  source: cosmos/slashing/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'query.pb.dart' as $0;
export 'query.pb.dart';

class QueryClient extends $grpc.Client {
  static final _$params =
      $grpc.ClientMethod<$0.QueryParamsRequest, $0.QueryParamsResponse>(
          '/cosmos.slashing.v1beta1.Query/Params',
          ($0.QueryParamsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryParamsResponse.fromBuffer(value));
  static final _$signingInfo = $grpc.ClientMethod<$0.QuerySigningInfoRequest,
          $0.QuerySigningInfoResponse>(
      '/cosmos.slashing.v1beta1.Query/SigningInfo',
      ($0.QuerySigningInfoRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QuerySigningInfoResponse.fromBuffer(value));
  static final _$signingInfos = $grpc.ClientMethod<$0.QuerySigningInfosRequest,
          $0.QuerySigningInfosResponse>(
      '/cosmos.slashing.v1beta1.Query/SigningInfos',
      ($0.QuerySigningInfosRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QuerySigningInfosResponse.fromBuffer(value));

  QueryClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.QueryParamsResponse> params(
      $0.QueryParamsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$params, request, options: options);
  }

  $grpc.ResponseFuture<$0.QuerySigningInfoResponse> signingInfo(
      $0.QuerySigningInfoRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$signingInfo, request, options: options);
  }

  $grpc.ResponseFuture<$0.QuerySigningInfosResponse> signingInfos(
      $0.QuerySigningInfosRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$signingInfos, request, options: options);
  }
}

abstract class QueryServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.slashing.v1beta1.Query';

  QueryServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$0.QueryParamsRequest, $0.QueryParamsResponse>(
            'Params',
            params_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.QueryParamsRequest.fromBuffer(value),
            ($0.QueryParamsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QuerySigningInfoRequest,
            $0.QuerySigningInfoResponse>(
        'SigningInfo',
        signingInfo_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QuerySigningInfoRequest.fromBuffer(value),
        ($0.QuerySigningInfoResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QuerySigningInfosRequest,
            $0.QuerySigningInfosResponse>(
        'SigningInfos',
        signingInfos_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QuerySigningInfosRequest.fromBuffer(value),
        ($0.QuerySigningInfosResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.QueryParamsResponse> params_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryParamsRequest> request) async {
    return params(call, await request);
  }

  $async.Future<$0.QuerySigningInfoResponse> signingInfo_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QuerySigningInfoRequest> request) async {
    return signingInfo(call, await request);
  }

  $async.Future<$0.QuerySigningInfosResponse> signingInfos_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QuerySigningInfosRequest> request) async {
    return signingInfos(call, await request);
  }

  $async.Future<$0.QueryParamsResponse> params(
      $grpc.ServiceCall call, $0.QueryParamsRequest request);
  $async.Future<$0.QuerySigningInfoResponse> signingInfo(
      $grpc.ServiceCall call, $0.QuerySigningInfoRequest request);
  $async.Future<$0.QuerySigningInfosResponse> signingInfos(
      $grpc.ServiceCall call, $0.QuerySigningInfosRequest request);
}
