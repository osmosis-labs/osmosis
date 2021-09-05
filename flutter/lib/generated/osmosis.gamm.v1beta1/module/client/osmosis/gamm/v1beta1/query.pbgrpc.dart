///
//  Generated code. Do not modify.
//  source: osmosis/gamm/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'query.pb.dart' as $1;
export 'query.pb.dart';

class QueryClient extends $grpc.Client {
  static final _$pools =
      $grpc.ClientMethod<$1.QueryPoolsRequest, $1.QueryPoolsResponse>(
          '/osmosis.gamm.v1beta1.Query/Pools',
          ($1.QueryPoolsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.QueryPoolsResponse.fromBuffer(value));
  static final _$numPools =
      $grpc.ClientMethod<$1.QueryNumPoolsRequest, $1.QueryNumPoolsResponse>(
          '/osmosis.gamm.v1beta1.Query/NumPools',
          ($1.QueryNumPoolsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.QueryNumPoolsResponse.fromBuffer(value));
  static final _$totalLiquidity = $grpc.ClientMethod<
          $1.QueryTotalLiquidityRequest, $1.QueryTotalLiquidityResponse>(
      '/osmosis.gamm.v1beta1.Query/TotalLiquidity',
      ($1.QueryTotalLiquidityRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.QueryTotalLiquidityResponse.fromBuffer(value));
  static final _$pool =
      $grpc.ClientMethod<$1.QueryPoolRequest, $1.QueryPoolResponse>(
          '/osmosis.gamm.v1beta1.Query/Pool',
          ($1.QueryPoolRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.QueryPoolResponse.fromBuffer(value));
  static final _$poolParams =
      $grpc.ClientMethod<$1.QueryPoolParamsRequest, $1.QueryPoolParamsResponse>(
          '/osmosis.gamm.v1beta1.Query/PoolParams',
          ($1.QueryPoolParamsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.QueryPoolParamsResponse.fromBuffer(value));
  static final _$totalShares = $grpc.ClientMethod<$1.QueryTotalSharesRequest,
          $1.QueryTotalSharesResponse>(
      '/osmosis.gamm.v1beta1.Query/TotalShares',
      ($1.QueryTotalSharesRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.QueryTotalSharesResponse.fromBuffer(value));
  static final _$poolAssets =
      $grpc.ClientMethod<$1.QueryPoolAssetsRequest, $1.QueryPoolAssetsResponse>(
          '/osmosis.gamm.v1beta1.Query/PoolAssets',
          ($1.QueryPoolAssetsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.QueryPoolAssetsResponse.fromBuffer(value));
  static final _$spotPrice =
      $grpc.ClientMethod<$1.QuerySpotPriceRequest, $1.QuerySpotPriceResponse>(
          '/osmosis.gamm.v1beta1.Query/SpotPrice',
          ($1.QuerySpotPriceRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.QuerySpotPriceResponse.fromBuffer(value));
  static final _$estimateSwapExactAmountIn = $grpc.ClientMethod<
          $1.QuerySwapExactAmountInRequest, $1.QuerySwapExactAmountInResponse>(
      '/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountIn',
      ($1.QuerySwapExactAmountInRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.QuerySwapExactAmountInResponse.fromBuffer(value));
  static final _$estimateSwapExactAmountOut = $grpc.ClientMethod<
          $1.QuerySwapExactAmountOutRequest,
          $1.QuerySwapExactAmountOutResponse>(
      '/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut',
      ($1.QuerySwapExactAmountOutRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.QuerySwapExactAmountOutResponse.fromBuffer(value));

  QueryClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.QueryPoolsResponse> pools(
      $1.QueryPoolsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$pools, request, options: options);
  }

  $grpc.ResponseFuture<$1.QueryNumPoolsResponse> numPools(
      $1.QueryNumPoolsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$numPools, request, options: options);
  }

  $grpc.ResponseFuture<$1.QueryTotalLiquidityResponse> totalLiquidity(
      $1.QueryTotalLiquidityRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$totalLiquidity, request, options: options);
  }

  $grpc.ResponseFuture<$1.QueryPoolResponse> pool($1.QueryPoolRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$pool, request, options: options);
  }

  $grpc.ResponseFuture<$1.QueryPoolParamsResponse> poolParams(
      $1.QueryPoolParamsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$poolParams, request, options: options);
  }

  $grpc.ResponseFuture<$1.QueryTotalSharesResponse> totalShares(
      $1.QueryTotalSharesRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$totalShares, request, options: options);
  }

  $grpc.ResponseFuture<$1.QueryPoolAssetsResponse> poolAssets(
      $1.QueryPoolAssetsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$poolAssets, request, options: options);
  }

  $grpc.ResponseFuture<$1.QuerySpotPriceResponse> spotPrice(
      $1.QuerySpotPriceRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$spotPrice, request, options: options);
  }

  $grpc.ResponseFuture<$1.QuerySwapExactAmountInResponse>
      estimateSwapExactAmountIn($1.QuerySwapExactAmountInRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$estimateSwapExactAmountIn, request,
        options: options);
  }

  $grpc.ResponseFuture<$1.QuerySwapExactAmountOutResponse>
      estimateSwapExactAmountOut($1.QuerySwapExactAmountOutRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$estimateSwapExactAmountOut, request,
        options: options);
  }
}

