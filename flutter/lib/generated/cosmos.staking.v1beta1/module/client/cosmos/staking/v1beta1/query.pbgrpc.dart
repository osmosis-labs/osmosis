///
//  Generated code. Do not modify.
//  source: cosmos/staking/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'query.pb.dart' as $0;
export 'query.pb.dart';

class QueryClient extends $grpc.Client {
  static final _$validators =
      $grpc.ClientMethod<$0.QueryValidatorsRequest, $0.QueryValidatorsResponse>(
          '/cosmos.staking.v1beta1.Query/Validators',
          ($0.QueryValidatorsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryValidatorsResponse.fromBuffer(value));
  static final _$validator =
      $grpc.ClientMethod<$0.QueryValidatorRequest, $0.QueryValidatorResponse>(
          '/cosmos.staking.v1beta1.Query/Validator',
          ($0.QueryValidatorRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryValidatorResponse.fromBuffer(value));
  static final _$validatorDelegations = $grpc.ClientMethod<
          $0.QueryValidatorDelegationsRequest,
          $0.QueryValidatorDelegationsResponse>(
      '/cosmos.staking.v1beta1.Query/ValidatorDelegations',
      ($0.QueryValidatorDelegationsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryValidatorDelegationsResponse.fromBuffer(value));
  static final _$validatorUnbondingDelegations = $grpc.ClientMethod<
          $0.QueryValidatorUnbondingDelegationsRequest,
          $0.QueryValidatorUnbondingDelegationsResponse>(
      '/cosmos.staking.v1beta1.Query/ValidatorUnbondingDelegations',
      ($0.QueryValidatorUnbondingDelegationsRequest value) =>
          value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryValidatorUnbondingDelegationsResponse.fromBuffer(value));
  static final _$delegation =
      $grpc.ClientMethod<$0.QueryDelegationRequest, $0.QueryDelegationResponse>(
          '/cosmos.staking.v1beta1.Query/Delegation',
          ($0.QueryDelegationRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryDelegationResponse.fromBuffer(value));
  static final _$unbondingDelegation = $grpc.ClientMethod<
          $0.QueryUnbondingDelegationRequest,
          $0.QueryUnbondingDelegationResponse>(
      '/cosmos.staking.v1beta1.Query/UnbondingDelegation',
      ($0.QueryUnbondingDelegationRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryUnbondingDelegationResponse.fromBuffer(value));
  static final _$delegatorDelegations = $grpc.ClientMethod<
          $0.QueryDelegatorDelegationsRequest,
          $0.QueryDelegatorDelegationsResponse>(
      '/cosmos.staking.v1beta1.Query/DelegatorDelegations',
      ($0.QueryDelegatorDelegationsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryDelegatorDelegationsResponse.fromBuffer(value));
  static final _$delegatorUnbondingDelegations = $grpc.ClientMethod<
          $0.QueryDelegatorUnbondingDelegationsRequest,
          $0.QueryDelegatorUnbondingDelegationsResponse>(
      '/cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations',
      ($0.QueryDelegatorUnbondingDelegationsRequest value) =>
          value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryDelegatorUnbondingDelegationsResponse.fromBuffer(value));
  static final _$redelegations = $grpc.ClientMethod<
          $0.QueryRedelegationsRequest, $0.QueryRedelegationsResponse>(
      '/cosmos.staking.v1beta1.Query/Redelegations',
      ($0.QueryRedelegationsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryRedelegationsResponse.fromBuffer(value));
  static final _$delegatorValidators = $grpc.ClientMethod<
          $0.QueryDelegatorValidatorsRequest,
          $0.QueryDelegatorValidatorsResponse>(
      '/cosmos.staking.v1beta1.Query/DelegatorValidators',
      ($0.QueryDelegatorValidatorsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryDelegatorValidatorsResponse.fromBuffer(value));
  static final _$delegatorValidator = $grpc.ClientMethod<
          $0.QueryDelegatorValidatorRequest,
          $0.QueryDelegatorValidatorResponse>(
      '/cosmos.staking.v1beta1.Query/DelegatorValidator',
      ($0.QueryDelegatorValidatorRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryDelegatorValidatorResponse.fromBuffer(value));
  static final _$historicalInfo = $grpc.ClientMethod<
          $0.QueryHistoricalInfoRequest, $0.QueryHistoricalInfoResponse>(
      '/cosmos.staking.v1beta1.Query/HistoricalInfo',
      ($0.QueryHistoricalInfoRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryHistoricalInfoResponse.fromBuffer(value));
  static final _$pool =
      $grpc.ClientMethod<$0.QueryPoolRequest, $0.QueryPoolResponse>(
          '/cosmos.staking.v1beta1.Query/Pool',
          ($0.QueryPoolRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryPoolResponse.fromBuffer(value));
  static final _$params =
      $grpc.ClientMethod<$0.QueryParamsRequest, $0.QueryParamsResponse>(
          '/cosmos.staking.v1beta1.Query/Params',
          ($0.QueryParamsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryParamsResponse.fromBuffer(value));

  QueryClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.QueryValidatorsResponse> validators(
      $0.QueryValidatorsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$validators, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryValidatorResponse> validator(
      $0.QueryValidatorRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$validator, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryValidatorDelegationsResponse>
      validatorDelegations($0.QueryValidatorDelegationsRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$validatorDelegations, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryValidatorUnbondingDelegationsResponse>
      validatorUnbondingDelegations(
          $0.QueryValidatorUnbondingDelegationsRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$validatorUnbondingDelegations, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.QueryDelegationResponse> delegation(
      $0.QueryDelegationRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$delegation, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryUnbondingDelegationResponse> unbondingDelegation(
      $0.QueryUnbondingDelegationRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$unbondingDelegation, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryDelegatorDelegationsResponse>
      delegatorDelegations($0.QueryDelegatorDelegationsRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$delegatorDelegations, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryDelegatorUnbondingDelegationsResponse>
      delegatorUnbondingDelegations(
          $0.QueryDelegatorUnbondingDelegationsRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$delegatorUnbondingDelegations, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.QueryRedelegationsResponse> redelegations(
      $0.QueryRedelegationsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$redelegations, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryDelegatorValidatorsResponse> delegatorValidators(
      $0.QueryDelegatorValidatorsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$delegatorValidators, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryDelegatorValidatorResponse> delegatorValidator(
      $0.QueryDelegatorValidatorRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$delegatorValidator, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryHistoricalInfoResponse> historicalInfo(
      $0.QueryHistoricalInfoRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$historicalInfo, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryPoolResponse> pool($0.QueryPoolRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$pool, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryParamsResponse> params(
      $0.QueryParamsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$params, request, options: options);
  }
}

abstract class QueryServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.staking.v1beta1.Query';

  QueryServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.QueryValidatorsRequest,
            $0.QueryValidatorsResponse>(
        'Validators',
        validators_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryValidatorsRequest.fromBuffer(value),
        ($0.QueryValidatorsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryValidatorRequest,
            $0.QueryValidatorResponse>(
        'Validator',
        validator_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryValidatorRequest.fromBuffer(value),
        ($0.QueryValidatorResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryValidatorDelegationsRequest,
            $0.QueryValidatorDelegationsResponse>(
        'ValidatorDelegations',
        validatorDelegations_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryValidatorDelegationsRequest.fromBuffer(value),
        ($0.QueryValidatorDelegationsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryValidatorUnbondingDelegationsRequest,
            $0.QueryValidatorUnbondingDelegationsResponse>(
        'ValidatorUnbondingDelegations',
        validatorUnbondingDelegations_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryValidatorUnbondingDelegationsRequest.fromBuffer(value),
        ($0.QueryValidatorUnbondingDelegationsResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryDelegationRequest,
            $0.QueryDelegationResponse>(
        'Delegation',
        delegation_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryDelegationRequest.fromBuffer(value),
        ($0.QueryDelegationResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryUnbondingDelegationRequest,
            $0.QueryUnbondingDelegationResponse>(
        'UnbondingDelegation',
        unbondingDelegation_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryUnbondingDelegationRequest.fromBuffer(value),
        ($0.QueryUnbondingDelegationResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryDelegatorDelegationsRequest,
            $0.QueryDelegatorDelegationsResponse>(
        'DelegatorDelegations',
        delegatorDelegations_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryDelegatorDelegationsRequest.fromBuffer(value),
        ($0.QueryDelegatorDelegationsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryDelegatorUnbondingDelegationsRequest,
            $0.QueryDelegatorUnbondingDelegationsResponse>(
        'DelegatorUnbondingDelegations',
        delegatorUnbondingDelegations_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryDelegatorUnbondingDelegationsRequest.fromBuffer(value),
        ($0.QueryDelegatorUnbondingDelegationsResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryRedelegationsRequest,
            $0.QueryRedelegationsResponse>(
        'Redelegations',
        redelegations_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryRedelegationsRequest.fromBuffer(value),
        ($0.QueryRedelegationsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryDelegatorValidatorsRequest,
            $0.QueryDelegatorValidatorsResponse>(
        'DelegatorValidators',
        delegatorValidators_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryDelegatorValidatorsRequest.fromBuffer(value),
        ($0.QueryDelegatorValidatorsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryDelegatorValidatorRequest,
            $0.QueryDelegatorValidatorResponse>(
        'DelegatorValidator',
        delegatorValidator_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryDelegatorValidatorRequest.fromBuffer(value),
        ($0.QueryDelegatorValidatorResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryHistoricalInfoRequest,
            $0.QueryHistoricalInfoResponse>(
        'HistoricalInfo',
        historicalInfo_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryHistoricalInfoRequest.fromBuffer(value),
        ($0.QueryHistoricalInfoResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryPoolRequest, $0.QueryPoolResponse>(
        'Pool',
        pool_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.QueryPoolRequest.fromBuffer(value),
        ($0.QueryPoolResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.QueryParamsRequest, $0.QueryParamsResponse>(
            'Params',
            params_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.QueryParamsRequest.fromBuffer(value),
            ($0.QueryParamsResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.QueryValidatorsResponse> validators_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryValidatorsRequest> request) async {
    return validators(call, await request);
  }

  $async.Future<$0.QueryValidatorResponse> validator_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryValidatorRequest> request) async {
    return validator(call, await request);
  }

  $async.Future<$0.QueryValidatorDelegationsResponse> validatorDelegations_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryValidatorDelegationsRequest> request) async {
    return validatorDelegations(call, await request);
  }

  $async.Future<$0.QueryValidatorUnbondingDelegationsResponse>
      validatorUnbondingDelegations_Pre(
          $grpc.ServiceCall call,
          $async.Future<$0.QueryValidatorUnbondingDelegationsRequest>
              request) async {
    return validatorUnbondingDelegations(call, await request);
  }

  $async.Future<$0.QueryDelegationResponse> delegation_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryDelegationRequest> request) async {
    return delegation(call, await request);
  }

  $async.Future<$0.QueryUnbondingDelegationResponse> unbondingDelegation_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryUnbondingDelegationRequest> request) async {
    return unbondingDelegation(call, await request);
  }

  $async.Future<$0.QueryDelegatorDelegationsResponse> delegatorDelegations_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryDelegatorDelegationsRequest> request) async {
    return delegatorDelegations(call, await request);
  }

  $async.Future<$0.QueryDelegatorUnbondingDelegationsResponse>
      delegatorUnbondingDelegations_Pre(
          $grpc.ServiceCall call,
          $async.Future<$0.QueryDelegatorUnbondingDelegationsRequest>
              request) async {
    return delegatorUnbondingDelegations(call, await request);
  }

  $async.Future<$0.QueryRedelegationsResponse> redelegations_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryRedelegationsRequest> request) async {
    return redelegations(call, await request);
  }

  $async.Future<$0.QueryDelegatorValidatorsResponse> delegatorValidators_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryDelegatorValidatorsRequest> request) async {
    return delegatorValidators(call, await request);
  }

  $async.Future<$0.QueryDelegatorValidatorResponse> delegatorValidator_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryDelegatorValidatorRequest> request) async {
    return delegatorValidator(call, await request);
  }

  $async.Future<$0.QueryHistoricalInfoResponse> historicalInfo_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryHistoricalInfoRequest> request) async {
    return historicalInfo(call, await request);
  }

  $async.Future<$0.QueryPoolResponse> pool_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryPoolRequest> request) async {
    return pool(call, await request);
  }

  $async.Future<$0.QueryParamsResponse> params_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryParamsRequest> request) async {
    return params(call, await request);
  }

  $async.Future<$0.QueryValidatorsResponse> validators(
      $grpc.ServiceCall call, $0.QueryValidatorsRequest request);
  $async.Future<$0.QueryValidatorResponse> validator(
      $grpc.ServiceCall call, $0.QueryValidatorRequest request);
  $async.Future<$0.QueryValidatorDelegationsResponse> validatorDelegations(
      $grpc.ServiceCall call, $0.QueryValidatorDelegationsRequest request);
  $async.Future<$0.QueryValidatorUnbondingDelegationsResponse>
      validatorUnbondingDelegations($grpc.ServiceCall call,
          $0.QueryValidatorUnbondingDelegationsRequest request);
  $async.Future<$0.QueryDelegationResponse> delegation(
      $grpc.ServiceCall call, $0.QueryDelegationRequest request);
  $async.Future<$0.QueryUnbondingDelegationResponse> unbondingDelegation(
      $grpc.ServiceCall call, $0.QueryUnbondingDelegationRequest request);
  $async.Future<$0.QueryDelegatorDelegationsResponse> delegatorDelegations(
      $grpc.ServiceCall call, $0.QueryDelegatorDelegationsRequest request);
  $async.Future<$0.QueryDelegatorUnbondingDelegationsResponse>
      delegatorUnbondingDelegations($grpc.ServiceCall call,
          $0.QueryDelegatorUnbondingDelegationsRequest request);
  $async.Future<$0.QueryRedelegationsResponse> redelegations(
      $grpc.ServiceCall call, $0.QueryRedelegationsRequest request);
  $async.Future<$0.QueryDelegatorValidatorsResponse> delegatorValidators(
      $grpc.ServiceCall call, $0.QueryDelegatorValidatorsRequest request);
  $async.Future<$0.QueryDelegatorValidatorResponse> delegatorValidator(
      $grpc.ServiceCall call, $0.QueryDelegatorValidatorRequest request);
  $async.Future<$0.QueryHistoricalInfoResponse> historicalInfo(
      $grpc.ServiceCall call, $0.QueryHistoricalInfoRequest request);
  $async.Future<$0.QueryPoolResponse> pool(
      $grpc.ServiceCall call, $0.QueryPoolRequest request);
  $async.Future<$0.QueryParamsResponse> params(
      $grpc.ServiceCall call, $0.QueryParamsRequest request);
}
