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

/////////////////////////////////////////////////////////////////////////////////
//                                  Iterators                                  //
/////////////////////////////////////////////////////////////////////////////////

type setCh chan *key

func (sc setCh) keyIter() (k *key) { return <-sc }
func (sc setCh) Iter() (v *int)    { return asData(<-sc) }

type setIterator interface {
	keyIter() (*key, bool)
	keyPop() *key
}

func newSliceIterator(keys []key) *sliceIterator {
	return &sliceIterator{i: 0, keys: keys}
}

type sliceIterator struct {
	i    int
	keys []key
}

func (si *sliceIterator) keyIter() (k *key, done bool) {
	done = si.i == len(si.keys)
	si.i++
	if done {
		return
	}
	k = &si.keys[si.i]
	return
}

func (si *sliceIterator) keyPop() (k *key) {
	k = &si.keys[len(si.keys)-1]
	return
}

type mapkeyIterator struct {
	iter *reflect.MapIter
}

func (si mapkeyIterator) keyIter() (k *key, done bool) {
	done = !si.iter.Next()
	if done {
		return
	}
	// TODO: avoid copying here by takinga pointer to the key perhaps <16-01-21, Max Schulte> //
	kCopy := si.iter.Key().Interface().(key)
	k = &kCopy
	return
}

// Save as key iter
func (si *mapkeyIterator) keyPop() (k *key) {
	k, _ = si.keyIter()
	return
}

type setIteratorOp struct {
	setIterator
}

func (si setIteratorOp) Iter() (v *int, done bool) {
	var k *key
	k, done = si.keyIter()
	v = asData(k)
	return
}

type Set interface {
	// Creates a new instance of the same type of set
	New() Set

	// 'key' methods
	keyHas(k *key) bool
	mapKeyAdd(k *key)
	mapKeyDel(k *key)

	// Size of the set
	Card() int

	// Channel and iterator based iteration
	Iterator() *setIteratorOp
	Chan() setCh
}

// SetOp gives all Interfaces that implement Set access to the following methods:
type SetOp struct {
	Set
}

/////////////////////////////////////////////////////////////////////////////////
//                                  Mutations                                  //
/////////////////////////////////////////////////////////////////////////////////

// Add adds a single element to the set returning if the operation was
// successful
func (s *SetOp) Add(v int) (ok bool) { return s.keyAdd(asKey(&v)) }
func (s *SetOp) keyAdd(k *key) (ok bool) {
	// Can only add if we don't have the key
	ok = !s.keyHas(k)
	if !ok {
		return
	}
	// Success
	s.mapKeyAdd(k)
	return
}

// Delete removes a single element to the set returning if the operation was
// successful
func (s *SetOp) Delete(v int) (ok bool) { return s.keyDelete(asKey(&v)) }
func (s *SetOp) keyDelete(k *key) (ok bool) {
	// Can only delete if we have the key
	ok = s.keyHas(k)
	if !ok {
		return
	}
	// Success
	s.mapKeyDel(k)
	return
}

// mutate is not for external use.  It is intended to make the code for 'Update'
// and 'Remove' smaller
func (s *SetOp) mutate(mutateFunc func(*key) bool, kk *[]key) (change int) {
	init := s.Card()
	for _, v := range *kk {
		mutateFunc(&v)
	}
	if init > s.Card() {
		return init - s.Card()
	}
	return s.Card() - init
}

func (s *SetOp) Update(vv ...int) (added int)   { return s.mutate(s.keyAdd, asKeys(&vv)) }
func (s *SetOp) Remove(vv ...int) (deleted int) { return s.mutate(s.keyDelete, asKeys(&vv)) }

// Pop returns the last item in ordered sets and is the same as calling 'Iter()'
// on unordered set iterator
func (s *SetOp) Pop() (v *int) { return asData(s.Iterator().keyPop()) }

/////////////////////////////////////////////////////////////////////////////////
//                                 Operations                                  //
/////////////////////////////////////////////////////////////////////////////////

