package keeper

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/rpc/jsonrpc/types"
	cmttypes "github.com/cometbft/cometbft/types"
	proto "github.com/cosmos/gogoproto/proto"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	bridge "github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

const (
	EventOutboundTransferType = "osmosis.bridge.v1beta1.EventOutboundTransfer"
)

var upgrader = websocket.Upgrader{}

func readNewBlockEvent(path string) coretypes.ResultEvent {
	dataStr, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	result := coretypes.ResultEvent{}
	err = cmtjson.Unmarshal([]byte(dataStr), &result)
	if err != nil {
		panic(err)
	}
	return result
}

func readBlockResults(path string) coretypes.ResultBlockResults {
	dataStr, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	result := coretypes.ResultBlockResults{}
	err = json.Unmarshal([]byte(dataStr), &result)
	if err != nil {
		panic(err)
	}
	return result
}

func success(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}
		defer c.Close()
		newBlock := readNewBlockEvent("./test_events/new_block_event.json")
		newBlockResp := types.NewRPCSuccessResponse(
			types.JSONRPCIntID(1),
			newBlock,
		)
		newBlockRaw, err := json.Marshal(newBlockResp)
		if err != nil {
			panic(err)
		}

		err = c.WriteMessage(1, newBlockRaw)
		if err != nil {
			panic(err)
		}
	} else if r.Method == "POST" {
		blockResults := readBlockResults("./test_events/block_results.json")
		blockResultsResp := types.NewRPCSuccessResponse(
			types.JSONRPCIntID(0),
			blockResults,
		)
		blockResultsRaw, err := json.Marshal(blockResultsResp)
		if err != nil {
			panic(err)
		}
		_, err = w.Write(blockResultsRaw)
		if err != nil {
			panic(err)
		}
	}
}

// Verify Observer properly reads subscribed messages and filters events
func TestObserver(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(success))
	defer s.Close()

	eventsOut := make(chan abcitypes.Event)
	observer, err := NewObserver(s.URL, eventsOut)
	require.NoError(t, err)

	query := cmttypes.QueryForEvent(cmttypes.EventNewBlock)
	observeEvents := []string{proto.MessageName(&bridge.EventOutboundTransfer{})}

	err = observer.Start(query.String(), observeEvents)
	require.NoError(t, err)

	// We expect Observer to receive 3 Txs with `EventOutboundTransferType` events in this test
	// Only 3 of the Txs are successful, so we should receive only 2 event through the channel
	events := [2]abcitypes.Event{}
	for i := 0; i < len(events); i++ {
		select {
		case e := <-eventsOut:
			events[i] = e
		case <-time.After(1 * time.Second):
			panic("Channel read timeout")
		}
	}

	require.Equal(t, EventOutboundTransferType, events[0].Type)
	require.Equal(t, EventOutboundTransferType, events[1].Type)
	require.Equal(t, 0, len(eventsOut))

	observer.Stop()
	close(eventsOut)
}

func TestObserverInvalidQuery(t *testing.T) {
	observer, err := NewObserver("http://localhost:26657", make(chan abcitypes.Event))
	require.NoError(t, err)
	query := "invalid"
	err = observer.Start(query, []string{})
	require.Error(t, err)
}
