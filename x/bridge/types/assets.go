package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
)

func DefaultAssetsWithStatuses() []AssetWithStatus {
	return []AssetWithStatus{
		{
			Asset: Asset{
				SourceChain: DefaultBitcoinChainName,
				Denom:       DefaultBitcoinDenomName,
				Precision:   DefaultBitcoinPrecision,
			},
			AssetStatus: AssetStatus_ASSET_STATUS_BLOCKED_BOTH,
		},
	}
}

func (m AssetWithStatus) Validate() error {
	err := m.Asset.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAsset, err.Error())
	}

	err = m.AssetStatus.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAssetStatus, err.Error())
	}

	return nil
}

func (m Asset) Validate() error {
	if len(m.SourceChain) == 0 {
		return errorsmod.Wrap(ErrInvalidSourceChain, "Source chain is empty")
	}
	if len(m.Denom) == 0 {
		return errorsmod.Wrap(ErrInvalidDenom, "Denom is empty")
	}
	return nil
}

func (m Asset) Name() string {
	return fmt.Sprintf("%s-%s", m.SourceChain, m.Denom)
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
