///
//  Generated code. Do not modify.
//  source: osmosis/incentives/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$createGauge =
      $grpc.ClientMethod<$1.MsgCreateGauge, $1.MsgCreateGaugeResponse>(
          '/osmosis.incentives.Msg/CreateGauge',
          ($1.MsgCreateGauge value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgCreateGaugeResponse.fromBuffer(value));
  static final _$addToGauge =
      $grpc.ClientMethod<$1.MsgAddToGauge, $1.MsgAddToGaugeResponse>(
          '/osmosis.incentives.Msg/AddToGauge',
          ($1.MsgAddToGauge value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgAddToGaugeResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgCreateGaugeResponse> createGauge(
      $1.MsgCreateGauge request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$createGauge, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgAddToGaugeResponse> addToGauge(
      $1.MsgAddToGauge request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$addToGauge, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'osmosis.incentives.Msg';

  MsgServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$1.MsgCreateGauge, $1.MsgCreateGaugeResponse>(
            'CreateGauge',
            createGauge_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.MsgCreateGauge.fromBuffer(value),
            ($1.MsgCreateGaugeResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgAddToGauge, $1.MsgAddToGaugeResponse>(
        'AddToGauge',
        addToGauge_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgAddToGauge.fromBuffer(value),
        ($1.MsgAddToGaugeResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgCreateGaugeResponse> createGauge_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgCreateGauge> request) async {
    return createGauge(call, await request);
  }

  $async.Future<$1.MsgAddToGaugeResponse> addToGauge_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgAddToGauge> request) async {
    return addToGauge(call, await request);
  }

  $async.Future<$1.MsgCreateGaugeResponse> createGauge(
      $grpc.ServiceCall call, $1.MsgCreateGauge request);
  $async.Future<$1.MsgAddToGaugeResponse> addToGauge(
      $grpc.ServiceCall call, $1.MsgAddToGauge request);
}
