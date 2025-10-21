package allocator

import (
	"sync"
)

const DEFAULT_SIZE uint32 = 65536
const START_UID_GID uint32 = 1000000

type Range struct {
	UidStart uint32
	UidSize  uint32
	GidStart uint32
	GidSize  uint32
}

type Allocator struct {
	used map[int]bool
	mex  int

	mu sync.Mutex
}

func NewAllocator() *Allocator {
	return &Allocator{
		used: make(map[int]bool),
		mex:  0,
	}
}

func (a *Allocator) Allocate() (int, Range) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.used[a.mex] = true
	r := Range{
		UidStart: START_UID_GID + uint32(a.mex)*DEFAULT_SIZE,
		UidSize:  DEFAULT_SIZE,
		GidStart: START_UID_GID + uint32(a.mex)*DEFAULT_SIZE,
		GidSize:  DEFAULT_SIZE,
	}
	use := a.mex

	for a.used[a.mex] {
		a.mex++
	}

	return use, r
}

func (a *Allocator) Free(i int) int {
	a.mu.Lock()
	defer a.mu.Unlock()

	if i < 0 || !a.used[i] {
		return -1
	}

	delete(a.used, i)
	if i < a.mex {
		a.mex = i
	}

	return 0
}
