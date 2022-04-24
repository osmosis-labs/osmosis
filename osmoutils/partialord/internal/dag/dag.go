package dag

import (
	"fmt"
	"sort"
)

// Dag struct maintains a directed acyclic graph, using adjacency lists to track edges.
type Dag struct {
	// there is a directed edge from u -> v, if directedEdgeList[u][v] = 1
	// there is a directed edge from v -> u, if directedEdgeList[u][v] = -1
	directedEdgeList []map[int]int8
	nodeNameToId     map[string]int
	idToNodeNames    map[int]string
}

func NewDag(nodes []string) Dag {
	nodeNameToId := make(map[string]int, len(nodes))
	idToNodeNames := make(map[int]string, len(nodes))
	directedEdgeList := make([]map[int]int8, len(nodes))
	for i, node := range nodes {
		nodeNameToId[node] = i
		idToNodeNames[i] = node
		directedEdgeList[i] = map[int]int8{}
	}
	if len(nodeNameToId) != len(nodes) {
		panic("provided multiple nodes with the same name")
	}
	return Dag{
		directedEdgeList: directedEdgeList,
		nodeNameToId:     nodeNameToId,
		idToNodeNames:    idToNodeNames,
	}
}

// Copy returns a new dag struct that is a copy of the original dag.
// Edges can be mutated in the copy, without altering the original.
func (dag Dag) Copy() Dag {
	directedEdgeList := make([]map[int]int8, len(dag.nodeNameToId))
	for i := 0; i < len(dag.nodeNameToId); i++ {
		originalEdgeList := dag.directedEdgeList[i]
		directedEdgeList[i] = make(map[int]int8, len(originalEdgeList))
		for k, v := range originalEdgeList {
			directedEdgeList[i][k] = v
		}
	}
	// we re-use nodeNameToId and idToNodeNames as these are fixed at dag creation.
	return Dag{
		directedEdgeList: directedEdgeList,
		nodeNameToId:     dag.nodeNameToId,
		idToNodeNames:    dag.idToNodeNames,
	}
}

func (dag Dag) hasDirectedEdge(u, v int) bool {
	uAdjacencyList := dag.directedEdgeList[u]
	_, exists := uAdjacencyList[v]
	return exists
}

// addEdge adds a directed edge from u -> v.
func (dag *Dag) addEdge(u, v int) error {
	if u == v {
		return fmt.Errorf("can't make self-edge")
	}
	if dag.hasDirectedEdge(v, u) {
		return fmt.Errorf("dag has conflicting edge")
	}
	dag.directedEdgeList[u][v] = 1
	dag.directedEdgeList[v][u] = -1
	return nil
}

// replaceEdge adds a directed edge from u -> v.
// it removes any edge that may already exist between the two.
func (dag *Dag) replaceEdge(u, v int) error {
	if u == v {
		return fmt.Errorf("can't make self-edge")
	}

	dag.directedEdgeList[u][v] = 1
	dag.directedEdgeList[v][u] = -1
	return nil
}

// deleteEdge deletes edges between u and v.
func (dag *Dag) deleteEdge(u, v int) {
	delete(dag.directedEdgeList[u], v)
	delete(dag.directedEdgeList[v], u)
}

func (dag *Dag) AddEdge(u, v string) error {
	uIndex, uExists := dag.nodeNameToId[u]
	vIndex, vExists := dag.nodeNameToId[v]
	if !uExists || !vExists {
		return fmt.Errorf("one of %s, %s does not exist in dag", u, v)
	}
	return dag.addEdge(uIndex, vIndex)
}

// ReplaceEdge adds a directed edge from u -> v.
// it removes any edge that may already exist between the two.
func (dag *Dag) ReplaceEdge(u, v string) error {
	uIndex, uExists := dag.nodeNameToId[u]
	vIndex, vExists := dag.nodeNameToId[v]
	if !uExists || !vExists {
		return fmt.Errorf("one of %s, %s does not exist in dag", u, v)
	}
	return dag.replaceEdge(uIndex, vIndex)
}

// AddFirstElements sets the provided elements to be first in all orderings.
// So if were making an ordering over {A, B, C, D, E}, and elems provided is {D, B, A}
// then we are guaranteed that the total ordering will begin with {D, B, A}
func (dag *Dag) AddFirstElements(nodes ...string) error {
	nodeIds, err := dag.namesToIds(nodes)
	if err != nil {
		return err
	}

	return dag.addFirst(nodeIds)
}

func (dag *Dag) addFirst(nodes []int) error {
	nodeMap := map[int]bool{}
	for i := 0; i < len(nodes); i++ {
		nodeMap[nodes[i]] = true
	}
	// First we add an edge from nodes[-1] to every node in the graph aside from the provided first nodes.
	// then we make nodes[-1] depend on nodes[-2], etc.
	lastOfFirstNodes := nodes[len(nodes)-1]
	for i := 0; i < len(dag.nodeNameToId); i++ {
		// skip any node in the 'first set'
		_, inMap := nodeMap[i]
		if inMap {
			continue
		}
		// We make everything on `lastOfFirstNodes`, and therefore have an edge from `lastOfFirstNodes` to it
		err := dag.replaceEdge(lastOfFirstNodes, i)
		// can't happen by above check
		if err != nil {
			return err
		}
	}

	// Make nodes[i+1] depend on nodes[i]
	for i := len(nodes) - 2; i >= 0; i-- {
		err := dag.replaceEdge(nodes[i], nodes[i+1])
		// can't happen by above check
		if err != nil {
			return err
		}
	}
	return nil
}

