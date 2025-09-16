package sandbox

import (
	"fmt"

	"github.com/joshjms/castletown/sandbox/allocator"
)

type Manager struct {
	Sandboxes map[string]*Sandbox

	Allocator *allocator.Allocator
}

func NewManager() (*Manager, error) {
	alloc, err := allocator.NewAllocator()
	if err != nil {
		return nil, err
	}

	return &Manager{
		Sandboxes: make(map[string]*Sandbox),
		Allocator: alloc,
	}, nil
}

func (m *Manager) NewSandbox(id string, cfg *Config, overlayfsDir string) (*Sandbox, error) {
	if _, exists := m.Sandboxes[id]; exists {
		return nil, fmt.Errorf("sandbox with id %q already exists", id)
	}

	sandbox := &Sandbox{
		id:           id,
		config:       cfg,
		overlayfsDir: overlayfsDir,
	}

	m.Sandboxes[id] = sandbox

	return sandbox, nil
}
