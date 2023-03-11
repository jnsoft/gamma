package helpers

type Ordered interface {
    type int, int8, int16, int32, int64,
        uint, uint8, uint16, uint32, uint64, uintptr,
        float32, float64,
        string
}

func sortedKeys[K Ordered, V any](m map[K]V) ([]K) {
        keys := make([]K, len(m))
        i := 0
        for k := range m {
            keys[i] = k
            i++
        }
        sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
        return keys
}