///
//  Generated code. Do not modify.
//  source: osmosis/lockup/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'query.pb.dart' as $0;
export 'query.pb.dart';

class QueryClient extends $grpc.Client {
  static final _$moduleBalance =
      $grpc.ClientMethod<$0.ModuleBalanceRequest, $0.ModuleBalanceResponse>(
          '/osmosis.lockup.Query/ModuleBalance',
          ($0.ModuleBalanceRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.ModuleBalanceResponse.fromBuffer(value));
  static final _$moduleLockedAmount = $grpc.ClientMethod<
          $0.ModuleLockedAmountRequest, $0.ModuleLockedAmountResponse>(
      '/osmosis.lockup.Query/ModuleLockedAmount',
      ($0.ModuleLockedAmountRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.ModuleLockedAmountResponse.fromBuffer(value));
  static final _$accountUnlockableCoins = $grpc.ClientMethod<
          $0.AccountUnlockableCoinsRequest, $0.AccountUnlockableCoinsResponse>(
      '/osmosis.lockup.Query/AccountUnlockableCoins',
      ($0.AccountUnlockableCoinsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.AccountUnlockableCoinsResponse.fromBuffer(value));
  static final _$accountUnlockingCoins = $grpc.ClientMethod<
          $0.AccountUnlockingCoinsRequest, $0.AccountUnlockingCoinsResponse>(
      '/osmosis.lockup.Query/AccountUnlockingCoins',
      ($0.AccountUnlockingCoinsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.AccountUnlockingCoinsResponse.fromBuffer(value));
  static final _$accountLockedCoins = $grpc.ClientMethod<
          $0.AccountLockedCoinsRequest, $0.AccountLockedCoinsResponse>(
      '/osmosis.lockup.Query/AccountLockedCoins',
      ($0.AccountLockedCoinsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.AccountLockedCoinsResponse.fromBuffer(value));
  static final _$accountLockedPastTime = $grpc.ClientMethod<
          $0.AccountLockedPastTimeRequest, $0.AccountLockedPastTimeResponse>(
      '/osmosis.lockup.Query/AccountLockedPastTime',
      ($0.AccountLockedPastTimeRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.AccountLockedPastTimeResponse.fromBuffer(value));
  static final _$accountLockedPastTimeNotUnlockingOnly = $grpc.ClientMethod<
          $0.AccountLockedPastTimeNotUnlockingOnlyRequest,
          $0.AccountLockedPastTimeNotUnlockingOnlyResponse>(
      '/osmosis.lockup.Query/AccountLockedPastTimeNotUnlockingOnly',
      ($0.AccountLockedPastTimeNotUnlockingOnlyRequest value) =>
          value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.AccountLockedPastTimeNotUnlockingOnlyResponse.fromBuffer(value));
  static final _$accountUnlockedBeforeTime = $grpc.ClientMethod<
          $0.AccountUnlockedBeforeTimeRequest,
          $0.AccountUnlockedBeforeTimeResponse>(
      '/osmosis.lockup.Query/AccountUnlockedBeforeTime',
      ($0.AccountUnlockedBeforeTimeRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.AccountUnlockedBeforeTimeResponse.fromBuffer(value));
  static final _$accountLockedPastTimeDenom = $grpc.ClientMethod<
          $0.AccountLockedPastTimeDenomRequest,
          $0.AccountLockedPastTimeDenomResponse>(
      '/osmosis.lockup.Query/AccountLockedPastTimeDenom',
      ($0.AccountLockedPastTimeDenomRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.AccountLockedPastTimeDenomResponse.fromBuffer(value));
  static final _$lockedDenom =
      $grpc.ClientMethod<$0.LockedDenomRequest, $0.LockedDenomResponse>(
          '/osmosis.lockup.Query/LockedDenom',
          ($0.LockedDenomRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.LockedDenomResponse.fromBuffer(value));
  static final _$lockedByID =
      $grpc.ClientMethod<$0.LockedRequest, $0.LockedResponse>(
          '/osmosis.lockup.Query/LockedByID',
          ($0.LockedRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $0.LockedResponse.fromBuffer(value));
  static final _$accountLockedLongerDuration = $grpc.ClientMethod<
          $0.AccountLockedLongerDurationRequest,
          $0.AccountLockedLongerDurationResponse>(
      '/osmosis.lockup.Query/AccountLockedLongerDuration',
      ($0.AccountLockedLongerDurationRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.AccountLockedLongerDurationResponse.fromBuffer(value));
  static final _$accountLockedLongerDurationNotUnlockingOnly =
      $grpc.ClientMethod<$0.AccountLockedLongerDurationNotUnlockingOnlyRequest,
              $0.AccountLockedLongerDurationNotUnlockingOnlyResponse>(
          '/osmosis.lockup.Query/AccountLockedLongerDurationNotUnlockingOnly',
          ($0.AccountLockedLongerDurationNotUnlockingOnlyRequest value) =>
              value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.AccountLockedLongerDurationNotUnlockingOnlyResponse.fromBuffer(
                  value));
  static final _$accountLockedLongerDurationDenom = $grpc.ClientMethod<
          $0.AccountLockedLongerDurationDenomRequest,
          $0.AccountLockedLongerDurationDenomResponse>(
      '/osmosis.lockup.Query/AccountLockedLongerDurationDenom',
      ($0.AccountLockedLongerDurationDenomRequest value) =>
          value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.AccountLockedLongerDurationDenomResponse.fromBuffer(value));

  QueryClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.ModuleBalanceResponse> moduleBalance(
      $0.ModuleBalanceRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$moduleBalance, request, options: options);
  }

  $grpc.ResponseFuture<$0.ModuleLockedAmountResponse> moduleLockedAmount(
      $0.ModuleLockedAmountRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$moduleLockedAmount, request, options: options);
  }

  $grpc.ResponseFuture<$0.AccountUnlockableCoinsResponse>
      accountUnlockableCoins($0.AccountUnlockableCoinsRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$accountUnlockableCoins, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.AccountUnlockingCoinsResponse> accountUnlockingCoins(
      $0.AccountUnlockingCoinsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$accountUnlockingCoins, request, options: options);
  }

  $grpc.ResponseFuture<$0.AccountLockedCoinsResponse> accountLockedCoins(
      $0.AccountLockedCoinsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$accountLockedCoins, request, options: options);
  }

  $grpc.ResponseFuture<$0.AccountLockedPastTimeResponse> accountLockedPastTime(
      $0.AccountLockedPastTimeRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$accountLockedPastTime, request, options: options);
  }

  $grpc.ResponseFuture<$0.AccountLockedPastTimeNotUnlockingOnlyResponse>
      accountLockedPastTimeNotUnlockingOnly(
          $0.AccountLockedPastTimeNotUnlockingOnlyRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$accountLockedPastTimeNotUnlockingOnly, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.AccountUnlockedBeforeTimeResponse>
      accountUnlockedBeforeTime($0.AccountUnlockedBeforeTimeRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$accountUnlockedBeforeTime, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.AccountLockedPastTimeDenomResponse>
      accountLockedPastTimeDenom($0.AccountLockedPastTimeDenomRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$accountLockedPastTimeDenom, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.LockedDenomResponse> lockedDenom(
      $0.LockedDenomRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$lockedDenom, request, options: options);
  }

  $grpc.ResponseFuture<$0.LockedResponse> lockedByID($0.LockedRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$lockedByID, request, options: options);
  }

  $grpc.ResponseFuture<$0.AccountLockedLongerDurationResponse>
      accountLockedLongerDuration($0.AccountLockedLongerDurationRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$accountLockedLongerDuration, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.AccountLockedLongerDurationNotUnlockingOnlyResponse>
      accountLockedLongerDurationNotUnlockingOnly(
          $0.AccountLockedLongerDurationNotUnlockingOnlyRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(
        _$accountLockedLongerDurationNotUnlockingOnly, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.AccountLockedLongerDurationDenomResponse>
      accountLockedLongerDurationDenom(
          $0.AccountLockedLongerDurationDenomRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$accountLockedLongerDurationDenom, request,
        options: options);
  }
}

abstract class QueryServiceBase extends $grpc.Service {
  $core.String get $name => 'osmosis.lockup.Query';

  QueryServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$0.ModuleBalanceRequest, $0.ModuleBalanceResponse>(
            'ModuleBalance',
            moduleBalance_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.ModuleBalanceRequest.fromBuffer(value),
            ($0.ModuleBalanceResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ModuleLockedAmountRequest,
            $0.ModuleLockedAmountResponse>(
        'ModuleLockedAmount',
        moduleLockedAmount_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.ModuleLockedAmountRequest.fromBuffer(value),
        ($0.ModuleLockedAmountResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AccountUnlockableCoinsRequest,
            $0.AccountUnlockableCoinsResponse>(
        'AccountUnlockableCoins',
        accountUnlockableCoins_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.AccountUnlockableCoinsRequest.fromBuffer(value),
        ($0.AccountUnlockableCoinsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AccountUnlockingCoinsRequest,
            $0.AccountUnlockingCoinsResponse>(
        'AccountUnlockingCoins',
        accountUnlockingCoins_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.AccountUnlockingCoinsRequest.fromBuffer(value),
        ($0.AccountUnlockingCoinsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AccountLockedCoinsRequest,
            $0.AccountLockedCoinsResponse>(
        'AccountLockedCoins',
        accountLockedCoins_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.AccountLockedCoinsRequest.fromBuffer(value),
        ($0.AccountLockedCoinsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AccountLockedPastTimeRequest,
            $0.AccountLockedPastTimeResponse>(
        'AccountLockedPastTime',
        accountLockedPastTime_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.AccountLockedPastTimeRequest.fromBuffer(value),
        ($0.AccountLockedPastTimeResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<
            $0.AccountLockedPastTimeNotUnlockingOnlyRequest,
            $0.AccountLockedPastTimeNotUnlockingOnlyResponse>(
        'AccountLockedPastTimeNotUnlockingOnly',
        accountLockedPastTimeNotUnlockingOnly_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.AccountLockedPastTimeNotUnlockingOnlyRequest.fromBuffer(value),
        ($0.AccountLockedPastTimeNotUnlockingOnlyResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AccountUnlockedBeforeTimeRequest,
            $0.AccountUnlockedBeforeTimeResponse>(
        'AccountUnlockedBeforeTime',
        accountUnlockedBeforeTime_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.AccountUnlockedBeforeTimeRequest.fromBuffer(value),
        ($0.AccountUnlockedBeforeTimeResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AccountLockedPastTimeDenomRequest,
            $0.AccountLockedPastTimeDenomResponse>(
        'AccountLockedPastTimeDenom',
        accountLockedPastTimeDenom_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.AccountLockedPastTimeDenomRequest.fromBuffer(value),
        ($0.AccountLockedPastTimeDenomResponse value) =>
            value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.LockedDenomRequest, $0.LockedDenomResponse>(
            'LockedDenom',
            lockedDenom_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.LockedDenomRequest.fromBuffer(value),
            ($0.LockedDenomResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.LockedRequest, $0.LockedResponse>(
        'LockedByID',
        lockedByID_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.LockedRequest.fromBuffer(value),
        ($0.LockedResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AccountLockedLongerDurationRequest,
            $0.AccountLockedLongerDurationResponse>(
        'AccountLockedLongerDuration',
        accountLockedLongerDuration_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.AccountLockedLongerDurationRequest.fromBuffer(value),
        ($0.AccountLockedLongerDurationResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<
            $0.AccountLockedLongerDurationNotUnlockingOnlyRequest,
            $0.AccountLockedLongerDurationNotUnlockingOnlyResponse>(
        'AccountLockedLongerDurationNotUnlockingOnly',
        accountLockedLongerDurationNotUnlockingOnly_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.AccountLockedLongerDurationNotUnlockingOnlyRequest.fromBuffer(
                value),
        ($0.AccountLockedLongerDurationNotUnlockingOnlyResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AccountLockedLongerDurationDenomRequest,
            $0.AccountLockedLongerDurationDenomResponse>(
        'AccountLockedLongerDurationDenom',
        accountLockedLongerDurationDenom_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.AccountLockedLongerDurationDenomRequest.fromBuffer(value),
        ($0.AccountLockedLongerDurationDenomResponse value) =>
            value.writeToBuffer()));
  }

  $async.Future<$0.ModuleBalanceResponse> moduleBalance_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.ModuleBalanceRequest> request) async {
    return moduleBalance(call, await request);
  }

  $async.Future<$0.ModuleLockedAmountResponse> moduleLockedAmount_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.ModuleLockedAmountRequest> request) async {
    return moduleLockedAmount(call, await request);
  }

  $async.Future<$0.AccountUnlockableCoinsResponse> accountUnlockableCoins_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.AccountUnlockableCoinsRequest> request) async {
    return accountUnlockableCoins(call, await request);
  }

  $async.Future<$0.AccountUnlockingCoinsResponse> accountUnlockingCoins_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.AccountUnlockingCoinsRequest> request) async {
    return accountUnlockingCoins(call, await request);
  }

  $async.Future<$0.AccountLockedCoinsResponse> accountLockedCoins_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.AccountLockedCoinsRequest> request) async {
    return accountLockedCoins(call, await request);
  }

  $async.Future<$0.AccountLockedPastTimeResponse> accountLockedPastTime_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.AccountLockedPastTimeRequest> request) async {
    return accountLockedPastTime(call, await request);
  }

  $async.Future<$0.AccountLockedPastTimeNotUnlockingOnlyResponse>
      accountLockedPastTimeNotUnlockingOnly_Pre(
          $grpc.ServiceCall call,
          $async.Future<$0.AccountLockedPastTimeNotUnlockingOnlyRequest>
              request) async {
    return accountLockedPastTimeNotUnlockingOnly(call, await request);
  }

  $async.Future<$0.AccountUnlockedBeforeTimeResponse>
      accountUnlockedBeforeTime_Pre($grpc.ServiceCall call,
          $async.Future<$0.AccountUnlockedBeforeTimeRequest> request) async {
    return accountUnlockedBeforeTime(call, await request);
  }

  $async.Future<$0.AccountLockedPastTimeDenomResponse>
      accountLockedPastTimeDenom_Pre($grpc.ServiceCall call,
          $async.Future<$0.AccountLockedPastTimeDenomRequest> request) async {
    return accountLockedPastTimeDenom(call, await request);
  }

  $async.Future<$0.LockedDenomResponse> lockedDenom_Pre($grpc.ServiceCall call,
      $async.Future<$0.LockedDenomRequest> request) async {
    return lockedDenom(call, await request);
  }

  $async.Future<$0.LockedResponse> lockedByID_Pre(
      $grpc.ServiceCall call, $async.Future<$0.LockedRequest> request) async {
    return lockedByID(call, await request);
  }

  $async.Future<$0.AccountLockedLongerDurationResponse>
      accountLockedLongerDuration_Pre($grpc.ServiceCall call,
          $async.Future<$0.AccountLockedLongerDurationRequest> request) async {
    return accountLockedLongerDuration(call, await request);
  }

  $async.Future<$0.AccountLockedLongerDurationNotUnlockingOnlyResponse>
      accountLockedLongerDurationNotUnlockingOnly_Pre(
          $grpc.ServiceCall call,
          $async.Future<$0.AccountLockedLongerDurationNotUnlockingOnlyRequest>
              request) async {
    return accountLockedLongerDurationNotUnlockingOnly(call, await request);
  }

  $async.Future<$0.AccountLockedLongerDurationDenomResponse>
      accountLockedLongerDurationDenom_Pre(
          $grpc.ServiceCall call,
          $async.Future<$0.AccountLockedLongerDurationDenomRequest>
              request) async {
    return accountLockedLongerDurationDenom(call, await request);
  }

  $async.Future<$0.ModuleBalanceResponse> moduleBalance(
      $grpc.ServiceCall call, $0.ModuleBalanceRequest request);
  $async.Future<$0.ModuleLockedAmountResponse> moduleLockedAmount(
      $grpc.ServiceCall call, $0.ModuleLockedAmountRequest request);
  $async.Future<$0.AccountUnlockableCoinsResponse> accountUnlockableCoins(
      $grpc.ServiceCall call, $0.AccountUnlockableCoinsRequest request);
  $async.Future<$0.AccountUnlockingCoinsResponse> accountUnlockingCoins(
      $grpc.ServiceCall call, $0.AccountUnlockingCoinsRequest request);
  $async.Future<$0.AccountLockedCoinsResponse> accountLockedCoins(
      $grpc.ServiceCall call, $0.AccountLockedCoinsRequest request);
  $async.Future<$0.AccountLockedPastTimeResponse> accountLockedPastTime(
      $grpc.ServiceCall call, $0.AccountLockedPastTimeRequest request);
  $async.Future<$0.AccountLockedPastTimeNotUnlockingOnlyResponse>
      accountLockedPastTimeNotUnlockingOnly($grpc.ServiceCall call,
          $0.AccountLockedPastTimeNotUnlockingOnlyRequest request);
  $async.Future<$0.AccountUnlockedBeforeTimeResponse> accountUnlockedBeforeTime(
      $grpc.ServiceCall call, $0.AccountUnlockedBeforeTimeRequest request);
  $async.Future<$0.AccountLockedPastTimeDenomResponse>
      accountLockedPastTimeDenom(
          $grpc.ServiceCall call, $0.AccountLockedPastTimeDenomRequest request);
  $async.Future<$0.LockedDenomResponse> lockedDenom(
      $grpc.ServiceCall call, $0.LockedDenomRequest request);
  $async.Future<$0.LockedResponse> lockedByID(
      $grpc.ServiceCall call, $0.LockedRequest request);
  $async.Future<$0.AccountLockedLongerDurationResponse>
      accountLockedLongerDuration($grpc.ServiceCall call,
          $0.AccountLockedLongerDurationRequest request);
  $async.Future<$0.AccountLockedLongerDurationNotUnlockingOnlyResponse>
      accountLockedLongerDurationNotUnlockingOnly($grpc.ServiceCall call,
          $0.AccountLockedLongerDurationNotUnlockingOnlyRequest request);
  $async.Future<$0.AccountLockedLongerDurationDenomResponse>
      accountLockedLongerDurationDenom($grpc.ServiceCall call,
          $0.AccountLockedLongerDurationDenomRequest request);
}
