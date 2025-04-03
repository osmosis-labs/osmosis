package chain

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	tmabcitypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/util"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	poolmanagerqueryproto "github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryproto"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v27/x/protorev/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
	twapqueryproto "github.com/osmosis-labs/osmosis/v27/x/twap/client/queryproto"
)

// PropTallyResult is the result of a proposal tally.
type PropTallyResult struct {
	Yes        osmomath.Int
	No         osmomath.Int
	Abstain    osmomath.Int
	NoWithVeto osmomath.Int
}

// QueryProtoRevNumberOfTrades gets the number of trades the protorev module has executed.
func (n *NodeConfig) QueryProtoRevNumberOfTrades() (osmomath.Int, error) {
	path := "/symphony/protorev/number_of_trades"

	bz, err := n.QueryGRPCGateway(path)
	if err != nil {
		return osmomath.Int{}, err
	}

	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevNumberOfTradesResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.NumberOfTrades, nil
}

// QueryProtoRevProfits gets the profits the protorev module has made.
func (n *NodeConfig) QueryProtoRevProfits() ([]sdk.Coin, error) {
	path := "/symphony/protorev/all_profits"

	bz, err := n.QueryGRPCGateway(path)
	if err != nil {
		return []sdk.Coin{}, err
	}

	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevAllProfitsResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.Profits, nil
}

// QueryProtoRevAllRouteStatistics gets all of the route statistics that the module has recorded.
func (n *NodeConfig) QueryProtoRevAllRouteStatistics() ([]protorevtypes.RouteStatistics, error) {
	path := "/symphony/protorev/all_route_statistics"

	bz, err := n.QueryGRPCGateway(path)
	if err != nil {
		return []protorevtypes.RouteStatistics{}, err
	}

	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevAllRouteStatisticsResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.Statistics, nil
}

// QueryProtoRevTokenPairArbRoutes gets all of the token pair hot routes that the module is currently using.
func (n *NodeConfig) QueryProtoRevTokenPairArbRoutes() ([]protorevtypes.TokenPairArbRoutes, error) {
	path := "/symphony/protorev/token_pair_arb_routes"

	bz, err := n.QueryGRPCGateway(path)
	if err != nil {
		return []protorevtypes.TokenPairArbRoutes{}, err
	}

	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevTokenPairArbRoutesResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.Routes, nil
}

// QueryProtoRevDeveloperAccount gets the developer account of the module.
func (n *NodeConfig) QueryProtoRevDeveloperAccount() (sdk.AccAddress, error) {
	path := "/symphony/protorev/developer_account"

	bz, err := n.QueryGRPCGateway(path)
	if err != nil {
		return nil, err
	}

	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevDeveloperAccountResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen

	account, err := sdk.AccAddressFromBech32(response.DeveloperAccount)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// QueryProtoRevInfoByPoolType gets information on how the module handles different pool types.
func (n *NodeConfig) QueryProtoRevInfoByPoolType() (*protorevtypes.InfoByPoolType, error) {
	path := "/symphony/protorev/info_by_pool_type"

	bz, err := n.QueryGRPCGateway(path)
	if err != nil {
		return nil, err
	}

	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevInfoByPoolTypeResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen
	return &response.InfoByPoolType, nil
}

// QueryProtoRevMaxPoolPointsPerTx gets the max pool points per tx of the module.
func (n *NodeConfig) QueryProtoRevMaxPoolPointsPerTx() (uint64, error) {
	path := "/symphony/protorev/max_pool_points_per_tx"

	bz, err := n.QueryGRPCGateway(path)
	if err != nil {
		return 0, err
	}

	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevMaxPoolPointsPerTxResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.MaxPoolPointsPerTx, nil
}

// QueryProtoRevMaxPoolPointsPerBlock gets the max pool points per block of the module.
func (n *NodeConfig) QueryProtoRevMaxPoolPointsPerBlock() (uint64, error) {
	path := "/symphony/protorev/max_pool_points_per_block"

	bz, err := n.QueryGRPCGateway(path)
	if err != nil {
		return 0, err
	}

	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevMaxPoolPointsPerBlockResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.MaxPoolPointsPerBlock, nil
}

// QueryProtoRevBaseDenoms gets the base denoms used to construct cyclic arbitrage routes.
func (n *NodeConfig) QueryProtoRevBaseDenoms() ([]protorevtypes.BaseDenom, error) {
	path := "/symphony/protorev/base_denoms"

	bz, err := n.QueryGRPCGateway(path)
	if err != nil {
		return []protorevtypes.BaseDenom{}, err
	}

	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevBaseDenomsResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.BaseDenoms, nil
}

// QueryProtoRevEnabled queries if the protorev module is enabled.
func (n *NodeConfig) QueryProtoRevEnabled() (bool, error) {
	path := "/symphony/protorev/enabled"

	bz, err := n.QueryGRPCGateway(path)
	if err != nil {
		return false, err
	}

	// nolint: staticcheck
	var response protorevtypes.QueryGetProtoRevEnabledResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.Enabled, nil
}

func (n *NodeConfig) QueryGRPCGateway(path string, parameters ...string) ([]byte, error) {
	if len(parameters)%2 != 0 {
		return nil, fmt.Errorf("invalid number of parameters, must follow the format of key + value")
	}

	// add the URL for the given validator ID, and prepend to to path.
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
	}, time.Minute, 10*time.Millisecond, "failed to execute HTTP request")

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
	path := "symphony/gamm/v1beta1/num_pools"

	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	//nolint:staticcheck
	var numPools gammtypes.QueryNumPoolsResponse
	err = util.Cdc.UnmarshalJSON(bz, &numPools)
	require.NoError(n.t, err)
	return numPools.NumPools
}

