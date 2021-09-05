///
//  Generated code. Do not modify.
//  source: cosmos/gov/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $1;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$submitProposal =
      $grpc.ClientMethod<$1.MsgSubmitProposal, $1.MsgSubmitProposalResponse>(
          '/cosmos.gov.v1beta1.Msg/SubmitProposal',
          ($1.MsgSubmitProposal value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgSubmitProposalResponse.fromBuffer(value));
  static final _$vote = $grpc.ClientMethod<$1.MsgVote, $1.MsgVoteResponse>(
      '/cosmos.gov.v1beta1.Msg/Vote',
      ($1.MsgVote value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.MsgVoteResponse.fromBuffer(value));
  static final _$deposit =
      $grpc.ClientMethod<$1.MsgDeposit, $1.MsgDepositResponse>(
          '/cosmos.gov.v1beta1.Msg/Deposit',
          ($1.MsgDeposit value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.MsgDepositResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.MsgSubmitProposalResponse> submitProposal(
      $1.MsgSubmitProposal request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$submitProposal, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgVoteResponse> vote($1.MsgVote request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$vote, request, options: options);
  }

  $grpc.ResponseFuture<$1.MsgDepositResponse> deposit($1.MsgDeposit request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$deposit, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.gov.v1beta1.Msg';

  MsgServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$1.MsgSubmitProposal, $1.MsgSubmitProposalResponse>(
            'SubmitProposal',
            submitProposal_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.MsgSubmitProposal.fromBuffer(value),
            ($1.MsgSubmitProposalResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgVote, $1.MsgVoteResponse>(
        'Vote',
        vote_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgVote.fromBuffer(value),
        ($1.MsgVoteResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.MsgDeposit, $1.MsgDepositResponse>(
        'Deposit',
        deposit_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.MsgDeposit.fromBuffer(value),
        ($1.MsgDepositResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MsgSubmitProposalResponse> submitProposal_Pre(
      $grpc.ServiceCall call,
      $async.Future<$1.MsgSubmitProposal> request) async {
    return submitProposal(call, await request);
  }

  $async.Future<$1.MsgVoteResponse> vote_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgVote> request) async {
    return vote(call, await request);
  }

  $async.Future<$1.MsgDepositResponse> deposit_Pre(
      $grpc.ServiceCall call, $async.Future<$1.MsgDeposit> request) async {
    return deposit(call, await request);
  }

  $async.Future<$1.MsgSubmitProposalResponse> submitProposal(
      $grpc.ServiceCall call, $1.MsgSubmitProposal request);
  $async.Future<$1.MsgVoteResponse> vote(
      $grpc.ServiceCall call, $1.MsgVote request);
  $async.Future<$1.MsgDepositResponse> deposit(
      $grpc.ServiceCall call, $1.MsgDeposit request);
}
