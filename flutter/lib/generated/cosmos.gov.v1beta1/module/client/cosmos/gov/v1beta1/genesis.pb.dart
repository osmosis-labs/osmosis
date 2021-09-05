///
//  Generated code. Do not modify.
//  source: cosmos/gov/v1beta1/genesis.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'gov.pb.dart' as $6;

class GenesisState extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GenesisState', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'startingProposalId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..pc<$6.Deposit>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'deposits', $pb.PbFieldType.PM, subBuilder: $6.Deposit.create)
    ..pc<$6.Vote>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'votes', $pb.PbFieldType.PM, subBuilder: $6.Vote.create)
    ..pc<$6.Proposal>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposals', $pb.PbFieldType.PM, subBuilder: $6.Proposal.create)
    ..aOM<$6.DepositParams>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'depositParams', subBuilder: $6.DepositParams.create)
    ..aOM<$6.VotingParams>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'votingParams', subBuilder: $6.VotingParams.create)
    ..aOM<$6.TallyParams>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tallyParams', subBuilder: $6.TallyParams.create)
    ..hasRequiredFields = false
  ;

  GenesisState._() : super();
  factory GenesisState() => create();
  factory GenesisState.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GenesisState.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GenesisState clone() => GenesisState()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GenesisState copyWith(void Function(GenesisState) updates) => super.copyWith((message) => updates(message as GenesisState)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static GenesisState create() => GenesisState._();
  GenesisState createEmptyInstance() => create();
  static $pb.PbList<GenesisState> createRepeated() => $pb.PbList<GenesisState>();
  @$core.pragma('dart2js:noInline')
  static GenesisState getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GenesisState>(create);
  static GenesisState _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get startingProposalId => $_getI64(0);
  @$pb.TagNumber(1)
  set startingProposalId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasStartingProposalId() => $_has(0);
  @$pb.TagNumber(1)
  void clearStartingProposalId() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$6.Deposit> get deposits => $_getList(1);

  @$pb.TagNumber(3)
  $core.List<$6.Vote> get votes => $_getList(2);

  @$pb.TagNumber(4)
  $core.List<$6.Proposal> get proposals => $_getList(3);

  @$pb.TagNumber(5)
  $6.DepositParams get depositParams => $_getN(4);
  @$pb.TagNumber(5)
  set depositParams($6.DepositParams v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasDepositParams() => $_has(4);
  @$pb.TagNumber(5)
  void clearDepositParams() => clearField(5);
  @$pb.TagNumber(5)
  $6.DepositParams ensureDepositParams() => $_ensure(4);

  @$pb.TagNumber(6)
  $6.VotingParams get votingParams => $_getN(5);
  @$pb.TagNumber(6)
  set votingParams($6.VotingParams v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasVotingParams() => $_has(5);
  @$pb.TagNumber(6)
  void clearVotingParams() => clearField(6);
  @$pb.TagNumber(6)
  $6.VotingParams ensureVotingParams() => $_ensure(5);

  @$pb.TagNumber(7)
  $6.TallyParams get tallyParams => $_getN(6);
  @$pb.TagNumber(7)
  set tallyParams($6.TallyParams v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasTallyParams() => $_has(6);
  @$pb.TagNumber(7)
  void clearTallyParams() => clearField(7);
  @$pb.TagNumber(7)
  $6.TallyParams ensureTallyParams() => $_ensure(6);
}

