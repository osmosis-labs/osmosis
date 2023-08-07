package main

import (
	"fmt"
	"sort"
)

// Define a struct for the validators
type Validator struct {
	Name          string
	Staked        float64
	RatioInValset float64
	UndelegateAmt float64
	VRatio        float64
}

func main() {
	// Initialize validators X and Y
	validators := []Validator{
		{Name: "X", Staked: 15.0, RatioInValset: 0.5},
		{Name: "Y", Staked: 5.0, RatioInValset: 0.5},
	}

	amount := 20.0

	// Step 1 and 2
	for i := range validators {
		validators[i].UndelegateAmt = validators[i].RatioInValset * amount
		validators[i].VRatio = validators[i].UndelegateAmt / validators[i].Staked
	}

	// Step 3
	maxVRatio := validators[0].VRatio
	for _, v := range validators {
		if v.VRatio > maxVRatio {
			maxVRatio = v.VRatio
		}
	}

	if maxVRatio <= 1 {
		fmt.Println("All validators can undelegate normally, falling to happy path.")
		return
	}

	// Step 4
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].VRatio > validators[j].VRatio
	})

	// Step 5
	targetRatio := 1.0
	amountRemaining := amount

	// Step 6
	for len(validators) > 0 && validators[0].VRatio > targetRatio {
		fmt.Printf("Undelegating fully from validator %s\n", validators[0].Name)
		amountRemaining -= validators[0].Staked
		targetRatio *= (1 - validators[0].RatioInValset)
		validators = validators[1:]

		fmt.Println(validators)
	}

	// Step 7
	fmt.Println("Distributing the remaining tokens normally amongst the remaining validators.")
	for _, v := range validators {
		undelegate := v.RatioInValset * amountRemaining
		fmt.Printf("Undelegate %.2f from validator %s\n", undelegate, v.Name)
	}
}
