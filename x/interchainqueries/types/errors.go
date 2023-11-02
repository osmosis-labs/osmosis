package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/interchainqueries module sentinel errors
var (
	ErrInvalidQueryID             = sdkerrors.Register(ModuleName, 1100, "invalid query id")
	ErrEmptyResult                = sdkerrors.Register(ModuleName, 1101, "empty result")
	ErrInvalidClientID            = sdkerrors.Register(ModuleName, 1102, "invalid client id")
	ErrInvalidUpdatePeriod        = sdkerrors.Register(ModuleName, 1103, "invalid update period")
	ErrInvalidConnectionID        = sdkerrors.Register(ModuleName, 1104, "invalid connection id")
	ErrInvalidQueryType           = sdkerrors.Register(ModuleName, 1105, "invalid query type")
	ErrInvalidTransactionsFilter  = sdkerrors.Register(ModuleName, 1106, "invalid transactions filter")
	ErrInvalidSubmittedResult     = sdkerrors.Register(ModuleName, 1107, "invalid result")
	ErrProtoMarshal               = sdkerrors.Register(ModuleName, 1108, "failed to marshal protobuf bytes")
	ErrProtoUnmarshal             = sdkerrors.Register(ModuleName, 1109, "failed to unmarshal protobuf bytes")
	ErrInvalidType                = sdkerrors.Register(ModuleName, 1110, "invalid type")
	ErrInternal                   = sdkerrors.Register(ModuleName, 1111, "internal error")
	ErrInvalidProof               = sdkerrors.Register(ModuleName, 1112, "merkle proof is invalid")
	ErrInvalidHeader              = sdkerrors.Register(ModuleName, 1113, "header is invalid")
	ErrInvalidHeight              = sdkerrors.Register(ModuleName, 1114, "height is invalid")
	ErrNoQueryResult              = sdkerrors.Register(ModuleName, 1115, "no query result")
	ErrNotContract                = sdkerrors.Register(ModuleName, 1116, "not a contract")
	ErrEmptyKeys                  = sdkerrors.Register(ModuleName, 1117, "keys are empty")
	ErrEmptyKeyPath               = sdkerrors.Register(ModuleName, 1118, "key path is empty")
	ErrEmptyKeyID                 = sdkerrors.Register(ModuleName, 1119, "key id is empty")
	ErrTooManyKVQueryKeys         = sdkerrors.Register(ModuleName, 1120, "too many keys")
	ErrUnexpectedQueryTypeGenesis = sdkerrors.Register(ModuleName, 1121, "unexpected query type")
)
