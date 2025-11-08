package allocator

import (
	"sync"
)

const DEFAULT_SIZE uint32 = 65536
const START_UID_GID uint32 = 1000000

type Range struct {
	UIDStart uint32
	UIDSize  uint32
	GIDStart uint32
	GIDSize  uint32
}

type AllocResult struct {
	index int
	CPU   int
	ID    Range
}

type Allocator struct {
	used map[int]bool
	free int

	mu sync.Mutex
}

func NewAllocator() *Allocator {
	return &Allocator{
		used: make(map[int]bool),
	}
}

func (a *Allocator) Allocate() AllocResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	r := Range{
		UIDStart: START_UID_GID + uint32(a.free)*DEFAULT_SIZE,
		UIDSize:  DEFAULT_SIZE,
		GIDStart: START_UID_GID + uint32(a.free)*DEFAULT_SIZE,
		GIDSize:  DEFAULT_SIZE,
	}

	res := AllocResult{
		index: a.free,
		CPU:   a.free,
		ID:    r,
	}

	use := a.free

	for a.used[use] {
		use++
	}

	return res
}

func (a *Allocator) Free(res AllocResult) int {
	a.mu.Lock()
	defer a.mu.Unlock()

	if res.index < 0 || !a.used[res.index] {
		return -1
	}

	delete(a.used, res.index)
	if res.index < a.free {
		a.free = res.index
	}

	return 0
}
