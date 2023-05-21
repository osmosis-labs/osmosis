package concentrated_liquidity

import (
	"fmt"

	iavlstore "github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/iavl"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/osmoutils/sumtree"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// func main() {
// 	tree, err := CreateLootTree()
// 	if err != nil {
// 		fmt.Println("ERROR OCCOURED: ", err)
// 		return
// 	}

// 	orders := []int64{7, 3, 2, 5, 6, 11, 1, 8}
// 	AddOrders(tree, orders)

// 	/************* STARTING TO FILL ORDERS **********/
// 	totalExistingOrders := getOrdersTotal(orders)

// 	// swap with 3 osmo happens
// 	newTotal, orderFilled, err := FillOrder(tree, totalExistingOrders, 3)
// 	if err != nil {
// 		fmt.Println("ERROR OCCOURED WHILE FILLING ORDER: ", err)
// 		return
// 	}

// 	fmt.Println(newTotal, orderFilled)

// 	// swap with 32 osmo happens
// 	newTotal, orderFilled, err = FillOrder(tree, newTotal, 32)
// 	if err != nil {
// 		fmt.Println("ERROR OCCOURED WHILE FILLING ORDER: ", err)
// 		return
// 	}

// 	// swap with 8 osmo happens
// 	newTotal, orderFilled, err = FillOrder(tree, newTotal, 8)
// 	if err != nil {
// 		fmt.Println("ERROR OCCOURED WHILE FILLING ORDER: ", err)
// 		return
// 	}

// 	fmt.Println(newTotal, orderFilled)

// 	/************* ORDER COMPLETELY FILLED **********/

// 	tree.DebugVisualize()

// }

func CreateLootTree() (sumtree.Tree, error) {
	db := dbm.NewMemDB()
	tree, err := iavl.NewMutableTree(db, 100, false)
	if err != nil {
		return sumtree.Tree{}, err
	}
	kvstore := iavlstore.UnsafeNewStore(tree)
	orderTree := sumtree.NewTree(kvstore, 3)

	return orderTree, nil
}

// add the order to the tree
func AddOrders(orderTree sumtree.Tree, orders []int64) error {
	for i, order := range orders {
		s := fmt.Sprintf("%d", i)
		byteKey := []byte(s)
		// orders to fill
		orderTree.Set(byteKey, sdk.NewIntFromUint64(uint64(order)))
	}
	return nil
}

// fillOrder is called after swap happens when the price is in that tick boundry
func FillOrder(orderTree sumtree.Tree, existingordersSum int64, amountToFill int64) (int64, int64, error) {
	var newOrderSum int64
	if amountToFill > existingordersSum {
		// This means that all orders have been filled
		return 0, 0, nil
	}

	// subtract from the parent sum
	newOrderSum = existingordersSum - amountToFill

	// ? once the order gets filled remove it from the tree
	return newOrderSum, amountToFill, nil
}

// claiming order results when ALL orders has been filled
// function get called when the current tick has been cross
func ClaimOrder(orderTree sumtree.Tree, key []byte, totalFilled int64) error {
	userPositionAmount, orderStart := orderTree.GetWithAccumulatedRange(key)
	//orderComplete := userPositionAmount.Add(orderStart)

	if totalFilled < orderStart {
		return fmt.Errorf("No limits executed yet")
	} else if totalFilled > orderStart+userPositionAmount {
		// order wants to be fully claimed

		totalFilled = totalFilled - userPositionAmount
		// claim leaf routine
		orderTree.ClaimLeafRoutine(key)

	} else {
		// TODO: explore option where leaf can have multiple values
		// the order wants to be partially claimed
		remainingAmount := userPositionAmount - (totalFilled - orderStart)

		if remainingAmount > 0 {
			// update the current node
			orderTree.Set(key, sdk.NewInt(remainingAmount))
		}
		amountClaimed := userPositionAmount - remainingAmount
		totalFilled = totalFilled - amountClaimed
	}

	return nil
}

// remove the order from the tree
func CancelOrder() {
	// You just have a sum-tree of things, if you want to cancel, you do:
	// Check have I been filled
	// 1. If filled, fail cancellation
	// 2. If partial filled, partial claim result, cancel remainder
	// 3. If not filled cancel remainder
	// To cancel remainder, you decrement the value in one node of the sumtree
}

func getOrdersTotal(orders []int64) int64 {
	sum := int64(0)
	for _, order := range orders {
		sum += order
	}

	return sum
}