func (n *NodeConfig) QueryPoolType(poolId string) string {
	path := fmt.Sprintf("/symphony/gamm/v1beta1/pool_type/%s", poolId)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var poolTypeResponse gammtypes.QueryPoolTypeResponse
	err = util.Cdc.UnmarshalJSON(bz, &poolTypeResponse)
	require.NoError(n.t, err)

	return poolTypeResponse.PoolType
}

func (n *NodeConfig) QueryConcentratedPositions(address string) []model.FullPositionBreakdown {
	path := fmt.Sprintf("/symphony/concentratedliquidity/v1beta1/positions/%s", address)

	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var positionsResponse queryproto.UserPositionsResponse
	err = util.Cdc.UnmarshalJSON(bz, &positionsResponse)
	require.NoError(n.t, err)
	return positionsResponse.Positions
}

func (n *NodeConfig) QueryConcentratedPool(poolId uint64) (cltypes.ConcentratedPoolExtension, error) {
	path := fmt.Sprintf("/symphony/poolmanager/v1beta1/pools/%d", poolId)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var poolResponse poolmanagerqueryproto.PoolResponse
	err = util.Cdc.UnmarshalJSON(bz, &poolResponse)
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

func (n *NodeConfig) QueryCFMMPool(poolId uint64) (gammtypes.CFMMPoolI, error) {
	path := fmt.Sprintf("/symphony/poolmanager/v1beta1/pools/%d", poolId)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var poolResponse poolmanagerqueryproto.PoolResponse
	err = util.Cdc.UnmarshalJSON(bz, &poolResponse)
	require.NoError(n.t, err)

	var pool poolmanagertypes.PoolI
	err = util.Cdc.UnpackAny(poolResponse.Pool, &pool)
	require.NoError(n.t, err)

	cfmmPool, ok := pool.(gammtypes.CFMMPoolI)

	if !ok {
		return nil, fmt.Errorf("invalid pool type: %T", pool)
	}

	return cfmmPool, nil
}

// QueryBalancer returns balances at the address.
func (n *NodeConfig) QueryBalances(address string) (sdk.Coins, error) {
	path := fmt.Sprintf("cosmos/bank/v1beta1/balances/%s", address)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var balancesResp banktypes.QueryAllBalancesResponse
	if err := util.Cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return sdk.Coins{}, err
	}
	return balancesResp.GetBalances(), nil
}

func (n *NodeConfig) QueryBalance(address, denom string) (sdk.Coin, error) {
	path := fmt.Sprintf("cosmos/bank/v1beta1/balances/%s/by_denom?denom=%s", address, denom)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var balancesResp banktypes.QueryBalanceResponse
	if err := util.Cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return sdk.Coin{}, err
	}
	return *balancesResp.GetBalance(), nil
}

