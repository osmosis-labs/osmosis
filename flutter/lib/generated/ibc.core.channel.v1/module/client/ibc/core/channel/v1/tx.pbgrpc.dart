///
//  Generated code. Do not modify.
//  source: ibc/core/channel/v1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$channelOpenInit =
      $grpc.ClientMethod<$1.MsgChannelOpenInit, $1.MsgChannelOpenInitResponse>(
          '/ibc.core.channel.v1.Msg/ChannelOpenInit',
          ($1.MsgChannelOpenInit value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgChannelOpenInitResponse.fromBuffer(value));
  static final _$channelOpenTry =
      $grpc.ClientMethod<$1.MsgChannelOpenTry, $1.MsgChannelOpenTryResponse>(
          '/ibc.core.channel.v1.Msg/ChannelOpenTry',
          ($1.MsgChannelOpenTry value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgChannelOpenTryResponse.fromBuffer(value));
  static final _$channelOpenAck =
      $grpc.ClientMethod<$1.MsgChannelOpenAck, $1.MsgChannelOpenAckResponse>(
          '/ibc.core.channel.v1.Msg/ChannelOpenAck',
          ($1.MsgChannelOpenAck value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgChannelOpenAckResponse.fromBuffer(value));
  static final _$channelOpenConfirm = $grpc.ClientMethod<
          $1.MsgChannelOpenConfirm, $1.MsgChannelOpenConfirmResponse>(
      '/ibc.core.channel.v1.Msg/ChannelOpenConfirm',
      ($1.MsgChannelOpenConfirm value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgChannelOpenConfirmResponse.fromBuffer(value));
  static final _$channelCloseInit = $grpc.ClientMethod<$1.MsgChannelCloseInit,
          $1.MsgChannelCloseInitResponse>(
      '/ibc.core.channel.v1.Msg/ChannelCloseInit',
      ($1.MsgChannelCloseInit value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgChannelCloseInitResponse.fromBuffer(value));
  static final _$channelCloseConfirm = $grpc.ClientMethod<
          $1.MsgChannelCloseConfirm, $1.MsgChannelCloseConfirmResponse>(
      '/ibc.core.channel.v1.Msg/ChannelCloseConfirm',
      ($1.MsgChannelCloseConfirm value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $1.MsgChannelCloseConfirmResponse.fromBuffer(value));
  static final _$recvPacket =
      $grpc.ClientMethod<$1.MsgRecvPacket, $1.MsgRecvPacketResponse>(
          '/ibc.core.channel.v1.Msg/RecvPacket',
          ($1.MsgRecvPacket value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgRecvPacketResponse.fromBuffer(value));
  static final _$timeout =
      $grpc.ClientMethod<$1.MsgTimeout, $1.MsgTimeoutResponse>(
          '/ibc.core.channel.v1.Msg/Timeout',
          ($1.MsgTimeout value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgTimeoutResponse.fromBuffer(value));
  static final _$timeoutOnClose =
      $grpc.ClientMethod<$1.MsgTimeoutOnClose, $1.MsgTimeoutOnCloseResponse>(
          '/ibc.core.channel.v1.Msg/TimeoutOnClose',
          ($1.MsgTimeoutOnClose value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgTimeoutOnCloseResponse.fromBuffer(value));
  static final _$acknowledgement =
      $grpc.ClientMethod<$1.MsgAcknowledgement, $1.MsgAcknowledgementResponse>(
          '/ibc.core.channel.v1.Msg/Acknowledgement',
          ($1.MsgAcknowledgement value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgAcknowledgementResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgChannelOpenInitResponse> channelOpenInit(
      $1.MsgChannelOpenInit request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$channelOpenInit, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgChannelOpenTryResponse> channelOpenTry(
      $1.MsgChannelOpenTry request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$channelOpenTry, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgChannelOpenAckResponse> channelOpenAck(
      $1.MsgChannelOpenAck request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$channelOpenAck, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgChannelOpenConfirmResponse> channelOpenConfirm(
      $1.MsgChannelOpenConfirm request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$channelOpenConfirm, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgChannelCloseInitResponse> channelCloseInit(
      $1.MsgChannelCloseInit request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$channelCloseInit, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgChannelCloseConfirmResponse> channelCloseConfirm(
      $1.MsgChannelCloseConfirm request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$channelCloseConfirm, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgRecvPacketResponse> recvPacket(
      $1.MsgRecvPacket request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$recvPacket, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgTimeoutResponse> timeout($1.MsgTimeout request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$timeout, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgTimeoutOnCloseResponse> timeoutOnClose(
      $1.MsgTimeoutOnClose request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$timeoutOnClose, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgAcknowledgementResponse> acknowledgement(
      $1.MsgAcknowledgement request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$acknowledgement, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'ibc.core.channel.v1.Msg';

  MsgServiceBase() {
    $addMethod($grpc.ServiceMethod<$1.MsgChannelOpenInit,
            $1.MsgChannelOpenInitResponse>(
        'ChannelOpenInit',
        channelOpenInit_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgChannelOpenInit.fromBuffer(value),
        ($1.MsgChannelOpenInitResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.MsgChannelOpenTry, $1.MsgChannelOpenTryResponse>(
            'ChannelOpenTry',
            channelOpenTry_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.MsgChannelOpenTry.fromBuffer(value),
            ($1.MsgChannelOpenTryResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.MsgChannelOpenAck, $1.MsgChannelOpenAckResponse>(
            'ChannelOpenAck',
            channelOpenAck_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.MsgChannelOpenAck.fromBuffer(value),
            ($1.MsgChannelOpenAckResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgChannelOpenConfirm,
            $1.MsgChannelOpenConfirmResponse>(
        'ChannelOpenConfirm',
        channelOpenConfirm_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgChannelOpenConfirm.fromBuffer(value),
        ($1.MsgChannelOpenConfirmResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgChannelCloseInit,
            $1.MsgChannelCloseInitResponse>(
        'ChannelCloseInit',
        channelCloseInit_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgChannelCloseInit.fromBuffer(value),
        ($1.MsgChannelCloseInitResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgChannelCloseConfirm,
            $1.MsgChannelCloseConfirmResponse>(
        'ChannelCloseConfirm',
        channelCloseConfirm_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgChannelCloseConfirm.fromBuffer(value),
        ($1.MsgChannelCloseConfirmResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgRecvPacket, $1.MsgRecvPacketResponse>(
        'RecvPacket',
        recvPacket_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgRecvPacket.fromBuffer(value),
        ($1.MsgRecvPacketResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgTimeout, $1.MsgTimeoutResponse>(
        'Timeout',
        timeout_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgTimeout.fromBuffer(value),
        ($1.MsgTimeoutResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.MsgTimeoutOnClose, $1.MsgTimeoutOnCloseResponse>(
            'TimeoutOnClose',
            timeoutOnClose_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.MsgTimeoutOnClose.fromBuffer(value),
            ($1.MsgTimeoutOnCloseResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgAcknowledgement,
            $1.MsgAcknowledgementResponse>(
        'Acknowledgement',
        acknowledgement_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.MsgAcknowledgement.fromBuffer(value),
        ($1.MsgAcknowledgementResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgChannelOpenInitResponse> channelOpenInit_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgChannelOpenInit> request) async {
    return channelOpenInit(call, await request);
  }

  $async.Future<$1.MsgChannelOpenTryResponse> channelOpenTry_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgChannelOpenTry> request) async {
    return channelOpenTry(call, await request);
  }

  $async.Future<$1.MsgChannelOpenAckResponse> channelOpenAck_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgChannelOpenAck> request) async {
    return channelOpenAck(call, await request);
  }

  $async.Future<$1.MsgChannelOpenConfirmResponse> channelOpenConfirm_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgChannelOpenConfirm> request) async {
    return channelOpenConfirm(call, await request);
  }

  $async.Future<$1.MsgChannelCloseInitResponse> channelCloseInit_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgChannelCloseInit> request) async {
    return channelCloseInit(call, await request);
  }

  $async.Future<$1.MsgChannelCloseConfirmResponse> channelCloseConfirm_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgChannelCloseConfirm> request) async {
    return channelCloseConfirm(call, await request);
  }

  $async.Future<$1.MsgRecvPacketResponse> recvPacket_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgRecvPacket> request) async {
    return recvPacket(call, await request);
  }

  $async.Future<$1.MsgTimeoutResponse> timeout_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgTimeout> request) async {
    return timeout(call, await request);
  }

  $async.Future<$1.MsgTimeoutOnCloseResponse> timeoutOnClose_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgTimeoutOnClose> request) async {
    return timeoutOnClose(call, await request);
  }

  $async.Future<$1.MsgAcknowledgementResponse> acknowledgement_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgAcknowledgement> request) async {
    return acknowledgement(call, await request);
  }

  $async.Future<$1.MsgChannelOpenInitResponse> channelOpenInit(
      $grpc.ServiceCall call, $1.MsgChannelOpenInit request);
  $async.Future<$1.MsgChannelOpenTryResponse> channelOpenTry(
      $grpc.ServiceCall call, $1.MsgChannelOpenTry request);
  $async.Future<$1.MsgChannelOpenAckResponse> channelOpenAck(
      $grpc.ServiceCall call, $1.MsgChannelOpenAck request);
  $async.Future<$1.MsgChannelOpenConfirmResponse> channelOpenConfirm(
      $grpc.ServiceCall call, $1.MsgChannelOpenConfirm request);
  $async.Future<$1.MsgChannelCloseInitResponse> channelCloseInit(
      $grpc.ServiceCall call, $1.MsgChannelCloseInit request);
  $async.Future<$1.MsgChannelCloseConfirmResponse> channelCloseConfirm(
      $grpc.ServiceCall call, $1.MsgChannelCloseConfirm request);
  $async.Future<$1.MsgRecvPacketResponse> recvPacket(
      $grpc.ServiceCall call, $1.MsgRecvPacket request);
  $async.Future<$1.MsgTimeoutResponse> timeout(
      $grpc.ServiceCall call, $1.MsgTimeout request);
  $async.Future<$1.MsgTimeoutOnCloseResponse> timeoutOnClose(
      $grpc.ServiceCall call, $1.MsgTimeoutOnClose request);
  $async.Future<$1.MsgAcknowledgementResponse> acknowledgement(
      $grpc.ServiceCall call, $1.MsgAcknowledgement request);
}
