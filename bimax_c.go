package main

import (
	"reflect"
	"unsafe"

	"github.com/maxsei/bimax/bimax"

	"C"
)

type BiMaxResult struct {
	bimax.BiMaxResult
}

// ToC converts the BiMaxResult to allocted data in C so for ease of use in
// returning results functions 'BiMaxBinaryMatrixC' and 'BiMaxBinaryVerticesC', that
// deal with returning to the C api.
func (r *BiMaxResult) ToC() (C.size_t, *C.longlong, C.size_t, *C.longlong) {
	// getSetValues64 returns the values of a set as []int64.
	getSetValues64 := func(set *SetOp) []int64 {
		result := make([]int64, 0, set.Card())
		set.Each(func(v int) (done bool) {
			result = append(result, int64(v))
			return
		})
		return result
	}
	// Get the values of row and column sets in the result as []int64.
	rows := getSetValues64(r.Rows)
	cols := getSetValues64(r.Cols)
	// toArrC convert slice of int64to a pointer to allocated C.longlong memory
	// and the size of length of the allocated memory measured in C.longlong.
	toArrC := func(x []int64) (C.size_t, *C.longlong) {
		bb := *(*[]byte)(unsafe.Pointer(&x))
		bbH := (*reflect.SliceHeader)(unsafe.Pointer(&bb))
		bbH.Len = len(x) * int(unsafe.Sizeof(x[0]))
		return C.size_t(len(x)), (*C.longlong)(C.CBytes(bb))
	}
	// Get the values of rows and cols as C memory pointers to C.longlong data and
	// lengths of said pointers to data
	lenRowsC, dataRowsC := toArrC(rows)
	lenColsC, dataColsC := toArrC(cols)

	return lenRowsC, dataRowsC, lenColsC, dataColsC
}

//export BiMaxBinaryMatrixC
func BiMaxBinaryMatrixC(nC, mC C.longlong, dataC *C.char) (C.size_t, *C.longlong, C.size_t, *C.longlong) {
	// Convert C input data into Go data
	n := int(nC)
	m := int(mC)

	var data []uint8
	dataH := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	dataH.Data = uintptr(unsafe.Pointer(dataC))
	dataH.Len = n * m

	result := BiMaxBinaryMatrix(n, m, data)
	return result.ToC()
}

//export BiMaxVerticesC
func BiMaxVerticesC(uuLenC C.size_t, uuC *C.longlong, vvLenC C.size_t, vvC *C.longlong) (C.size_t, *C.longlong, C.size_t, *C.longlong) {
	var uu, vv []int
	pointSliceToCData := func(length C.size_t, data *C.longlong, sl *[]int) {
		header := (*reflect.SliceHeader)(unsafe.Pointer(sl))
		header.Data = uintptr(unsafe.Pointer(data))
		header.Len = int(length)
	}
	pointSliceToCData(uuLenC, uuC, &uu)
	pointSliceToCData(vvLenC, vvC, &vv)
	result := BiMaxVertices(uu, vv)
	return result.ToC()
}
