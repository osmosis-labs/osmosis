package types

import errorsmod "cosmossdk.io/errors"

func (m Asset) Validate() error {
	if len(m.SourceChain) == 0 {
		return errorsmod.Wrap(ErrInvalidSourceChain, "Source chain is empty")
	}
	if len(m.Denom) == 0 {
		return errorsmod.Wrap(ErrInvalidDenom, "Denom is empty")
	}
	return nil
}

func DefaultAssets() []Asset {
	return []Asset{
		{
			SourceChain: DefaultBitcoinChainName,
			Denom:       DefaultBitcoinDenomName,
		},
	}
}
