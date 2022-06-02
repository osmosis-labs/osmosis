package configurer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

func (bc *baseConfigurer) CreatePool(chainId string, valIdx int, poolFile string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	containerId := bc.valResources[chainId][valIdx].Container.ID

	bc.t.Logf("running create pool on chain id: %s with container: %s", chainId, containerId)
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	require.Eventually(
		bc.t,
		func() bool {
			exec, err := bc.dockerPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    containerId,
				User:         "root",
				Cmd: []string{
					"osmosisd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/osmosis/%s", poolFile), fmt.Sprintf("--chain-id=%s", chainId), "--from=val", "-b=block", "--yes", "--keyring-backend=test",
				},
			})
			require.NoError(bc.t, err)
			err = bc.dockerPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			require.NoError(bc.t, err)
			return strings.Contains(outBuf.String(), "code: 0")
		},
		time.Minute,
		time.Second,
		"tx returned non code 0; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	bc.t.Logf("successfully created pool on chain id: %s with container: %s", chainId, containerId)
}

func (bc *baseConfigurer) SendIBC(srcChain *chain.Chain, dstChain *chain.Chain, recipient string, token sdk.Coin) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	bc.t.Logf("sending %s from %s to %s (%s)", token, srcChain.ChainMeta.Id, dstChain.ChainMeta.Id, recipient)
	balancesBPre, err := bc.queryBalances(bc.valResources[dstChain.ChainMeta.Id][0].Container.ID, recipient)
	require.NoError(bc.t, err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	require.Eventually(
		bc.t,
		func() bool {
			exec, err := bc.dockerPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    bc.hermesResource.Container.ID,
				User:         "root",
				Cmd: []string{
					"hermes",
					"tx",
					"raw",
					"ft-transfer",
					dstChain.ChainMeta.Id,
					srcChain.ChainMeta.Id,
					"transfer",  // source chain port ID
					"channel-0", // since only one connection/channel exists, assume 0
					token.Amount.String(),
					fmt.Sprintf("--denom=%s", token.Denom),
					fmt.Sprintf("--receiver=%s", recipient),
					"--timeout-height-offset=1000",
				},
			})
			require.NoError(bc.t, err)

			err = bc.dockerPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			require.NoError(bc.t, err)
			return strings.Contains(outBuf.String(), "Success")
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	require.Eventually(
		bc.t,
		func() bool {
			balancesBPost, err := bc.queryBalances(bc.valResources[dstChain.ChainMeta.Id][0].Container.ID, recipient)
			require.NoError(bc.t, err)
			ibcCoin := balancesBPost.Sub(balancesBPre)
			if ibcCoin.Len() == 1 {
				tokenPre := balancesBPre.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				tokenPost := balancesBPost.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				resPre := chain.OsmoToken.Amount
				resPost := tokenPost.Sub(tokenPre)
				return resPost.Uint64() == resPre.Uint64()
			} else {
				return false
			}
		},
		5*time.Minute,
		time.Second,
		"tx not received on destination chain",
	)

	bc.t.Log("successfully sent IBC tokens")
}

