///
//  Generated code. Do not modify.
//  source: cosmos/gov/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'query.pb.dart' as $0;
export 'query.pb.dart';

class QueryClient extends $grpc.Client {
  static final _$proposal =
      $grpc.ClientMethod<$0.QueryProposalRequest, $0.QueryProposalResponse>(
          '/cosmos.gov.v1beta1.Query/Proposal',
          ($0.QueryProposalRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryProposalResponse.fromBuffer(value));
  static final _$proposals =
      $grpc.ClientMethod<$0.QueryProposalsRequest, $0.QueryProposalsResponse>(
          '/cosmos.gov.v1beta1.Query/Proposals',
          ($0.QueryProposalsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryProposalsResponse.fromBuffer(value));
  static final _$vote =
      $grpc.ClientMethod<$0.QueryVoteRequest, $0.QueryVoteResponse>(
          '/cosmos.gov.v1beta1.Query/Vote',
          ($0.QueryVoteRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryVoteResponse.fromBuffer(value));
  static final _$votes =
      $grpc.ClientMethod<$0.QueryVotesRequest, $0.QueryVotesResponse>(
          '/cosmos.gov.v1beta1.Query/Votes',
          ($0.QueryVotesRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryVotesResponse.fromBuffer(value));
  static final _$params =
      $grpc.ClientMethod<$0.QueryParamsRequest, $0.QueryParamsResponse>(
          '/cosmos.gov.v1beta1.Query/Params',
          ($0.QueryParamsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryParamsResponse.fromBuffer(value));
  static final _$deposit =
      $grpc.ClientMethod<$0.QueryDepositRequest, $0.QueryDepositResponse>(
          '/cosmos.gov.v1beta1.Query/Deposit',
          ($0.QueryDepositRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryDepositResponse.fromBuffer(value));
  static final _$deposits =
      $grpc.ClientMethod<$0.QueryDepositsRequest, $0.QueryDepositsResponse>(
          '/cosmos.gov.v1beta1.Query/Deposits',
          ($0.QueryDepositsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $0.QueryDepositsResponse.fromBuffer(value));
  static final _$tallyResult = $grpc.ClientMethod<$0.QueryTallyResultRequest,
          $0.QueryTallyResultResponse>(
      '/cosmos.gov.v1beta1.Query/TallyResult',
      ($0.QueryTallyResultRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.QueryTallyResultResponse.fromBuffer(value));

  QueryClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.QueryProposalResponse> proposal(
      $0.QueryProposalRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$proposal, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryProposalsResponse> proposals(
      $0.QueryProposalsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$proposals, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryVoteResponse> vote($0.QueryVoteRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$vote, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryVotesResponse> votes(
      $0.QueryVotesRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$votes, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryParamsResponse> params(
      $0.QueryParamsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$params, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryDepositResponse> deposit(
      $0.QueryDepositRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$deposit, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryDepositsResponse> deposits(
      $0.QueryDepositsRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$deposits, request, options: options);
  }

  $grpc.ResponseFuture<$0.QueryTallyResultResponse> tallyResult(
      $0.QueryTallyResultRequest request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$tallyResult, request, options: options);
  }
}

abstract class QueryServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.gov.v1beta1.Query';

  QueryServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$0.QueryProposalRequest, $0.QueryProposalResponse>(
            'Proposal',
            proposal_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.QueryProposalRequest.fromBuffer(value),
            ($0.QueryProposalResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryProposalsRequest,
            $0.QueryProposalsResponse>(
        'Proposals',
        proposals_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryProposalsRequest.fromBuffer(value),
        ($0.QueryProposalsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryVoteRequest, $0.QueryVoteResponse>(
        'Vote',
        vote_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.QueryVoteRequest.fromBuffer(value),
        ($0.QueryVoteResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryVotesRequest, $0.QueryVotesResponse>(
        'Votes',
        votes_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.QueryVotesRequest.fromBuffer(value),
        ($0.QueryVotesResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.QueryParamsRequest, $0.QueryParamsResponse>(
            'Params',
            params_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.QueryParamsRequest.fromBuffer(value),
            ($0.QueryParamsResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.QueryDepositRequest, $0.QueryDepositResponse>(
            'Deposit',
            deposit_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.QueryDepositRequest.fromBuffer(value),
            ($0.QueryDepositResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.QueryDepositsRequest, $0.QueryDepositsResponse>(
            'Deposits',
            deposits_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.QueryDepositsRequest.fromBuffer(value),
            ($0.QueryDepositsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.QueryTallyResultRequest,
            $0.QueryTallyResultResponse>(
        'TallyResult',
        tallyResult_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.QueryTallyResultRequest.fromBuffer(value),
        ($0.QueryTallyResultResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.QueryProposalResponse> proposal_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryProposalRequest> request) async {
    return proposal(call, await request);
  }

  $async.Future<$0.QueryProposalsResponse> proposals_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryProposalsRequest> request) async {
    return proposals(call, await request);
  }

  $async.Future<$0.QueryVoteResponse> vote_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryVoteRequest> request) async {
    return vote(call, await request);
  }

  $async.Future<$0.QueryVotesResponse> votes_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryVotesRequest> request) async {
    return votes(call, await request);
  }

  $async.Future<$0.QueryParamsResponse> params_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryParamsRequest> request) async {
    return params(call, await request);
  }

  $async.Future<$0.QueryDepositResponse> deposit_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryDepositRequest> request) async {
    return deposit(call, await request);
  }

  $async.Future<$0.QueryDepositsResponse> deposits_Pre($grpc.ServiceCall call,
      $async.Future<$0.QueryDepositsRequest> request) async {
    return deposits(call, await request);
  }

  $async.Future<$0.QueryTallyResultResponse> tallyResult_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.QueryTallyResultRequest> request) async {
    return tallyResult(call, await request);
  }

  $async.Future<$0.QueryProposalResponse> proposal(
      $grpc.ServiceCall call, $0.QueryProposalRequest request);
  $async.Future<$0.QueryProposalsResponse> proposals(
      $grpc.ServiceCall call, $0.QueryProposalsRequest request);
  $async.Future<$0.QueryVoteResponse> vote(
      $grpc.ServiceCall call, $0.QueryVoteRequest request);
  $async.Future<$0.QueryVotesResponse> votes(
      $grpc.ServiceCall call, $0.QueryVotesRequest request);
  $async.Future<$0.QueryParamsResponse> params(
      $grpc.ServiceCall call, $0.QueryParamsRequest request);
  $async.Future<$0.QueryDepositResponse> deposit(
      $grpc.ServiceCall call, $0.QueryDepositRequest request);
  $async.Future<$0.QueryDepositsResponse> deposits(
      $grpc.ServiceCall call, $0.QueryDepositsRequest request);
  $async.Future<$0.QueryTallyResultResponse> tallyResult(
      $grpc.ServiceCall call, $0.QueryTallyResultRequest request);
}
