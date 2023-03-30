package chain

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	tmabcitypes "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v15/tests/e2e/util"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/query"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagerqueryproto "github.com/osmosis-labs/osmosis/v15/x/poolmanager/client/queryproto"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v15/x/protorev/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
	twapqueryproto "github.com/osmosis-labs/osmosis/v15/x/twap/client/queryproto"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

func (n *NodeConfig) genNodeQuery(path string, resp proto.Message, parameters ...string) error {
	bz, err := n.QueryGRPCGateway(path, parameters...)
	fmt.Println("QueryGRPCGateway", bz, err)
	if err != nil {
		return err
	}
	err = util.Cdc.UnmarshalJSON(bz, resp)
	fmt.Println("unmarsshall", resp, err)
	if err != nil {
		return err
	}
	fmt.Println("res in", resp)

	return nil
}

func (n *NodeConfig) QueryProtoRevNumberOfTrades() (sdk.Int, error) {
	path := "/osmosis/v14/protorev/number_of_trades"
	var resp protorevtypes.QueryGetProtoRevNumberOfTradesResponse
	err := n.genNodeQuery(path, &resp)
	fmt.Println("res out", resp)
	require.NoError(n.t, err) // this error should not happen
	return resp.NumberOfTrades, nil
}

// QueryProtoRevProfits gets the profits the protorev module has made.
func (n *NodeConfig) QueryProtoRevProfits() ([]sdk.Coin, error) {
	path := "/osmosis/v14/protorev/all_profits"
	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevAllProfitsResponse
	err := n.genNodeQuery(path, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.Profits, nil
}

// QueryProtoRevAllRouteStatistics gets all of the route statistics that the module has recorded.
func (n *NodeConfig) QueryProtoRevAllRouteStatistics() ([]protorevtypes.RouteStatistics, error) {
	path := "/osmosis/v14/protorev/all_route_statistics"
	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevAllRouteStatisticsResponse
	err := n.genNodeQuery(path, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.Statistics, nil
}

// QueryProtoRevTokenPairArbRoutes gets all of the token pair hot routes that the module is currently using.
func (n *NodeConfig) QueryProtoRevTokenPairArbRoutes() ([]protorevtypes.TokenPairArbRoutes, error) {
	path := "/osmosis/v14/protorev/token_pair_arb_routes"
	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevTokenPairArbRoutesResponse
	err := n.genNodeQuery(path, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.Routes, nil
}

// QueryProtoRevDeveloperAccount gets the developer account of the module.
func (n *NodeConfig) QueryProtoRevDeveloperAccount() (sdk.AccAddress, error) {
	path := "/osmosis/v14/protorev/developer_account"
	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevDeveloperAccountResponse
	err := n.genNodeQuery(path, &response)
	require.NoError(n.t, err) // this error should not happen

	account, err := sdk.AccAddressFromBech32(response.DeveloperAccount)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// QueryProtoRevPoolWeights gets the pool point weights of the module.
func (n *NodeConfig) QueryProtoRevPoolWeights() (protorevtypes.PoolWeights, error) {
	path := "/osmosis/v14/protorev/pool_weights"
	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevPoolWeightsResponse
	err := n.genNodeQuery(path, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.PoolWeights, nil
}

// QueryProtoRevMaxPoolPointsPerTx gets the max pool points per tx of the module.
func (n *NodeConfig) QueryProtoRevMaxPoolPointsPerTx() (uint64, error) {
	path := "/osmosis/v14/protorev/max_pool_points_per_tx"
	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevMaxPoolPointsPerTxResponse
	err := n.genNodeQuery(path, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.MaxPoolPointsPerTx, nil
}

// QueryProtoRevMaxPoolPointsPerBlock gets the max pool points per block of the module.
func (n *NodeConfig) QueryProtoRevMaxPoolPointsPerBlock() (uint64, error) {
	path := "/osmosis/v14/protorev/max_pool_points_per_block"
	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevMaxPoolPointsPerBlockResponse
	err := n.genNodeQuery(path, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.MaxPoolPointsPerBlock, nil
}

// QueryProtoRevBaseDenoms gets the base denoms used to construct cyclic arbitrage routes.
func (n *NodeConfig) QueryProtoRevBaseDenoms() ([]protorevtypes.BaseDenom, error) {
	path := "/osmosis/v14/protorev/base_denoms"
	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevBaseDenomsResponse
	err := n.genNodeQuery(path, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.BaseDenoms, nil
}

// QueryProtoRevEnabled queries if the protorev module is enabled.
func (n *NodeConfig) QueryProtoRevEnabled() (bool, error) {
	path := "/osmosis/v14/protorev/enabled"
	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevEnabledResponse
	err := n.genNodeQuery(path, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.Enabled, nil
}

func (n *NodeConfig) QueryGRPCGateway(path string, parameters ...string) ([]byte, error) {
	if len(parameters)%2 != 0 {
		return nil, fmt.Errorf("invalid number of parameters, must follow the format of key + value")
	}

	// add the URL for the given validator ID, and pre-pend to to path.
	hostPort, err := n.containerManager.GetHostPort(n.Name, "1317/tcp")
	require.NoError(n.t, err)
	endpoint := fmt.Sprintf("http://%s", hostPort)
	fullQueryPath := fmt.Sprintf("%s/%s", endpoint, path)

	var resp *http.Response
	require.Eventually(n.t, func() bool {
		req, err := http.NewRequest("GET", fullQueryPath, nil)
		if err != nil {
			return false
		}

		if len(parameters) > 0 {
			q := req.URL.Query()
			for i := 0; i < len(parameters); i += 2 {
				q.Add(parameters[i], parameters[i+1])
			}
			req.URL.RawQuery = q.Encode()
		}

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			n.t.Logf("error while executing HTTP request: %s", err.Error())
			return false
		}

		return resp.StatusCode != http.StatusServiceUnavailable
	}, time.Minute, time.Millisecond*10, "failed to execute HTTP request")

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bz))
	}
	return bz, nil
}

func (n *NodeConfig) QueryNumPools() uint64 {
	path := "osmosis/gamm/v1beta1/num_pools"
	//nolint:staticcheck
	var numPools gammtypes.QueryNumPoolsResponse
	err := n.genNodeQuery(path, &numPools)
	require.NoError(n.t, err)
	return numPools.NumPools
}

func (n *NodeConfig) QueryPoolType(poolId string) string {
	path := fmt.Sprintf("/osmosis/gamm/v1beta1/pool_type/%s", poolId)
	var poolTypeResponse gammtypes.QueryPoolTypeResponse
	err := n.genNodeQuery(path, &poolTypeResponse)
	require.NoError(n.t, err)

	return poolTypeResponse.PoolType
}

func (n *NodeConfig) QueryConcentratedPositions(address string) []model.PositionWithUnderlyingAssetBreakdown {
	path := fmt.Sprintf("/osmosis/concentratedliquidity/v1beta1/positions/%s", address)
	var positionsResponse query.QueryUserPositionsResponse
	err := n.genNodeQuery(path, &positionsResponse)
	require.NoError(n.t, err)
	return positionsResponse.Positions
}
func (n *NodeConfig) QueryConcentratedPool(poolId uint64) (cltypes.ConcentratedPoolExtension, error) {
	path := fmt.Sprintf("/osmosis/poolmanager/v1beta1/pools/%d", poolId)
	var poolResponse poolmanagerqueryproto.PoolResponse
	err := n.genNodeQuery(path, &poolResponse)
	require.NoError(n.t, err)

	var pool poolmanagertypes.PoolI
	err = util.Cdc.UnpackAny(poolResponse.Pool, &pool)
	require.NoError(n.t, err)

	poolCLextension, ok := pool.(cltypes.ConcentratedPoolExtension)

	if !ok {
		return nil, fmt.Errorf("invalid pool type: %T", pool)
	}

	return poolCLextension, nil
}

// QueryBalancer returns balances at the address.
func (n *NodeConfig) QueryBalances(address string) (sdk.Coins, error) {
	path := fmt.Sprintf("cosmos/bank/v1beta1/balances/%s", address)
	var balancesResp banktypes.QueryAllBalancesResponse
	if err := n.genNodeQuery(path, &balancesResp); err != nil {
		return sdk.Coins{}, err
	}
	return balancesResp.GetBalances(), nil
}

func (n *NodeConfig) QuerySupplyOf(denom string) (sdk.Int, error) {
	path := fmt.Sprintf("cosmos/bank/v1beta1/supply/%s", denom)
	var supplyResp banktypes.QuerySupplyOfResponse
	if err := n.genNodeQuery(path, &supplyResp); err != nil {
		return sdk.NewInt(0), err
	}
	return supplyResp.Amount.Amount, nil
}

func (n *NodeConfig) QuerySupply() (sdk.Coins, error) {
	path := "cosmos/bank/v1beta1/supply"
	var supplyResp banktypes.QueryTotalSupplyResponse
	if err := n.genNodeQuery(path, &supplyResp); err != nil {
		return nil, err
	}
	return supplyResp.Supply, nil
}

func (n *NodeConfig) QueryContractsFromId(codeId int) ([]string, error) {
	path := fmt.Sprintf("/cosmwasm/wasm/v1/code/%d/contracts", codeId)
	var contractsResponse wasmtypes.QueryContractsByCodeResponse
	if err := n.genNodeQuery(path, &contractsResponse); err != nil {
		return nil, err
	}

	return contractsResponse.Contracts, nil
}

func (n *NodeConfig) QueryLatestWasmCodeID() uint64 {
	path := "/cosmwasm/wasm/v1/code"
	var response wasmtypes.QueryCodesResponse
	err := n.genNodeQuery(path, &response)
	require.NoError(n.t, err)
	if len(response.CodeInfos) == 0 {
		return 0
	}
	return response.CodeInfos[len(response.CodeInfos)-1].CodeID
}

func (n *NodeConfig) QueryWasmSmart(contract string, msg string, result any) error {
	// base64-encode the msg
	encodedMsg := base64.StdEncoding.EncodeToString([]byte(msg))
	path := fmt.Sprintf("/cosmwasm/wasm/v1/contract/%s/smart/%s", contract, encodedMsg)
	var response wasmtypes.QuerySmartContractStateResponse
	err := n.genNodeQuery(path, &response)
	if err != nil {
		return err
	}

	err = json.Unmarshal(response.Data, &result)
	if err != nil {
		return err
	}
	return nil
}

func (n *NodeConfig) QueryWasmSmartObject(contract string, msg string) (resultObject map[string]interface{}, err error) {
	err = n.QueryWasmSmart(contract, msg, &resultObject)
	if err != nil {
		return nil, err
	}
	return resultObject, nil
}

func (n *NodeConfig) QueryWasmSmartArray(contract string, msg string) (resultArray []interface{}, err error) {
	err = n.QueryWasmSmart(contract, msg, &resultArray)
	if err != nil {
		return nil, err
	}
	return resultArray, nil
}

func (n *NodeConfig) QueryPropTally(proposalNumber int) (sdk.Int, sdk.Int, sdk.Int, sdk.Int, error) {
	path := fmt.Sprintf("cosmos/gov/v1beta1/proposals/%d/tally", proposalNumber)
	var balancesResp govtypes.QueryTallyResultResponse
	if err := n.genNodeQuery(path, &balancesResp); err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), err
	}
	noTotal := balancesResp.Tally.No
	yesTotal := balancesResp.Tally.Yes
	noWithVetoTotal := balancesResp.Tally.NoWithVeto
	abstainTotal := balancesResp.Tally.Abstain

	return noTotal, yesTotal, noWithVetoTotal, abstainTotal, nil
}

func (n *NodeConfig) QueryPropStatus(proposalNumber int) (string, error) {
	path := fmt.Sprintf("cosmos/gov/v1beta1/proposals/%d", proposalNumber)
	var propResp govtypes.QueryProposalResponse
	if err := n.genNodeQuery(path, &propResp); err != nil {
		return "", err
	}
	proposalStatus := propResp.Proposal.Status

	return proposalStatus.String(), nil
}

func (n *NodeConfig) QueryIntermediaryAccount(denom string, valAddr string) (int, error) {
	intAccount := superfluidtypes.GetSuperfluidIntermediaryAccountAddr(denom, valAddr)
	path := fmt.Sprintf(
		"cosmos/staking/v1beta1/validators/%s/delegations/%s",
		valAddr, intAccount,
	)
	var stakingResp stakingtypes.QueryDelegationResponse
	err := n.genNodeQuery(path, &stakingResp)
	require.NoError(n.t, err)

	intAccBalance := stakingResp.DelegationResponse.Balance.Amount.String()
	intAccountBalance, err := strconv.Atoi(intAccBalance)
	require.NoError(n.t, err)
	return intAccountBalance, err
}

func (n *NodeConfig) QueryCurrentEpoch(identifier string) int64 {
	path := "osmosis/epochs/v1beta1/current_epoch"
	var response epochstypes.QueryCurrentEpochResponse
	err := n.genNodeQuery(path, &response, "identifier", identifier)
	require.NoError(n.t, err)
	return response.CurrentEpoch
}

func (n *NodeConfig) QueryArithmeticTwapToNow(poolId uint64, baseAsset, quoteAsset string, startTime time.Time) (sdk.Dec, error) {
	path := "osmosis/twap/v1beta1/ArithmeticTwapToNow"
	var response twapqueryproto.ArithmeticTwapToNowResponse
	err := n.genNodeQuery(path, &response,
		"pool_id", strconv.FormatInt(int64(poolId), 10),
		"base_asset", baseAsset,
		"quote_asset", quoteAsset,
		"start_time", startTime.Format(time.RFC3339Nano),
	)
	require.NoError(n.t, err) // this error should not happen
	return response.ArithmeticTwap, nil
}

func (n *NodeConfig) QueryArithmeticTwap(poolId uint64, baseAsset, quoteAsset string, startTime time.Time, endTime time.Time) (sdk.Dec, error) {
	path := "osmosis/twap/v1beta1/ArithmeticTwap"
	var response twapqueryproto.ArithmeticTwapResponse
	err := n.genNodeQuery(path, &response,
		"pool_id", strconv.FormatInt(int64(poolId), 10),
		"base_asset", baseAsset,
		"quote_asset", quoteAsset,
		"start_time", startTime.Format(time.RFC3339Nano),
		"end_time", endTime.Format(time.RFC3339Nano),
	)
	require.NoError(n.t, err) // this error should not happen
	return response.ArithmeticTwap, nil
}

func (n *NodeConfig) QueryGeometricTwapToNow(poolId uint64, baseAsset, quoteAsset string, startTime time.Time) (sdk.Dec, error) {
	path := "osmosis/twap/v1beta1/GeometricTwapToNow"
	var response twapqueryproto.GeometricTwapToNowResponse
	err := n.genNodeQuery(path, &response,
		"pool_id", strconv.FormatInt(int64(poolId), 10),
		"base_asset", baseAsset,
		"quote_asset", quoteAsset,
		"start_time", startTime.Format(time.RFC3339Nano),
	)
	require.NoError(n.t, err)
	return response.GeometricTwap, nil
}

func (n *NodeConfig) QueryGeometricTwap(poolId uint64, baseAsset, quoteAsset string, startTime time.Time, endTime time.Time) (sdk.Dec, error) {
	path := "osmosis/twap/v1beta1/GeometricTwap"
	var response twapqueryproto.GeometricTwapResponse
	err := n.genNodeQuery(path, &response,
		"pool_id", strconv.FormatInt(int64(poolId), 10),
		"base_asset", baseAsset,
		"quote_asset", quoteAsset,
		"start_time", startTime.Format(time.RFC3339Nano),
		"end_time", endTime.Format(time.RFC3339Nano),
	)
	require.NoError(n.t, err)
	return response.GeometricTwap, nil
}

// QueryHashFromBlock gets block hash at a specific height. Otherwise, error.
func (n *NodeConfig) QueryHashFromBlock(height int64) (string, error) {
	block, err := n.rpcClient.Block(context.Background(), &height)
	if err != nil {
		return "", err
	}
	return block.BlockID.Hash.String(), nil
}

// QueryCurrentHeight returns the current block height of the node or error.
func (n *NodeConfig) QueryCurrentHeight() (int64, error) {
	status, err := n.rpcClient.Status(context.Background())
	if err != nil {
		return 0, err
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

// QueryLatestBlockTime returns the latest block time.
func (n *NodeConfig) QueryLatestBlockTime() time.Time {
	status, err := n.rpcClient.Status(context.Background())
	require.NoError(n.t, err)
	return status.SyncInfo.LatestBlockTime
}

// QueryListSnapshots gets all snapshots currently created for a node.
func (n *NodeConfig) QueryListSnapshots() ([]*tmabcitypes.Snapshot, error) {
	abciResponse, err := n.rpcClient.ABCIQuery(context.Background(), "/app/snapshots", nil)
	if err != nil {
		return nil, err
	}

	var listSnapshots tmabcitypes.ResponseListSnapshots
	if err := json.Unmarshal(abciResponse.Response.Value, &listSnapshots); err != nil {
		return nil, err
	}

	return listSnapshots.Snapshots, nil
}
