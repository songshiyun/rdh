package rdh

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRbhMap(t *testing.T) {
	h := NewRbhMap(15, func(string2 string) uint64 {
		h := fnv.New64()
		_, _ = h.Write([]byte(string2))
		return h.Sum64()
	})
	t.Log(h.size)
	t.Log(h.shift)
}

func TestRbhMap_Get(t *testing.T) {
	//a := assert.New(t)
	hash := NewRbhMap(10, func(string2 string) uint64 {
		h := fnv.New64()
		_, _ = h.Write([]byte(string2))
		return h.Sum64()
	})
	for i := 0; i < 11; i++ {
		d := strconv.Itoa(i)
		hash.Put(d, d)
	}
	entries(hash.Entries()).Iterator(printKV)
	for i := 0; i < 11; i++ {
		d := strconv.Itoa(i)
		hash.Delete(d)
	}
	println("++++++++++++=")
	entries(hash.Entries()).Iterator(printKV)
	/*for i := 0; i < 30; i++ {
		d := strconv.Itoa(i)
		v, ok := hash.Get(d)
		if i < 15 {
			a.Equal(false, ok)
			a.Equal(nil, v)
		} else {
			a.Equal(true, ok)
			a.Equal(d, v)
		}
	}*/
}

func printKV(item *entry) {
	if item == nil {
		println("---")
		return
	}
	fmt.Println(item.key, item.value)
}

func TestPutAndGet(t *testing.T) {
	a := assert.New(t)
	rng := rand.New(rand.NewSource(int64(uint64(time.Now().UnixNano()))))
	keys := make([]string, 1<<15)
	m := NewRbhMap(len(keys), hashfuncT)
	for i := range keys {
		str := strconv.Itoa(rng.Intn(1 << 20))
		keys[i] = str
		m.Put(keys[i], str)
	}
	for i := range keys {
		k := keys[i]
		v, ok1 := m.Get(k)
		a.Equal(true, ok1)
		a.Equal(k, v)
	}
}

func BenchmarkRbhMap_Put(b *testing.B) {
	var m *rbhMap
	for i := 0; i < b.N; i++ {
		if m == nil {
			b.StopTimer()
			m = NewRbhMap(1<<15, func(string2 string) uint64 {
				h := fnv.New64()
				_, _ = h.Write([]byte(string2))
				return h.Sum64()
			})
			b.StartTimer()
		}
		m.Put(strconv.Itoa(i), i)
	}
}

func BenchmarkRbhGet(b *testing.B) {
	rng := rand.New(rand.NewSource(int64(uint64(time.Now().UnixNano()))))
	keys := make([]string, 1<<20)
	m := NewRbhMap(len(keys), hashfuncT)
	for i := range keys {
		str := strconv.Itoa(rng.Intn(1 << 20))
		keys[i] = str
		m.Put(keys[i], str)
	}
	b.ResetTimer()
	for i, j := 0, 0; i < b.N; i, j = i+1, j+1 {
		if j == len(keys) {
			j = 0
		}
		_, _ = m.Get(keys[j])
	}
}

func hashfuncT(string2 string) uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(string2))
	return h.Sum64()
}

func TestGet127(t *testing.T) {
	hash := NewRbhMap(16, func(string2 string) uint64 {
		h := fnv.New64()
		_, _ = h.Write([]byte(string2))
		return h.Sum64()
	})
	hash.Put("127", "127")
	t.Log(hash.Get("127"))
}
