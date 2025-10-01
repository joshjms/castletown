package sandbox

import (
	"fmt"
	"sync"

	"github.com/joshjms/castletown/sandbox/allocator"
)

type Manager struct {
	sandboxes       map[string]*Sandbox
	allocatedRanges map[string]int

	allocator *allocator.Allocator

	mu sync.Mutex
}

func NewManager() (*Manager, error) {
	alloc, err := allocator.NewAllocator()
	if err != nil {
		return nil, err
	}

	return &Manager{
		sandboxes:       make(map[string]*Sandbox),
		allocatedRanges: make(map[string]int),
		allocator:       alloc,
	}, nil
}

func (m *Manager) NewSandbox(id string, cfg *Config) (*Sandbox, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sandboxes[id]; exists {
		return nil, fmt.Errorf("sandbox with id %q already exists", id)
	}

	idx, rng := m.allocator.Allocate()
	if idx == -1 {
		return nil, fmt.Errorf("no available uid/gid ranges")
	}

	cfg.UserNamespace = &UserNamespaceConfig{
		HostUID:      uint32(rng.UidStart),
		ContainerUID: 0,
		UIDMapCount:  uint32(rng.UidSize),
		HostGID:      uint32(rng.GidStart),
		ContainerGID: 0,
		GIDMapCount:  uint32(rng.GidSize),
	}

	sandbox := &Sandbox{
		id:     id,
		config: cfg,
	}

	m.sandboxes[id] = sandbox
	m.allocatedRanges[id] = idx

	return sandbox, nil
}

func (m *Manager) DestroySandbox(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sandbox, exists := m.sandboxes[id]
	if !exists {
		return fmt.Errorf("sandbox with id %q does not exist", id)
	}

	if err := sandbox.Destroy(); err != nil {
		return fmt.Errorf("error destroying sandbox: %w", err)
	}

	idx, ok := m.allocatedRanges[id]
	if !ok {
		return fmt.Errorf("no allocated range found for sandbox id %q", id)
	}

	if freedIdx := m.allocator.Free(idx); freedIdx == -1 {
		return fmt.Errorf("failed to free allocated range for sandbox id %q", id)
	}

	delete(m.sandboxes, id)
	delete(m.allocatedRanges, id)
	return nil
}
