package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
)

func DefaultAssets() []Asset {
	return []Asset{
		{
			Id: AssetID{
				SourceChain: DefaultBitcoinChainName,
				Denom:       DefaultBitcoinDenomName,
			},
			Status:   AssetStatus_ASSET_STATUS_BLOCKED_BOTH,
			Exponent: DefaultBitcoinExponent,
		},
	}
}

func (m Asset) Name() string {
	return m.Id.Name()
}

func (m Asset) Validate() error {
	err := m.Id.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAssetID, err.Error())
	}

	err = m.Status.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAssetStatus, err.Error())
	}

	// don't check m.Exponent and m.LastTransferHeight since they are always valid

	return nil
}

func (m AssetID) Name() string {
	return fmt.Sprintf("%s-%s", m.SourceChain, m.Denom)
}

func (m AssetID) Validate() error {
	if len(m.SourceChain) == 0 {
		return errorsmod.Wrap(ErrInvalidSourceChain, "Source chain is empty")
	}
	if len(m.Denom) == 0 {
		return errorsmod.Wrap(ErrInvalidDenom, "Denom is empty")
	}
	return nil
}

func (m AssetStatus) InboundActive() bool {
	switch m {
	case AssetStatus_ASSET_STATUS_OK,
		AssetStatus_ASSET_STATUS_BLOCKED_OUTBOUND:
		return true
	case AssetStatus_ASSET_STATUS_BLOCKED_INBOUND,
		AssetStatus_ASSET_STATUS_BLOCKED_BOTH:
		return false
	default:
		return false
	}
}

func (m AssetStatus) OutboundActive() bool {
	switch m {
	case AssetStatus_ASSET_STATUS_OK,
		AssetStatus_ASSET_STATUS_BLOCKED_INBOUND:
		return true
	case AssetStatus_ASSET_STATUS_BLOCKED_OUTBOUND,
		AssetStatus_ASSET_STATUS_BLOCKED_BOTH:
		return false
	default:
		return false
	}
}

func (m AssetStatus) Validate() error {
	switch m {
	case AssetStatus_ASSET_STATUS_OK,
		AssetStatus_ASSET_STATUS_BLOCKED_INBOUND,
		AssetStatus_ASSET_STATUS_BLOCKED_OUTBOUND,
		AssetStatus_ASSET_STATUS_BLOCKED_BOTH:
		return nil
	case AssetStatus_ASSET_STATUS_UNSPECIFIED:
		return fmt.Errorf("invalid asset status: %v", m)
	default:
		return fmt.Errorf("unknown asset status: %v", m)
	}
}
