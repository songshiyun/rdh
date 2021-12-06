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

type HashFunc func(string2 string) uint64

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
	//size = next2Power(size)
	r.size = size
	r.shift = uint32(64 - bits.Len32(r.size-1))
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
	for i := r.robinHoodHash(key); ; i++ {
		e := r.entries.at(i)
		if maybeExist && key == e.key {
			//exist
			e.value = n.value
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
			i = r.robinHoodHash(key) - 1
			n.dist = 0
			maybeExist = false
		}
	}

}

func (r *rbhMap) robinHoodHash(key string) uint32 {
	return uint32(r.hash(key) >> r.shift)
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
	for i := r.robinHoodHash(key); ; i++ {
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
	for i := r.robinHoodHash(k); ; i++ {
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

type entries []*entry

func (es entries) Iterator(f func(item *entry)) {
	for _, v := range es {
		f(v)
	}
}

func (r *rbhMap) Entries() []*entry {
	if r.entries.len == 0 {
		return nil
	}
	res := make([]*entry, r.entries.len)
	for i := uint32(0); i < r.entries.len; i++ {
		rbEntry := r.entries.at(i)
		res = append(res, rbEntry.value)
	}
	return res
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
