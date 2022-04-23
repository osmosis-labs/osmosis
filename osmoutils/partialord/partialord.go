package partialord

import "github.com/osmosis-labs/osmosis/v7/osmoutils/partialord/internal/dag"

type PartialOrdering struct {
	dag         dag.Dag
	firstSealed bool
	lastSealed  bool
}

func NewPartialOrdering(elements []string) PartialOrdering {
	// TODO: Ensure elements has no duplicates
	// TODO: sort elements in case caller obtains it via map iteration
	return PartialOrdering{
		dag:         dag.NewDag(elements),
		firstSealed: false,
		lastSealed:  false,
	}
}

func handleDagErr(err error) {
	// all dag errors are logic errors that the intended users of this package should not make.
	if err != nil {
		panic(err)
	}
}

// After marks that A should come after B
func (ord *PartialOrdering) After(A string, B string) {
	// Set that A depends on B / an edge from B -> A
	err := ord.dag.AddEdge(B, A)
	handleDagErr(err)
}

// After marks that A should come before B
func (ord *PartialOrdering) Before(A string, B string) {
	// Set that B depends on A / an edge from A -> B
	err := ord.dag.AddEdge(A, B)
	handleDagErr(err)
}

// Sets elems to be the first elements in the ordering.
// So if were making an ordering over {A, B, C, D, E}, and elems provided is {D, B, A}
// then we are guaranteed that the total ordering will begin with {D, B, A}
func (ord *PartialOrdering) FirstElements(elems ...string) {
	if ord.firstSealed {
		panic("FirstElements has already been called")
	}
	// We make every node in the dag have a dependency on elems[-1]
	// then we change elems[-1] to depend on elems[-2], and so forth.
	err := ord.dag.AddFirstElements(elems...)
	handleDagErr(err)
	ord.firstSealed = true
}

// Sets elems to be the last elements in the ordering.
// So if were making an ordering over {A, B, C, D, E}, and elems provided is {D, B, A}
// then we are guaranteed that the total ordering will end with {D, B, A}
func (ord *PartialOrdering) LastElements(elems ...string) {
	if ord.lastSealed {
		panic("FirstElements has already been called")
	}
	// We make every node in the dag have a dependency on elems[0]
	// then we make elems[1] depend on elems[0], and so forth.
	err := ord.dag.AddLastElements(elems...)
	handleDagErr(err)
	ord.lastSealed = true
}

func (ord *PartialOrdering) TotalOrdering() []string {
	return ord.dag.TopologicalSort()
}
