package keeper

func sliceIndex(IDs []uint64, ID uint64) int {
	for index, id := range IDs {
		if id == ID {
			return index
		}
	}
	return -1
}

func removeValue(IDs []uint64, ID uint64) ([]uint64, int) {
	index := sliceIndex(IDs, ID)
	if index < 0 {
		return IDs, index
	}
	IDs[index] = IDs[len(IDs)-1] // set last element to index
	return IDs[:len(IDs)-1], index
}

// combineKeys combine bytes array into a single bytes
func combineKeys(keys ...[]byte) []byte {
	// TODO: should add test for combineKeys, like bytes ordering
	combined := []byte{}
	for _, key := range keys {
		combined = append(combined, key...)
	}
	return combined
}