package configurer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
	superfluidtypes "github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (bc *baseConfigurer) QueryRPC(path string) ([]byte, error) {
	var err error
	var resp *http.Response
	retriesLeft := 5
	for {
		resp, err = http.Get(path)

		if resp.StatusCode == http.StatusServiceUnavailable {
			retriesLeft--
			if retriesLeft == 0 {
				return nil, err
			}
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func (bc *baseConfigurer) QueryChainStatus(c *chain.Config, validatorIdx int) []byte {
	cmd := []string{"osmosisd", "status"}
	_, errBuf, err := bc.containerManager.ExecCmd(bc.t, c.Id, validatorIdx, cmd, "")
	require.NoError(bc.t, err)
	return errBuf.Bytes()
}

func (bc *baseConfigurer) QueryCurrentChainHeightFromValidator(c *chain.Config, validatorIdx int) int {
	var block syncInfo
	require.Eventually(
		bc.t,
		func() bool {
			out := bc.QueryChainStatus(c, validatorIdx)
			err := json.Unmarshal(out, &block)
			if err != nil {
				return false
			}
			return true
		},
		1*time.Minute,
		time.Second,
		"Osmosis node failed to retrieve height info",
	)
	currentHeight, err := strconv.Atoi(block.SyncInfo.LatestHeight)
	require.NoError(bc.t, err)
	return currentHeight
}

func (bc *baseConfigurer) QueryBalances(c *chain.Config, i int, addr string) (sdk.Coins, error) {
	cmd := []string{"osmosisd", "query", "bank", "balances", addr, "--output=json"}
	outBuf, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, i, cmd, "")
	require.NoError(bc.t, err)

	var balancesResp banktypes.QueryAllBalancesResponse
	err = util.Cdc.UnmarshalJSON(outBuf.Bytes(), &balancesResp)
	require.NoError(bc.t, err)

	return balancesResp.GetBalances(), nil
}

func (bc *baseConfigurer) QueryPropTally(chainId string, validatorIdx int, addr string) (sdk.Int, sdk.Int, sdk.Int, sdk.Int, error) {
	hostPort, err := bc.containerManager.GetValidatorHostPort(chainId, validatorIdx, "1317/tcp")
	require.NoError(bc.t, err)

	endpoint := fmt.Sprintf("http://%s", hostPort)

	path := fmt.Sprintf(
		"%s/cosmos/gov/v1beta1/proposals/%s/tally",
		endpoint, addr,
	)
	bz, err := bc.QueryRPC(path)
	require.NoError(bc.t, err)

	var balancesResp govtypes.QueryTallyResultResponse
	if err := util.Cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), err
	}
	noTotal := balancesResp.Tally.No
	yesTotal := balancesResp.Tally.Yes
	noWithVetoTotal := balancesResp.Tally.NoWithVeto
	abstainTotal := balancesResp.Tally.Abstain

	return noTotal, yesTotal, noWithVetoTotal, abstainTotal, nil
}

func (bc *baseConfigurer) QueryValidatorOperatorAddresses(c *chain.Config) {
	for i, val := range c.ValidatorConfigs {
		cmd := []string{"osmosisd", "debug", "addr", val.PublicKey}
		bc.t.Logf("extracting validator operator addresses for chain-id: %s", c.Id)
		_, errBuf, err := bc.containerManager.ExecCmd(bc.t, c.Id, i, cmd, "")
		require.NoError(bc.t, err)
		re := regexp.MustCompile("osmovaloper(.{39})")
		operAddr := fmt.Sprintf("%s\n", re.FindString(errBuf.String()))
		c.ValidatorConfigs[i].OperatorAddress = strings.TrimSuffix(operAddr, "\n")
	}
}

func (bc *baseConfigurer) QueryIntermediaryAccount(chainId string, validatorIdx int, denom string, valAddr string) (int, error) {
	hostPort, err := bc.containerManager.GetValidatorHostPort(chainId, validatorIdx, "1317/tcp")
	require.NoError(bc.t, err)

	endpoint := fmt.Sprintf("http://%s", hostPort)

	intAccount := superfluidtypes.GetSuperfluidIntermediaryAccountAddr(denom, valAddr)
	path := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations/%s",
		endpoint, valAddr, intAccount,
	)
	bz, err := bc.QueryRPC(path)
	require.NoError(bc.t, err)

	var stakingResp stakingtypes.QueryDelegationResponse
	err = util.Cdc.UnmarshalJSON(bz, &stakingResp)
	require.NoError(bc.t, err)

	intAccBalance := stakingResp.DelegationResponse.Balance.Amount.String()
	intAccountBalance, err := strconv.Atoi(intAccBalance)
	require.NoError(bc.t, err)
	return intAccountBalance, err
}
