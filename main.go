package main

import "fmt"

func main() {
	order := &[]uint64{}

	// added 5 orders
	newOrder := uint64(2)
	AddOrder(newOrder, order)
	newOrder = uint64(3)
	AddOrder(newOrder, order)
	newOrder = uint64(5)
	AddOrder(newOrder, order)
	newOrder = uint64(25)
	AddOrder(newOrder, order)
	newOrder = uint64(1)
	AddOrder(newOrder, order)

	fmt.Println("OrdersAdded: ", order)

	// O(n) time complexity
	// keeps track of the prefix sum
	prefixSumList := PrefixSum(order)
	fmt.Println("Prefix Sum List: ", prefixSumList)
	// orderSum AKA total order amount to Fill
	orderSum := (*prefixSumList)[len(*order)-1]

	// Fill "5amt" of order
	fmt.Println("****FilledAmt: 6****")
	_, amtFilled := FillOrder(6, orderSum)

	// Claim order id = 1
	fmt.Println("****Claimorder Id: 1****")
	ClaimOrder(1, order, prefixSumList, amtFilled)

	fmt.Println("OrderList: ", order)
	fmt.Println("PrefixSum list: ", prefixSumList)

	// Cancel Order id = 3
	fmt.Println("****Remove order id = 3****")
	CancelOrder(3, order, prefixSumList)

	fmt.Println("OrderList: ", order)
	fmt.Println("PrefixSum list: ", prefixSumList)

}

// AddOrder adds a new order to the order book
func AddOrder(newOrder uint64, order *[]uint64) {
	*order = append(*order, newOrder)
}

// FillOrder fills the order and returns the remaining order
func FillOrder(orderAmountToFill uint64, totalOrderSum uint64) (uint64, uint64) {
	if orderAmountToFill >= totalOrderSum {
		return 0, totalOrderSum
	}

	amtLeftTofill := totalOrderSum - orderAmountToFill
	return amtLeftTofill, orderAmountToFill
}

// ClaimOrder claims the order and updates the prefix sum array.
// NOTE: this takes care of partial claim too
func ClaimOrder(orderToClaimIdx uint64, order *[]uint64, prefixSum *[]uint64, amtFilled uint64) {
	orderFromIndex := (*order)[orderToClaimIdx]
	prefixSumFromIndex := (*prefixSum)[orderToClaimIdx]

	// orderToClaim = AccumSum - TotalFilled
	amtToClaim := orderFromIndex - (prefixSumFromIndex - amtFilled)
	(*order)[orderToClaimIdx] = orderFromIndex - amtToClaim

	// update prefix sum array
	(*prefixSum)[orderToClaimIdx] = prefixSumFromIndex - amtToClaim
}

func CancelOrder(orderToCancel uint64, order *[]uint64, prefixSum *[]uint64) {

	(*prefixSum)[orderToCancel+1] = (*prefixSum)[orderToCancel+1] - (*order)[orderToCancel]
	// remove the value from the order
	*order = append((*order)[:orderToCancel], (*order)[orderToCancel+1:]...)

	// remove the value from the prefix sum array
	*prefixSum = append((*prefixSum)[:orderToCancel], (*prefixSum)[orderToCancel+1:]...)

}

func PrefixSum(order *[]uint64) *[]uint64 {
	// Prefix sum array
	sumPrefix := make([]uint64, len(*order))
	total := uint64(0)
	for i, val := range *order {
		total += val
		sumPrefix[i] = total
	}

	return &sumPrefix
}
