package sumtree

var nodeKeyPrefix []byte

const nodeKeyPrefixLen = 5

func init() {
	// nodeKeyPrefix is assumed to be 5 bytes
	nodeKeyPrefix = []byte("node/")
	if len(nodeKeyPrefix) != nodeKeyPrefixLen {
		panic("Invalid constants in accumulation store")
	}
}
