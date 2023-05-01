package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
)

func GetUniV3SubgraphData(pathToSaveAt string) {
	// Set the subgraph URL and query.
	const (
		subgraphURL = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v3"
		query       = `{
			positions(
				first: 1000
				skip: %d,
				orderBy: id
				orderDirection: asc
				where: {pool: "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", liquidity_gt: 0}
				block: {number: 17033391}
			  ) {
				id
				liquidity
				tickLower {
				  tickIdx
				  price0
				  price1
				}
				tickUpper {
				  tickIdx
				  price0
				  price1
				}
				depositedToken0
				depositedToken1
			  }
			}`
	)

	// Set initial skip value and slice for storing positions.
	skip := 0
	allPositions := make([]SubgraphPosition, 0)

	for {
		formatterQuery := fmt.Sprintf(query, skip)

		// Create a GraphQL request
		req := GraphqlRequest{Query: formatterQuery}

		// Encode the request as JSON
		reqBytes, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}

		// Make the HTTP request to the subgraph.
		resp, err := http.Post(subgraphURL, "application/json", bytes.NewBuffer(reqBytes))
		if err != nil {
			fmt.Println("Error making request:", err)
			return
		}
		defer resp.Body.Close()

		// Decode the response from JSON
		var data GraphqlResponse
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			panic(err)
		}

		currentPositions := data.Data.Positions

		if len(currentPositions) == 0 {
			break
		}

		allPositions = append(allPositions, currentPositions...)

		skip += 1000
	}

	fmt.Printf("found %d positions", len(allPositions))

	// sort by id since the subgraph since to be broken
	sort.Slice(allPositions, func(i, j int) bool {
		return allPositions[i].ID < allPositions[j].ID
	})

	// Write the allPositions slice to a JSON file.
	jsonData, err := json.MarshalIndent(allPositions, "", "    ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	if err := os.WriteFile(pathToSaveAt, jsonData, 0644); err != nil {
		fmt.Println("Error writing JSON file:", err)
		return
	}

	fmt.Printf("Data written to %s\n", pathToSaveAt)
}
