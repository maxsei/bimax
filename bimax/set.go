package main

import (
	"fmt"
	"sort"
)

/////////////////////////////////////////////////////////////////////////////////
//                                  Iterators                                  //
/////////////////////////////////////////////////////////////////////////////////

func newSetCh() *setCh {
	return &setCh{ch: make(chan int)}
}

type setCh struct {
	ch chan int
}

func (sc setCh) send(k int) { sc.ch <- k }
func (sc setCh) Iter() int  { return (<-sc.ch) }
func (sc setCh) Close()     { close(sc.ch) }

type Set interface {
	// Set creation
	New() Set
	copySet() Set
	// Key related operations
	keyHas(k int) bool
	mapKeyAdd(k int)
	mapKeyDel(k int)
	// Set cardinality
	Card() int
	// iteration
	keyEach(do func(k int) (done bool))
	Chan() *setCh
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
func (s *SetOp) Add(v int) (ok bool) { return s.keyAdd((v)) }
func (s *SetOp) keyAdd(k int) (ok bool) {
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
func (s *SetOp) Delete(v int) (ok bool) { return s.keyDelete((v)) }
func (s *SetOp) keyDelete(k int) (ok bool) {
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
func (s *SetOp) mutate(mutateFunc func(int) bool, vv []int) (change int) {
	init := s.Card()
	for _, v := range vv {
		mutateFunc((v))
	}
	if init > s.Card() {
		return init - s.Card()
	}
	return s.Card() - init
}

func (s *SetOp) Update(vv ...int) (added int)   { return s.mutate(s.keyAdd, vv) }
func (s *SetOp) Remove(vv ...int) (deleted int) { return s.mutate(s.keyDelete, vv) }

/////////////////////////////////////////////////////////////////////////////////
//                                 Operations                                  //
/////////////////////////////////////////////////////////////////////////////////

// predicateSet compares one set to another. If all is set to true then the
// function will try and find the union of the two sets else it will find the
// difference
func (s *SetOp) predicateSet(other Set, all bool) (product Set) {
	product = s.New()
	// Iterate over smaller set if unionPredicate
	otherOp := &SetOp{other}
	a, b := s, otherOp
	if all && (b.Card() < a.Card()) {
		b = s
		a = otherOp
	}
	a.keyEach(func(k int) (_ bool) {
		// Add keys to product, skipping if predicate matches.
		if b.keyHas(k) != all {
			return
		}
		product.mapKeyAdd(k)
		return
	})
	return
}
func (s *SetOp) intersection(other Set) (product Set) { return s.predicateSet(other, true) }
func (s *SetOp) difference(other Set) (product Set)   { return s.predicateSet(other, false) }
func (s *SetOp) symmetricDifference(other Set) (product Set) {
	union := s.predicateSet(other, true)
	// Diff 1
	diff1 := &SetOp{(&SetOp{other}).predicateSet(union, false)}
	// Diff 2
	diff2 := s.predicateSet(union, false)

	diff2.keyEach(func(k int) (_ bool) {
		diff1.keyAdd(k)
		return
	})
	product = diff1.Set
	return
}

func (s *SetOp) union(other Set) (product Set) {
	smol, larg := s.Set, other
	if larg.Card() < smol.Card() {
		smol = other
		larg = s.Set
	}
	c := &SetOp{larg.copySet()}
	smol.keyEach(func(k int) (_ bool) {
		c.keyAdd(k)
		return
	})
	return c.Set
}
func (s *SetOp) Each(do func(v int) (done bool)) {
	keyDo := func(k int) (done bool) {
		return do((k))
	}
	s.keyEach(keyDo)
}

func (s *SetOp) Get(i int) (v int) {
	if s.Card() <= i {
		panic(fmt.Sprintf("%d out of range of set with cardinality %d", i, s.Card()))
	}
	j := 0
	var result int
	s.keyEach(func(k int) (done bool) {
		if j == i {
			result = k
			return true
		}
		j++
		return
	})
	v = (result)
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
	smol, larg := Set(s), other
	if s.Card() > other.Card() {
		smol, larg = other, s
	}

	// See if the set should include or exclude
	var predicate bool
	smol.keyEach(func(k int) (done bool) {
		predicate = other.keyHas(k)
		done = true
		return
	})

	// Iterate over the smallest set and check for items in other set
	predicateFailed := false
	smol.keyEach(func(k int) (done bool) {
		if larg.keyHas(k) != predicate {
			predicateFailed = true
			return true
		}
		return
	})

	// Return what we were trying to prove all along
	// If the predicate failed to be proven return no joint set.
	if predicateFailed {
		return JointSetNone
	}
	// If predicate is false then neither two sets had similiar elements.
	if predicate == false {
		return JointSetDisJoint
	}
	// Otherwise the sets are joint based on some relation of cardinality
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

func (s *SetOp) Has(v int) bool { return s.keyHas((v)) }

func (s *SetOp) Values() []int {
	vv := make([]int, 0, s.Card())
	s.keyEach(func(k int) (_ bool) {
		vv = append(vv, (k))
		return
	})
	return vv
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
		set: make(map[int]struct{}),
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
type unorderedSet struct {
	set map[int]struct{}
}

/////////////////////////////////////////
//  Start Set Interface Implmentation  //
/////////////////////////////////////////
// Set creation
func (s *unorderedSet) New() Set { return NewSet() }
func (s *unorderedSet) copySet() Set {
	product := NewSet()
	for k, _ := range s.set {
		product.set.set[k] = struct{}{}
	}
	return product
}

// Key related operations
func (s *unorderedSet) keyHas(k int) bool { _, has := s.set[k]; return has }
func (s *unorderedSet) mapKeyAdd(k int)   { s.set[k] = struct{}{} }
func (s *unorderedSet) mapKeyDel(k int)   { delete(s.set, k) }

// Set cardinality
func (s *unorderedSet) Card() int { return len(s.set) }

// Iteration
func (s *unorderedSet) keyEach(do func(k int) (done bool)) {
	for k, _ := range s.set {
		done := do(k)
		if done {
			return
		}
	}
}
func (s *unorderedSet) Chan() (iterator *setCh) {
	iterator = newSetCh()
	go func() {
		i := 0
		for k, _ := range s.set {
			iterator.send(k)
			i++
		}
		// Close channel if not already closed
		if i == s.Card() {
			iterator.Close()
		}
	}()
	return
}

/////////////////////////////////////////
//   End Set Interface Implmentation   //
/////////////////////////////////////////

// Operations that require type assertion this set's type
func (s *UnorderedSet) Intersection(other Set) (product *UnorderedSet) {
	return s.intersection(other).(*UnorderedSet)
}
func (s *UnorderedSet) Difference(other Set) (product *UnorderedSet) {
	return s.difference(other).(*UnorderedSet)
}
func (s *UnorderedSet) SymmetricDifference(other Set) (product *UnorderedSet) {
	return s.symmetricDifference(other).(*UnorderedSet)
}

func (s *UnorderedSet) Copy() (product *UnorderedSet) { return s.copySet().(*UnorderedSet) }
func (s *UnorderedSet) Union(other Set) (product *UnorderedSet) {
	return s.union(other).(*UnorderedSet)
}

// Order returns copy of the current set as an ordered set
func (s *UnorderedSet) Order(cmp func(v1, v2 int) bool) *OrderedSet {
	result := NewOrderedSetWithCapacity(cmp, s.Card())
	for k, _ := range s.set.set {
		result.set.mapKeyAdd(k)
		result.keyAdd(k)
	}
	return result
}

/////////////////////////////////////////////////////////////////////////////////
//                                 Ordered Set                                 //
/////////////////////////////////////////////////////////////////////////////////

func NewOrderedSet(cmp func(v1, v2 int) bool) *OrderedSet {
	return NewOrderedSetWithCapacity(cmp, 0)
}

func NewOrderedSetFromSlice(cmp func(v1, v2 int) bool, vv []int) *OrderedSet {
	return NewOrderedSetWith(cmp, vv...)
}
func NewOrderedSetWith(cmp func(v1, v2 int) bool, vv ...int) *OrderedSet {
	result := NewOrderedSetWithCapacity(cmp, 0)
	result.Update(vv...)
	return result
}

func NewOrderedSetWithCapacity(cmp func(v1, v2 int) bool, capacity int) *OrderedSet {
	set := &orderedSet{
		set:     make(map[int]struct{}),
		keys:    make([]int, 0, capacity),
		compare: cmp,
	}
	return &OrderedSet{&SetOp{set}, set}
}

type OrderedSet struct {
	*SetOp
	set *orderedSet
}

type orderedSet struct {
	set     map[int]struct{}
	keys    []int
	compare func(v1, v2 int) bool
}

func (o *orderedSet) search(k int) int {
	// Find the user defined sort comparison index
	return sort.Search(len(o.keys), func(i int) bool {
		return o.compare((k), (o.keys[i]))
	})
}

/////////////////////////////////////////
//  Start Set Interface Implmentation  //
/////////////////////////////////////////
// Set creation
func (o *orderedSet) New() Set { return NewOrderedSet(o.compare) }
func (o *orderedSet) copySet() Set {
	product := NewOrderedSetWithCapacity(o.compare, o.Card())
	product.set.keys = append(product.set.keys, o.keys...)
	for k, _ := range o.set {
		o.set[k] = struct{}{}
	}
	return product
}

// Key related operations
func (o *orderedSet) keyHas(k int) bool { _, ok := o.set[k]; return ok }
func (o *orderedSet) mapKeyAdd(k int) {
	i := o.search(k)
	// Shift over, copy mem, and insert element at i
	o.keys = append(o.keys, k)
	copy(o.keys[i+1:], o.keys[i:len(o.keys)-1])
	o.keys[i] = k
	// Add to map
	o.set[k] = struct{}{}
}
func (o *orderedSet) mapKeyDel(k int) {
	i := o.search(k)
	// Get the end slice of keys
	// Remove k from the sorted set
	o.keys = append(o.keys[:i], o.keys[i+1:]...)
	// Remove from map
	delete(o.set, k)
}

// Set cardinality
func (o *orderedSet) Card() int { return len(o.keys) }

// Iteration
func (o *orderedSet) keyEach(do func(k int) (done bool)) {
	for _, k := range o.keys {
		done := do(k)
		if done {
			return
		}
	}
}
func (o *orderedSet) Chan() (iterator *setCh) {
	iterator = newSetCh()
	go func() {
		i := 0
		for k, _ := range o.set {
			iterator.send(k)
			i++
		}
		// Close channel if not already closed
		if i == o.Card() {
			iterator.Close()
		}
	}()
	return
}

/////////////////////////////////////////
//   End Set Interface Implmentation   //
/////////////////////////////////////////

// Operations that require type assertion this set's type
func (o *OrderedSet) Intersection(other Set) (product *OrderedSet) {
	return o.intersection(other).(*OrderedSet)
}
func (o *OrderedSet) Difference(other Set) (product *OrderedSet) {
	return o.difference(other).(*OrderedSet)
}
func (o *OrderedSet) SymmetricDifference(other Set) (product *OrderedSet) {
	return o.symmetricDifference(other).(*OrderedSet)
}
func (o *OrderedSet) Copy() (product *OrderedSet) { return o.copySet().(*OrderedSet) }
func (o *OrderedSet) Union(other Set) (product *OrderedSet) {
	return o.union(other).(*OrderedSet)
}

// UnOrder returns copy of the current ordered set as a set
func (o *OrderedSet) Unorder() *UnorderedSet {
	result := NewSet()
	for _, k := range o.set.keys {
		result.keyAdd(k)
	}
	return result
}
