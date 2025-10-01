package allocator

import (
	"sync"

	"github.com/joshjms/castletown/config"
)

const DEFAULT_SIZE uint32 = 65536

type Range struct {
	UidStart uint32
	UidSize  uint32
	GidStart uint32
	GidSize  uint32
}

type Allocator struct {
	ranges []Range
	used   []bool

	mu sync.Mutex
}

func NewAllocator() (*Allocator, error) {
	startUid := uint32(100000)
	startGid := uint32(100000)
	maxContainers := config.MaxConcurrency

	ranges := make([]Range, maxContainers)
	used := make([]bool, maxContainers)
	for i := range maxContainers {
		ranges[i] = Range{
			UidStart: startUid + uint32(i)*DEFAULT_SIZE,
			UidSize:  DEFAULT_SIZE,
			GidStart: startGid + uint32(i)*DEFAULT_SIZE,
			GidSize:  DEFAULT_SIZE,
		}
		used[i] = false
	}

	return &Allocator{
		ranges: ranges,
		used:   used,
	}, nil
}

func (a *Allocator) Allocate() (int, Range) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, used := range a.used {
		if !used {
			a.used[i] = true
			return i, a.ranges[i]
		}
	}

	return -1, Range{}
}

func (a *Allocator) Free(i int) int {
	a.mu.Lock()
	defer a.mu.Unlock()

	if i < 0 || i >= len(a.used) {
		return -1
	}

	if !a.used[i] {
		return -1
	}

	a.used[i] = false
	return i
}
