package chain

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/common"
)

func decodeTx(txBytes []byte) (*sdktx.Tx, error) {
	var raw sdktx.TxRaw

	// reject all unknown proto fields in the root TxRaw
	err := unknownproto.RejectUnknownFieldsStrict(txBytes, &raw, common.EncodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := common.Cdc.Unmarshal(txBytes, &raw); err != nil {
		return nil, err
	}

	var body sdktx.TxBody
	if err := common.Cdc.Unmarshal(raw.BodyBytes, &body); err != nil {
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	var authInfo sdktx.AuthInfo

	// reject all unknown proto fields in AuthInfo
	err = unknownproto.RejectUnknownFieldsStrict(raw.AuthInfoBytes, &authInfo, common.EncodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := common.Cdc.Unmarshal(raw.AuthInfoBytes, &authInfo); err != nil {
		return nil, fmt.Errorf("failed to decode auth info: %w", err)
	}

	return &sdktx.Tx{
		Body:       &body,
		AuthInfo:   &authInfo,
		Signatures: raw.Signatures,
	}, nil
}
