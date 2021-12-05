package rdh

import (
	"math/bits"
	"unsafe"
)

type rbhEntry struct {
	key   string
	value *entry
	dist  uint32
}

type entry struct {
	key   string
	value interface{}
}

type rbhEntries struct {
	ptr unsafe.Pointer
	len uint32
}

func newRbhEntries(n uint32) rbhEntries {
	size := uintptr(n) * unsafe.Sizeof(rbhEntry{})
	ptr := make([]byte, size)
	return rbhEntries{
		ptr: unsafe.Pointer(&ptr[0]),
		len: n,
	}
}

func (e rbhEntries) at(i uint32) *rbhEntry {
	return (*rbhEntry)(unsafe.Pointer(uintptr(e.ptr) +
		uintptr(i)*unsafe.Sizeof(rbhEntry{})))
}

type HashFunc func(string2 string) uint32

type rbhMap struct {
	entries rbhEntries
	size    uint32
	shift   uint32
	count   uint32
	maxDist uint32
	hash    HashFunc
}

func maxDistForSize(size uint32) uint32 {
	desired := uint32(bits.Len32(size))
	if desired < 4 {
		desired = 4
	}
	return desired
}

func NewRbhMap(cap int, hash HashFunc) *rbhMap {
	m := &rbhMap{}
	m.hash = hash
	m.init(cap)
	return m
}

func (r *rbhMap) init(cap int) {
	if cap < 1 {
		cap = 1
	}
	target := 1 << (uint(bits.Len(uint(2*cap-1))) - 1)
	r.rehash(uint32(target))
}

func (r *rbhMap) rehash(size uint32) {
	oldEntries := r.entries
	size = next2Power(size)
	r.size = size
	r.shift = size - 1
	r.maxDist = maxDistForSize(size)
	r.entries = newRbhEntries(size + r.maxDist)
	r.count = 0

	for i := uint32(0); i < oldEntries.len; i++ {
		e := oldEntries.at(i)
		if e.value != nil {
			r.put(e.key, e.value)
		}
	}
	// discard old entry
}

func (r *rbhMap) Put(k string, v interface{}) {
	ent := &entry{k, v}
	r.put(k, ent)
	return
}

func (r *rbhMap) put(key string, val *entry) {
	maybeExist := true
	n := rbhEntry{key: key, value: val, dist: 0}
	for i := uint32(r.hash(key) & r.shift); ; i++ {
		e := r.entries.at(i)
		if maybeExist && key == e.key {
			//exist
			e.value = val
			return
		}
		if e.value == nil {
			//empty entry
			*e = n
			r.count++
			return
		}
		if e.dist < n.dist {
			n, *e = *e, n
			maybeExist = false
		}
		n.dist++
		if n.dist == r.maxDist {
			r.rehash(2 * r.size)
			i = uint32(r.hash(key)>>r.shift) - 1
			n.dist = 0
			maybeExist = false
		}
	}

}

func (r *rbhMap) Get(k string) (interface{}, bool) {
	v, ok := r.get(k)
	if !ok {
		return nil, ok
	}
	return v.value, ok
}

func (r *rbhMap) get(key string) (*entry, bool) {
	var dist uint32
	for i := uint32(r.hash(key) & r.shift); ; i++ {
		e := r.entries.at(i)
		if key == e.key {
			return e.value, true
		}
		if e.dist < dist {
			// not found
			return nil, false
		}
		dist++
	}
}

func (r *rbhMap) Delete(k string) {
	var dist uint32
	for i := uint32(r.hash(k) & r.shift); ; i++ {
		e := r.entries.at(i)
		if e.key == k {
			r.count--
			for j := i + 1; ; j++ {
				t := r.entries.at(j)
				if t.value == nil {
					//
					*e = rbhEntry{}
					return
				}
				e.key = t.key
				e.value = t.value
				e.dist = t.dist - 1
				e = t // forward back
			}
		}
		if dist > e.dist {
			// not found
			return
		}
		dist++
	}
}

func (r *rbhMap) findByValue(ent *entry) *rbhEntry {
	for i := uint32(0); i < r.entries.len; i++ {
		e := r.entries.at(i)
		if e.value == ent {
			return e
		}
	}
	return nil
}

func next2Power(x uint32) uint32 {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x++
	return x
}
