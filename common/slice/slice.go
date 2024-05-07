package slice

func ChunkSlice[T any](slice []T, size int) [][]T {
	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func RmDupl[T comparable](slice []T) []T {
	encountered := map[T]bool{}
	result := make([]T, 0)
	for _, e := range slice {
		if !encountered[e] {
			encountered[e] = true
			result = append(result, e)
		}
	}
	return result
}
