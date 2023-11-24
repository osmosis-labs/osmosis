package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientkeeper "github.com/cosmos/ibc-go/v7/modules/core/02-client/keeper"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	tendermintLightClient "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

type HeaderVerifier interface {
	VerifyHeaders(ctx sdk.Context, cleintkeeper clientkeeper.Keeper, clientID string, header, nextHeader exported.ClientMessage) error
	UnpackHeader(any *codectypes.Any) (exported.ClientMessage, error)
}

type TransactionVerifier interface {
	VerifyTransaction(
		header *tendermintLightClient.Header,
		nextHeader *tendermintLightClient.Header,
		tx *TxValue,
	) error
}
