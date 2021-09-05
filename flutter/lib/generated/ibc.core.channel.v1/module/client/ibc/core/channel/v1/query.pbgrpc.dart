///
//  Generated code. Do not modify.
//  source: ibc/core/channel/v1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'query.pb.dart' as $0;
export 'query.pb.dart';

class QueryClient extends $grpc.Client {
  static final _$channel =
      $grpc.ClientMethod<$0.QueryChannelRequest, $0.QueryChannelResponse>(
          '/ibc.core.channel.v1.Query/Channel',
          ($0.QueryChannelRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryChannelResponse.fromBuffer(value));
  static final _$channels =
      $grpc.ClientMethod<$0.QueryChannelsRequest, $0.QueryChannelsResponse>(
          '/ibc.core.channel.v1.Query/Channels',
          ($0.QueryChannelsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryChannelsResponse.fromBuffer(value));
  static final _$connectionChannels = $grpc.ClientMethod<
          $0.QueryConnectionChannelsRequest,
          $0.QueryConnectionChannelsResponse>(
      '/ibc.core.channel.v1.Query/ConnectionChannels',
      ($0.QueryConnectionChannelsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryConnectionChannelsResponse.fromBuffer(value));
  static final _$channelClientState = $grpc.ClientMethod<
          $0.QueryChannelClientStateRequest,
          $0.QueryChannelClientStateResponse>(
      '/ibc.core.channel.v1.Query/ChannelClientState',
      ($0.QueryChannelClientStateRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryChannelClientStateResponse.fromBuffer(value));
  static final _$channelConsensusState = $grpc.ClientMethod<
          $0.QueryChannelConsensusStateRequest,
          $0.QueryChannelConsensusStateResponse>(
      '/ibc.core.channel.v1.Query/ChannelConsensusState',
      ($0.QueryChannelConsensusStateRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryChannelConsensusStateResponse.fromBuffer(value));
  static final _$packetCommitment = $grpc.ClientMethod<
          $0.QueryPacketCommitmentRequest, $0.QueryPacketCommitmentResponse>(
      '/ibc.core.channel.v1.Query/PacketCommitment',
      ($0.QueryPacketCommitmentRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryPacketCommitmentResponse.fromBuffer(value));
  static final _$packetCommitments = $grpc.ClientMethod<
          $0.QueryPacketCommitmentsRequest, $0.QueryPacketCommitmentsResponse>(
      '/ibc.core.channel.v1.Query/PacketCommitments',
      ($0.QueryPacketCommitmentsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryPacketCommitmentsResponse.fromBuffer(value));
  static final _$packetReceipt = $grpc.ClientMethod<
          $0.QueryPacketReceiptRequest, $0.QueryPacketReceiptResponse>(
      '/ibc.core.channel.v1.Query/PacketReceipt',
      ($0.QueryPacketReceiptRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryPacketReceiptResponse.fromBuffer(value));
  static final _$packetAcknowledgement = $grpc.ClientMethod<
          $0.QueryPacketAcknowledgementRequest,
          $0.QueryPacketAcknowledgementResponse>(
      '/ibc.core.channel.v1.Query/PacketAcknowledgement',
      ($0.QueryPacketAcknowledgementRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryPacketAcknowledgementResponse.fromBuffer(value));
  static final _$packetAcknowledgements = $grpc.ClientMethod<
          $0.QueryPacketAcknowledgementsRequest,
          $0.QueryPacketAcknowledgementsResponse>(
      '/ibc.core.channel.v1.Query/PacketAcknowledgements',
      ($0.QueryPacketAcknowledgementsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryPacketAcknowledgementsResponse.fromBuffer(value));
  static final _$unreceivedPackets = $grpc.ClientMethod<
          $0.QueryUnreceivedPacketsRequest, $0.QueryUnreceivedPacketsResponse>(
      '/ibc.core.channel.v1.Query/UnreceivedPackets',
      ($0.QueryUnreceivedPacketsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryUnreceivedPacketsResponse.fromBuffer(value));
  static final _$unreceivedAcks = $grpc.ClientMethod<
          $0.QueryUnreceivedAcksRequest, $0.QueryUnreceivedAcksResponse>(
      '/ibc.core.channel.v1.Query/UnreceivedAcks',
      ($0.QueryUnreceivedAcksRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryUnreceivedAcksResponse.fromBuffer(value));
  static final _$nextSequenceReceive = $grpc.ClientMethod<
          $0.QueryNextSequenceReceiveRequest,
          $0.QueryNextSequenceReceiveResponse>(
      '/ibc.core.channel.v1.Query/NextSequenceReceive',
      ($0.QueryNextSequenceReceiveRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryNextSequenceReceiveResponse.fromBuffer(value));

  QueryClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.QueryChannelResponse> channel(
      $0.QueryChannelRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$channel, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryChannelsResponse> channels(
      $0.QueryChannelsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$channels, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryConnectionChannelsResponse> connectionChannels(
      $0.QueryConnectionChannelsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$connectionChannels, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryChannelClientStateResponse> channelClientState(
      $0.QueryChannelClientStateRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$channelClientState, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryChannelConsensusStateResponse>
      channelConsensusState($0.QueryChannelConsensusStateRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$channelConsensusState, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryPacketCommitmentResponse> packetCommitment(
      $0.QueryPacketCommitmentRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$packetCommitment, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryPacketCommitmentsResponse> packetCommitments(
      $0.QueryPacketCommitmentsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$packetCommitments, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryPacketReceiptResponse> packetReceipt(
      $0.QueryPacketReceiptRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$packetReceipt, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryPacketAcknowledgementResponse>
      packetAcknowledgement($0.QueryPacketAcknowledgementRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$packetAcknowledgement, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryPacketAcknowledgementsResponse>
      packetAcknowledgements($0.QueryPacketAcknowledgementsRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$packetAcknowledgements, request,
        options: options);
  }

  $grpc.ResponseFuture<$0.QueryUnreceivedPacketsResponse> unreceivedPackets(
      $0.QueryUnreceivedPacketsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$unreceivedPackets, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryUnreceivedAcksResponse> unreceivedAcks(
      $0.QueryUnreceivedAcksRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$unreceivedAcks, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryNextSequenceReceiveResponse> nextSequenceReceive(
      $0.QueryNextSequenceReceiveRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$nextSequenceReceive, request, options: options);
  }
}

abstract class QueryServiceBase extends $grpc.Service {
  $core.String get $name => 'ibc.core.channel.v1.Query';

  QueryServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$0.QueryChannelRequest, $0.QueryChannelResponse>(
            'Channel',
            channel_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.QueryChannelRequest.fromBuffer(value),
            ($0.QueryChannelResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.QueryChannelsRequest, $0.QueryChannelsResponse>(
            'Channels',
            channels_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.QueryChannelsRequest.fromBuffer(value),
            ($0.QueryChannelsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryConnectionChannelsRequest,
            $0.QueryConnectionChannelsResponse>(
        'ConnectionChannels',
        connectionChannels_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryConnectionChannelsRequest.fromBuffer(value),
        ($0.QueryConnectionChannelsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryChannelClientStateRequest,
            $0.QueryChannelClientStateResponse>(
        'ChannelClientState',
        channelClientState_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryChannelClientStateRequest.fromBuffer(value),
        ($0.QueryChannelClientStateResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryChannelConsensusStateRequest,
            $0.QueryChannelConsensusStateResponse>(
        'ChannelConsensusState',
        channelConsensusState_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryChannelConsensusStateRequest.fromBuffer(value),
        ($0.QueryChannelConsensusStateResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryPacketCommitmentRequest,
            $0.QueryPacketCommitmentResponse>(
        'PacketCommitment',
        packetCommitment_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryPacketCommitmentRequest.fromBuffer(value),
        ($0.QueryPacketCommitmentResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryPacketCommitmentsRequest,
            $0.QueryPacketCommitmentsResponse>(
        'PacketCommitments',
        packetCommitments_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryPacketCommitmentsRequest.fromBuffer(value),
        ($0.QueryPacketCommitmentsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryPacketReceiptRequest,
            $0.QueryPacketReceiptResponse>(
        'PacketReceipt',
        packetReceipt_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryPacketReceiptRequest.fromBuffer(value),
        ($0.QueryPacketReceiptResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryPacketAcknowledgementRequest,
            $0.QueryPacketAcknowledgementResponse>(
        'PacketAcknowledgement',
        packetAcknowledgement_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryPacketAcknowledgementRequest.fromBuffer(value),
        ($0.QueryPacketAcknowledgementResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryPacketAcknowledgementsRequest,
            $0.QueryPacketAcknowledgementsResponse>(
        'PacketAcknowledgements',
        packetAcknowledgements_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryPacketAcknowledgementsRequest.fromBuffer(value),
        ($0.QueryPacketAcknowledgementsResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryUnreceivedPacketsRequest,
            $0.QueryUnreceivedPacketsResponse>(
        'UnreceivedPackets',
        unreceivedPackets_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryUnreceivedPacketsRequest.fromBuffer(value),
        ($0.QueryUnreceivedPacketsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryUnreceivedAcksRequest,
            $0.QueryUnreceivedAcksResponse>(
        'UnreceivedAcks',
        unreceivedAcks_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryUnreceivedAcksRequest.fromBuffer(value),
        ($0.QueryUnreceivedAcksResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryNextSequenceReceiveRequest,
            $0.QueryNextSequenceReceiveResponse>(
        'NextSequenceReceive',
        nextSequenceReceive_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryNextSequenceReceiveRequest.fromBuffer(value),
        ($0.QueryNextSequenceReceiveResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.QueryChannelResponse> channel_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryChannelRequest> request) async {
    return channel(call, await request);
  }

  $async.Future<$0.QueryChannelsResponse> channels_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryChannelsRequest> request) async {
    return channels(call, await request);
  }

  $async.Future<$0.QueryConnectionChannelsResponse> connectionChannels_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryConnectionChannelsRequest> request) async {
    return connectionChannels(call, await request);
  }

  $async.Future<$0.QueryChannelClientStateResponse> channelClientState_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryChannelClientStateRequest> request) async {
    return channelClientState(call, await request);
  }

  $async.Future<$0.QueryChannelConsensusStateResponse>
      channelConsensusState_Pre($grpc.ServiceCall call,
          $async.Future<$0.QueryChannelConsensusStateRequest> request) async {
    return channelConsensusState(call, await request);
  }

  $async.Future<$0.QueryPacketCommitmentResponse> packetCommitment_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryPacketCommitmentRequest> request) async {
    return packetCommitment(call, await request);
  }

  $async.Future<$0.QueryPacketCommitmentsResponse> packetCommitments_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryPacketCommitmentsRequest> request) async {
    return packetCommitments(call, await request);
  }

  $async.Future<$0.QueryPacketReceiptResponse> packetReceipt_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryPacketReceiptRequest> request) async {
    return packetReceipt(call, await request);
  }

  $async.Future<$0.QueryPacketAcknowledgementResponse>
      packetAcknowledgement_Pre($grpc.ServiceCall call,
          $async.Future<$0.QueryPacketAcknowledgementRequest> request) async {
    return packetAcknowledgement(call, await request);
  }

  $async.Future<$0.QueryPacketAcknowledgementsResponse>
      packetAcknowledgements_Pre($grpc.ServiceCall call,
          $async.Future<$0.QueryPacketAcknowledgementsRequest> request) async {
    return packetAcknowledgements(call, await request);
  }

  $async.Future<$0.QueryUnreceivedPacketsResponse> unreceivedPackets_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryUnreceivedPacketsRequest> request) async {
    return unreceivedPackets(call, await request);
  }

  $async.Future<$0.QueryUnreceivedAcksResponse> unreceivedAcks_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryUnreceivedAcksRequest> request) async {
    return unreceivedAcks(call, await request);
  }

  $async.Future<$0.QueryNextSequenceReceiveResponse> nextSequenceReceive_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryNextSequenceReceiveRequest> request) async {
    return nextSequenceReceive(call, await request);
  }

  $async.Future<$0.QueryChannelResponse> channel(
      $grpc.ServiceCall call, $0.QueryChannelRequest request);
  $async.Future<$0.QueryChannelsResponse> channels(
      $grpc.ServiceCall call, $0.QueryChannelsRequest request);
  $async.Future<$0.QueryConnectionChannelsResponse> connectionChannels(
      $grpc.ServiceCall call, $0.QueryConnectionChannelsRequest request);
  $async.Future<$0.QueryChannelClientStateResponse> channelClientState(
      $grpc.ServiceCall call, $0.QueryChannelClientStateRequest request);
  $async.Future<$0.QueryChannelConsensusStateResponse> channelConsensusState(
      $grpc.ServiceCall call, $0.QueryChannelConsensusStateRequest request);
  $async.Future<$0.QueryPacketCommitmentResponse> packetCommitment(
      $grpc.ServiceCall call, $0.QueryPacketCommitmentRequest request);
  $async.Future<$0.QueryPacketCommitmentsResponse> packetCommitments(
      $grpc.ServiceCall call, $0.QueryPacketCommitmentsRequest request);
  $async.Future<$0.QueryPacketReceiptResponse> packetReceipt(
      $grpc.ServiceCall call, $0.QueryPacketReceiptRequest request);
  $async.Future<$0.QueryPacketAcknowledgementResponse> packetAcknowledgement(
      $grpc.ServiceCall call, $0.QueryPacketAcknowledgementRequest request);
  $async.Future<$0.QueryPacketAcknowledgementsResponse> packetAcknowledgements(
      $grpc.ServiceCall call, $0.QueryPacketAcknowledgementsRequest request);
  $async.Future<$0.QueryUnreceivedPacketsResponse> unreceivedPackets(
      $grpc.ServiceCall call, $0.QueryUnreceivedPacketsRequest request);
  $async.Future<$0.QueryUnreceivedAcksResponse> unreceivedAcks(
      $grpc.ServiceCall call, $0.QueryUnreceivedAcksRequest request);
  $async.Future<$0.QueryNextSequenceReceiveResponse> nextSequenceReceive(
      $grpc.ServiceCall call, $0.QueryNextSequenceReceiveRequest request);
}