// AddLastElements sets the provided elements to be last in all orderings.
// So if were making an ordering over {A, B, C, D, E}, and elems provided is {D, B, A}
// then we are guaranteed that the total ordering will end with {D, B, A}
func (dag *Dag) AddLastElements(nodes ...string) error {
	nodeIds, err := dag.namesToIds(nodes)
	if err != nil {
		return err
	}

	return dag.addLast(nodeIds)
}

func (dag *Dag) addLast(nodes []int) error {
	nodeMap := map[int]bool{}
	for i := 0; i < len(nodes); i++ {
		nodeMap[nodes[i]] = true
	}
	// First we add an edge from every node in the graph aside from the provided last nodes, to nodes[0]
	// then we make nodes[1] depend on nodes[0], etc.
	firstOfLastNodes := nodes[0]
	for i := 0; i < len(dag.nodeNameToId); i++ {
		// skip any node in the 'last set'
		_, inMap := nodeMap[i]
		if inMap {
			continue
		}
		// We make `firstOfLastNodes` depend on every node, and therefore have an edge from each node to `firstOfLastNodes`
		err := dag.replaceEdge(i, firstOfLastNodes)
		// can't happen by above check
		if err != nil {
			return err
		}
	}

	// Make nodes[i] depend on nodes[i-1]
	for i := 1; i < len(nodes); i++ {
		err := dag.replaceEdge(nodes[i-1], nodes[i])
		// can't happen by above check
		if err != nil {
			return err
		}
	}
	return nil
}

func (dag Dag) hasEdges() bool {
	for _, m := range dag.directedEdgeList {
		if len(m) > 0 {
			return true
		}
	}
	return false
}

func (dag *Dag) namesToIds(names []string) ([]int, error) {
	nodeIds := []int{}
	for _, name := range names {
		nodeIndex, nodeExists := dag.nodeNameToId[name]
		if !nodeExists {
			return []int{}, fmt.Errorf("%s does not exist in dag", name)
		}
		nodeIds = append(nodeIds, nodeIndex)
	}
	return nodeIds, nil
}

func (dag Dag) idsToNames(ids []int) []string {
	sortedNames := make([]string, 0, len(ids))
	for i := 0; i < len(dag.nodeNameToId); i++ {
		id := ids[i]
		sortedNames = append(sortedNames, dag.idToNodeNames[id])
	}
	return sortedNames
}

func (dag Dag) hasIncomingEdge(u int) bool {
	adjacencyList := dag.directedEdgeList[u]
	for _, v := range adjacencyList {
		if v == -1 {
			return true
		}
	}
	return false
}

// returns nodes with no incoming edges.
func (dag *Dag) topologicalTopLevelNodes() []int {
	topLevelNodes := []int{}

	for i := 0; i < len(dag.nodeNameToId); i++ {
		if !dag.hasIncomingEdge(i) {
			topLevelNodes = append(topLevelNodes, i)
		}
	}

	return topLevelNodes
}

// Returns a Topological Sort of the DAG, using Kahn's algorithm.
// https://en.wikipedia.org/wiki/Topological_sorting#Kahn's_algorithm
func (dag Dag) TopologicalSort() []string {
	// G is the mutable graph we work on, which we remove edges from.
	G := dag.Copy()
	// L in pseudocode
	sortedIDs := make([]int, 0, len(dag.nodeNameToId))
	topLevelNodes := dag.topologicalTopLevelNodes()

	// while len(topLevelNodes) != 0
	for {
		if len(topLevelNodes) == 0 {
			break
		}
		// pop a node `n`` off of topLevelNodes
		n := topLevelNodes[0]
		topLevelNodes = topLevelNodes[1:]
		// add it to the sorted list
		sortedIDs = append(sortedIDs, n)
		nEdgeList := G.directedEdgeList[n]

		// normally we'd do map iteration, but because we need cross-machine determinism,
		// we gather all the nodes M for which there is an edge n -> m, sort that list,
		// and then iterate over it.
		nodesM := make([]int, 0, len(nEdgeList))
		for m, direction := range nEdgeList {
			if direction != 1 {
				panic("dag: topological sort correctness error. " +
					"Popped node n was expected to have no incoming edges")
			}
			nodesM = append(nodesM, m)
		}

		sort.Ints(nodesM)

		for _, m := range nodesM {
			// remove edge from n -> m
			G.deleteEdge(n, m)
			// if m has no incomingEdges, add to topLevelNodes
			if !G.hasIncomingEdge(m) {
				topLevelNodes = append(topLevelNodes, m)
			}
		}
	}

	if G.hasEdges() {
		fmt.Println(G)
		panic("dag: invalid construction, attempted to topologically sort a tree that is not a dag. A cycle exists")
	}

	return dag.idsToNames(sortedIDs)
}
