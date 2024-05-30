package osmoutils_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
)

func (s *TestSuite) TestCreateModuleAccount() {
	baseWithAddr := func(addr sdk.AccAddress) sdk.AccountI {
		acc := authtypes.ProtoBaseAccount()
		acc.SetAddress(addr)
		return acc
	}
	userAccViaSeqnum := func(addr sdk.AccAddress) sdk.AccountI {
		base := baseWithAddr(addr)
		base.SetSequence(2)
		return base
	}
	userAccViaPubkey := func(addr sdk.AccAddress) sdk.AccountI {
		base := baseWithAddr(addr)
		base.SetPubKey(secp256k1.GenPrivKey().PubKey())
		return base
	}
	defaultModuleAccAddr := address.Module("dummy module", []byte{1})
	testcases := map[string]struct {
		priorAccounts []sdk.AccountI
		moduleAccAddr sdk.AccAddress
		expErr        bool
	}{
		"no prior acc": {
			priorAccounts: []sdk.AccountI{},
			moduleAccAddr: defaultModuleAccAddr,
			expErr:        false,
		},
		"prior empty acc at addr": {
			priorAccounts: []sdk.AccountI{baseWithAddr(defaultModuleAccAddr)},
			moduleAccAddr: defaultModuleAccAddr,
			expErr:        false,
		},
		"prior user acc at addr (sequence)": {
			priorAccounts: []sdk.AccountI{userAccViaSeqnum(defaultModuleAccAddr)},
			moduleAccAddr: defaultModuleAccAddr,
			expErr:        true,
		},
		"prior user acc at addr (pubkey)": {
			priorAccounts: []sdk.AccountI{userAccViaPubkey(defaultModuleAccAddr)},
			moduleAccAddr: defaultModuleAccAddr,
			expErr:        true,
		},
	}
	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()
			for _, priorAcc := range tc.priorAccounts {
				s.accountKeeper.SetAccount(s.ctx, priorAcc)
			}
			err := osmoutils.CreateModuleAccount(s.ctx, s.accountKeeper, tc.moduleAccAddr)
			osmoassert.ConditionalError(s.T(), tc.expErr, err)
		})
	}
}