abstract class QueryServiceBase extends $grpc.Service {
  $core.String get $name => 'osmosis.gamm.v1beta1.Query';

  QueryServiceBase() {
    $addMethod($grpc.ServiceMethod<$1.QueryPoolsRequest, $1.QueryPoolsResponse>(
        'Pools',
        pools_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.QueryPoolsRequest.fromBuffer(value),
        ($1.QueryPoolsResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.QueryNumPoolsRequest, $1.QueryNumPoolsResponse>(
            'NumPools',
            numPools_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.QueryNumPoolsRequest.fromBuffer(value),
            ($1.QueryNumPoolsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.QueryTotalLiquidityRequest,
            $1.QueryTotalLiquidityResponse>(
        'TotalLiquidity',
        totalLiquidity_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.QueryTotalLiquidityRequest.fromBuffer(value),
        ($1.QueryTotalLiquidityResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.QueryPoolRequest, $1.QueryPoolResponse>(
        'Pool',
        pool_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.QueryPoolRequest.fromBuffer(value),
        ($1.QueryPoolResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.QueryPoolParamsRequest,
            $1.QueryPoolParamsResponse>(
        'PoolParams',
        poolParams_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.QueryPoolParamsRequest.fromBuffer(value),
        ($1.QueryPoolParamsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.QueryTotalSharesRequest,
            $1.QueryTotalSharesResponse>(
        'TotalShares',
        totalShares_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.QueryTotalSharesRequest.fromBuffer(value),
        ($1.QueryTotalSharesResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.QueryPoolAssetsRequest,
            $1.QueryPoolAssetsResponse>(
        'PoolAssets',
        poolAssets_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.QueryPoolAssetsRequest.fromBuffer(value),
        ($1.QueryPoolAssetsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.QuerySpotPriceRequest,
            $1.QuerySpotPriceResponse>(
        'SpotPrice',
        spotPrice_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.QuerySpotPriceRequest.fromBuffer(value),
        ($1.QuerySpotPriceResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.QuerySwapExactAmountInRequest,
            $1.QuerySwapExactAmountInResponse>(
        'EstimateSwapExactAmountIn',
        estimateSwapExactAmountIn_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.QuerySwapExactAmountInRequest.fromBuffer(value),
        ($1.QuerySwapExactAmountInResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.QuerySwapExactAmountOutRequest,
            $1.QuerySwapExactAmountOutResponse>(
        'EstimateSwapExactAmountOut',
        estimateSwapExactAmountOut_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.QuerySwapExactAmountOutRequest.fromBuffer(value),
        ($1.QuerySwapExactAmountOutResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.QueryPoolsResponse> pools_Pre($grpc.ServiceCall call,
      $async.Future<$1.QueryPoolsRequest> request) async {
    return pools(call, await request);
  }

  $async.Future<$1.QueryNumPoolsResponse> numPools_Pre($grpc.ServiceCall call,
      $async.Future<$1.QueryNumPoolsRequest> request) async {
    return numPools(call, await request);
  }

  $async.Future<$1.QueryTotalLiquidityResponse> totalLiquidity_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.QueryTotalLiquidityRequest> request) async {
    return totalLiquidity(call, await request);
  }

  $async.Future<$1.QueryPoolResponse> pool_Pre($grpc.ServiceCall call,
      $async.Future<$1.QueryPoolRequest> request) async {
    return pool(call, await request);
  }

  $async.Future<$1.QueryPoolParamsResponse> poolParams_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.QueryPoolParamsRequest> request) async {
    return poolParams(call, await request);
  }

  $async.Future<$1.QueryTotalSharesResponse> totalShares_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.QueryTotalSharesRequest> request) async {
    return totalShares(call, await request);
  }

  $async.Future<$1.QueryPoolAssetsResponse> poolAssets_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.QueryPoolAssetsRequest> request) async {
    return poolAssets(call, await request);
  }

  $async.Future<$1.QuerySpotPriceResponse> spotPrice_Pre($grpc.ServiceCall call,
      $async.Future<$1.QuerySpotPriceRequest> request) async {
    return spotPrice(call, await request);
  }

  $async.Future<$1.QuerySwapExactAmountInResponse>
      estimateSwapExactAmountIn_Pre($grpc.ServiceCall call,
          $async.Future<$1.QuerySwapExactAmountInRequest> request) async {
    return estimateSwapExactAmountIn(call, await request);
  }

  $async.Future<$1.QuerySwapExactAmountOutResponse>
      estimateSwapExactAmountOut_Pre($grpc.ServiceCall call,
          $async.Future<$1.QuerySwapExactAmountOutRequest> request) async {
    return estimateSwapExactAmountOut(call, await request);
  }

  $async.Future<$1.QueryPoolsResponse> pools(
      $grpc.ServiceCall call, $1.QueryPoolsRequest request);
  $async.Future<$1.QueryNumPoolsResponse> numPools(
      $grpc.ServiceCall call, $1.QueryNumPoolsRequest request);
  $async.Future<$1.QueryTotalLiquidityResponse> totalLiquidity(
      $grpc.ServiceCall call, $1.QueryTotalLiquidityRequest request);
  $async.Future<$1.QueryPoolResponse> pool(
      $grpc.ServiceCall call, $1.QueryPoolRequest request);
  $async.Future<$1.QueryPoolParamsResponse> poolParams(
      $grpc.ServiceCall call, $1.QueryPoolParamsRequest request);
  $async.Future<$1.QueryTotalSharesResponse> totalShares(
      $grpc.ServiceCall call, $1.QueryTotalSharesRequest request);
  $async.Future<$1.QueryPoolAssetsResponse> poolAssets(
      $grpc.ServiceCall call, $1.QueryPoolAssetsRequest request);
  $async.Future<$1.QuerySpotPriceResponse> spotPrice(
      $grpc.ServiceCall call, $1.QuerySpotPriceRequest request);
  $async.Future<$1.QuerySwapExactAmountInResponse> estimateSwapExactAmountIn(
      $grpc.ServiceCall call, $1.QuerySwapExactAmountInRequest request);
  $async.Future<$1.QuerySwapExactAmountOutResponse> estimateSwapExactAmountOut(
      $grpc.ServiceCall call, $1.QuerySwapExactAmountOutRequest request);
}
