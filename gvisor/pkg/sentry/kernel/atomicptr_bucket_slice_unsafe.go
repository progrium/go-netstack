package kernel

import (
	"sync/atomic"
	"unsafe"
)

// An AtomicPtr is a pointer to a value of type Value that can be atomically
// loaded and stored. The zero value of an AtomicPtr represents nil.
//
// Note that copying AtomicPtr by value performs a non-atomic read of the
// stored pointer, which is unsafe if Store() can be called concurrently; in
// this case, do `dst.Store(src.Load())` instead.
//
// +stateify savable
type descriptorBucketSliceAtomicPtr struct {
	ptr unsafe.Pointer `state:".(*descriptorBucketSlice)"`
}

func (p *descriptorBucketSliceAtomicPtr) savePtr() *descriptorBucketSlice {
	return p.Load()
}

func (p *descriptorBucketSliceAtomicPtr) loadPtr(v *descriptorBucketSlice) {
	p.Store(v)
}

// Load returns the value set by the most recent Store. It returns nil if there
// has been no previous call to Store.
//
//go:nosplit
func (p *descriptorBucketSliceAtomicPtr) Load() *descriptorBucketSlice {
	return (*descriptorBucketSlice)(atomic.LoadPointer(&p.ptr))
}

// Store sets the value returned by Load to x.
func (p *descriptorBucketSliceAtomicPtr) Store(x *descriptorBucketSlice) {
	atomic.StorePointer(&p.ptr, (unsafe.Pointer)(x))
}

// Swap atomically stores `x` into *p and returns the previous *p value.
func (p *descriptorBucketSliceAtomicPtr) Swap(x *descriptorBucketSlice) *descriptorBucketSlice {
	return (*descriptorBucketSlice)(atomic.SwapPointer(&p.ptr, (unsafe.Pointer)(x)))
}