func (n *NodeConfig) QuerySupplyOf(denom string) (osmomath.Int, error) {
	path := fmt.Sprintf("cosmos/bank/v1beta1/supply/%s", denom)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var supplyResp banktypes.QuerySupplyOfResponse
	if err := util.Cdc.UnmarshalJSON(bz, &supplyResp); err != nil {
		return osmomath.NewInt(0), err
	}
	return supplyResp.Amount.Amount, nil
}

func (n *NodeConfig) QuerySupply() (sdk.Coins, error) {
	path := "cosmos/bank/v1beta1/supply"
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var supplyResp banktypes.QueryTotalSupplyResponse
	if err := util.Cdc.UnmarshalJSON(bz, &supplyResp); err != nil {
		return nil, err
	}
	return supplyResp.Supply, nil
}

func (n *NodeConfig) QueryContractsFromId(codeId int) ([]string, error) {
	path := fmt.Sprintf("/cosmwasm/wasm/v1/code/%d/contracts", codeId)
	bz, err := n.QueryGRPCGateway(path)

	require.NoError(n.t, err)

	var contractsResponse wasmtypes.QueryContractsByCodeResponse
	if err := util.Cdc.UnmarshalJSON(bz, &contractsResponse); err != nil {
		return nil, err
	}

	return contractsResponse.Contracts, nil
}

func (n *NodeConfig) QueryLatestWasmCodeID() uint64 {
	path := "/cosmwasm/wasm/v1/code"

	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var response wasmtypes.QueryCodesResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
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

	bz, err := n.QueryGRPCGateway(path)
	if err != nil {
		return err
	}

	var response wasmtypes.QuerySmartContractStateResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
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

func (n *NodeConfig) QueryPropTally(proposalNumber int) (PropTallyResult, error) {
	path := fmt.Sprintf("cosmos/gov/v1beta1/proposals/%d/tally", proposalNumber)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var balancesResp govtypesv1.QueryTallyResultResponse
	if err := util.Cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return PropTallyResult{
			Yes:        osmomath.ZeroInt(),
			No:         osmomath.ZeroInt(),
			Abstain:    osmomath.ZeroInt(),
			NoWithVeto: osmomath.ZeroInt(),
		}, err
	}
	noTotal := balancesResp.Tally.No
	yesTotal := balancesResp.Tally.Yes
	noWithVetoTotal := balancesResp.Tally.NoWithVeto
	abstainTotal := balancesResp.Tally.Abstain

	return PropTallyResult{
		Yes:        yesTotal,
		No:         noTotal,
		Abstain:    abstainTotal,
		NoWithVeto: noWithVetoTotal,
	}, nil
}

func (n *NodeConfig) QueryPropStatus(proposalNumber int) (string, error) {
	path := fmt.Sprintf("cosmos/gov/v1beta1/proposals/%d", proposalNumber)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var propResp govtypesv1.QueryProposalResponse
	err = util.Cdc.UnmarshalJSON(bz, &propResp)
	if err != nil && !strings.Contains(err.Error(), "is_expedited") {
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

	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var stakingResp stakingtypes.QueryDelegationResponse
	err = util.Cdc.UnmarshalJSON(bz, &stakingResp)
	require.NoError(n.t, err)

	intAccBalance := stakingResp.DelegationResponse.Balance.Amount.String()
	intAccountBalance, err := strconv.Atoi(intAccBalance)
	require.NoError(n.t, err)
	return intAccountBalance, err
}

func (n *NodeConfig) QueryCurrentEpoch(identifier string) int64 {
	path := "symphony/epochs/v1beta1/current_epoch"

	bz, err := n.QueryGRPCGateway(path, "identifier", identifier)
	require.NoError(n.t, err)

	var response epochstypes.QueryCurrentEpochResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err)
	return response.CurrentEpoch
}

func (n *NodeConfig) QueryConcentratedPooIdLinkFromCFMM(cfmmPoolId uint64) uint64 {
	path := fmt.Sprintf("/symphony/gamm/v1beta1/concentrated_pool_id_link_from_cfmm/%d", cfmmPoolId)

	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	//nolint:staticcheck
	var response gammtypes.QueryConcentratedPoolIdLinkFromCFMMResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err)
	return response.ConcentratedPoolId
}

