///
//  Generated code. Do not modify.
//  source: ibc/core/client/v1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$createClient =
      $grpc.ClientMethod<$1.MsgCreateClient, $1.MsgCreateClientResponse>(
          '/ibc.core.client.v1.Msg/CreateClient',
          ($1.MsgCreateClient value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgCreateClientResponse.fromBuffer(value));
  static final _$updateClient =
      $grpc.ClientMethod<$1.MsgUpdateClient, $1.MsgUpdateClientResponse>(
          '/ibc.core.client.v1.Msg/UpdateClient',
          ($1.MsgUpdateClient value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgUpdateClientResponse.fromBuffer(value));
  static final _$upgradeClient =
      $grpc.ClientMethod<$1.MsgUpgradeClient, $1.MsgUpgradeClientResponse>(
          '/ibc.core.client.v1.Msg/UpgradeClient',
          ($1.MsgUpgradeClient value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgUpgradeClientResponse.fromBuffer(value));
  static final _$submitMisbehaviour = $grpc.ClientMethod<
          $1.MsgSubmitMisbehaviour, $1.MsgSubmitMisbehaviourResponse>(
      '/ibc.core.client.v1.Msg/SubmitMisbehaviour',
      ($1.MsgSubmitMisbehaviour value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgSubmitMisbehaviourResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgCreateClientResponse> createClient(
      $1.MsgCreateClient request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$createClient, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgUpdateClientResponse> updateClient(
      $1.MsgUpdateClient request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$updateClient, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgUpgradeClientResponse> upgradeClient(
      $1.MsgUpgradeClient request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$upgradeClient, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgSubmitMisbehaviourResponse> submitMisbehaviour(
      $1.MsgSubmitMisbehaviour request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$submitMisbehaviour, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'ibc.core.client.v1.Msg';

  MsgServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$1.MsgCreateClient, $1.MsgCreateClientResponse>(
            'CreateClient',
            createClient_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.MsgCreateClient.fromBuffer(value),
            ($1.MsgCreateClientResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.MsgUpdateClient, $1.MsgUpdateClientResponse>(
            'UpdateClient',
            updateClient_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.MsgUpdateClient.fromBuffer(value),
            ($1.MsgUpdateClientResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.MsgUpgradeClient, $1.MsgUpgradeClientResponse>(
            'UpgradeClient',
            upgradeClient_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.MsgUpgradeClient.fromBuffer(value),
            ($1.MsgUpgradeClientResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgSubmitMisbehaviour,
            $1.MsgSubmitMisbehaviourResponse>(
        'SubmitMisbehaviour',
        submitMisbehaviour_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgSubmitMisbehaviour.fromBuffer(value),
        ($1.MsgSubmitMisbehaviourResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgCreateClientResponse> createClient_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgCreateClient> request) async {
    return createClient(call, await request);
  }

  $async.Future<$1.MsgUpdateClientResponse> updateClient_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgUpdateClient> request) async {
    return updateClient(call, await request);
  }

  $async.Future<$1.MsgUpgradeClientResponse> upgradeClient_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgUpgradeClient> request) async {
    return upgradeClient(call, await request);
  }

  $async.Future<$1.MsgSubmitMisbehaviourResponse> submitMisbehaviour_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgSubmitMisbehaviour> request) async {
    return submitMisbehaviour(call, await request);
  }

  $async.Future<$1.MsgCreateClientResponse> createClient(
      $grpc.ServiceCall call, $1.MsgCreateClient request);
  $async.Future<$1.MsgUpdateClientResponse> updateClient(
      $grpc.ServiceCall call, $1.MsgUpdateClient request);
  $async.Future<$1.MsgUpgradeClientResponse> upgradeClient(
      $grpc.ServiceCall call, $1.MsgUpgradeClient request);
  $async.Future<$1.MsgSubmitMisbehaviourResponse> submitMisbehaviour(
      $grpc.ServiceCall call, $1.MsgSubmitMisbehaviour request);
}
