package main

import (
	"fmt"
	"reflect"
	"sort"
	"unsafe"
)

/////////////////////////////////////////////////////////////////////////////////
//                                     Key                                     //
/////////////////////////////////////////////////////////////////////////////////
type key struct {
	data [unsafe.Sizeof(*new(int))]byte
}

func asKey(vPtr *int) *key      { return (*key)(unsafe.Pointer(vPtr)) }
func asKeys(vPtr *[]int) *[]key { return (*[]key)(unsafe.Pointer(vPtr)) }
func asData(k *key) *int        { return (*int)(unsafe.Pointer(k)) }
func asDatum(k *[]key) *[]int   { return (*[]int)(unsafe.Pointer(k)) }

// type Set interface {
// 	Has(k *key) bool
// 	Card() int
// }

/////////////////////////////////////////////////////////////////////////////////
//                                  Builders                                   //
/////////////////////////////////////////////////////////////////////////////////

// NewSet returns an emtpy set
func NewSet() *Set { return NewSetWith() }

// NewSetFromSlice returns a set from ints
func NewSetFromSlice(vv []int) *Set { return NewSetWith(vv...) }

// NewSetWith returns a set with the passed int
func NewSetWith(vv ...int) *Set {
	result := &Set{
		set: make(map[key]struct{}),
	}
	result.Update(vv...)
	return result
}

// Set represent a unique collection of int
// TODO: make a thread safe version <16-01-21, Max Schulte> //
// TODO: make the set ordered based on udfs <16-01-21, Max Schulte> //
type Set struct {
	set map[key]struct{}
}

/////////////////////////////////////////////////////////////////////////////////
//                                  Mutations                                  //
/////////////////////////////////////////////////////////////////////////////////

// Add adds a single element to the set returning if the operation was
// successful
func (s *Set) Add(v int) (ok bool) { return s.keyAdd(asKey(&v)) }
func (s *Set) keyAdd(k *key) (ok bool) {
	// Can only add if we don't have the key
	ok = !s.keyHas(k)
	if !ok {
		return
	}
	// Success
	s.set[*k] = struct{}{}
	return
}

// Delete removes a single element to the set returning if the operation was
// successful
func (s *Set) Delete(v int) (ok bool) { return s.keyDelete(asKey(&v)) }
func (s *Set) keyDelete(k *key) (ok bool) {
	// Can only delete if we have the key
	ok = s.keyHas(k)
	if !ok {
		return
	}
	// Success
	delete(s.set, *k)
	return
}

// mutate is not for external use.  It is intended to make the code for 'Update'
// and 'Remove' smaller
func (s *Set) mutate(mutateFunc func(*key) bool, kk *[]key) (change int) {
	init := len(s.set)
	for _, v := range *kk {
		mutateFunc(&v)
	}
	if init > len(s.set) {
		return init - len(s.set)
	}
	return len(s.set) - init
}

func (s *Set) Update(vv ...int) (added int)   { return s.mutate(s.keyAdd, asKeys(&vv)) }
func (s *Set) Remove(vv ...int) (deleted int) { return s.mutate(s.keyDelete, asKeys(&vv)) }

func (s *Set) Clear() { s = NewSet() }

/////////////////////////////////////////////////////////////////////////////////
//                                 Operations                                  //
/////////////////////////////////////////////////////////////////////////////////

func (s *Set) predicateSet(other *Set, predIn bool) (product *Set) {
	product = NewSet()
	for v, _ := range s.set {
		_, in := other.set[v]
		if in != predIn {
			continue
		}
		product.set[v] = struct{}{}
	}
	return
}

func (s *Set) Intersection(other *Set) (product *Set) { return s.predicateSet(other, true) }
func (s *Set) Difference(other *Set) (product *Set)   { return s.predicateSet(other, false) }

/////////////////////////////////////////////////////////////////////////////////
//                                 Properties                                  //
/////////////////////////////////////////////////////////////////////////////////

type JointSetCategory int

const (
	JointSets JointSetCategory = iota
	JointSetDisJoint
	JointSetNone
	// Joint Set Category
	JointSetSubset
	JointSetSuperset
	JointSetEqualset
)

func (s *Set) JointSetCategory(other *Set) JointSetCategory {
	// TODO: could make this function variadice for multiple set comparison <15-01-21, Max Schulte> //
	// See if the set should include or exclude
	var predicate bool
	for k, _ := range s.set {
		predicate = other.keyHas(&k)
		break
	}

	// Separate set into what is smaller and larger set
	small, large := s, other
	if s.Card() > other.Card() {
		small, large = other, s
	}

	// Iterate over the smallest set and check for items in other set
	for v, _ := range small.set {
		_, in := large.set[v]
		if in != predicate {
			return JointSetNone
		}
	}

	// Return what we were trying to prove all along
	if predicate == false {
		return JointSetDisJoint
	}
	// Joint Set Category
	switch {
	case s.Card() == other.Card():
		return JointSetEqualset
	case s.Card() < other.Card():
		return JointSetSubset
	case s.Card() > other.Card():
		return JointSetSuperset
	}
	panic("unreachable")
}

