package lockd

// It's size is 32bit on 32bit platform and 64bit on 64bit platform.
// It's uint.

// HashResult contains result of hash function.
type HashResult = uint64

// Hasher takes key and heshes it into HashResult.
type Hasher func(key string) HashResult

// DefaultHasher provides default Hash function for mutexes.
func DefaultHasher(key string) HashResult {
	hash := HashResult(1315423911)
	for i := 0; i < len(key); i++ {
		charByte := HashResult(key[i])
		hash ^= ((hash << 5) + charByte + (hash >> 2))
	}
	return hash
}
