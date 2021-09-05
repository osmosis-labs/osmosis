///
//  Generated code. Do not modify.
//  source: confio/proofs.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'proofs.pbenum.dart';

export 'proofs.pbenum.dart';

class ExistenceProof extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ExistenceProof', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'key', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'value', $pb.PbFieldType.OY)
    ..aOM<LeafOp>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'leaf', subBuilder: LeafOp.create)
    ..pc<InnerOp>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'path', $pb.PbFieldType.PM, subBuilder: InnerOp.create)
    ..hasRequiredFields = false
  ;

  ExistenceProof._() : super();
  factory ExistenceProof() => create();
  factory ExistenceProof.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ExistenceProof.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ExistenceProof clone() => ExistenceProof()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ExistenceProof copyWith(void Function(ExistenceProof) updates) => super.copyWith((message) => updates(message as ExistenceProof)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ExistenceProof create() => ExistenceProof._();
  ExistenceProof createEmptyInstance() => create();
  static $pb.PbList<ExistenceProof> createRepeated() => $pb.PbList<ExistenceProof>();
  @$core.pragma('dart2js:noInline')
  static ExistenceProof getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ExistenceProof>(create);
  static ExistenceProof _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get key => $_getN(0);
  @$pb.TagNumber(1)
  set key($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasKey() => $_has(0);
  @$pb.TagNumber(1)
  void clearKey() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get value => $_getN(1);
  @$pb.TagNumber(2)
  set value($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasValue() => $_has(1);
  @$pb.TagNumber(2)
  void clearValue() => clearField(2);

  @$pb.TagNumber(3)
  LeafOp get leaf => $_getN(2);
  @$pb.TagNumber(3)
  set leaf(LeafOp v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasLeaf() => $_has(2);
  @$pb.TagNumber(3)
  void clearLeaf() => clearField(3);
  @$pb.TagNumber(3)
  LeafOp ensureLeaf() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.List<InnerOp> get path => $_getList(3);
}

class NonExistenceProof extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'NonExistenceProof', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'key', $pb.PbFieldType.OY)
    ..aOM<ExistenceProof>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'left', subBuilder: ExistenceProof.create)
    ..aOM<ExistenceProof>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'right', subBuilder: ExistenceProof.create)
    ..hasRequiredFields = false
  ;

  NonExistenceProof._() : super();
  factory NonExistenceProof() => create();
  factory NonExistenceProof.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory NonExistenceProof.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  NonExistenceProof clone() => NonExistenceProof()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  NonExistenceProof copyWith(void Function(NonExistenceProof) updates) => super.copyWith((message) => updates(message as NonExistenceProof)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static NonExistenceProof create() => NonExistenceProof._();
  NonExistenceProof createEmptyInstance() => create();
  static $pb.PbList<NonExistenceProof> createRepeated() => $pb.PbList<NonExistenceProof>();
  @$core.pragma('dart2js:noInline')
  static NonExistenceProof getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<NonExistenceProof>(create);
  static NonExistenceProof _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get key => $_getN(0);
  @$pb.TagNumber(1)
  set key($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasKey() => $_has(0);
  @$pb.TagNumber(1)
  void clearKey() => clearField(1);

  @$pb.TagNumber(2)
  ExistenceProof get left => $_getN(1);
  @$pb.TagNumber(2)
  set left(ExistenceProof v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasLeft() => $_has(1);
  @$pb.TagNumber(2)
  void clearLeft() => clearField(2);
  @$pb.TagNumber(2)
  ExistenceProof ensureLeft() => $_ensure(1);

  @$pb.TagNumber(3)
  ExistenceProof get right => $_getN(2);
  @$pb.TagNumber(3)
  set right(ExistenceProof v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasRight() => $_has(2);
  @$pb.TagNumber(3)
  void clearRight() => clearField(3);
  @$pb.TagNumber(3)
  ExistenceProof ensureRight() => $_ensure(2);
}

enum CommitmentProof_Proof {
  exist, 
  nonexist, 
  batch, 
  compressed, 
  notSet
}

class CommitmentProof extends $pb.GeneratedMessage {
  static const $core.Map<$core.int, CommitmentProof_Proof> _CommitmentProof_ProofByTag = {
    1 : CommitmentProof_Proof.exist,
    2 : CommitmentProof_Proof.nonexist,
    3 : CommitmentProof_Proof.batch,
    4 : CommitmentProof_Proof.compressed,
    0 : CommitmentProof_Proof.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CommitmentProof', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..oo(0, [1, 2, 3, 4])
    ..aOM<ExistenceProof>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'exist', subBuilder: ExistenceProof.create)
    ..aOM<NonExistenceProof>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nonexist', subBuilder: NonExistenceProof.create)
    ..aOM<BatchProof>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'batch', subBuilder: BatchProof.create)
    ..aOM<CompressedBatchProof>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'compressed', subBuilder: CompressedBatchProof.create)
    ..hasRequiredFields = false
  ;

  CommitmentProof._() : super();
  factory CommitmentProof() => create();
  factory CommitmentProof.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CommitmentProof.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CommitmentProof clone() => CommitmentProof()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CommitmentProof copyWith(void Function(CommitmentProof) updates) => super.copyWith((message) => updates(message as CommitmentProof)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CommitmentProof create() => CommitmentProof._();
  CommitmentProof createEmptyInstance() => create();
  static $pb.PbList<CommitmentProof> createRepeated() => $pb.PbList<CommitmentProof>();
  @$core.pragma('dart2js:noInline')
  static CommitmentProof getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CommitmentProof>(create);
  static CommitmentProof _defaultInstance;

  CommitmentProof_Proof whichProof() => _CommitmentProof_ProofByTag[$_whichOneof(0)];
  void clearProof() => clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  ExistenceProof get exist => $_getN(0);
  @$pb.TagNumber(1)
  set exist(ExistenceProof v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasExist() => $_has(0);
  @$pb.TagNumber(1)
  void clearExist() => clearField(1);
  @$pb.TagNumber(1)
  ExistenceProof ensureExist() => $_ensure(0);

  @$pb.TagNumber(2)
  NonExistenceProof get nonexist => $_getN(1);
  @$pb.TagNumber(2)
  set nonexist(NonExistenceProof v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasNonexist() => $_has(1);
  @$pb.TagNumber(2)
  void clearNonexist() => clearField(2);
  @$pb.TagNumber(2)
  NonExistenceProof ensureNonexist() => $_ensure(1);

  @$pb.TagNumber(3)
  BatchProof get batch => $_getN(2);
  @$pb.TagNumber(3)
  set batch(BatchProof v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasBatch() => $_has(2);
  @$pb.TagNumber(3)
  void clearBatch() => clearField(3);
  @$pb.TagNumber(3)
  BatchProof ensureBatch() => $_ensure(2);

  @$pb.TagNumber(4)
  CompressedBatchProof get compressed => $_getN(3);
  @$pb.TagNumber(4)
  set compressed(CompressedBatchProof v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasCompressed() => $_has(3);
  @$pb.TagNumber(4)
  void clearCompressed() => clearField(4);
  @$pb.TagNumber(4)
  CompressedBatchProof ensureCompressed() => $_ensure(3);
}

class LeafOp extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LeafOp', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..e<HashOp>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'hash', $pb.PbFieldType.OE, defaultOrMaker: HashOp.NO_HASH, valueOf: HashOp.valueOf, enumValues: HashOp.values)
    ..e<HashOp>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'prehashKey', $pb.PbFieldType.OE, defaultOrMaker: HashOp.NO_HASH, valueOf: HashOp.valueOf, enumValues: HashOp.values)
    ..e<HashOp>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'prehashValue', $pb.PbFieldType.OE, defaultOrMaker: HashOp.NO_HASH, valueOf: HashOp.valueOf, enumValues: HashOp.values)
    ..e<LengthOp>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'length', $pb.PbFieldType.OE, defaultOrMaker: LengthOp.NO_PREFIX, valueOf: LengthOp.valueOf, enumValues: LengthOp.values)
    ..a<$core.List<$core.int>>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'prefix', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  LeafOp._() : super();
  factory LeafOp() => create();
  factory LeafOp.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LeafOp.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LeafOp clone() => LeafOp()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LeafOp copyWith(void Function(LeafOp) updates) => super.copyWith((message) => updates(message as LeafOp)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static LeafOp create() => LeafOp._();
  LeafOp createEmptyInstance() => create();
  static $pb.PbList<LeafOp> createRepeated() => $pb.PbList<LeafOp>();
  @$core.pragma('dart2js:noInline')
  static LeafOp getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LeafOp>(create);
  static LeafOp _defaultInstance;

  @$pb.TagNumber(1)
  HashOp get hash => $_getN(0);
  @$pb.TagNumber(1)
  set hash(HashOp v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasHash() => $_has(0);
  @$pb.TagNumber(1)
  void clearHash() => clearField(1);

  @$pb.TagNumber(2)
  HashOp get prehashKey => $_getN(1);
  @$pb.TagNumber(2)
  set prehashKey(HashOp v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPrehashKey() => $_has(1);
  @$pb.TagNumber(2)
  void clearPrehashKey() => clearField(2);

  @$pb.TagNumber(3)
  HashOp get prehashValue => $_getN(2);
  @$pb.TagNumber(3)
  set prehashValue(HashOp v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasPrehashValue() => $_has(2);
  @$pb.TagNumber(3)
  void clearPrehashValue() => clearField(3);

  @$pb.TagNumber(4)
  LengthOp get length => $_getN(3);
  @$pb.TagNumber(4)
  set length(LengthOp v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasLength() => $_has(3);
  @$pb.TagNumber(4)
  void clearLength() => clearField(4);

  @$pb.TagNumber(5)
  $core.List<$core.int> get prefix => $_getN(4);
  @$pb.TagNumber(5)
  set prefix($core.List<$core.int> v) { $_setBytes(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasPrefix() => $_has(4);
  @$pb.TagNumber(5)
  void clearPrefix() => clearField(5);
}

class InnerOp extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'InnerOp', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..e<HashOp>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'hash', $pb.PbFieldType.OE, defaultOrMaker: HashOp.NO_HASH, valueOf: HashOp.valueOf, enumValues: HashOp.values)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'prefix', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'suffix', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  InnerOp._() : super();
  factory InnerOp() => create();
  factory InnerOp.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory InnerOp.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  InnerOp clone() => InnerOp()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  InnerOp copyWith(void Function(InnerOp) updates) => super.copyWith((message) => updates(message as InnerOp)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static InnerOp create() => InnerOp._();
  InnerOp createEmptyInstance() => create();
  static $pb.PbList<InnerOp> createRepeated() => $pb.PbList<InnerOp>();
  @$core.pragma('dart2js:noInline')
  static InnerOp getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<InnerOp>(create);
  static InnerOp _defaultInstance;

  @$pb.TagNumber(1)
  HashOp get hash => $_getN(0);
  @$pb.TagNumber(1)
  set hash(HashOp v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasHash() => $_has(0);
  @$pb.TagNumber(1)
  void clearHash() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get prefix => $_getN(1);
  @$pb.TagNumber(2)
  set prefix($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPrefix() => $_has(1);
  @$pb.TagNumber(2)
  void clearPrefix() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get suffix => $_getN(2);
  @$pb.TagNumber(3)
  set suffix($core.List<$core.int> v) { $_setBytes(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSuffix() => $_has(2);
  @$pb.TagNumber(3)
  void clearSuffix() => clearField(3);
}

class ProofSpec extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ProofSpec', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..aOM<LeafOp>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'leafSpec', subBuilder: LeafOp.create)
    ..aOM<InnerSpec>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'innerSpec', subBuilder: InnerSpec.create)
    ..a<$core.int>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxDepth', $pb.PbFieldType.O3)
    ..a<$core.int>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'minDepth', $pb.PbFieldType.O3)
    ..hasRequiredFields = false
  ;

  ProofSpec._() : super();
  factory ProofSpec() => create();
  factory ProofSpec.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ProofSpec.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ProofSpec clone() => ProofSpec()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ProofSpec copyWith(void Function(ProofSpec) updates) => super.copyWith((message) => updates(message as ProofSpec)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ProofSpec create() => ProofSpec._();
  ProofSpec createEmptyInstance() => create();
  static $pb.PbList<ProofSpec> createRepeated() => $pb.PbList<ProofSpec>();
  @$core.pragma('dart2js:noInline')
  static ProofSpec getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProofSpec>(create);
  static ProofSpec _defaultInstance;

  @$pb.TagNumber(1)
  LeafOp get leafSpec => $_getN(0);
  @$pb.TagNumber(1)
  set leafSpec(LeafOp v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasLeafSpec() => $_has(0);
  @$pb.TagNumber(1)
  void clearLeafSpec() => clearField(1);
  @$pb.TagNumber(1)
  LeafOp ensureLeafSpec() => $_ensure(0);

  @$pb.TagNumber(2)
  InnerSpec get innerSpec => $_getN(1);
  @$pb.TagNumber(2)
  set innerSpec(InnerSpec v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasInnerSpec() => $_has(1);
  @$pb.TagNumber(2)
  void clearInnerSpec() => clearField(2);
  @$pb.TagNumber(2)
  InnerSpec ensureInnerSpec() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.int get maxDepth => $_getIZ(2);
  @$pb.TagNumber(3)
  set maxDepth($core.int v) { $_setSignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasMaxDepth() => $_has(2);
  @$pb.TagNumber(3)
  void clearMaxDepth() => clearField(3);

  @$pb.TagNumber(4)
  $core.int get minDepth => $_getIZ(3);
  @$pb.TagNumber(4)
  set minDepth($core.int v) { $_setSignedInt32(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasMinDepth() => $_has(3);
  @$pb.TagNumber(4)
  void clearMinDepth() => clearField(4);
}

class InnerSpec extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'InnerSpec', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..p<$core.int>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'childOrder', $pb.PbFieldType.P3)
    ..a<$core.int>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'childSize', $pb.PbFieldType.O3)
    ..a<$core.int>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'minPrefixLength', $pb.PbFieldType.O3)
    ..a<$core.int>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'maxPrefixLength', $pb.PbFieldType.O3)
    ..a<$core.List<$core.int>>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'emptyChild', $pb.PbFieldType.OY)
    ..e<HashOp>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'hash', $pb.PbFieldType.OE, defaultOrMaker: HashOp.NO_HASH, valueOf: HashOp.valueOf, enumValues: HashOp.values)
    ..hasRequiredFields = false
  ;

  InnerSpec._() : super();
  factory InnerSpec() => create();
  factory InnerSpec.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory InnerSpec.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  InnerSpec clone() => InnerSpec()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  InnerSpec copyWith(void Function(InnerSpec) updates) => super.copyWith((message) => updates(message as InnerSpec)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static InnerSpec create() => InnerSpec._();
  InnerSpec createEmptyInstance() => create();
  static $pb.PbList<InnerSpec> createRepeated() => $pb.PbList<InnerSpec>();
  @$core.pragma('dart2js:noInline')
  static InnerSpec getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<InnerSpec>(create);
  static InnerSpec _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get childOrder => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get childSize => $_getIZ(1);
  @$pb.TagNumber(2)
  set childSize($core.int v) { $_setSignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChildSize() => $_has(1);
  @$pb.TagNumber(2)
  void clearChildSize() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get minPrefixLength => $_getIZ(2);
  @$pb.TagNumber(3)
  set minPrefixLength($core.int v) { $_setSignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasMinPrefixLength() => $_has(2);
  @$pb.TagNumber(3)
  void clearMinPrefixLength() => clearField(3);

  @$pb.TagNumber(4)
  $core.int get maxPrefixLength => $_getIZ(3);
  @$pb.TagNumber(4)
  set maxPrefixLength($core.int v) { $_setSignedInt32(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasMaxPrefixLength() => $_has(3);
  @$pb.TagNumber(4)
  void clearMaxPrefixLength() => clearField(4);

  @$pb.TagNumber(5)
  $core.List<$core.int> get emptyChild => $_getN(4);
  @$pb.TagNumber(5)
  set emptyChild($core.List<$core.int> v) { $_setBytes(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasEmptyChild() => $_has(4);
  @$pb.TagNumber(5)
  void clearEmptyChild() => clearField(5);

  @$pb.TagNumber(6)
  HashOp get hash => $_getN(5);
  @$pb.TagNumber(6)
  set hash(HashOp v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasHash() => $_has(5);
  @$pb.TagNumber(6)
  void clearHash() => clearField(6);
}

class BatchProof extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'BatchProof', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..pc<BatchEntry>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'entries', $pb.PbFieldType.PM, subBuilder: BatchEntry.create)
    ..hasRequiredFields = false
  ;

  BatchProof._() : super();
  factory BatchProof() => create();
  factory BatchProof.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory BatchProof.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  BatchProof clone() => BatchProof()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  BatchProof copyWith(void Function(BatchProof) updates) => super.copyWith((message) => updates(message as BatchProof)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static BatchProof create() => BatchProof._();
  BatchProof createEmptyInstance() => create();
  static $pb.PbList<BatchProof> createRepeated() => $pb.PbList<BatchProof>();
  @$core.pragma('dart2js:noInline')
  static BatchProof getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<BatchProof>(create);
  static BatchProof _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<BatchEntry> get entries => $_getList(0);
}

enum BatchEntry_Proof {
  exist, 
  nonexist, 
  notSet
}

class BatchEntry extends $pb.GeneratedMessage {
  static const $core.Map<$core.int, BatchEntry_Proof> _BatchEntry_ProofByTag = {
    1 : BatchEntry_Proof.exist,
    2 : BatchEntry_Proof.nonexist,
    0 : BatchEntry_Proof.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'BatchEntry', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..aOM<ExistenceProof>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'exist', subBuilder: ExistenceProof.create)
    ..aOM<NonExistenceProof>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nonexist', subBuilder: NonExistenceProof.create)
    ..hasRequiredFields = false
  ;

  BatchEntry._() : super();
  factory BatchEntry() => create();
  factory BatchEntry.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory BatchEntry.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  BatchEntry clone() => BatchEntry()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  BatchEntry copyWith(void Function(BatchEntry) updates) => super.copyWith((message) => updates(message as BatchEntry)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static BatchEntry create() => BatchEntry._();
  BatchEntry createEmptyInstance() => create();
  static $pb.PbList<BatchEntry> createRepeated() => $pb.PbList<BatchEntry>();
  @$core.pragma('dart2js:noInline')
  static BatchEntry getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<BatchEntry>(create);
  static BatchEntry _defaultInstance;

  BatchEntry_Proof whichProof() => _BatchEntry_ProofByTag[$_whichOneof(0)];
  void clearProof() => clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  ExistenceProof get exist => $_getN(0);
  @$pb.TagNumber(1)
  set exist(ExistenceProof v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasExist() => $_has(0);
  @$pb.TagNumber(1)
  void clearExist() => clearField(1);
  @$pb.TagNumber(1)
  ExistenceProof ensureExist() => $_ensure(0);

  @$pb.TagNumber(2)
  NonExistenceProof get nonexist => $_getN(1);
  @$pb.TagNumber(2)
  set nonexist(NonExistenceProof v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasNonexist() => $_has(1);
  @$pb.TagNumber(2)
  void clearNonexist() => clearField(2);
  @$pb.TagNumber(2)
  NonExistenceProof ensureNonexist() => $_ensure(1);
}

class CompressedBatchProof extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CompressedBatchProof', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..pc<CompressedBatchEntry>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'entries', $pb.PbFieldType.PM, subBuilder: CompressedBatchEntry.create)
    ..pc<InnerOp>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lookupInners', $pb.PbFieldType.PM, subBuilder: InnerOp.create)
    ..hasRequiredFields = false
  ;

  CompressedBatchProof._() : super();
  factory CompressedBatchProof() => create();
  factory CompressedBatchProof.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CompressedBatchProof.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CompressedBatchProof clone() => CompressedBatchProof()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CompressedBatchProof copyWith(void Function(CompressedBatchProof) updates) => super.copyWith((message) => updates(message as CompressedBatchProof)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CompressedBatchProof create() => CompressedBatchProof._();
  CompressedBatchProof createEmptyInstance() => create();
  static $pb.PbList<CompressedBatchProof> createRepeated() => $pb.PbList<CompressedBatchProof>();
  @$core.pragma('dart2js:noInline')
  static CompressedBatchProof getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CompressedBatchProof>(create);
  static CompressedBatchProof _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<CompressedBatchEntry> get entries => $_getList(0);

  @$pb.TagNumber(2)
  $core.List<InnerOp> get lookupInners => $_getList(1);
}

enum CompressedBatchEntry_Proof {
  exist, 
  nonexist, 
  notSet
}

class CompressedBatchEntry extends $pb.GeneratedMessage {
  static const $core.Map<$core.int, CompressedBatchEntry_Proof> _CompressedBatchEntry_ProofByTag = {
    1 : CompressedBatchEntry_Proof.exist,
    2 : CompressedBatchEntry_Proof.nonexist,
    0 : CompressedBatchEntry_Proof.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CompressedBatchEntry', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..aOM<CompressedExistenceProof>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'exist', subBuilder: CompressedExistenceProof.create)
    ..aOM<CompressedNonExistenceProof>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nonexist', subBuilder: CompressedNonExistenceProof.create)
    ..hasRequiredFields = false
  ;

  CompressedBatchEntry._() : super();
  factory CompressedBatchEntry() => create();
  factory CompressedBatchEntry.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CompressedBatchEntry.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CompressedBatchEntry clone() => CompressedBatchEntry()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CompressedBatchEntry copyWith(void Function(CompressedBatchEntry) updates) => super.copyWith((message) => updates(message as CompressedBatchEntry)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CompressedBatchEntry create() => CompressedBatchEntry._();
  CompressedBatchEntry createEmptyInstance() => create();
  static $pb.PbList<CompressedBatchEntry> createRepeated() => $pb.PbList<CompressedBatchEntry>();
  @$core.pragma('dart2js:noInline')
  static CompressedBatchEntry getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CompressedBatchEntry>(create);
  static CompressedBatchEntry _defaultInstance;

  CompressedBatchEntry_Proof whichProof() => _CompressedBatchEntry_ProofByTag[$_whichOneof(0)];
  void clearProof() => clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  CompressedExistenceProof get exist => $_getN(0);
  @$pb.TagNumber(1)
  set exist(CompressedExistenceProof v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasExist() => $_has(0);
  @$pb.TagNumber(1)
  void clearExist() => clearField(1);
  @$pb.TagNumber(1)
  CompressedExistenceProof ensureExist() => $_ensure(0);

  @$pb.TagNumber(2)
  CompressedNonExistenceProof get nonexist => $_getN(1);
  @$pb.TagNumber(2)
  set nonexist(CompressedNonExistenceProof v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasNonexist() => $_has(1);
  @$pb.TagNumber(2)
  void clearNonexist() => clearField(2);
  @$pb.TagNumber(2)
  CompressedNonExistenceProof ensureNonexist() => $_ensure(1);
}

class CompressedExistenceProof extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CompressedExistenceProof', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'key', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'value', $pb.PbFieldType.OY)
    ..aOM<LeafOp>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'leaf', subBuilder: LeafOp.create)
    ..p<$core.int>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'path', $pb.PbFieldType.P3)
    ..hasRequiredFields = false
  ;

  CompressedExistenceProof._() : super();
  factory CompressedExistenceProof() => create();
  factory CompressedExistenceProof.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CompressedExistenceProof.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CompressedExistenceProof clone() => CompressedExistenceProof()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CompressedExistenceProof copyWith(void Function(CompressedExistenceProof) updates) => super.copyWith((message) => updates(message as CompressedExistenceProof)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CompressedExistenceProof create() => CompressedExistenceProof._();
  CompressedExistenceProof createEmptyInstance() => create();
  static $pb.PbList<CompressedExistenceProof> createRepeated() => $pb.PbList<CompressedExistenceProof>();
  @$core.pragma('dart2js:noInline')
  static CompressedExistenceProof getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CompressedExistenceProof>(create);
  static CompressedExistenceProof _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get key => $_getN(0);
  @$pb.TagNumber(1)
  set key($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasKey() => $_has(0);
  @$pb.TagNumber(1)
  void clearKey() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get value => $_getN(1);
  @$pb.TagNumber(2)
  set value($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasValue() => $_has(1);
  @$pb.TagNumber(2)
  void clearValue() => clearField(2);

  @$pb.TagNumber(3)
  LeafOp get leaf => $_getN(2);
  @$pb.TagNumber(3)
  set leaf(LeafOp v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasLeaf() => $_has(2);
  @$pb.TagNumber(3)
  void clearLeaf() => clearField(3);
  @$pb.TagNumber(3)
  LeafOp ensureLeaf() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.List<$core.int> get path => $_getList(3);
}

class CompressedNonExistenceProof extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CompressedNonExistenceProof', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ics23'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'key', $pb.PbFieldType.OY)
    ..aOM<CompressedExistenceProof>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'left', subBuilder: CompressedExistenceProof.create)
    ..aOM<CompressedExistenceProof>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'right', subBuilder: CompressedExistenceProof.create)
    ..hasRequiredFields = false
  ;

  CompressedNonExistenceProof._() : super();
  factory CompressedNonExistenceProof() => create();
  factory CompressedNonExistenceProof.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CompressedNonExistenceProof.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CompressedNonExistenceProof clone() => CompressedNonExistenceProof()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CompressedNonExistenceProof copyWith(void Function(CompressedNonExistenceProof) updates) => super.copyWith((message) => updates(message as CompressedNonExistenceProof)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CompressedNonExistenceProof create() => CompressedNonExistenceProof._();
  CompressedNonExistenceProof createEmptyInstance() => create();
  static $pb.PbList<CompressedNonExistenceProof> createRepeated() => $pb.PbList<CompressedNonExistenceProof>();
  @$core.pragma('dart2js:noInline')
  static CompressedNonExistenceProof getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CompressedNonExistenceProof>(create);
  static CompressedNonExistenceProof _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get key => $_getN(0);
  @$pb.TagNumber(1)
  set key($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasKey() => $_has(0);
  @$pb.TagNumber(1)
  void clearKey() => clearField(1);

  @$pb.TagNumber(2)
  CompressedExistenceProof get left => $_getN(1);
  @$pb.TagNumber(2)
  set left(CompressedExistenceProof v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasLeft() => $_has(1);
  @$pb.TagNumber(2)
  void clearLeft() => clearField(2);
  @$pb.TagNumber(2)
  CompressedExistenceProof ensureLeft() => $_ensure(1);

  @$pb.TagNumber(3)
  CompressedExistenceProof get right => $_getN(2);
  @$pb.TagNumber(3)
  set right(CompressedExistenceProof v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasRight() => $_has(2);
  @$pb.TagNumber(3)
  void clearRight() => clearField(3);
  @$pb.TagNumber(3)
  CompressedExistenceProof ensureRight() => $_ensure(2);
}