func (s *Set) IsDisjoint(other *Set) bool { return s.JointSetCategory(other) == JointSetDisJoint }
func (s *Set) IsSubset(other *Set) bool   { return s.JointSetCategory(other) == JointSetSubset }
func (s *Set) IsSuperset(other *Set) bool { return s.JointSetCategory(other) == JointSetSuperset }
func (s *Set) IsEqual(other *Set) bool    { return s.JointSetCategory(other) == JointSetEqualset }

func (s *Set) keyHas(k *key) bool { _, in := s.set[*k]; return in }
func (s *Set) Has(v int) bool     { return s.keyHas(asKey(&v)) }

func (s *Set) Card() int { return len(s.set) }

/////////////////////////////////////////////////////////////////////////////////
//                         Values and Value Iterators                          //
/////////////////////////////////////////////////////////////////////////////////

func (s *Set) Values() []int {
	vv := make([]int, 0, len(s.set))
	iterator := s.Iterator()
	for {
		// Get next value
		v, done := iterator.Iter()
		if done {
			break
		}
		vv = append(vv, v)
	}
	return vv
}

type setIterator struct {
	set  *Set
	iter *reflect.MapIter
}

func (si setIterator) Iter() (v int, done bool) {
	done = si.iter.Next()
	done = !done
	if done {
		return
	}
	key := si.iter.Key().Interface().(key)
	v = *(*int)(unsafe.Pointer(&key))
	return
}

func (s *Set) Iterator() *setIterator {
	return &setIterator{set: s, iter: reflect.ValueOf(s.set).MapRange()}
}

func (s *Set) IteratorChan() (iterator chan int) {
	go func() {
		for key, _ := range s.set {
			iterator <- *(asData(&key))
		}
		close(iterator)
	}()
	return
}

/////////////////////////////////////////////////////////////////////////////////
//                                    Misc                                     //
/////////////////////////////////////////////////////////////////////////////////

// String returns set{<values>}
func (s *Set) String() string {
	vv := s.Values()
	vvStr := []byte(fmt.Sprint(vv))
	vvStr[0] = '{'
	vvStr[len(vvStr)-1] = '}'
	return "set" + string(vvStr)
}

// Copy returns a copy of the current set
func (s *Set) Copy() *Set {
	result := NewSet()
	for k, v := range s.set {
		result.set[k] = v
	}
	return result
}

// Order returns copy of the current set as an ordered set
func (s *Set) Order(cmp func(v1, v2 *int) bool) *OrderedSet {
	return nil
}

/////////////////////////////////////////////////////////////////////////////////
//                                 Ordered Set                                 //
/////////////////////////////////////////////////////////////////////////////////

func NewOrderedSet(cmp func(v1, v2 *int) bool) *OrderedSet {
	return NewOrderedSetWithCapacity(cmp, 0)
}

func NewOrderedSetWithCapacity(cmp func(v1, v2 *int) bool, capacity int) *OrderedSet {
	set := NewSet()
	result := OrderedSet{
		set,
		set,
		make([]key, 0, capacity),
		cmp,
	}
	return &result
}

type OrderedSet struct {
	*Set
	set     *Set
	keys    []key
	compare func(v1, v2 *int) bool
}

func (o *OrderedSet) keyIdx(k *key) int {
	// Send key and ith key to comparison function
	return sort.Search(len(o.keys), func(i int) bool { return o.compare(asData(k), asData(&o.keys[i])) })
}

// Add adds a single element to the set returning if the operation was
// successful
func (o *OrderedSet) Add(v int) (ok bool) { return o.keyAdd(asKey(&v)) }
func (o *OrderedSet) keyAdd(k *key) (ok bool) {
	// Try adding key to unordered set
	ok = o.set.keyAdd(k)
	if !ok {
		return
	}
	i := o.keyIdx(k)
	// If not found in vv just append it
	if i == len(o.keys) {
		o.keys = append(o.keys, *k)
		return
	}
	// Shift over, copy mem, and insert element at i
	o.keys = append(o.keys, key{})
	copy(o.keys[i+1:], o.keys[i:len(o.keys)-1])
	o.keys[i] = *k
	return true
}

// Delete removes a single element to the set returning if the operation was
// successful
func (o *OrderedSet) Delete(v int) (ok bool) { return o.keyDelete(asKey(&v)) }
func (o *OrderedSet) keyDelete(k *key) (ok bool) {
	// Try deleting from unordered set
	ok = o.set.keyDelete(k)
	if !ok {
		return ok
	}
	// Remove the ith key from keys
	i := o.keyIdx(k)
	o.keys = append(o.keys[:i], o.keys[i+1:]...)
	return
}
