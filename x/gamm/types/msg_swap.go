package types

// SwapMsg defines a simple interface for getting the token denoms on a swap message route.
type SwapMsgRoute interface {
	TokenInDenom() string
	TokenOutDenom() string
	TokenDenomsOnPath() []string
}
