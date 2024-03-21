package observer_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/rpc/jsonrpc/types"
	cmttypes "github.com/cometbft/cometbft/types"
	proto "github.com/cosmos/gogoproto/proto"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/observer"
	bridge "github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

var upgrader = websocket.Upgrader{}

func readNewBlockEvent(t *testing.T, path string) coretypes.ResultEvent {
	dataStr, err := os.ReadFile(path)
	require.NoError(t, err)
	result := coretypes.ResultEvent{}
	err = cmtjson.Unmarshal([]byte(dataStr), &result)
	require.NoError(t, err)
	return result
}

func readBlockResults(t *testing.T, path string) coretypes.ResultBlockResults {
	dataStr, err := os.ReadFile(path)
	require.NoError(t, err)
	result := coretypes.ResultBlockResults{}
	err = json.Unmarshal([]byte(dataStr), &result)
	require.NoError(t, err)
	return result
}

func success(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			c, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer c.Close()
			newBlock := readNewBlockEvent(t, "./test_events/new_block_event.json")
			newBlockResp := types.NewRPCSuccessResponse(
				types.JSONRPCIntID(1),
				newBlock,
			)
			newBlockRaw, err := json.Marshal(newBlockResp)
			require.NoError(t, err)
			err = c.WriteMessage(1, newBlockRaw)
			require.NoError(t, err)
		case http.MethodPost:
			blockResults := readBlockResults(t, "./test_events/block_results.json")
			blockResultsResp := types.NewRPCSuccessResponse(
				types.JSONRPCIntID(0),
				blockResults,
			)
			blockResultsRaw, err := json.Marshal(blockResultsResp)
			require.NoError(t, err)
			_, err = w.Write(blockResultsRaw)
			require.NoError(t, err)
		default:
			t.Fatal("Unexpected request method", r.Method)
		}
	}
}

// TestObserver verifies Observer properly reads subscribed messages and filters events
func TestObserver(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(success(t)))
	defer s.Close()

	observer, err := observer.NewObserver(log.NewNopLogger(), s.URL)
	require.NoError(t, err)

	ctx := context.Background()
	query := cmttypes.QueryForEvent(cmttypes.EventNewBlock)
	observeEvents := []string{proto.MessageName(&bridge.EventOutboundTransfer{})}

	err = observer.Start(ctx, query.String(), observeEvents)
	require.NoError(t, err)

	// We expect Observer to receive 3 Txs with `EventOutboundTransferType` events in this test
	// Only 2 of the Txs are successful, so we should receive only 2 event through the channel
	eventsOut := observer.Events()
	events := [2]abcitypes.Event{}
	for i := 0; i < len(events); i++ {
		require.Eventually(t, func() bool {
			events[i] = <-eventsOut
			return true
		}, time.Second, 100*time.Millisecond, "Timeout reading events from observer")
	}

	expectedEventType := proto.MessageName(&bridge.EventOutboundTransfer{})
	require.Equal(t, expectedEventType, events[0].Type)
	require.Equal(t, expectedEventType, events[1].Type)
	require.Equal(t, 0, len(eventsOut))

	err = observer.Stop(ctx)
	require.NoError(t, err)
}

// TestObserverEmptyRpcUrl verifies NewObserver returns an error if RPC URL is empty
func TestObserverEmptyRpcUrl(t *testing.T) {
	_, err := observer.NewObserver(log.NewNopLogger(), "")
	require.Error(t, err)
}

// TestObserverInvalidQuery verifies observer Start returns an error if query is invalid
func TestObserverInvalidQuery(t *testing.T) {
	observer, err := observer.NewObserver(log.NewNopLogger(), "http://localhost:26657")
	require.NoError(t, err)
	query := "invalid"
	err = observer.Start(context.Background(), query, []string{})
	require.Error(t, err)
}
