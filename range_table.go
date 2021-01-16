package main

import (
	"fmt"
	"sort"
	"unsafe"
)

// NewRange safely constructs a range returning and error if the range is valid
// or not.  If you know what you are doing you can always create a Range by
// manually constructing it with 'Result{start, stop}' and calling
// 'Range.Valid()' later to check its validity as a range
func NewRange(start, stop int) (*Range, error) {
	result := Range([2]int{start, stop})
	if err := result.Valid(); err != nil {
		return nil, err
	}
	return &result, nil
}

// Range represents a range of numbers or a way to slice array's, matracies,
// tensors, etc.
type Range [2]int

// Start gets the first value of the range
func (r *Range) Start() int { return r[0] }

// Stop gets the end value of the range
func (r *Range) Stop() int { return r[1] }

// Len of the range
func (r *Range) Len() int { return r.Stop() - r.Start() }

// Step implements Gorgonia Slicing
func (r *Range) Step() int { return 1 }

// String
func (r *Range) String() string { return fmt.Sprintf("Range [%d, %d)", r.Start(), r.Stop()) }

// In check to see if a value is in a range
func (r *Range) In(x int) bool { return (r.Start() <= x) && (x < r.Stop()) }

// Valid returns if the range is a valid range or not
func (r *Range) Valid() error {
	if r.Start() > r.Stop() {
		return fmt.Errorf("%v invalid", r)
	}
	return nil
}

// Offset adds an offset to the range by increasing/decreasing the start and
// stop simultaneously.
// Same as but without any error:
// r.SetStart(r.Start() + offset)
// r.SetEnd(r.End() + offset)
func (r *Range) Offset(offset int) {
	r[0] += offset
	r[1] += offset
}

// set is for internal use within funcs 'SetStart' and 'SetStop' is
// responsible for setting each range. Returns whter or not setting the value is
// valid or not after setting the value
func (r *Range) set(x, i int) error {
	r[i] = x
	return r.Valid()
}

// SetStart see func 'set' but do not it call directly
func (r *Range) SetStart(x int) error { return r.set(x, 0) }

// SetStop see func 'set' but do not it call directly
func (r *Range) SetStop(x int) error { return r.set(x, 1) }

// SetStop see func 'set' but do not it call directly

// Same as
func (r *Range) SetLen(x int) { r[1] = r.Start() + x }

func NewRangeTable() *RangeTable {
	return &RangeTable{ap: make(map[string]int)}
}

type RangeTable struct {
	ap          map[string]int // access patern for mapping names
	names       []string       // Names of the ranges
	ranges      [][]int        // Points to 'rangesStack'
	rangesStack []int          // Underlying range data whose length is equal to len(Range) * len(names) and is randomly accessed
}

func (rt *RangeTable) Ranges() []Range {
	resultData := make([]int, len(rt.rangesStack))
	copy(resultData, rt.rangesStack)
	result := make([]Range, len(rt.ranges))
	for i := range result {
		result[i] = *(*Range)(unsafe.Pointer(&resultData[i*len(result[i])]))
	}
	return result
}
func (rt *RangeTable) Names() []string { return rt.names }
func (rt *RangeTable) Len() int        { return len(rt.ranges) }
func (rt *RangeTable) Empty() bool     { return len(rt.ranges) == 0 }

