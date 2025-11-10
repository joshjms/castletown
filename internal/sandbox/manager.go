package sandbox

import (
	"context"
	"fmt"
	"sync"

	"github.com/joshjms/castletown/internal/config"
	"github.com/joshjms/castletown/internal/sandbox/allocator"
)

var m *Manager

type Manager struct {
	cfg *config.Config

	sandboxes map[string]*Sandbox
	allocated map[string]allocator.AllocResult

	allocator      *allocator.Allocator
	maxConcurrency int

	mu  sync.Mutex
	sem chan struct{}
}

func NewManager(cfg *config.Config) *Manager {
	alloc := allocator.NewAllocator()
	maxConc := cfg.MaxParallelSandboxes
	if maxConc <= 0 {
		maxConc = 1
	}

	m = &Manager{
		cfg:            cfg,
		sandboxes:      make(map[string]*Sandbox),
		allocated:      make(map[string]allocator.AllocResult),
		allocator:      alloc,
		maxConcurrency: maxConc,
		sem:            make(chan struct{}, maxConc),
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
		id:         id,
		config:     cfg,
		runtimeCfg: m.cfg,
	}

	m.sandboxes[id] = sandbox
	m.allocated[id] = res

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

	if err := removeCgroup(id); err != nil {
		return fmt.Errorf("failed to remove cgroup for sandbox %q: %w", id, err)
	}

	idx, ok := m.allocated[id]
	if !ok {
		return fmt.Errorf("no allocated range found for sandbox id %q", id)
	}

	if freedIdx := m.allocator.Free(idx); freedIdx == -1 {
		return fmt.Errorf("failed to free allocated range for sandbox id %q", id)
	}

	delete(m.sandboxes, id)
	delete(m.allocated, id)

	return nil
}
