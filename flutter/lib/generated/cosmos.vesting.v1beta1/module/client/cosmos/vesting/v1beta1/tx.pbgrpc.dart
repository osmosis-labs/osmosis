///
//  Generated code. Do not modify.
//  source: cosmos/vesting/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'tx.pb.dart' as $0;
export 'tx.pb.dart';

class MsgClient extends $grpc.Client {
  static final _$createVestingAccount = $grpc.ClientMethod<
          $0.MsgCreateVestingAccount, $0.MsgCreateVestingAccountResponse>(
      '/cosmos.vesting.v1beta1.Msg/CreateVestingAccount',
      ($0.MsgCreateVestingAccount value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $0.MsgCreateVestingAccountResponse.fromBuffer(value));

  MsgClient($grpc.ClientChannel channel,
      {$grpc.CallOptions options,
      $core.Iterable<$grpc.ClientInterceptor> interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.MsgCreateVestingAccountResponse> createVestingAccount(
      $0.MsgCreateVestingAccount request,
      {$grpc.CallOptions options}) {
    return $createUnaryCall(_$createVestingAccount, request, options: options);
  }
}

abstract class MsgServiceBase extends $grpc.Service {
  $core.String get $name => 'cosmos.vesting.v1beta1.Msg';

  MsgServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.MsgCreateVestingAccount,
            $0.MsgCreateVestingAccountResponse>(
        'CreateVestingAccount',
        createVestingAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.MsgCreateVestingAccount.fromBuffer(value),
        ($0.MsgCreateVestingAccountResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.MsgCreateVestingAccountResponse> createVestingAccount_Pre(
      $grpc.ServiceCall call,
      $async.Future<$0.MsgCreateVestingAccount> request) async {
    return createVestingAccount(call, await request);
  }

  $async.Future<$0.MsgCreateVestingAccountResponse> createVestingAccount(
      $grpc.ServiceCall call, $0.MsgCreateVestingAccount request);
}
