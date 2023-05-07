package util

func ChunkSlice[V any](slice []V, chunkSize int) [][]V {
	var chunks [][]V
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

func ChunkMap[K comparable, V any](myMap map[K]V, chunkSize int) []map[K]V {
	var chunks []map[K]V

	currentSubMap := make(map[K]V)

	for k, v := range myMap {
		currentSubMap[k] = v
		if len(currentSubMap) >= chunkSize {
			chunks = append(chunks, currentSubMap)
			currentSubMap = map[K]V{}
		}
	}
	if len(currentSubMap) > 0 {
		chunks = append(chunks, currentSubMap)
	}

	return chunks
}
