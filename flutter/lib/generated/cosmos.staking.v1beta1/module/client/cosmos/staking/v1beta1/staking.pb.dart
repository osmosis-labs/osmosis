///
//  Generated code. Do not modify.
//  source: cosmos/staking/v1beta1/staking.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../../tendermint/types/types.pb.dart' as $7;
import '../../../google/protobuf/timestamp.pb.dart' as $5;
import '../../../google/protobuf/any.pb.dart' as $8;
import '../../../google/protobuf/duration.pb.dart' as $9;
import '../../base/v1beta1/coin.pb.dart' as $10;

import 'staking.pbenum.dart';

export 'staking.pbenum.dart';

class HistoricalInfo extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'HistoricalInfo', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOM<$7.Header>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'header', subBuilder: $7.Header.create)
    ..pc<Validator>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'valset', $pb.PbFieldType.PM, subBuilder: Validator.create)
    ..hasRequiredFields = false
  ;

  HistoricalInfo._() : super();
  factory HistoricalInfo() => create();
  factory HistoricalInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory HistoricalInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  HistoricalInfo clone() => HistoricalInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  HistoricalInfo copyWith(void Function(HistoricalInfo) updates) => super.copyWith((message) => updates(message as HistoricalInfo)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static HistoricalInfo create() => HistoricalInfo._();
  HistoricalInfo createEmptyInstance() => create();
  static $pb.PbList<HistoricalInfo> createRepeated() => $pb.PbList<HistoricalInfo>();
  @$core.pragma('dart2js:noInline')
  static HistoricalInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<HistoricalInfo>(create);
  static HistoricalInfo _defaultInstance;

  @$pb.TagNumber(1)
  $7.Header get header => $_getN(0);
  @$pb.TagNumber(1)
  set header($7.Header v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasHeader() => $_has(0);
  @$pb.TagNumber(1)
  void clearHeader() => clearField(1);
  @$pb.TagNumber(1)
  $7.Header ensureHeader() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<Validator> get valset => $_getList(1);
}

class CommissionRates extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CommissionRates', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'rate')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxRate')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxChangeRate')
    ..hasRequiredFields = false
  ;

  CommissionRates._() : super();
  factory CommissionRates() => create();
  factory CommissionRates.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CommissionRates.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CommissionRates clone() => CommissionRates()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CommissionRates copyWith(void Function(CommissionRates) updates) => super.copyWith((message) => updates(message as CommissionRates)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CommissionRates create() => CommissionRates._();
  CommissionRates createEmptyInstance() => create();
  static $pb.PbList<CommissionRates> createRepeated() => $pb.PbList<CommissionRates>();
  @$core.pragma('dart2js:noInline')
  static CommissionRates getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CommissionRates>(create);
  static CommissionRates _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get rate => $_getSZ(0);
  @$pb.TagNumber(1)
  set rate($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasRate() => $_has(0);
  @$pb.TagNumber(1)
  void clearRate() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get maxRate => $_getSZ(1);
  @$pb.TagNumber(2)
  set maxRate($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasMaxRate() => $_has(1);
  @$pb.TagNumber(2)
  void clearMaxRate() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get maxChangeRate => $_getSZ(2);
  @$pb.TagNumber(3)
  set maxChangeRate($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasMaxChangeRate() => $_has(2);
  @$pb.TagNumber(3)
  void clearMaxChangeRate() => clearField(3);
}

class Commission extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Commission', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOM<CommissionRates>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'commissionRates', subBuilder: CommissionRates.create)
    ..aOM<$5.Timestamp>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'updateTime', subBuilder: $5.Timestamp.create)
    ..hasRequiredFields = false
  ;

  Commission._() : super();
  factory Commission() => create();
  factory Commission.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Commission.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Commission clone() => Commission()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Commission copyWith(void Function(Commission) updates) => super.copyWith((message) => updates(message as Commission)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Commission create() => Commission._();
  Commission createEmptyInstance() => create();
  static $pb.PbList<Commission> createRepeated() => $pb.PbList<Commission>();
  @$core.pragma('dart2js:noInline')
  static Commission getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Commission>(create);
  static Commission _defaultInstance;

  @$pb.TagNumber(1)
  CommissionRates get commissionRates => $_getN(0);
  @$pb.TagNumber(1)
  set commissionRates(CommissionRates v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasCommissionRates() => $_has(0);
  @$pb.TagNumber(1)
  void clearCommissionRates() => clearField(1);
  @$pb.TagNumber(1)
  CommissionRates ensureCommissionRates() => $_ensure(0);

  @$pb.TagNumber(2)
  $5.Timestamp get updateTime => $_getN(1);
  @$pb.TagNumber(2)
  set updateTime($5.Timestamp v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasUpdateTime() => $_has(1);
  @$pb.TagNumber(2)
  void clearUpdateTime() => clearField(2);
  @$pb.TagNumber(2)
  $5.Timestamp ensureUpdateTime() => $_ensure(1);
}

class Description extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Description', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'moniker')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'identity')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'website')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'securityContact')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'details')
    ..hasRequiredFields = false
  ;

  Description._() : super();
  factory Description() => create();
  factory Description.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Description.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Description clone() => Description()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Description copyWith(void Function(Description) updates) => super.copyWith((message) => updates(message as Description)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Description create() => Description._();
  Description createEmptyInstance() => create();
  static $pb.PbList<Description> createRepeated() => $pb.PbList<Description>();
  @$core.pragma('dart2js:noInline')
  static Description getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Description>(create);
  static Description _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get moniker => $_getSZ(0);
  @$pb.TagNumber(1)
  set moniker($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMoniker() => $_has(0);
  @$pb.TagNumber(1)
  void clearMoniker() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get identity => $_getSZ(1);
  @$pb.TagNumber(2)
  set identity($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasIdentity() => $_has(1);
  @$pb.TagNumber(2)
  void clearIdentity() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get website => $_getSZ(2);
  @$pb.TagNumber(3)
  set website($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasWebsite() => $_has(2);
  @$pb.TagNumber(3)
  void clearWebsite() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get securityContact => $_getSZ(3);
  @$pb.TagNumber(4)
  set securityContact($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasSecurityContact() => $_has(3);
  @$pb.TagNumber(4)
  void clearSecurityContact() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get details => $_getSZ(4);
  @$pb.TagNumber(5)
  set details($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasDetails() => $_has(4);
  @$pb.TagNumber(5)
  void clearDetails() => clearField(5);
}

class Validator extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Validator', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'operatorAddress')
    ..aOM<$8.Any>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'consensusPubkey', subBuilder: $8.Any.create)
    ..aOB(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'jailed')
    ..e<BondStatus>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'status', $pb.PbFieldType.OE, defaultOrMaker: BondStatus.BOND_STATUS_UNSPECIFIED, valueOf: BondStatus.valueOf, enumValues: BondStatus.values)
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokens')
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorShares')
    ..aOM<Description>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'description', subBuilder: Description.create)
    ..aInt64(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'unbondingHeight')
    ..aOM<$5.Timestamp>(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'unbondingTime', subBuilder: $5.Timestamp.create)
    ..aOM<Commission>(10, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'commission', subBuilder: Commission.create)
    ..aOS(11, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'minSelfDelegation')
    ..hasRequiredFields = false
  ;

  Validator._() : super();
  factory Validator() => create();
  factory Validator.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Validator.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Validator clone() => Validator()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Validator copyWith(void Function(Validator) updates) => super.copyWith((message) => updates(message as Validator)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Validator create() => Validator._();
  Validator createEmptyInstance() => create();
  static $pb.PbList<Validator> createRepeated() => $pb.PbList<Validator>();
  @$core.pragma('dart2js:noInline')
  static Validator getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Validator>(create);
  static Validator _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get operatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set operatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOperatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearOperatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $8.Any get consensusPubkey => $_getN(1);
  @$pb.TagNumber(2)
  set consensusPubkey($8.Any v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasConsensusPubkey() => $_has(1);
  @$pb.TagNumber(2)
  void clearConsensusPubkey() => clearField(2);
  @$pb.TagNumber(2)
  $8.Any ensureConsensusPubkey() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.bool get jailed => $_getBF(2);
  @$pb.TagNumber(3)
  set jailed($core.bool v) { $_setBool(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasJailed() => $_has(2);
  @$pb.TagNumber(3)
  void clearJailed() => clearField(3);

  @$pb.TagNumber(4)
  BondStatus get status => $_getN(3);
  @$pb.TagNumber(4)
  set status(BondStatus v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasStatus() => $_has(3);
  @$pb.TagNumber(4)
  void clearStatus() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get tokens => $_getSZ(4);
  @$pb.TagNumber(5)
  set tokens($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasTokens() => $_has(4);
  @$pb.TagNumber(5)
  void clearTokens() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get delegatorShares => $_getSZ(5);
  @$pb.TagNumber(6)
  set delegatorShares($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasDelegatorShares() => $_has(5);
  @$pb.TagNumber(6)
  void clearDelegatorShares() => clearField(6);

  @$pb.TagNumber(7)
  Description get description => $_getN(6);
  @$pb.TagNumber(7)
  set description(Description v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasDescription() => $_has(6);
  @$pb.TagNumber(7)
  void clearDescription() => clearField(7);
  @$pb.TagNumber(7)
  Description ensureDescription() => $_ensure(6);

  @$pb.TagNumber(8)
  $fixnum.Int64 get unbondingHeight => $_getI64(7);
  @$pb.TagNumber(8)
  set unbondingHeight($fixnum.Int64 v) { $_setInt64(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasUnbondingHeight() => $_has(7);
  @$pb.TagNumber(8)
  void clearUnbondingHeight() => clearField(8);

  @$pb.TagNumber(9)
  $5.Timestamp get unbondingTime => $_getN(8);
  @$pb.TagNumber(9)
  set unbondingTime($5.Timestamp v) { setField(9, v); }
  @$pb.TagNumber(9)
  $core.bool hasUnbondingTime() => $_has(8);
  @$pb.TagNumber(9)
  void clearUnbondingTime() => clearField(9);
  @$pb.TagNumber(9)
  $5.Timestamp ensureUnbondingTime() => $_ensure(8);

  @$pb.TagNumber(10)
  Commission get commission => $_getN(9);
  @$pb.TagNumber(10)
  set commission(Commission v) { setField(10, v); }
  @$pb.TagNumber(10)
  $core.bool hasCommission() => $_has(9);
  @$pb.TagNumber(10)
  void clearCommission() => clearField(10);
  @$pb.TagNumber(10)
  Commission ensureCommission() => $_ensure(9);

  @$pb.TagNumber(11)
  $core.String get minSelfDelegation => $_getSZ(10);
  @$pb.TagNumber(11)
  set minSelfDelegation($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasMinSelfDelegation() => $_has(10);
  @$pb.TagNumber(11)
  void clearMinSelfDelegation() => clearField(11);
}

class ValAddresses extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ValAddresses', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..pPS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'addresses')
    ..hasRequiredFields = false
  ;

  ValAddresses._() : super();
  factory ValAddresses() => create();
  factory ValAddresses.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ValAddresses.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ValAddresses clone() => ValAddresses()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ValAddresses copyWith(void Function(ValAddresses) updates) => super.copyWith((message) => updates(message as ValAddresses)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ValAddresses create() => ValAddresses._();
  ValAddresses createEmptyInstance() => create();
  static $pb.PbList<ValAddresses> createRepeated() => $pb.PbList<ValAddresses>();
  @$core.pragma('dart2js:noInline')
  static ValAddresses getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ValAddresses>(create);
  static ValAddresses _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.String> get addresses => $_getList(0);
}

class DVPair extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DVPair', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..hasRequiredFields = false
  ;

  DVPair._() : super();
  factory DVPair() => create();
  factory DVPair.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DVPair.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DVPair clone() => DVPair()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DVPair copyWith(void Function(DVPair) updates) => super.copyWith((message) => updates(message as DVPair)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static DVPair create() => DVPair._();
  DVPair createEmptyInstance() => create();
  static $pb.PbList<DVPair> createRepeated() => $pb.PbList<DVPair>();
  @$core.pragma('dart2js:noInline')
  static DVPair getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DVPair>(create);
  static DVPair _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get delegatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set delegatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get validatorAddress => $_getSZ(1);
  @$pb.TagNumber(2)
  set validatorAddress($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasValidatorAddress() => $_has(1);
  @$pb.TagNumber(2)
  void clearValidatorAddress() => clearField(2);
}

class DVPairs extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DVPairs', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..pc<DVPair>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pairs', $pb.PbFieldType.PM, subBuilder: DVPair.create)
    ..hasRequiredFields = false
  ;

  DVPairs._() : super();
  factory DVPairs() => create();
  factory DVPairs.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DVPairs.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DVPairs clone() => DVPairs()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DVPairs copyWith(void Function(DVPairs) updates) => super.copyWith((message) => updates(message as DVPairs)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static DVPairs create() => DVPairs._();
  DVPairs createEmptyInstance() => create();
  static $pb.PbList<DVPairs> createRepeated() => $pb.PbList<DVPairs>();
  @$core.pragma('dart2js:noInline')
  static DVPairs getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DVPairs>(create);
  static DVPairs _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<DVPair> get pairs => $_getList(0);
}

class DVVTriplet extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DVVTriplet', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorSrcAddress')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorDstAddress')
    ..hasRequiredFields = false
  ;

  DVVTriplet._() : super();
  factory DVVTriplet() => create();
  factory DVVTriplet.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DVVTriplet.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DVVTriplet clone() => DVVTriplet()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DVVTriplet copyWith(void Function(DVVTriplet) updates) => super.copyWith((message) => updates(message as DVVTriplet)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static DVVTriplet create() => DVVTriplet._();
  DVVTriplet createEmptyInstance() => create();
  static $pb.PbList<DVVTriplet> createRepeated() => $pb.PbList<DVVTriplet>();
  @$core.pragma('dart2js:noInline')
  static DVVTriplet getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DVVTriplet>(create);
  static DVVTriplet _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get delegatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set delegatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get validatorSrcAddress => $_getSZ(1);
  @$pb.TagNumber(2)
  set validatorSrcAddress($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasValidatorSrcAddress() => $_has(1);
  @$pb.TagNumber(2)
  void clearValidatorSrcAddress() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get validatorDstAddress => $_getSZ(2);
  @$pb.TagNumber(3)
  set validatorDstAddress($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasValidatorDstAddress() => $_has(2);
  @$pb.TagNumber(3)
  void clearValidatorDstAddress() => clearField(3);
}

class DVVTriplets extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DVVTriplets', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..pc<DVVTriplet>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'triplets', $pb.PbFieldType.PM, subBuilder: DVVTriplet.create)
    ..hasRequiredFields = false
  ;

  DVVTriplets._() : super();
  factory DVVTriplets() => create();
  factory DVVTriplets.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DVVTriplets.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DVVTriplets clone() => DVVTriplets()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DVVTriplets copyWith(void Function(DVVTriplets) updates) => super.copyWith((message) => updates(message as DVVTriplets)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static DVVTriplets create() => DVVTriplets._();
  DVVTriplets createEmptyInstance() => create();
  static $pb.PbList<DVVTriplets> createRepeated() => $pb.PbList<DVVTriplets>();
  @$core.pragma('dart2js:noInline')
  static DVVTriplets getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DVVTriplets>(create);
  static DVVTriplets _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<DVVTriplet> get triplets => $_getList(0);
}

class Delegation extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Delegation', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'shares')
    ..hasRequiredFields = false
  ;

  Delegation._() : super();
  factory Delegation() => create();
  factory Delegation.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Delegation.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Delegation clone() => Delegation()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Delegation copyWith(void Function(Delegation) updates) => super.copyWith((message) => updates(message as Delegation)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Delegation create() => Delegation._();
  Delegation createEmptyInstance() => create();
  static $pb.PbList<Delegation> createRepeated() => $pb.PbList<Delegation>();
  @$core.pragma('dart2js:noInline')
  static Delegation getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Delegation>(create);
  static Delegation _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get delegatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set delegatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get validatorAddress => $_getSZ(1);
  @$pb.TagNumber(2)
  set validatorAddress($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasValidatorAddress() => $_has(1);
  @$pb.TagNumber(2)
  void clearValidatorAddress() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get shares => $_getSZ(2);
  @$pb.TagNumber(3)
  set shares($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasShares() => $_has(2);
  @$pb.TagNumber(3)
  void clearShares() => clearField(3);
}

class UnbondingDelegation extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'UnbondingDelegation', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..pc<UnbondingDelegationEntry>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'entries', $pb.PbFieldType.PM, subBuilder: UnbondingDelegationEntry.create)
    ..hasRequiredFields = false
  ;

  UnbondingDelegation._() : super();
  factory UnbondingDelegation() => create();
  factory UnbondingDelegation.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UnbondingDelegation.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UnbondingDelegation clone() => UnbondingDelegation()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UnbondingDelegation copyWith(void Function(UnbondingDelegation) updates) => super.copyWith((message) => updates(message as UnbondingDelegation)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static UnbondingDelegation create() => UnbondingDelegation._();
  UnbondingDelegation createEmptyInstance() => create();
  static $pb.PbList<UnbondingDelegation> createRepeated() => $pb.PbList<UnbondingDelegation>();
  @$core.pragma('dart2js:noInline')
  static UnbondingDelegation getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UnbondingDelegation>(create);
  static UnbondingDelegation _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get delegatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set delegatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get validatorAddress => $_getSZ(1);
  @$pb.TagNumber(2)
  set validatorAddress($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasValidatorAddress() => $_has(1);
  @$pb.TagNumber(2)
  void clearValidatorAddress() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<UnbondingDelegationEntry> get entries => $_getList(2);
}

class UnbondingDelegationEntry extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'UnbondingDelegationEntry', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aInt64(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'creationHeight')
    ..aOM<$5.Timestamp>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'completionTime', subBuilder: $5.Timestamp.create)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'initialBalance')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'balance')
    ..hasRequiredFields = false
  ;

  UnbondingDelegationEntry._() : super();
  factory UnbondingDelegationEntry() => create();
  factory UnbondingDelegationEntry.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UnbondingDelegationEntry.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UnbondingDelegationEntry clone() => UnbondingDelegationEntry()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UnbondingDelegationEntry copyWith(void Function(UnbondingDelegationEntry) updates) => super.copyWith((message) => updates(message as UnbondingDelegationEntry)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static UnbondingDelegationEntry create() => UnbondingDelegationEntry._();
  UnbondingDelegationEntry createEmptyInstance() => create();
  static $pb.PbList<UnbondingDelegationEntry> createRepeated() => $pb.PbList<UnbondingDelegationEntry>();
  @$core.pragma('dart2js:noInline')
  static UnbondingDelegationEntry getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UnbondingDelegationEntry>(create);
  static UnbondingDelegationEntry _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get creationHeight => $_getI64(0);
  @$pb.TagNumber(1)
  set creationHeight($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCreationHeight() => $_has(0);
  @$pb.TagNumber(1)
  void clearCreationHeight() => clearField(1);

  @$pb.TagNumber(2)
  $5.Timestamp get completionTime => $_getN(1);
  @$pb.TagNumber(2)
  set completionTime($5.Timestamp v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasCompletionTime() => $_has(1);
  @$pb.TagNumber(2)
  void clearCompletionTime() => clearField(2);
  @$pb.TagNumber(2)
  $5.Timestamp ensureCompletionTime() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.String get initialBalance => $_getSZ(2);
  @$pb.TagNumber(3)
  set initialBalance($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasInitialBalance() => $_has(2);
  @$pb.TagNumber(3)
  void clearInitialBalance() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get balance => $_getSZ(3);
  @$pb.TagNumber(4)
  set balance($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasBalance() => $_has(3);
  @$pb.TagNumber(4)
  void clearBalance() => clearField(4);
}

class RedelegationEntry extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RedelegationEntry', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aInt64(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'creationHeight')
    ..aOM<$5.Timestamp>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'completionTime', subBuilder: $5.Timestamp.create)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'initialBalance')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sharesDst')
    ..hasRequiredFields = false
  ;

  RedelegationEntry._() : super();
  factory RedelegationEntry() => create();
  factory RedelegationEntry.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RedelegationEntry.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RedelegationEntry clone() => RedelegationEntry()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RedelegationEntry copyWith(void Function(RedelegationEntry) updates) => super.copyWith((message) => updates(message as RedelegationEntry)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RedelegationEntry create() => RedelegationEntry._();
  RedelegationEntry createEmptyInstance() => create();
  static $pb.PbList<RedelegationEntry> createRepeated() => $pb.PbList<RedelegationEntry>();
  @$core.pragma('dart2js:noInline')
  static RedelegationEntry getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RedelegationEntry>(create);
  static RedelegationEntry _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get creationHeight => $_getI64(0);
  @$pb.TagNumber(1)
  set creationHeight($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCreationHeight() => $_has(0);
  @$pb.TagNumber(1)
  void clearCreationHeight() => clearField(1);

  @$pb.TagNumber(2)
  $5.Timestamp get completionTime => $_getN(1);
  @$pb.TagNumber(2)
  set completionTime($5.Timestamp v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasCompletionTime() => $_has(1);
  @$pb.TagNumber(2)
  void clearCompletionTime() => clearField(2);
  @$pb.TagNumber(2)
  $5.Timestamp ensureCompletionTime() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.String get initialBalance => $_getSZ(2);
  @$pb.TagNumber(3)
  set initialBalance($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasInitialBalance() => $_has(2);
  @$pb.TagNumber(3)
  void clearInitialBalance() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get sharesDst => $_getSZ(3);
  @$pb.TagNumber(4)
  set sharesDst($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasSharesDst() => $_has(3);
  @$pb.TagNumber(4)
  void clearSharesDst() => clearField(4);
}

class Redelegation extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Redelegation', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorSrcAddress')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorDstAddress')
    ..pc<RedelegationEntry>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'entries', $pb.PbFieldType.PM, subBuilder: RedelegationEntry.create)
    ..hasRequiredFields = false
  ;

  Redelegation._() : super();
  factory Redelegation() => create();
  factory Redelegation.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Redelegation.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Redelegation clone() => Redelegation()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Redelegation copyWith(void Function(Redelegation) updates) => super.copyWith((message) => updates(message as Redelegation)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Redelegation create() => Redelegation._();
  Redelegation createEmptyInstance() => create();
  static $pb.PbList<Redelegation> createRepeated() => $pb.PbList<Redelegation>();
  @$core.pragma('dart2js:noInline')
  static Redelegation getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Redelegation>(create);
  static Redelegation _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get delegatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set delegatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get validatorSrcAddress => $_getSZ(1);
  @$pb.TagNumber(2)
  set validatorSrcAddress($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasValidatorSrcAddress() => $_has(1);
  @$pb.TagNumber(2)
  void clearValidatorSrcAddress() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get validatorDstAddress => $_getSZ(2);
  @$pb.TagNumber(3)
  set validatorDstAddress($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasValidatorDstAddress() => $_has(2);
  @$pb.TagNumber(3)
  void clearValidatorDstAddress() => clearField(3);

  @$pb.TagNumber(4)
  $core.List<RedelegationEntry> get entries => $_getList(3);
}

class Params extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Params', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOM<$9.Duration>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'unbondingTime', subBuilder: $9.Duration.create)
    ..a<$core.int>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxValidators', $pb.PbFieldType.OU3)
    ..a<$core.int>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxEntries', $pb.PbFieldType.OU3)
    ..a<$core.int>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'historicalEntries', $pb.PbFieldType.OU3)
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'bondDenom')
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'minCommissionRate')
    ..hasRequiredFields = false
  ;

  Params._() : super();
  factory Params() => create();
  factory Params.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Params.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Params clone() => Params()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Params copyWith(void Function(Params) updates) => super.copyWith((message) => updates(message as Params)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Params create() => Params._();
  Params createEmptyInstance() => create();
  static $pb.PbList<Params> createRepeated() => $pb.PbList<Params>();
  @$core.pragma('dart2js:noInline')
  static Params getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Params>(create);
  static Params _defaultInstance;

  @$pb.TagNumber(1)
  $9.Duration get unbondingTime => $_getN(0);
  @$pb.TagNumber(1)
  set unbondingTime($9.Duration v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasUnbondingTime() => $_has(0);
  @$pb.TagNumber(1)
  void clearUnbondingTime() => clearField(1);
  @$pb.TagNumber(1)
  $9.Duration ensureUnbondingTime() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.int get maxValidators => $_getIZ(1);
  @$pb.TagNumber(2)
  set maxValidators($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasMaxValidators() => $_has(1);
  @$pb.TagNumber(2)
  void clearMaxValidators() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get maxEntries => $_getIZ(2);
  @$pb.TagNumber(3)
  set maxEntries($core.int v) { $_setUnsignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasMaxEntries() => $_has(2);
  @$pb.TagNumber(3)
  void clearMaxEntries() => clearField(3);

  @$pb.TagNumber(4)
  $core.int get historicalEntries => $_getIZ(3);
  @$pb.TagNumber(4)
  set historicalEntries($core.int v) { $_setUnsignedInt32(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasHistoricalEntries() => $_has(3);
  @$pb.TagNumber(4)
  void clearHistoricalEntries() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get bondDenom => $_getSZ(4);
  @$pb.TagNumber(5)
  set bondDenom($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasBondDenom() => $_has(4);
  @$pb.TagNumber(5)
  void clearBondDenom() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get minCommissionRate => $_getSZ(5);
  @$pb.TagNumber(6)
  set minCommissionRate($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasMinCommissionRate() => $_has(5);
  @$pb.TagNumber(6)
  void clearMinCommissionRate() => clearField(6);
}

class DelegationResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DelegationResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOM<Delegation>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegation', subBuilder: Delegation.create)
    ..aOM<$10.Coin>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'balance', subBuilder: $10.Coin.create)
    ..hasRequiredFields = false
  ;

  DelegationResponse._() : super();
  factory DelegationResponse() => create();
  factory DelegationResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DelegationResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DelegationResponse clone() => DelegationResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DelegationResponse copyWith(void Function(DelegationResponse) updates) => super.copyWith((message) => updates(message as DelegationResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static DelegationResponse create() => DelegationResponse._();
  DelegationResponse createEmptyInstance() => create();
  static $pb.PbList<DelegationResponse> createRepeated() => $pb.PbList<DelegationResponse>();
  @$core.pragma('dart2js:noInline')
  static DelegationResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DelegationResponse>(create);
  static DelegationResponse _defaultInstance;

  @$pb.TagNumber(1)
  Delegation get delegation => $_getN(0);
  @$pb.TagNumber(1)
  set delegation(Delegation v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegation() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegation() => clearField(1);
  @$pb.TagNumber(1)
  Delegation ensureDelegation() => $_ensure(0);

  @$pb.TagNumber(2)
  $10.Coin get balance => $_getN(1);
  @$pb.TagNumber(2)
  set balance($10.Coin v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasBalance() => $_has(1);
  @$pb.TagNumber(2)
  void clearBalance() => clearField(2);
  @$pb.TagNumber(2)
  $10.Coin ensureBalance() => $_ensure(1);
}

class RedelegationEntryResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RedelegationEntryResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOM<RedelegationEntry>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'redelegationEntry', subBuilder: RedelegationEntry.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'balance')
    ..hasRequiredFields = false
  ;

  RedelegationEntryResponse._() : super();
  factory RedelegationEntryResponse() => create();
  factory RedelegationEntryResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RedelegationEntryResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RedelegationEntryResponse clone() => RedelegationEntryResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RedelegationEntryResponse copyWith(void Function(RedelegationEntryResponse) updates) => super.copyWith((message) => updates(message as RedelegationEntryResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RedelegationEntryResponse create() => RedelegationEntryResponse._();
  RedelegationEntryResponse createEmptyInstance() => create();
  static $pb.PbList<RedelegationEntryResponse> createRepeated() => $pb.PbList<RedelegationEntryResponse>();
  @$core.pragma('dart2js:noInline')
  static RedelegationEntryResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RedelegationEntryResponse>(create);
  static RedelegationEntryResponse _defaultInstance;

  @$pb.TagNumber(1)
  RedelegationEntry get redelegationEntry => $_getN(0);
  @$pb.TagNumber(1)
  set redelegationEntry(RedelegationEntry v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasRedelegationEntry() => $_has(0);
  @$pb.TagNumber(1)
  void clearRedelegationEntry() => clearField(1);
  @$pb.TagNumber(1)
  RedelegationEntry ensureRedelegationEntry() => $_ensure(0);

  @$pb.TagNumber(4)
  $core.String get balance => $_getSZ(1);
  @$pb.TagNumber(4)
  set balance($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(4)
  $core.bool hasBalance() => $_has(1);
  @$pb.TagNumber(4)
  void clearBalance() => clearField(4);
}

class RedelegationResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RedelegationResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOM<Redelegation>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'redelegation', subBuilder: Redelegation.create)
    ..pc<RedelegationEntryResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'entries', $pb.PbFieldType.PM, subBuilder: RedelegationEntryResponse.create)
    ..hasRequiredFields = false
  ;

  RedelegationResponse._() : super();
  factory RedelegationResponse() => create();
  factory RedelegationResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RedelegationResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RedelegationResponse clone() => RedelegationResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RedelegationResponse copyWith(void Function(RedelegationResponse) updates) => super.copyWith((message) => updates(message as RedelegationResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RedelegationResponse create() => RedelegationResponse._();
  RedelegationResponse createEmptyInstance() => create();
  static $pb.PbList<RedelegationResponse> createRepeated() => $pb.PbList<RedelegationResponse>();
  @$core.pragma('dart2js:noInline')
  static RedelegationResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RedelegationResponse>(create);
  static RedelegationResponse _defaultInstance;

  @$pb.TagNumber(1)
  Redelegation get redelegation => $_getN(0);
  @$pb.TagNumber(1)
  set redelegation(Redelegation v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasRedelegation() => $_has(0);
  @$pb.TagNumber(1)
  void clearRedelegation() => clearField(1);
  @$pb.TagNumber(1)
  Redelegation ensureRedelegation() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<RedelegationEntryResponse> get entries => $_getList(1);
}

class Pool extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Pool', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.staking.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'notBondedTokens')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'bondedTokens')
    ..hasRequiredFields = false
  ;

  Pool._() : super();
  factory Pool() => create();
  factory Pool.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Pool.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Pool clone() => Pool()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Pool copyWith(void Function(Pool) updates) => super.copyWith((message) => updates(message as Pool)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Pool create() => Pool._();
  Pool createEmptyInstance() => create();
  static $pb.PbList<Pool> createRepeated() => $pb.PbList<Pool>();
  @$core.pragma('dart2js:noInline')
  static Pool getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Pool>(create);
  static Pool _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get notBondedTokens => $_getSZ(0);
  @$pb.TagNumber(1)
  set notBondedTokens($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasNotBondedTokens() => $_has(0);
  @$pb.TagNumber(1)
  void clearNotBondedTokens() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get bondedTokens => $_getSZ(1);
  @$pb.TagNumber(2)
  set bondedTokens($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasBondedTokens() => $_has(1);
  @$pb.TagNumber(2)
  void clearBondedTokens() => clearField(2);
}

