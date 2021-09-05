///
//  Generated code. Do not modify.
//  source: cosmos/staking/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$createValidator =
      $grpc.ClientMethod<$1.MsgCreateValidator, $1.MsgCreateValidatorResponse>(
          '/cosmos.staking.v1beta1.Msg/CreateValidator',
          ($1.MsgCreateValidator value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgCreateValidatorResponse.fromBuffer(value));
  static final _$editValidator =
      $grpc.ClientMethod<$1.MsgEditValidator, $1.MsgEditValidatorResponse>(
          '/cosmos.staking.v1beta1.Msg/EditValidator',
          ($1.MsgEditValidator value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgEditValidatorResponse.fromBuffer(value));
  static final _$delegate =
      $grpc.ClientMethod<$1.MsgDelegate, $1.MsgDelegateResponse>(
          '/cosmos.staking.v1beta1.Msg/Delegate',
          ($1.MsgDelegate value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgDelegateResponse.fromBuffer(value));
  static final _$beginRedelegate =
      $grpc.ClientMethod<$1.MsgBeginRedelegate, $1.MsgBeginRedelegateResponse>(
          '/cosmos.staking.v1beta1.Msg/BeginRedelegate',
          ($1.MsgBeginRedelegate value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgBeginRedelegateResponse.fromBuffer(value));
  static final _$undelegate =
      $grpc.ClientMethod<$1.MsgUndelegate, $1.MsgUndelegateResponse>(
          '/cosmos.staking.v1beta1.Msg/Undelegate',
          ($1.MsgUndelegate value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgUndelegateResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgCreateValidatorResponse> createValidator(
      $1.MsgCreateValidator request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$createValidator, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgEditValidatorResponse> editValidator(
      $1.MsgEditValidator request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$editValidator, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgDelegateResponse> delegate($1.MsgDelegate request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$delegate, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgBeginRedelegateResponse> beginRedelegate(
      $1.MsgBeginRedelegate request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$beginRedelegate, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgUndelegateResponse> undelegate(
      $1.MsgUndelegate request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$undelegate, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.staking.v1beta1.Msg';

  MsgServiceBase() {
    $addMethod($grpc.ServiceMethod<$1.MsgCreateValidator,
            $1.MsgCreateValidatorResponse>(
        'CreateValidator',
        createValidator_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgCreateValidator.fromBuffer(value),
        ($1.MsgCreateValidatorResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.MsgEditValidator, $1.MsgEditValidatorResponse>(
            'EditValidator',
            editValidator_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.MsgEditValidator.fromBuffer(value),
            ($1.MsgEditValidatorResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgDelegate, $1.MsgDelegateResponse>(
        'Delegate',
        delegate_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgDelegate.fromBuffer(value),
        ($1.MsgDelegateResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgBeginRedelegate,
            $1.MsgBeginRedelegateResponse>(
        'BeginRedelegate',
        beginRedelegate_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgBeginRedelegate.fromBuffer(value),
        ($1.MsgBeginRedelegateResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgUndelegate, $1.MsgUndelegateResponse>(
        'Undelegate',
        undelegate_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgUndelegate.fromBuffer(value),
        ($1.MsgUndelegateResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgCreateValidatorResponse> createValidator_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgCreateValidator> request) async {
    return createValidator(call, await request);
  }

  $async.Future<$1.MsgEditValidatorResponse> editValidator_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgEditValidator> request) async {
    return editValidator(call, await request);
  }

  $async.Future<$1.MsgDelegateResponse> delegate_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgDelegate> request) async {
    return delegate(call, await request);
  }

  $async.Future<$1.MsgBeginRedelegateResponse> beginRedelegate_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgBeginRedelegate> request) async {
    return beginRedelegate(call, await request);
  }

  $async.Future<$1.MsgUndelegateResponse> undelegate_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgUndelegate> request) async {
    return undelegate(call, await request);
  }

  $async.Future<$1.MsgCreateValidatorResponse> createValidator(
      $grpc.ServiceCall call, $1.MsgCreateValidator request);
  $async.Future<$1.MsgEditValidatorResponse> editValidator(
      $grpc.ServiceCall call, $1.MsgEditValidator request);
  $async.Future<$1.MsgDelegateResponse> delegate(
      $grpc.ServiceCall call, $1.MsgDelegate request);
  $async.Future<$1.MsgBeginRedelegateResponse> beginRedelegate(
      $grpc.ServiceCall call, $1.MsgBeginRedelegate request);
  $async.Future<$1.MsgUndelegateResponse> undelegate(
      $grpc.ServiceCall call, $1.MsgUndelegate request);
}