func (s *SetOp) predicateSet(other Set, predIn bool) (product Set) {
	product = s.New()
	for iterator := s.Iterator(); ; {
		k, done := iterator.keyIter()
		if done {
			break
		}
		if other.keyHas(k) != predIn {
			continue
		}
		product.mapKeyAdd(k)
	}
	return
}

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

func (s *SetOp) JointSetCategory(other Set) JointSetCategory {
	// TODO: could make this function variadice for multiple set comparison <15-01-21, Max Schulte> //
	// TODO: deal with empty set which is both a disjoint and a subset of other <16-01-21, Max Schulte> //

	// Separate set into what is smaller and larger set
	small, large := Set(s), other
	if s.Card() > other.Card() {
		small, large = other, s
	}

	// See if the set should include or exclude
	var predicate bool
	iterator := small.Iterator()
	k, done := iterator.keyIter()
	predicate = other.keyHas(k)

	// Iterate over the smallest set and check for items in other set
	for done {
		k, done := iterator.keyIter()
		if done {
			break
		}
		if large.keyHas(k) != predicate {
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

func (s *SetOp) IsDisjoint(other Set) bool {
	return s.JointSetCategory(other) == JointSetDisJoint
}
func (s *SetOp) IsSubset(other Set) bool {
	return s.JointSetCategory(other) == JointSetSubset
}
func (s *SetOp) IsSuperset(other Set) bool {
	return s.JointSetCategory(other) == JointSetSuperset
}
func (s *SetOp) IsEqual(other Set) bool {
	return s.JointSetCategory(other) == JointSetEqualset
}

func (s *SetOp) Has(v int) bool { return s.keyHas(asKey(&v)) }

func (s *SetOp) Values() []int {
	vv := make([]int, 0, s.Card())

	for iterator := s.Iterator(); ; {
		// Get next value
		k, done := iterator.keyIter()
		if done {
			break
		}
		vv = append(vv, *asData(k))
	}
	return vv
}

// Copy returns a copy of the current set
func (s *SetOp) Copy() Set {
	result := s.New()
	for iterator := s.Iterator(); ; {
		k, done := iterator.keyIter()
		if done {
			break
		}
		result.mapKeyAdd(k)
	}
	return result
}

// String returns set{<values>}
func (s *SetOp) String() string {
	vv := s.Values()
	vvStr := []byte(fmt.Sprint(vv))
	vvStr[0] = '{'
	vvStr[len(vvStr)-1] = '}'
	return "set" + string(vvStr)
}

/////////////////////////////////////////////////////////////////////////////////
//                                  Builders                                   //
/////////////////////////////////////////////////////////////////////////////////

// NewSet returns an emtpy set
func NewSet() *UnorderedSet { return NewSetWith() }

// NewSetFromSlice returns a set from ints
func NewSetFromSlice(vv []int) *UnorderedSet { return NewSetWith(vv...) }

// NewSetWith returns a set with the passed int
func NewSetWith(vv ...int) *UnorderedSet {
	set := &unorderedSet{
		set: make(map[key]struct{}),
	}
	result := &UnorderedSet{&SetOp{set}, set}
	result.Update(vv...)
	return result
}

type UnorderedSet struct {
	*SetOp
	set *unorderedSet
}

// UnorderedSet represent a unique collection of int
// TODO: make a thread safe version <16-01-21, Max Schulte> //
// TODO: make the set ordered based on udfs <16-01-21, Max Schulte> //
type unorderedSet struct {
	set map[key]struct{}
}

// Set interface implementation
func (s *unorderedSet) New() Set           { return NewSet() }
func (s *unorderedSet) keyHas(k *key) bool { _, has := s.set[*k]; return has }
func (s *unorderedSet) mapKeyAdd(k *key)   { s.set[*k] = struct{}{} }
func (s *unorderedSet) mapKeyDel(k *key)   { delete(s.set, *k) }
func (s *unorderedSet) Card() int          { return len(s.set) }
func (s *unorderedSet) Iterator() *setIteratorOp {
	return &setIteratorOp{&mapkeyIterator{reflect.ValueOf(s.set).MapRange()}}
}
func (s *unorderedSet) Chan() (iterator setCh) {
	iterator = make(setCh)
	go func() {
		for key, _ := range s.set {
			iterator <- &key
		}
		close(iterator)
	}()
	return
}

func (s *unorderedSet) Clear() { s = s.New().(*unorderedSet) }

// Type assertions for set operations
func (s *UnorderedSet) Intersection(other Set) (product *UnorderedSet) {
	return s.predicateSet(other, true).(*UnorderedSet)
}
func (s *UnorderedSet) Difference(other Set) (product *UnorderedSet) {
	return s.predicateSet(other, false).(*UnorderedSet)
}

// Order returns copy of the current set as an ordered set
func (s *UnorderedSet) Order(cmp func(v1, v2 *int) bool) *OrderedSet {
	result := NewOrderedSetWithCapacity(cmp, s.Card())
	for k, _ := range s.set.set {
		result.set.mapKeyAdd(&k)
		result.keyAdd(&k)
	}
	return result
}

/////////////////////////////////////////////////////////////////////////////////
//                                 Ordered Set                                 //
/////////////////////////////////////////////////////////////////////////////////

func NewOrderedSet(cmp func(v1, v2 *int) bool) *OrderedSet {
	return NewOrderedSetWithCapacity(cmp, 0)
}

func NewOrderedSetFromSlice(cmp func(v1, v2 *int) bool, vv []int) *OrderedSet {
	return NewOrderedSetWith(cmp, vv...)
}
func NewOrderedSetWith(cmp func(v1, v2 *int) bool, vv ...int) *OrderedSet {
	result := NewOrderedSetWithCapacity(cmp, 0)
	result.Update(vv...)
	return result
}

func NewOrderedSetWithCapacity(cmp func(v1, v2 *int) bool, capacity int) *OrderedSet {
	set := &orderedSet{
		set:     make(map[key]struct{}),
		keys:    make([]key, 0, capacity),
		compare: cmp,
	}
	return &OrderedSet{&SetOp{set}, set}
}

type OrderedSet struct {
	*SetOp
	set *orderedSet
}

type orderedSet struct {
	set     map[key]struct{}
	keys    []key
	compare func(v1, v2 *int) bool
}

func (o *orderedSet) search(k *key) int {
	// Find the user defined sort comparison index
	i := sort.Search(len(o.keys), func(i int) bool {
		return o.compare(asData(k), asData(&o.keys[i]))
	})
	return i
}

// Set interface implementation
func (o *orderedSet) New() Set           { return NewOrderedSet(o.compare) }
func (o *orderedSet) keyHas(k *key) bool { _, ok := o.set[*k]; return ok }
func (o *orderedSet) mapKeyAdd(k *key) {
	i := o.search(k)
	// Shift over, copy mem, and insert element at i
	o.keys = append(o.keys, *k)
	copy(o.keys[i+1:], o.keys[i:len(o.keys)-1])
	o.keys[i] = *k
	// Add to map
	o.set[*k] = struct{}{}
}
func (o *orderedSet) mapKeyDel(k *key) {
	i := o.search(k)
	// Get the end slice of keys
	// Remove k from the sorted set
	o.keys = append(o.keys[i:], o.keys[i+1:]...)
	// Remove from map
	delete(o.set, *k)
}
func (o *orderedSet) Card() int { return len(o.keys) }
func (o *orderedSet) Iterator() *setIteratorOp {
	return &setIteratorOp{newSliceIterator(o.keys)}
}
func (o *orderedSet) Chan() (iterator setCh) {
	iterator = make(setCh)
	go func() {
		for _, k := range o.keys {
			iterator <- &k
		}
		close(iterator)
	}()
	return
}

// Type assertions for set operations
func (o *OrderedSet) Intersection(other Set) (product *OrderedSet) {
	return o.predicateSet(other, true).(*OrderedSet)
}
func (o *OrderedSet) Difference(other Set) (product *OrderedSet) {
	return o.predicateSet(other, false).(*OrderedSet)
}

// UnOrder returns copy of the current ordered set as a set
func (o *OrderedSet) Unorder() *UnorderedSet {
	result := NewSet()
	for _, k := range o.set.keys {
		result.keyAdd(&k)
	}
	return result
}
