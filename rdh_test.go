package rdh

import (
	"hash/fnv"
	"strconv"
	"testing"
)

func TestNewRbhMap(t *testing.T) {
	h := NewRbhMap(15, func(string2 string) uint32 {
		h := fnv.New32()
		_, _ = h.Write([]byte(string2))
		return h.Sum32()
	})
	t.Log(h.size)
	t.Log(h.shift)
}

func TestRbhMap_Get(t *testing.T) {
	hash := NewRbhMap(16, func(string2 string) uint32 {
		h := fnv.New32()
		_, _ = h.Write([]byte(string2))
		return h.Sum32()
	})
	for i := 0; i < 1000; i++ {
		d := strconv.Itoa(i)
		hash.Put(d, d)
	}
	for i := 0; i < 1000; i++ {
		d := strconv.Itoa(i)
		v, ok := hash.Get(d)
		t.Log(ok, v)
	}
}

func TestGet127(t *testing.T) {
	hash := NewRbhMap(16, func(string2 string) uint32 {
		h := fnv.New32()
		_, _ = h.Write([]byte(string2))
		return h.Sum32()
	})
	hash.Put("127", "127")
	t.Log(hash.Get("127"))
}

func TestHash(t *testing.T) {
	for i := 0; i < 1000; i++ {
		d := strconv.Itoa(i)
		t.Log(hashx(d) >> 60)
	}
}

func hashx(string2 string) uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(string2))
	return h.Sum64()
}
