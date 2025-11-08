package sandbox

import (
	"context"
	"fmt"
	"sync"

	"github.com/joshjms/castletown/sandbox/allocator"
)

var m *Manager

type Manager struct {
	sandboxes       map[string]*Sandbox
	allocatedRanges map[string]allocator.AllocResult

	allocator      *allocator.Allocator
	maxConcurrency int

	mu  sync.Mutex
	sem chan struct{}
}

func NewManager(maxConcurrency int) *Manager {
	alloc := allocator.NewAllocator()

	m = &Manager{
		sandboxes:       make(map[string]*Sandbox),
		allocatedRanges: make(map[string]allocator.AllocResult),
		allocator:       alloc,
		maxConcurrency:  maxConcurrency,
		sem:             make(chan struct{}, maxConcurrency),
	}

	return m
}

func GetManager() *Manager {
	return m
}

func (m *Manager) NewSandbox(id string, cfg *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sandboxes[id]; exists {
		return fmt.Errorf("sandbox with id %q already exists", id)
	}

	res := m.allocator.Allocate()

	cfg.UserNamespace = &UserNamespaceConfig{
		HostUID:      uint32(res.ID.UIDStart),
		ContainerUID: 0,
		UIDMapCount:  uint32(res.ID.UIDSize),
		HostGID:      uint32(res.ID.GIDStart),
		ContainerGID: 0,
		GIDMapCount:  uint32(res.ID.GIDSize),
	}

	cfg.Cgroup.CpusetCpus = fmt.Sprintf("%d", res.CPU)
	cfg.Cgroup.CpusetMems = "0"

	sandbox := &Sandbox{
		id:     id,
		config: cfg,
	}

	m.sandboxes[id] = sandbox
	m.allocatedRanges[id] = res

	return nil
}

func (m *Manager) RunSandbox(ctx context.Context, id string) (Report, error) {
	m.sem <- struct{}{}
	defer func() { <-m.sem }()

	m.mu.Lock()
	sandbox, exists := m.sandboxes[id]
	m.mu.Unlock()

	if !exists {
		return Report{}, fmt.Errorf("sandbox with id %q does not exist", id)
	}

	report, err := sandbox.Run(ctx)
	if err != nil {
		return Report{}, fmt.Errorf("error running sandbox %q: %w", id, err)
	}

	return report, nil
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
