package types

// CandidatePool is a data structure representing a
// candidate pool to be used for routing.
type CandidatePool struct {
	ID            uint64
	TokenOutDenom string
}

// CandidateRoute is a data structure representing a
// candidate route to be used for routing.
type CandidateRoute struct {
	Pools                     []CandidatePool
	IsCanonicalOrderboolRoute bool
}

// CandidateRoutes is a data structure representing a
// list of candidate routes to be used for routing.
// Additionally, it encapsulates a map of unique pool IDs
// contained in the routes.
type CandidateRoutes struct {
	Routes                     []CandidateRoute
	UniquePoolIDs              map[uint64]struct{}
	ContainsCanonicalOrderbook bool
}
