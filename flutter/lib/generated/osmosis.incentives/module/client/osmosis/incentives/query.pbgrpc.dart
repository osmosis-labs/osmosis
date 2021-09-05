///
//  Generated code. Do not modify.
//  source: osmosis/incentives/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'query.pb.dart' as $0;
export 'query.pb.dart';

class QueryClient extends $grpc.Client {
  static final _$moduleToDistributeCoins = $grpc.ClientMethod<
          $0.ModuleToDistributeCoinsRequest,
          $0.ModuleToDistributeCoinsResponse>(
      '/osmosis.incentives.Query/ModuleToDistributeCoins',
      ($0.ModuleToDistributeCoinsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.ModuleToDistributeCoinsResponse.fromBuffer(value));
  static final _$moduleDistributedCoins = $grpc.ClientMethod<
          $0.ModuleDistributedCoinsRequest, $0.ModuleDistributedCoinsResponse>(
      '/osmosis.incentives.Query/ModuleDistributedCoins',
      ($0.ModuleDistributedCoinsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.ModuleDistributedCoinsResponse.fromBuffer(value));
  static final _$gaugeByID =
      $grpc.ClientMethod<$0.GaugeByIDRequest, $0.GaugeByIDResponse>(
          '/osmosis.incentives.Query/GaugeByID',
          ($0.GaugeByIDRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.GaugeByIDResponse.fromBuffer(value));
  static final _$gauges =
      $grpc.ClientMethod<$0.GaugesRequest, $0.GaugesResponse>(
          '/osmosis.incentives.Query/Gauges',
          ($0.GaugesRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $0.GaugesResponse.fromBuffer(value));
  static final _$activeGauges =
      $grpc.ClientMethod<$0.ActiveGaugesRequest, $0.ActiveGaugesResponse>(
          '/osmosis.incentives.Query/ActiveGauges',
          ($0.ActiveGaugesRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.ActiveGaugesResponse.fromBuffer(value));
  static final _$upcomingGauges =
      $grpc.ClientMethod<$0.UpcomingGaugesRequest, $0.UpcomingGaugesResponse>(
          '/osmosis.incentives.Query/UpcomingGauges',
          ($0.UpcomingGaugesRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.UpcomingGaugesResponse.fromBuffer(value));
  static final _$rewardsEst =
      $grpc.ClientMethod<$0.RewardsEstRequest, $0.RewardsEstResponse>(
          '/osmosis.incentives.Query/RewardsEst',
          ($0.RewardsEstRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.RewardsEstResponse.fromBuffer(value));
  static final _$lockableDurations = $grpc.ClientMethod<
          $0.QueryLockableDurationsRequest, $0.QueryLockableDurationsResponse>(
      '/osmosis.incentives.Query/LockableDurations',
      ($0.QueryLockableDurationsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryLockableDurationsResponse.fromBuffer(value));

  QueryClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.ModuleToDistributeCoinsResponse>
      moduleToDistributeCoins($0.ModuleToDistributeCoinsRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$moduleToDistributeCoins, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.ModuleDistributedCoinsResponse>
      moduleDistributedCoins($0.ModuleDistributedCoinsRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$moduleDistributedCoins, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.GaugeByIDResponse> gaugeByID(
      $0.GaugeByIDRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$gaugeByID, request, options: options);
  }

  $grpc.ResponseFuture<$0.GaugesResponse> gauges($0.GaugesRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$gauges, request, options: options);
  }

  $grpc.ResponseFuture<$0.ActiveGaugesResponse> activeGauges(
      $0.ActiveGaugesRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$activeGauges, request, options: options);
  }

  $grpc.ResponseFuture<$0.UpcomingGaugesResponse> upcomingGauges(
      $0.UpcomingGaugesRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$upcomingGauges, request, options: options);
  }

  $grpc.ResponseFuture<$0.RewardsEstResponse> rewardsEst(
      $0.RewardsEstRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$rewardsEst, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryLockableDurationsResponse> lockableDurations(
      $0.QueryLockableDurationsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$lockableDurations, request, options: options);
  }
}

abstract class QueryServiceBase extends $grpc.Service {
  $core.String get $name => 'osmosis.incentives.Query';

  QueryServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.ModuleToDistributeCoinsRequest,
            $0.ModuleToDistributeCoinsResponse>(
        'ModuleToDistributeCoins',
        moduleToDistributeCoins_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.ModuleToDistributeCoinsRequest.fromBuffer(value),
        ($0.ModuleToDistributeCoinsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ModuleDistributedCoinsRequest,
            $0.ModuleDistributedCoinsResponse>(
        'ModuleDistributedCoins',
        moduleDistributedCoins_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.ModuleDistributedCoinsRequest.fromBuffer(value),
        ($0.ModuleDistributedCoinsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.GaugeByIDRequest, $0.GaugeByIDResponse>(
        'GaugeByID',
        gaugeByID_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.GaugeByIDRequest.fromBuffer(value),
        ($0.GaugeByIDResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.GaugesRequest, $0.GaugesResponse>(
        'Gauges',
        gauges_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.GaugesRequest.fromBuffer(value),
        ($0.GaugesResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.ActiveGaugesRequest, $0.ActiveGaugesResponse>(
            'ActiveGauges',
            activeGauges_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.ActiveGaugesRequest.fromBuffer(value),
            ($0.ActiveGaugesResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UpcomingGaugesRequest,
            $0.UpcomingGaugesResponse>(
        'UpcomingGauges',
        upcomingGauges_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.UpcomingGaugesRequest.fromBuffer(value),
        ($0.UpcomingGaugesResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RewardsEstRequest, $0.RewardsEstResponse>(
        'RewardsEst',
        rewardsEst_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RewardsEstRequest.fromBuffer(value),
        ($0.RewardsEstResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryLockableDurationsRequest,
            $0.QueryLockableDurationsResponse>(
        'LockableDurations',
        lockableDurations_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryLockableDurationsRequest.fromBuffer(value),
        ($0.QueryLockableDurationsResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.ModuleToDistributeCoinsResponse> moduleToDistributeCoins_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.ModuleToDistributeCoinsRequest> request) async {
    return moduleToDistributeCoins(call, await request);
  }

  $async.Future<$0.ModuleDistributedCoinsResponse> moduleDistributedCoins_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.ModuleDistributedCoinsRequest> request) async {
    return moduleDistributedCoins(call, await request);
  }

  $async.Future<$0.GaugeByIDResponse> gaugeByID_Pre($grpc.ServiceCall call,
      $async.Future<$0.GaugeByIDRequest> request) async {
    return gaugeByID(call, await request);
  }

  $async.Future<$0.GaugesResponse> gauges_Pre(
      $grpc.ServiceCall call, $async.Future<$0.GaugesRequest> request) async {
    return gauges(call, await request);
  }

  $async.Future<$0.ActiveGaugesResponse> activeGauges_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.ActiveGaugesRequest> request) async {
    return activeGauges(call, await request);
  }

  $async.Future<$0.UpcomingGaugesResponse> upcomingGauges_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.UpcomingGaugesRequest> request) async {
    return upcomingGauges(call, await request);
  }

  $async.Future<$0.RewardsEstResponse> rewardsEst_Pre($grpc.ServiceCall call,
      $async.Future<$0.RewardsEstRequest> request) async {
    return rewardsEst(call, await request);
  }

  $async.Future<$0.QueryLockableDurationsResponse> lockableDurations_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryLockableDurationsRequest> request) async {
    return lockableDurations(call, await request);
  }

  $async.Future<$0.ModuleToDistributeCoinsResponse> moduleToDistributeCoins(
      $grpc.ServiceCall call, $0.ModuleToDistributeCoinsRequest request);
  $async.Future<$0.ModuleDistributedCoinsResponse> moduleDistributedCoins(
      $grpc.ServiceCall call, $0.ModuleDistributedCoinsRequest request);
  $async.Future<$0.GaugeByIDResponse> gaugeByID(
      $grpc.ServiceCall call, $0.GaugeByIDRequest request);
  $async.Future<$0.GaugesResponse> gauges(
      $grpc.ServiceCall call, $0.GaugesRequest request);
  $async.Future<$0.ActiveGaugesResponse> activeGauges(
      $grpc.ServiceCall call, $0.ActiveGaugesRequest request);
  $async.Future<$0.UpcomingGaugesResponse> upcomingGauges(
      $grpc.ServiceCall call, $0.UpcomingGaugesRequest request);
  $async.Future<$0.RewardsEstResponse> rewardsEst(
      $grpc.ServiceCall call, $0.RewardsEstRequest request);
  $async.Future<$0.QueryLockableDurationsResponse> lockableDurations(
      $grpc.ServiceCall call, $0.QueryLockableDurationsRequest request);
}
