package types_test

import (
	"testing"

	proto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

func TestDenomPairTakerFeeProposalMarshalUnmarshal(t *testing.T) {
	records := []types.DenomPairTakerFee{
		{
			TokenInDenom:  "uion",
			TokenOutDenom: appparams.BaseCoinUnit,
			TakerFee:      osmomath.MustNewDecFromStr("0.0013"),
		},
		{
			TokenInDenom:  appparams.BaseCoinUnit,
			TokenOutDenom: "uion",
			TakerFee:      osmomath.MustNewDecFromStr("0.0013"),
		},
		{
			TokenInDenom:  "stake",
			TokenOutDenom: appparams.BaseCoinUnit,
			TakerFee:      osmomath.MustNewDecFromStr("0.0016"),
		},
		{
			TokenInDenom:  "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			TokenOutDenom: appparams.BaseCoinUnit,
			TakerFee:      osmomath.MustNewDecFromStr("0.0017"),
		},
	}

	tests := []struct {
		proposal *types.DenomPairTakerFeeProposal
	}{
		{ // empty title
			proposal: &types.DenomPairTakerFeeProposal{
				Title:       "",
				Description: "proposal to add denom pair taker fee records",
			},
		},
		{ // empty description
			proposal: &types.DenomPairTakerFeeProposal{
				Title:       "title",
				Description: "",
			},
		},
		{ // happy path
			proposal: &types.DenomPairTakerFeeProposal{
				Title:             "title",
				Description:       "proposal to add denom pair taker fee records",
				DenomPairTakerFee: records,
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.proposal)
		require.NoError(t, err)
		decoded := types.DenomPairTakerFeeProposal{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.proposal, decoded)
	}
}

func TestDenomPairTakerFeeProposal_ValidateBasic(t *testing.T) {
	baseRecord := types.DenomPairTakerFee{
		TokenInDenom:  "uion",
		TokenOutDenom: appparams.BaseCoinUnit,
		TakerFee:      osmomath.MustNewDecFromStr("0.0013"),
	}

	withSameDenom := func(record types.DenomPairTakerFee) types.DenomPairTakerFee {
		record.TokenOutDenom = record.TokenInDenom
		return record
	}

	withInvalidTokenInDenom := func(record types.DenomPairTakerFee) types.DenomPairTakerFee {
		record.TokenInDenom = "0"
		return record
	}

	withInvalidTokenOutDenom := func(record types.DenomPairTakerFee) types.DenomPairTakerFee {
		record.TokenOutDenom = "1"
		return record
	}

	withInvalidTakerFee := func(record types.DenomPairTakerFee) types.DenomPairTakerFee {
		record.TakerFee = osmomath.MustNewDecFromStr("1.01")
		return record
	}

	withInvalidRecord := func(record types.DenomPairTakerFee) types.DenomPairTakerFee {
		record = types.DenomPairTakerFee{}
		return record
	}

	tests := []struct {
		name       string
		modifyFunc func(types.DenomPairTakerFee) types.DenomPairTakerFee
		expectPass bool
	}{
		{
			name:       "proper msg",
			modifyFunc: func(record types.DenomPairTakerFee) types.DenomPairTakerFee { return record },
			expectPass: true,
		},
		{
			name:       "invalid denom pair",
			modifyFunc: withSameDenom,
			expectPass: false,
		},
		{
			name:       "invalid TokenInDenom",
			modifyFunc: withInvalidTokenInDenom,
			expectPass: false,
		},
		{
			name:       "invalid TokenOutDenom",
			modifyFunc: withInvalidTokenOutDenom,
			expectPass: false,
		},
		{
			name:       "invalid taker fee",
			modifyFunc: withInvalidTakerFee,
			expectPass: false,
		},
		{
			name:       "invalid record",
			modifyFunc: withInvalidRecord,
			expectPass: false,
		},
	}

	for _, test := range tests {
		records := []types.DenomPairTakerFee{test.modifyFunc(baseRecord)}

		denomPairTakerFeeProposal := types.NewDenomPairTakerFeeProposal("title", "description", records)

		if test.expectPass {
			require.NoError(t, denomPairTakerFeeProposal.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, denomPairTakerFeeProposal.ValidateBasic(), "test: %v", test.name)
		}
	}
}