func (n *NodeConfig) QueryArithmeticTwapToNow(poolId uint64, baseAsset, quoteAsset string, startTime time.Time) (osmomath.Dec, error) {
	path := "symphony/twap/v1beta1/ArithmeticTwapToNow"

	bz, err := n.QueryGRPCGateway(
		path,
		"pool_id", strconv.FormatInt(int64(poolId), 10),
		"base_asset", baseAsset,
		"quote_asset", quoteAsset,
		"start_time", startTime.Format(time.RFC3339Nano),
	)
	if err != nil {
		return osmomath.Dec{}, err
	}

	var response twapqueryproto.ArithmeticTwapToNowResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.ArithmeticTwap, nil
}

func (n *NodeConfig) QueryArithmeticTwap(poolId uint64, baseAsset, quoteAsset string, startTime time.Time, endTime time.Time) (osmomath.Dec, error) {
	path := "symphony/twap/v1beta1/ArithmeticTwap"

	bz, err := n.QueryGRPCGateway(
		path,
		"pool_id", strconv.FormatInt(int64(poolId), 10),
		"base_asset", baseAsset,
		"quote_asset", quoteAsset,
		"start_time", startTime.Format(time.RFC3339Nano),
		"end_time", endTime.Format(time.RFC3339Nano),
	)
	if err != nil {
		return osmomath.Dec{}, err
	}

	var response twapqueryproto.ArithmeticTwapResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err) // this error should not happen
	return response.ArithmeticTwap, nil
}

func (n *NodeConfig) QueryGeometricTwapToNow(poolId uint64, baseAsset, quoteAsset string, startTime time.Time) (osmomath.Dec, error) {
	path := "symphony/twap/v1beta1/GeometricTwapToNow"

	bz, err := n.QueryGRPCGateway(
		path,
		"pool_id", strconv.FormatInt(int64(poolId), 10),
		"base_asset", baseAsset,
		"quote_asset", quoteAsset,
		"start_time", startTime.Format(time.RFC3339Nano),
	)
	if err != nil {
		return osmomath.Dec{}, err
	}

	var response twapqueryproto.GeometricTwapToNowResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err)
	return response.GeometricTwap, nil
}

func (n *NodeConfig) QueryGeometricTwap(poolId uint64, baseAsset, quoteAsset string, startTime time.Time, endTime time.Time) (osmomath.Dec, error) {
	path := "symphony/twap/v1beta1/GeometricTwap"

	bz, err := n.QueryGRPCGateway(
		path,
		"pool_id", strconv.FormatInt(int64(poolId), 10),
		"base_asset", baseAsset,
		"quote_asset", quoteAsset,
		"start_time", startTime.Format(time.RFC3339Nano),
		"end_time", endTime.Format(time.RFC3339Nano),
	)
	if err != nil {
		return osmomath.Dec{}, err
	}

	var response twapqueryproto.GeometricTwapResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
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

// QueryAllSuperfluidAssets returns all authorized superfluid assets.
func (n *NodeConfig) QueryAllSuperfluidAssets() []superfluidtypes.SuperfluidAsset {
	path := "/symphony/superfluid/v1beta1/all_assets"

	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	//nolint:staticcheck
	var response superfluidtypes.AllAssetsResponse
	err = util.Cdc.UnmarshalJSON(bz, &response)
	require.NoError(n.t, err)
	return response.Assets
}

func (n *NodeConfig) QueryCommunityPoolModuleAccount() string {
	cmd := []string{"symphonyd", "query", "auth", "module-accounts", "--output=json"}

	out, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "", false, false)
	require.NoError(n.t, err)
	var result map[string][]interface{}
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(n.t, err)
	for _, acc := range result["accounts"] {
		account, ok := acc.(map[string]interface{})
		require.True(n.t, ok)
		if account["name"] == "distribution" {
			moduleAccount, ok := account["base_account"].(map[string]interface{})["address"].(string)
			require.True(n.t, ok)
			return moduleAccount
		}
	}
	require.True(n.t, false, "distribution module account not found")
	return ""
}