func (bc *baseConfigurer) queryBalances(containerId string, addr string) (sdk.Coins, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := bc.dockerPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    containerId,
		User:         "root",
		Cmd: []string{
			"osmosisd", "query", "bank", "balances", addr, "--output=json",
		},
	})
	require.NoError(bc.t, err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = bc.dockerPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	require.NoErrorf(
		bc.t,
		err,
		"failed to query height; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	outBufByte := outBuf.Bytes()
	var balancesResp banktypes.QueryAllBalancesResponse
	if err := util.Cdc.UnmarshalJSON(outBufByte, &balancesResp); err != nil {
		return nil, err
	}

	return balancesResp.GetBalances(), nil
}

func (bc *baseConfigurer) getCurrentChainHeight(containerId string) int {
	var block syncInfo
	out := bc.chainStatus(containerId)
	err := json.Unmarshal(out, &block)
	require.NoError(bc.t, err)
	currentHeight, err := strconv.Atoi(block.SyncInfo.LatestHeight)
	require.NoError(bc.t, err)
	return currentHeight
}

func (bc *baseConfigurer) chainStatus(containerId string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := bc.dockerPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    containerId,
		User:         "root",
		Cmd: []string{
			"osmosisd", "status",
		},
	})
	require.NoError(bc.t, err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = bc.dockerPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	require.NoErrorf(bc.t,
		err,
		"failed to query height; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	errBufByte := errBuf.Bytes()
	return errBufByte
}

func (bc *baseConfigurer) submitProposal(c *chain.Chain, upgradeHeight int) {
	upgradeHeightStr := strconv.Itoa(upgradeHeight)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	bc.t.Logf("submitting upgrade proposal on %s container: %s", bc.valResources[c.ChainMeta.Id][0].Container.Name[1:], bc.valResources[c.ChainMeta.Id][0].Container.ID)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	require.Eventually(
		bc.t,
		func() bool {
			exec, err := bc.dockerPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    bc.valResources[c.ChainMeta.Id][0].Container.ID,
				User:         "root",
				Cmd: []string{
					"osmosisd", "tx", "gov", "submit-proposal", "software-upgrade", UpgradeVersion, fmt.Sprintf("--title=\"%s upgrade\"", UpgradeVersion), "--description=\"upgrade proposal submission\"", fmt.Sprintf("--upgrade-height=%s", upgradeHeightStr), "--upgrade-info=\"\"", fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json",
				},
			})
			require.NoError(bc.t, err)

			err = bc.dockerPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			require.NoError(bc.t, err)
			return strings.Contains(outBuf.String(), "code: 0")
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	bc.t.Log("successfully submitted proposal")
}

func (bc *baseConfigurer) depositProposal(c *chain.Chain) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	bc.t.Logf("depositing to upgrade proposal from %s container: %s", bc.valResources[c.ChainMeta.Id][0].Container.Name[1:], bc.valResources[c.ChainMeta.Id][0].Container.ID)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	require.Eventually(
		bc.t,
		func() bool {
			exec, err := bc.dockerPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    bc.valResources[c.ChainMeta.Id][0].Container.ID,
				User:         "root",
				Cmd: []string{
					"osmosisd", "tx", "gov", "deposit", "1", "10000000stake", "--from=val", fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), "-b=block", "--yes", "--keyring-backend=test",
				},
			})
			require.NoError(bc.t, err)

			err = bc.dockerPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			require.NoError(bc.t, err)
			return strings.Contains(outBuf.String(), "code: 0")
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	bc.t.Log("successfully deposited to proposal")
}

func (bc *baseConfigurer) voteProposal(chainConfig *ChainConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	chain := chainConfig.chain

	bc.t.Logf("voting for upgrade proposal for chain-id: %s", chain.ChainMeta.Id)
	for i := range chain.Validators {
		if _, ok := chainConfig.skipRunValidatorIndexes[i]; ok {
			continue
		}

		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer
		)

		require.Eventually(
			bc.t,
			func() bool {
				exec, err := bc.dockerPool.Client.CreateExec(docker.CreateExecOptions{
					Context:      ctx,
					AttachStdout: true,
					AttachStderr: true,
					Container:    bc.valResources[chain.ChainMeta.Id][i].Container.ID,
					User:         "root",
					Cmd: []string{
						"osmosisd", "tx", "gov", "vote", "1", "yes", "--from=val", fmt.Sprintf("--chain-id=%s", chain.ChainMeta.Id), "-b=block", "--yes", "--keyring-backend=test",
					},
				})
				require.NoError(bc.t, err)

				err = bc.dockerPool.Client.StartExec(exec.ID, docker.StartExecOptions{
					Context:      ctx,
					Detach:       false,
					OutputStream: &outBuf,
					ErrorStream:  &errBuf,
				})
				require.NoError(bc.t, err)
				return strings.Contains(outBuf.String(), "code: 0")
			},
			time.Minute,
			time.Second,
			"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
		)

		bc.t.Logf("successfully voted for proposal from %s container: %s", bc.valResources[chain.ChainMeta.Id][i].Container.Name[1:], bc.valResources[chain.ChainMeta.Id][i].Container.ID)
	}
}
