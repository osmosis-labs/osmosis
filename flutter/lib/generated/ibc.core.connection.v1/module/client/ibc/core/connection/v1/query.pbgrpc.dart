///
//  Generated code. Do not modify.
//  source: ibc/core/connection/v1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'query.pb.dart' as $0;
export 'query.pb.dart';

class QueryClient extends $grpc.Client {
  static final _$connection =
      $grpc.ClientMethod<$0.QueryConnectionRequest, $0.QueryConnectionResponse>(
          '/ibc.core.connection.v1.Query/Connection',
          ($0.QueryConnectionRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryConnectionResponse.fromBuffer(value));
  static final _$connections = $grpc.ClientMethod<$0.QueryConnectionsRequest,
          $0.QueryConnectionsResponse>(
      '/ibc.core.connection.v1.Query/Connections',
      ($0.QueryConnectionsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryConnectionsResponse.fromBuffer(value));
  static final _$clientConnections = $grpc.ClientMethod<
          $0.QueryClientConnectionsRequest, $0.QueryClientConnectionsResponse>(
      '/ibc.core.connection.v1.Query/ClientConnections',
      ($0.QueryClientConnectionsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryClientConnectionsResponse.fromBuffer(value));
  static final _$connectionClientState = $grpc.ClientMethod<
          $0.QueryConnectionClientStateRequest,
          $0.QueryConnectionClientStateResponse>(
      '/ibc.core.connection.v1.Query/ConnectionClientState',
      ($0.QueryConnectionClientStateRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryConnectionClientStateResponse.fromBuffer(value));
  static final _$connectionConsensusState = $grpc.ClientMethod<
          $0.QueryConnectionConsensusStateRequest,
          $0.QueryConnectionConsensusStateResponse>(
      '/ibc.core.connection.v1.Query/ConnectionConsensusState',
      ($0.QueryConnectionConsensusStateRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryConnectionConsensusStateResponse.fromBuffer(value));

  QueryClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.QueryConnectionResponse> connection(
      $0.QueryConnectionRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$connection, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryConnectionsResponse> connections(
      $0.QueryConnectionsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$connections, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryClientConnectionsResponse> clientConnections(
      $0.QueryClientConnectionsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$clientConnections, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryConnectionClientStateResponse>
      connectionClientState($0.QueryConnectionClientStateRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$connectionClientState, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryConnectionConsensusStateResponse>
      connectionConsensusState($0.QueryConnectionConsensusStateRequest request,
          {$grpc.CallOptions options}) {
    return $createUnaryCall(_$connectionConsensusState, request,
        options: options);
  }
}

abstract class QueryServiceBase extends $grpc.Service {
  $core.String get $name => 'ibc.core.connection.v1.Query';

  QueryServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.QueryConnectionRequest,
            $0.QueryConnectionResponse>(
        'Connection',
        connection_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryConnectionRequest.fromBuffer(value),
        ($0.QueryConnectionResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryConnectionsRequest,
            $0.QueryConnectionsResponse>(
        'Connections',
        connections_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryConnectionsRequest.fromBuffer(value),
        ($0.QueryConnectionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryClientConnectionsRequest,
            $0.QueryClientConnectionsResponse>(
        'ClientConnections',
        clientConnections_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryClientConnectionsRequest.fromBuffer(value),
        ($0.QueryClientConnectionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryConnectionClientStateRequest,
            $0.QueryConnectionClientStateResponse>(
        'ConnectionClientState',
        connectionClientState_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryConnectionClientStateRequest.fromBuffer(value),
        ($0.QueryConnectionClientStateResponse value) =>
            value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryConnectionConsensusStateRequest,
            $0.QueryConnectionConsensusStateResponse>(
        'ConnectionConsensusState',
        connectionConsensusState_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryConnectionConsensusStateRequest.fromBuffer(value),
        ($0.QueryConnectionConsensusStateResponse value) =>
            value.writeToBuffer()));
  }

  $async.Future<$0.QueryConnectionResponse> connection_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryConnectionRequest> request) async {
    return connection(call, await request);
  }

  $async.Future<$0.QueryConnectionsResponse> connections_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryConnectionsRequest> request) async {
    return connections(call, await request);
  }

  $async.Future<$0.QueryClientConnectionsResponse> clientConnections_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryClientConnectionsRequest> request) async {
    return clientConnections(call, await request);
  }

  $async.Future<$0.QueryConnectionClientStateResponse>
      connectionClientState_Pre($grpc.ServiceCall call,
          $async.Future<$0.QueryConnectionClientStateRequest> request) async {
    return connectionClientState(call, await request);
  }

  $async.Future<$0.QueryConnectionConsensusStateResponse>
      connectionConsensusState_Pre(
          $grpc.ServiceCall call,
          $async.Future<$0.QueryConnectionConsensusStateRequest>
              request) async {
    return connectionConsensusState(call, await request);
  }

  $async.Future<$0.QueryConnectionResponse> connection(
      $grpc.ServiceCall call, $0.QueryConnectionRequest request);
  $async.Future<$0.QueryConnectionsResponse> connections(
      $grpc.ServiceCall call, $0.QueryConnectionsRequest request);
  $async.Future<$0.QueryClientConnectionsResponse> clientConnections(
      $grpc.ServiceCall call, $0.QueryClientConnectionsRequest request);
  $async.Future<$0.QueryConnectionClientStateResponse> connectionClientState(
      $grpc.ServiceCall call, $0.QueryConnectionClientStateRequest request);
  $async.Future<$0.QueryConnectionConsensusStateResponse>
      connectionConsensusState($grpc.ServiceCall call,
          $0.QueryConnectionConsensusStateRequest request);
}
