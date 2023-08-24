package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	hash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("key is %s, expect %v, got %s\n", k, v, hash.Get(k))
		}
	}

	hash.Add("8")
	testCases["27"] = "8"

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("key is %s, expect %v, got %s\n", k, v, hash.Get(k))
		}
	}

	// check default hash function
	hash = New(3, nil)
	hash.Add("6", "4", "2")
	fn := crc32.ChecksumIEEE
	keys := []string{
		"2", "11", "23", "27",
	}
	for _, v := range keys {
		hashed := fn([]byte(v))
		idx := sort.Search(len(hash.keys), func(i int) bool {
			return hash.keys[i] >= int(hashed)
		})
		expect := hash.hashMap[hash.keys[idx%len(hash.keys)]]
		if hash.Get(v) != expect {
			t.Errorf("key is %s, expect %v, got %s\n", v, expect, hash.Get(v))
		}
	}

	// check empty
	hash = New(3, nil)
	if hash.Get("1") != "" {
		t.Errorf("empty consistent hash should return empty")
	}
}
