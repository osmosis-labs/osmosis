package ibc_rate_limit_test

import (
	"fmt"
	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	ibc_rate_limit "github.com/osmosis-labs/osmosis/v12/x/ibc-rate-limit"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DenomTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestDenomTestSuite(t *testing.T) {
	suite.Run(t, new(DenomTestSuite))
}

var (
	osmoToOther = "channel-0"
	otherToOsmo = "channel-37"
)

func (s *DenomTestSuite) TestIBCDenoms() {
	testCases := map[string]struct {
		packetDenom string
		send        bool
		expRLDenom  string
	}{
		"send native": {
			packetDenom: "osmo",
			send:        true,
			expRLDenom:  "osmo",
		},
		"receive native": {
			packetDenom: "transfer/" + otherToOsmo + "/osmo",
			send:        false,
			expRLDenom:  "osmo",
		},
		"send foreign": {
			packetDenom: "transfer/" + osmoToOther + "/weirdo",
			send:        true,
			expRLDenom:  "ibc/something",
		},
		"receive foreign": {
			packetDenom: "weirdo",
			send:        false,
			expRLDenom:  "weirdo",
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			fmt.Println(name, tc)
			var sourceChannel, destChannel string
			if tc.send {
				sourceChannel, destChannel = osmoToOther, otherToOsmo
			} else {
				sourceChannel, destChannel = otherToOsmo, osmoToOther
			}
			local := ibc_rate_limit.GetIBCDenom(sourceChannel, destChannel, tc.packetDenom)
			s.Require().Equal(tc.expRLDenom, local)

		})

	}
}
