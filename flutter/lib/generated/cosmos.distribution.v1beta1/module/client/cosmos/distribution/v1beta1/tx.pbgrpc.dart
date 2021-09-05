///
//  Generated code. Do not modify.
//  source: cosmos/distribution/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$setWithdrawAddress = $grpc.ClientMethod<
          $1.MsgSetWithdrawAddress, $1.MsgSetWithdrawAddressResponse>(
      '/cosmos.distribution.v1beta1.Msg/SetWithdrawAddress',
      ($1.MsgSetWithdrawAddress value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgSetWithdrawAddressResponse.fromBuffer(value));
  static final _$withdrawDelegatorReward = $grpc.ClientMethod<
          $1.MsgWithdrawDelegatorReward, $1.MsgWithdrawDelegatorRewardResponse>(
      '/cosmos.distribution.v1beta1.Msg/WithdrawDelegatorReward',
      ($1.MsgWithdrawDelegatorReward value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgWithdrawDelegatorRewardResponse.fromBuffer(value));
  static final _$withdrawValidatorCommission = $grpc.ClientMethod<
          $1.MsgWithdrawValidatorCommission,
          $1.MsgWithdrawValidatorCommissionResponse>(
      '/cosmos.distribution.v1beta1.Msg/WithdrawValidatorCommission',
      ($1.MsgWithdrawValidatorCommission value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgWithdrawValidatorCommissionResponse.fromBuffer(value));
  static final _$fundCommunityPool = $grpc.ClientMethod<$1.MsgFundCommunityPool,
          $1.MsgFundCommunityPoolResponse>(
      '/cosmos.distribution.v1beta1.Msg/FundCommunityPool',
      ($1.MsgFundCommunityPool value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgFundCommunityPoolResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgSetWithdrawAddressResponse> setWithdrawAddress(
      $1.MsgSetWithdrawAddress request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$setWithdrawAddress, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgWithdrawDelegatorRewardResponse>
      withdrawDelegatorReward($1.MsgWithdrawDelegatorReward request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$withdrawDelegatorReward, request,
        options: options);
  }

  $grpc.ResponseFuture<$1.MsgWithdrawValidatorCommissionResponse>
      withdrawValidatorCommission($1.MsgWithdrawValidatorCommission request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$withdrawValidatorCommission, request,
        options: options);
  }

  $grpc.ResponseFuture<$1.MsgFundCommunityPoolResponse> fundCommunityPool(
      $1.MsgFundCommunityPool request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$fundCommunityPool, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.distribution.v1beta1.Msg';

  MsgServiceBase() {
    $addMethod($grpc.ServiceMethod<$1.MsgSetWithdrawAddress,
            $1.MsgSetWithdrawAddressResponse>(
        'SetWithdrawAddress',
        setWithdrawAddress_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgSetWithdrawAddress.fromBuffer(value),
        ($1.MsgSetWithdrawAddressResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgWithdrawDelegatorReward,
            $1.MsgWithdrawDelegatorRewardResponse>(
        'WithdrawDelegatorReward',
        withdrawDelegatorReward_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgWithdrawDelegatorReward.fromBuffer(value),
        ($1.MsgWithdrawDelegatorRewardResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgWithdrawValidatorCommission,
            $1.MsgWithdrawValidatorCommissionResponse>(
        'WithdrawValidatorCommission',
        withdrawValidatorCommission_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgWithdrawValidatorCommission.fromBuffer(value),
        ($1.MsgWithdrawValidatorCommissionResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgFundCommunityPool,
            $1.MsgFundCommunityPoolResponse>(
        'FundCommunityPool',
        fundCommunityPool_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgFundCommunityPool.fromBuffer(value),
        ($1.MsgFundCommunityPoolResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgSetWithdrawAddressResponse> setWithdrawAddress_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgSetWithdrawAddress> request) async {
    return setWithdrawAddress(call, await request);
  }

  $async.Future<$1.MsgWithdrawDelegatorRewardResponse>
      withdrawDelegatorReward_Pre($grpc.ServiceCall call,
          $async.Future<$1.MsgWithdrawDelegatorReward> request) async {
    return withdrawDelegatorReward(call, await request);
  }

  $async.Future<$1.MsgWithdrawValidatorCommissionResponse>
      withdrawValidatorCommission_Pre($grpc.ServiceCall call,
          $async.Future<$1.MsgWithdrawValidatorCommission> request) async {
    return withdrawValidatorCommission(call, await request);
  }

  $async.Future<$1.MsgFundCommunityPoolResponse> fundCommunityPool_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgFundCommunityPool> request) async {
    return fundCommunityPool(call, await request);
  }

  $async.Future<$1.MsgSetWithdrawAddressResponse> setWithdrawAddress(
      $grpc.ServiceCall call, $1.MsgSetWithdrawAddress request);
  $async.Future<$1.MsgWithdrawDelegatorRewardResponse> withdrawDelegatorReward(
      $grpc.ServiceCall call, $1.MsgWithdrawDelegatorReward request);
  $async.Future<$1.MsgWithdrawValidatorCommissionResponse>
      withdrawValidatorCommission(
          $grpc.ServiceCall call, $1.MsgWithdrawValidatorCommission request);
  $async.Future<$1.MsgFundCommunityPoolResponse> fundCommunityPool(
      $grpc.ServiceCall call, $1.MsgFundCommunityPool request);
}
