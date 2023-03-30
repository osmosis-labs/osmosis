// This file defines requests and responses for execute
// cosmwasm messages of a transmuter contract.
package transmuter

import "github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/cosmwasm/msg"

// JoinPoolExecuteMsg is the execute message to join a pool.
type JoinPoolExecuteMsgRequest struct {
	JoinPoolMsg msg.EmptyStruct `json:"join_pool"`
}

type JoinPoolExecuteMsgResponse = msg.EmptyStruct

// ExitPoolExecuteMsg is the execute message to exit a pool.
type ExitPoolExecuteMsg struct {
	ExitPoolMsg msg.EmptyStruct `json:"exit_pool"`
}

type ExitPoolExecuteMsgResponse = msg.EmptyStruct
