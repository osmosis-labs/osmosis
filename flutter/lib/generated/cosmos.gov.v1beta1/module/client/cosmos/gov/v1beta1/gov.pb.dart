///
//  Generated code. Do not modify.
//  source: cosmos/gov/v1beta1/gov.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../base/v1beta1/coin.pb.dart' as $2;
import '../../../google/protobuf/any.pb.dart' as $3;
import '../../../google/protobuf/timestamp.pb.dart' as $4;
import '../../../google/protobuf/duration.pb.dart' as $5;

import 'gov.pbenum.dart';

export 'gov.pbenum.dart';

class TextProposal extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'TextProposal', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'title')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'description')
    ..hasRequiredFields = false
  ;

  TextProposal._() : super();
  factory TextProposal() => create();
  factory TextProposal.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory TextProposal.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  TextProposal clone() => TextProposal()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  TextProposal copyWith(void Function(TextProposal) updates) => super.copyWith((message) => updates(message as TextProposal)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static TextProposal create() => TextProposal._();
  TextProposal createEmptyInstance() => create();
  static $pb.PbList<TextProposal> createRepeated() => $pb.PbList<TextProposal>();
  @$core.pragma('dart2js:noInline')
  static TextProposal getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TextProposal>(create);
  static TextProposal _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get title => $_getSZ(0);
  @$pb.TagNumber(1)
  set title($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTitle() => $_has(0);
  @$pb.TagNumber(1)
  void clearTitle() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => clearField(2);
}

class Deposit extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Deposit', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposalId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'depositor')
    ..pc<$2.Coin>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', $pb.PbFieldType.PM, subBuilder: $2.Coin.create)
    ..hasRequiredFields = false
  ;

  Deposit._() : super();
  factory Deposit() => create();
  factory Deposit.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Deposit.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Deposit clone() => Deposit()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Deposit copyWith(void Function(Deposit) updates) => super.copyWith((message) => updates(message as Deposit)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Deposit create() => Deposit._();
  Deposit createEmptyInstance() => create();
  static $pb.PbList<Deposit> createRepeated() => $pb.PbList<Deposit>();
  @$core.pragma('dart2js:noInline')
  static Deposit getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Deposit>(create);
  static Deposit _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get proposalId => $_getI64(0);
  @$pb.TagNumber(1)
  set proposalId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasProposalId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProposalId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get depositor => $_getSZ(1);
  @$pb.TagNumber(2)
  set depositor($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasDepositor() => $_has(1);
  @$pb.TagNumber(2)
  void clearDepositor() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$2.Coin> get amount => $_getList(2);
}

class Proposal extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Proposal', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposalId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOM<$3.Any>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'content', subBuilder: $3.Any.create)
    ..e<ProposalStatus>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'status', $pb.PbFieldType.OE, defaultOrMaker: ProposalStatus.PROPOSAL_STATUS_UNSPECIFIED, valueOf: ProposalStatus.valueOf, enumValues: ProposalStatus.values)
    ..aOM<TallyResult>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'finalTallyResult', subBuilder: TallyResult.create)
    ..aOM<$4.Timestamp>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'submitTime', subBuilder: $4.Timestamp.create)
    ..aOM<$4.Timestamp>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'depositEndTime', subBuilder: $4.Timestamp.create)
    ..pc<$2.Coin>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'totalDeposit', $pb.PbFieldType.PM, subBuilder: $2.Coin.create)
    ..aOM<$4.Timestamp>(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'votingStartTime', subBuilder: $4.Timestamp.create)
    ..aOM<$4.Timestamp>(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'votingEndTime', subBuilder: $4.Timestamp.create)
    ..hasRequiredFields = false
  ;

  Proposal._() : super();
  factory Proposal() => create();
  factory Proposal.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Proposal.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Proposal clone() => Proposal()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Proposal copyWith(void Function(Proposal) updates) => super.copyWith((message) => updates(message as Proposal)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Proposal create() => Proposal._();
  Proposal createEmptyInstance() => create();
  static $pb.PbList<Proposal> createRepeated() => $pb.PbList<Proposal>();
  @$core.pragma('dart2js:noInline')
  static Proposal getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Proposal>(create);
  static Proposal _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get proposalId => $_getI64(0);
  @$pb.TagNumber(1)
  set proposalId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasProposalId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProposalId() => clearField(1);

  @$pb.TagNumber(2)
  $3.Any get content => $_getN(1);
  @$pb.TagNumber(2)
  set content($3.Any v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasContent() => $_has(1);
  @$pb.TagNumber(2)
  void clearContent() => clearField(2);
  @$pb.TagNumber(2)
  $3.Any ensureContent() => $_ensure(1);

  @$pb.TagNumber(3)
  ProposalStatus get status => $_getN(2);
  @$pb.TagNumber(3)
  set status(ProposalStatus v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasStatus() => $_has(2);
  @$pb.TagNumber(3)
  void clearStatus() => clearField(3);

  @$pb.TagNumber(4)
  TallyResult get finalTallyResult => $_getN(3);
  @$pb.TagNumber(4)
  set finalTallyResult(TallyResult v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasFinalTallyResult() => $_has(3);
  @$pb.TagNumber(4)
  void clearFinalTallyResult() => clearField(4);
  @$pb.TagNumber(4)
  TallyResult ensureFinalTallyResult() => $_ensure(3);

  @$pb.TagNumber(5)
  $4.Timestamp get submitTime => $_getN(4);
  @$pb.TagNumber(5)
  set submitTime($4.Timestamp v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasSubmitTime() => $_has(4);
  @$pb.TagNumber(5)
  void clearSubmitTime() => clearField(5);
  @$pb.TagNumber(5)
  $4.Timestamp ensureSubmitTime() => $_ensure(4);

  @$pb.TagNumber(6)
  $4.Timestamp get depositEndTime => $_getN(5);
  @$pb.TagNumber(6)
  set depositEndTime($4.Timestamp v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasDepositEndTime() => $_has(5);
  @$pb.TagNumber(6)
  void clearDepositEndTime() => clearField(6);
  @$pb.TagNumber(6)
  $4.Timestamp ensureDepositEndTime() => $_ensure(5);

  @$pb.TagNumber(7)
  $core.List<$2.Coin> get totalDeposit => $_getList(6);

  @$pb.TagNumber(8)
  $4.Timestamp get votingStartTime => $_getN(7);
  @$pb.TagNumber(8)
  set votingStartTime($4.Timestamp v) { setField(8, v); }
  @$pb.TagNumber(8)
  $core.bool hasVotingStartTime() => $_has(7);
  @$pb.TagNumber(8)
  void clearVotingStartTime() => clearField(8);
  @$pb.TagNumber(8)
  $4.Timestamp ensureVotingStartTime() => $_ensure(7);

  @$pb.TagNumber(9)
  $4.Timestamp get votingEndTime => $_getN(8);
  @$pb.TagNumber(9)
  set votingEndTime($4.Timestamp v) { setField(9, v); }
  @$pb.TagNumber(9)
  $core.bool hasVotingEndTime() => $_has(8);
  @$pb.TagNumber(9)
  void clearVotingEndTime() => clearField(9);
  @$pb.TagNumber(9)
  $4.Timestamp ensureVotingEndTime() => $_ensure(8);
}

class TallyResult extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'TallyResult', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'yes')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'abstain')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'no')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'noWithVeto')
    ..hasRequiredFields = false
  ;

  TallyResult._() : super();
  factory TallyResult() => create();
  factory TallyResult.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory TallyResult.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  TallyResult clone() => TallyResult()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  TallyResult copyWith(void Function(TallyResult) updates) => super.copyWith((message) => updates(message as TallyResult)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static TallyResult create() => TallyResult._();
  TallyResult createEmptyInstance() => create();
  static $pb.PbList<TallyResult> createRepeated() => $pb.PbList<TallyResult>();
  @$core.pragma('dart2js:noInline')
  static TallyResult getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TallyResult>(create);
  static TallyResult _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get yes => $_getSZ(0);
  @$pb.TagNumber(1)
  set yes($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasYes() => $_has(0);
  @$pb.TagNumber(1)
  void clearYes() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get abstain => $_getSZ(1);
  @$pb.TagNumber(2)
  set abstain($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasAbstain() => $_has(1);
  @$pb.TagNumber(2)
  void clearAbstain() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get no => $_getSZ(2);
  @$pb.TagNumber(3)
  set no($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasNo() => $_has(2);
  @$pb.TagNumber(3)
  void clearNo() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get noWithVeto => $_getSZ(3);
  @$pb.TagNumber(4)
  set noWithVeto($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasNoWithVeto() => $_has(3);
  @$pb.TagNumber(4)
  void clearNoWithVeto() => clearField(4);
}

class Vote extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Vote', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposalId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'voter')
    ..e<VoteOption>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'option', $pb.PbFieldType.OE, defaultOrMaker: VoteOption.VOTE_OPTION_UNSPECIFIED, valueOf: VoteOption.valueOf, enumValues: VoteOption.values)
    ..hasRequiredFields = false
  ;

  Vote._() : super();
  factory Vote() => create();
  factory Vote.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Vote.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Vote clone() => Vote()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Vote copyWith(void Function(Vote) updates) => super.copyWith((message) => updates(message as Vote)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Vote create() => Vote._();
  Vote createEmptyInstance() => create();
  static $pb.PbList<Vote> createRepeated() => $pb.PbList<Vote>();
  @$core.pragma('dart2js:noInline')
  static Vote getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Vote>(create);
  static Vote _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get proposalId => $_getI64(0);
  @$pb.TagNumber(1)
  set proposalId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasProposalId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProposalId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get voter => $_getSZ(1);
  @$pb.TagNumber(2)
  set voter($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasVoter() => $_has(1);
  @$pb.TagNumber(2)
  void clearVoter() => clearField(2);

  @$pb.TagNumber(3)
  VoteOption get option => $_getN(2);
  @$pb.TagNumber(3)
  set option(VoteOption v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasOption() => $_has(2);
  @$pb.TagNumber(3)
  void clearOption() => clearField(3);
}

class DepositParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DepositParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..pc<$2.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'minDeposit', $pb.PbFieldType.PM, subBuilder: $2.Coin.create)
    ..aOM<$5.Duration>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxDepositPeriod', subBuilder: $5.Duration.create)
    ..hasRequiredFields = false
  ;

  DepositParams._() : super();
  factory DepositParams() => create();
  factory DepositParams.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DepositParams.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DepositParams clone() => DepositParams()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DepositParams copyWith(void Function(DepositParams) updates) => super.copyWith((message) => updates(message as DepositParams)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static DepositParams create() => DepositParams._();
  DepositParams createEmptyInstance() => create();
  static $pb.PbList<DepositParams> createRepeated() => $pb.PbList<DepositParams>();
  @$core.pragma('dart2js:noInline')
  static DepositParams getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DepositParams>(create);
  static DepositParams _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.Coin> get minDeposit => $_getList(0);

  @$pb.TagNumber(2)
  $5.Duration get maxDepositPeriod => $_getN(1);
  @$pb.TagNumber(2)
  set maxDepositPeriod($5.Duration v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasMaxDepositPeriod() => $_has(1);
  @$pb.TagNumber(2)
  void clearMaxDepositPeriod() => clearField(2);
  @$pb.TagNumber(2)
  $5.Duration ensureMaxDepositPeriod() => $_ensure(1);
}

class VotingParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'VotingParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..aOM<$5.Duration>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'votingPeriod', subBuilder: $5.Duration.create)
    ..hasRequiredFields = false
  ;

  VotingParams._() : super();
  factory VotingParams() => create();
  factory VotingParams.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory VotingParams.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  VotingParams clone() => VotingParams()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  VotingParams copyWith(void Function(VotingParams) updates) => super.copyWith((message) => updates(message as VotingParams)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static VotingParams create() => VotingParams._();
  VotingParams createEmptyInstance() => create();
  static $pb.PbList<VotingParams> createRepeated() => $pb.PbList<VotingParams>();
  @$core.pragma('dart2js:noInline')
  static VotingParams getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VotingParams>(create);
  static VotingParams _defaultInstance;

  @$pb.TagNumber(1)
  $5.Duration get votingPeriod => $_getN(0);
  @$pb.TagNumber(1)
  set votingPeriod($5.Duration v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasVotingPeriod() => $_has(0);
  @$pb.TagNumber(1)
  void clearVotingPeriod() => clearField(1);
  @$pb.TagNumber(1)
  $5.Duration ensureVotingPeriod() => $_ensure(0);
}

class TallyParams extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'TallyParams', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'quorum', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'threshold', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'vetoThreshold', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  TallyParams._() : super();
  factory TallyParams() => create();
  factory TallyParams.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory TallyParams.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  TallyParams clone() => TallyParams()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  TallyParams copyWith(void Function(TallyParams) updates) => super.copyWith((message) => updates(message as TallyParams)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static TallyParams create() => TallyParams._();
  TallyParams createEmptyInstance() => create();
  static $pb.PbList<TallyParams> createRepeated() => $pb.PbList<TallyParams>();
  @$core.pragma('dart2js:noInline')
  static TallyParams getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TallyParams>(create);
  static TallyParams _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get quorum => $_getN(0);
  @$pb.TagNumber(1)
  set quorum($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasQuorum() => $_has(0);
  @$pb.TagNumber(1)
  void clearQuorum() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get threshold => $_getN(1);
  @$pb.TagNumber(2)
  set threshold($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasThreshold() => $_has(1);
  @$pb.TagNumber(2)
  void clearThreshold() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get vetoThreshold => $_getN(2);
  @$pb.TagNumber(3)
  set vetoThreshold($core.List<$core.int> v) { $_setBytes(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasVetoThreshold() => $_has(2);
  @$pb.TagNumber(3)
  void clearVetoThreshold() => clearField(3);
}