func (rt *RangeTable) Insert(name string, r Range) error {
	// Check range validity
	if err := r.Valid(); err != nil {
		return err
	}
	// Find the index of r.Start and r.Stop belong to
	i, foundStart := rt.searchIdx(r.Start())
	j, foundStop := rt.searchIdx(r.Stop())

	// If name exists in the name we need to mutate data
	if k, ok := rt.indexByName(name); ok {
		// oldRange := (*Range)(unsafe.Pointer(&rt.ranges[k][0]))
		// Overwrite oldRange data
		copy(rt.ranges[k], r[:])
		// No moving result and name indicies necessary
		if (k == i) || (i == rt.Len()-1) {
			return nil
		}

		// Move all name mapped indicies back one in range (i, k)
		// Reorder names in the map
		toFrom := [2]int{i, k}
		fmt.Printf("toFrom = %+v\n", toFrom)
		sort.Ints(toFrom[:])
		sign := (toFrom[1] - toFrom[0]) / (k - i)
		for _, name := range rt.names[toFrom[0]:toFrom[1]] {
			rt.ap[name] += sign
		}
		rt.ap[name] = i

		// Swap ranges and names
		// Ranges
		rTemp := rt.ranges[k]
		rt.ranges[k] = rt.ranges[i]
		rt.ranges[i] = rTemp
		// Names
		nameTemp := rt.names[k]
		rt.names[k] = rt.names[i]
		rt.names[i] = nameTemp
		return nil
	}

	// If new range lies in two different ranges cannot insert
	if i != j {
		return fmt.Errorf("could not insert %v", r)
	}

	// If new range is within or equal to range in existing in rt
	if foundStop && foundStart {
		// Rename map entry
		delete(rt.ap, rt.names[i])
		copy([]byte(rt.names[i]), []byte(name))
		rt.ap[rt.names[i]] = i
		// Copy new data over
		if n := copy(rt.ranges[i], r[:]); n != len(r) {
			panic("copy failed")
		}
		return nil
	}

	// Move all name mapped indicies up one in range i:
	for _, name := range rt.names[i:] {
		rt.ap[name] += 1
	}
	// Insert r to rt.ranges at i
	// Append range to rangesStack
	newRange := Range{}
	copy(newRange[:], r[:])
	rt.rangesStack = append(rt.rangesStack, newRange[:]...)
	// Point a new range at this new data
	rt.ranges = append(rt.ranges[:i], append([][]int{newRange[:]}, rt.ranges[i:]...)...)

	// Insert name to rt.names at i
	rt.names = append(rt.names[:i], append([]string{name}, rt.names[i:]...)...)
	rt.ap[rt.names[i]] = i
	return nil
}

func (rt *RangeTable) InsertMust(name string, r Range) {
	if err := rt.Insert(name, r); err != nil {
		panic(err)
	}
}

func (rt *RangeTable) Search(x int) (name string, ok bool) {
	idx, ok := rt.searchIdx(x)
	if !ok {
		return "", ok
	}
	return rt.names[idx], ok
}

func (rt *RangeTable) Delete(name string) (ok bool) {
	// If empty return false
	if rt.Empty() {
		return false
	}
	idx, ok := rt.indexByName(name)
	if !ok {
		return false
	}

	// Decrement all mappings greater than i
	if idx < (rt.Len() - 1) {
		for _, name := range rt.names[idx+1:] {
			rt.ap[name] -= 1
		}
	}
	// Remove range
	rt.ranges = append(rt.ranges[:idx], rt.ranges[idx+1:]...)
	// Remove remove name
	rt.names = append(rt.names[:idx], rt.names[idx+1:]...)
	// Remove data from ranges Stack
	rt.rangesStack = append(rt.rangesStack[:idx*len(Range{})], rt.rangesStack[(idx+1)*len(Range{}):]...)
	// Remove from map
	delete(rt.ap, name)

	return true
}

func (rt *RangeTable) searchIdx(x int) (idx int, ok bool) {
	var foundIdx int
	idx = sort.Search(len(rt.ranges), func(i int) bool {
		r := (*Range)(unsafe.Pointer(&rt.ranges[i][0]))
		if !ok {
			ok = r.In(x)
			foundIdx = i
		}
		return x <= r.Start()
	})
	if ok {
		return foundIdx, ok
	}
	return idx, ok
}

func (rt *RangeTable) RangeByName(name string) (r Range, ok bool) {
	var result Range
	rPtr, ok := rt.unsafeRangeByName(name)
	copy(result[:], rPtr[:])
	return result, ok
}

func (rt *RangeTable) unsafeRangeByName(name string) (r *Range, ok bool) {
	i, ok := rt.indexByName(name)
	if !ok {
		return
	}
	r = (*Range)(unsafe.Pointer(&rt.ranges[i][0]))
	return r, true
}

func (rt *RangeTable) indexByName(name string) (i int, ok bool) {
	if rt.Empty() {
		return -1, false
	}
	i, ok = rt.ap[name]
	return
}

func (rt *RangeTable) String() string {
	result := fmt.Sprintf("RangeTable %p\n", rt)
	for _, name := range rt.names {
		r, ok := rt.unsafeRangeByName(name)
		if !ok {
			panic(fmt.Sprintf("no such name %s", name))
		}
		result += fmt.Sprintf("%s: %s\n", name, r.String())
	}
	// Result result minux the last newline character
	return result[:len(result)-1]
}
