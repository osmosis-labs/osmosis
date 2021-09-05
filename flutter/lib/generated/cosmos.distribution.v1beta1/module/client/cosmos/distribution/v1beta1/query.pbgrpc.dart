///
//  Generated code. Do not modify.
//  source: cosmos/distribution/v1beta1/query.proto
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
          '/cosmos.distribution.v1beta1.Query/Params',
          ($0.QueryParamsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryParamsResponse.fromBuffer(value));
  static final _$validatorOutstandingRewards = $grpc.ClientMethod<
          $0.QueryValidatorOutstandingRewardsRequest,
          $0.QueryValidatorOutstandingRewardsResponse>(
      '/cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards',
      ($0.QueryValidatorOutstandingRewardsRequest value) =>
          value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryValidatorOutstandingRewardsResponse.fromBuffer(value));
  static final _$validatorCommission = $grpc.ClientMethod<
          $0.QueryValidatorCommissionRequest,
          $0.QueryValidatorCommissionResponse>(
      '/cosmos.distribution.v1beta1.Query/ValidatorCommission',
      ($0.QueryValidatorCommissionRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryValidatorCommissionResponse.fromBuffer(value));
  static final _$validatorSlashes = $grpc.ClientMethod<
          $0.QueryValidatorSlashesRequest, $0.QueryValidatorSlashesResponse>(
      '/cosmos.distribution.v1beta1.Query/ValidatorSlashes',
      ($0.QueryValidatorSlashesRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryValidatorSlashesResponse.fromBuffer(value));
  static final _$delegationRewards = $grpc.ClientMethod<
          $0.QueryDelegationRewardsRequest, $0.QueryDelegationRewardsResponse>(
      '/cosmos.distribution.v1beta1.Query/DelegationRewards',
      ($0.QueryDelegationRewardsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryDelegationRewardsResponse.fromBuffer(value));
  static final _$delegationTotalRewards = $grpc.ClientMethod<
          $0.QueryDelegationTotalRewardsRequest,
          $0.QueryDelegationTotalRewardsResponse>(
      '/cosmos.distribution.v1beta1.Query/DelegationTotalRewards',
      ($0.QueryDelegationTotalRewardsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryDelegationTotalRewardsResponse.fromBuffer(value));
  static final _$delegatorValidators = $grpc.ClientMethod<
          $0.QueryDelegatorValidatorsRequest,
          $0.QueryDelegatorValidatorsResponse>(
      '/cosmos.distribution.v1beta1.Query/DelegatorValidators',
      ($0.QueryDelegatorValidatorsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryDelegatorValidatorsResponse.fromBuffer(value));
  static final _$delegatorWithdrawAddress = $grpc.ClientMethod<
          $0.QueryDelegatorWithdrawAddressRequest,
          $0.QueryDelegatorWithdrawAddressResponse>(
      '/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress',
      ($0.QueryDelegatorWithdrawAddressRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryDelegatorWithdrawAddressResponse.fromBuffer(value));
  static final _$communityPool = $grpc.ClientMethod<
          $0.QueryCommunityPoolRequest, $0.QueryCommunityPoolResponse>(
      '/cosmos.distribution.v1beta1.Query/CommunityPool',
      ($0.QueryCommunityPoolRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryCommunityPoolResponse.fromBuffer(value));

  QueryClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.QueryParamsResponse> params(
      $0.QueryParamsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$params, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryValidatorOutstandingRewardsResponse>
      validatorOutstandingRewards(
          $0.QueryValidatorOutstandingRewardsRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$validatorOutstandingRewards, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.QueryValidatorCommissionResponse> validatorCommission(
      $0.QueryValidatorCommissionRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$validatorCommission, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryValidatorSlashesResponse> validatorSlashes(
      $0.QueryValidatorSlashesRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$validatorSlashes, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryDelegationRewardsResponse> delegationRewards(
      $0.QueryDelegationRewardsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$delegationRewards, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryDelegationTotalRewardsResponse>
      delegationTotalRewards($0.QueryDelegationTotalRewardsRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$delegationTotalRewards, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.QueryDelegatorValidatorsResponse> delegatorValidators(
      $0.QueryDelegatorValidatorsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$delegatorValidators, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryDelegatorWithdrawAddressResponse>
      delegatorWithdrawAddress($0.QueryDelegatorWithdrawAddressRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$delegatorWithdrawAddress, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.QueryCommunityPoolResponse> communityPool(
      $0.QueryCommunityPoolRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$communityPool, request, options: options);
  }
}

abstract class QueryServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.distribution.v1beta1.Query';

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
    $addMethod($grpc.ServiceMethod<$0.QueryValidatorOutstandingRewardsRequest,
            $0.QueryValidatorOutstandingRewardsResponse>(
        'ValidatorOutstandingRewards',
        validatorOutstandingRewards_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryValidatorOutstandingRewardsRequest.fromBuffer(value),
        ($0.QueryValidatorOutstandingRewardsResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryValidatorCommissionRequest,
            $0.QueryValidatorCommissionResponse>(
        'ValidatorCommission',
        validatorCommission_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryValidatorCommissionRequest.fromBuffer(value),
        ($0.QueryValidatorCommissionResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryValidatorSlashesRequest,
            $0.QueryValidatorSlashesResponse>(
        'ValidatorSlashes',
        validatorSlashes_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryValidatorSlashesRequest.fromBuffer(value),
        ($0.QueryValidatorSlashesResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryDelegationRewardsRequest,
            $0.QueryDelegationRewardsResponse>(
        'DelegationRewards',
        delegationRewards_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryDelegationRewardsRequest.fromBuffer(value),
        ($0.QueryDelegationRewardsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryDelegationTotalRewardsRequest,
            $0.QueryDelegationTotalRewardsResponse>(
        'DelegationTotalRewards',
        delegationTotalRewards_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryDelegationTotalRewardsRequest.fromBuffer(value),
        ($0.QueryDelegationTotalRewardsResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryDelegatorValidatorsRequest,
            $0.QueryDelegatorValidatorsResponse>(
        'DelegatorValidators',
        delegatorValidators_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryDelegatorValidatorsRequest.fromBuffer(value),
        ($0.QueryDelegatorValidatorsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryDelegatorWithdrawAddressRequest,
            $0.QueryDelegatorWithdrawAddressResponse>(
        'DelegatorWithdrawAddress',
        delegatorWithdrawAddress_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryDelegatorWithdrawAddressRequest.fromBuffer(value),
        ($0.QueryDelegatorWithdrawAddressResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryCommunityPoolRequest,
            $0.QueryCommunityPoolResponse>(
        'CommunityPool',
        communityPool_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryCommunityPoolRequest.fromBuffer(value),
        ($0.QueryCommunityPoolResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.QueryParamsResponse> params_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryParamsRequest> request) async {
    return params(call, await request);
  }

  $async.Future<$0.QueryValidatorOutstandingRewardsResponse>
      validatorOutstandingRewards_Pre(
          $grpc.ServiceCall call,
          $async.Future<$0.QueryValidatorOutstandingRewardsRequest>
              request) async {
    return validatorOutstandingRewards(call, await request);
  }

  $async.Future<$0.QueryValidatorCommissionResponse> validatorCommission_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryValidatorCommissionRequest> request) async {
    return validatorCommission(call, await request);
  }

  $async.Future<$0.QueryValidatorSlashesResponse> validatorSlashes_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryValidatorSlashesRequest> request) async {
    return validatorSlashes(call, await request);
  }

  $async.Future<$0.QueryDelegationRewardsResponse> delegationRewards_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryDelegationRewardsRequest> request) async {
    return delegationRewards(call, await request);
  }

  $async.Future<$0.QueryDelegationTotalRewardsResponse>
      delegationTotalRewards_Pre($grpc.ServiceCall call,
          $async.Future<$0.QueryDelegationTotalRewardsRequest> request) async {
    return delegationTotalRewards(call, await request);
  }

  $async.Future<$0.QueryDelegatorValidatorsResponse> delegatorValidators_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryDelegatorValidatorsRequest> request) async {
    return delegatorValidators(call, await request);
  }

  $async.Future<$0.QueryDelegatorWithdrawAddressResponse>
      delegatorWithdrawAddress_Pre(
          $grpc.ServiceCall call,
          $async.Future<$0.QueryDelegatorWithdrawAddressRequest>
              request) async {
    return delegatorWithdrawAddress(call, await request);
  }

  $async.Future<$0.QueryCommunityPoolResponse> communityPool_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryCommunityPoolRequest> request) async {
    return communityPool(call, await request);
  }

  $async.Future<$0.QueryParamsResponse> params(
      $grpc.ServiceCall call, $0.QueryParamsRequest request);
  $async.Future<$0.QueryValidatorOutstandingRewardsResponse>
      validatorOutstandingRewards($grpc.ServiceCall call,
          $0.QueryValidatorOutstandingRewardsRequest request);
  $async.Future<$0.QueryValidatorCommissionResponse> validatorCommission(
      $grpc.ServiceCall call, $0.QueryValidatorCommissionRequest request);
  $async.Future<$0.QueryValidatorSlashesResponse> validatorSlashes(
      $grpc.ServiceCall call, $0.QueryValidatorSlashesRequest request);
  $async.Future<$0.QueryDelegationRewardsResponse> delegationRewards(
      $grpc.ServiceCall call, $0.QueryDelegationRewardsRequest request);
  $async.Future<$0.QueryDelegationTotalRewardsResponse> delegationTotalRewards(
      $grpc.ServiceCall call, $0.QueryDelegationTotalRewardsRequest request);
  $async.Future<$0.QueryDelegatorValidatorsResponse> delegatorValidators(
      $grpc.ServiceCall call, $0.QueryDelegatorValidatorsRequest request);
  $async.Future<$0.QueryDelegatorWithdrawAddressResponse>
      delegatorWithdrawAddress($grpc.ServiceCall call,
          $0.QueryDelegatorWithdrawAddressRequest request);
  $async.Future<$0.QueryCommunityPoolResponse> communityPool(
      $grpc.ServiceCall call, $0.QueryCommunityPoolRequest request);
}
