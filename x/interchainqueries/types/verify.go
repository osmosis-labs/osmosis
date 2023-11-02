package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientkeeper "github.com/cosmos/ibc-go/v4/modules/core/02-client/keeper"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	tendermintLightClientTypes "github.com/cosmos/ibc-go/v4/modules/light-clients/07-tendermint/types"
)

type HeaderVerifier interface {
	VerifyHeaders(ctx sdk.Context, cleintkeeper clientkeeper.Keeper, clientID string, header, nextHeader exported.Header) error
	UnpackHeader(any *codectypes.Any) (exported.Header, error)
}

type TransactionVerifier interface {
	VerifyTransaction(
		header *tendermintLightClientTypes.Header,
		nextHeader *tendermintLightClientTypes.Header,
		tx *TxValue,
	) error
}
