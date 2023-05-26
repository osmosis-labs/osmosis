package main

type SubgraphPosition struct {
	ID        string `json:"id"`
	Liquidity string `json:"liquidity"`
	TickLower struct {
		TickIdx string `json:"tickIdx"`
		Price0  string "json:\"price0\""
		Price1  string "json:\"price1\""
	} `json:"tickLower"`
	TickUpper struct {
		TickIdx string `json:"tickIdx"`
		Price0  string "json:\"price0\""
		Price1  string "json:\"price1\""
	} `json:"tickUpper"`
	DepositedToken0 string `json:"depositedToken0"`
	DepositedToken1 string `json:"depositedToken1"`
}

type GraphqlRequest struct {
	Query string `json:"query"`
}

type GraphqlResponse struct {
	Data struct {
		Positions []SubgraphPosition `json:"positions"`
	} `json:"data"`
}
