package cache

import (
	"container/list"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/joshjms/castletown/internal/models"
	"github.com/joshjms/castletown/internal/repository"
)

// ProblemStore caches problem metadata and its associated testcases using an LRU eviction policy.
type ProblemStore struct {
	repo repository.Repository

	mu       sync.Mutex
	capacity int
	ll       *list.List
	entries  map[int64]*list.Element
	diskDir  string
}

type problemEntry struct {
	id        int64
	problem   *models.Problem
	testcases []models.Testcase
}

func NewProblemStore(repo repository.Repository, capacity int, diskDir string) *ProblemStore {
	if capacity <= 0 {
		capacity = 128
	}
	if diskDir == "" {
		diskDir = "/var/castletown/problems"
	}
	return &ProblemStore{
		repo:     repo,
		capacity: capacity,
		ll:       list.New(),
		entries:  make(map[int64]*list.Element),
		diskDir:  diskDir,
	}
}

// GetProblem returns a problem and its testcases either from cache or by querying the repository.
func (ps *ProblemStore) GetProblem(ctx context.Context, id int64) (*models.Problem, []models.Testcase, error) {
	if problem, testcases, ok := ps.getFromCache(id); ok {
		if err := ps.ensureTestcasesOnDisk(id, testcases); err != nil {
			return nil, nil, err
		}
		return cloneProblem(problem), cloneTestcases(testcases), nil
	}

	problem, err := ps.repo.GetProblemDetails(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	testcases, err := ps.repo.GetTestcases(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	if err := ps.ensureTestcasesOnDisk(id, testcases); err != nil {
		return nil, nil, err
	}

	ps.addToCache(id, problem, testcases)

	return cloneProblem(problem), cloneTestcases(testcases), nil
}

func (ps *ProblemStore) getFromCache(id int64) (*models.Problem, []models.Testcase, bool) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	elem, ok := ps.entries[id]
	if !ok {
		return nil, nil, false
	}

	ps.ll.MoveToFront(elem)
	entry := elem.Value.(*problemEntry)
	return entry.problem, entry.testcases, true
}

func (ps *ProblemStore) addToCache(id int64, problem *models.Problem, testcases []models.Testcase) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if elem, ok := ps.entries[id]; ok {
		ps.ll.MoveToFront(elem)
		entry := elem.Value.(*problemEntry)
		entry.problem = cloneProblem(problem)
		entry.testcases = cloneTestcases(testcases)
		return
	}

	if ps.ll.Len() >= ps.capacity {
		ps.evictOldest()
	}

	entry := &problemEntry{
		id:        id,
		problem:   cloneProblem(problem),
		testcases: cloneTestcases(testcases),
	}
	elem := ps.ll.PushFront(entry)
	ps.entries[id] = elem
}

func (ps *ProblemStore) evictOldest() {
	elem := ps.ll.Back()
	if elem == nil {
		return
	}

	ps.ll.Remove(elem)
	entry := elem.Value.(*problemEntry)
	delete(ps.entries, entry.id)
}

func (ps *ProblemStore) ensureTestcasesOnDisk(problemID int64, testcases []models.Testcase) error {
	baseDir := filepath.Join(ps.diskDir, fmt.Sprintf("%d", problemID), "testcase")

	for _, tc := range testcases {
		tcDir := filepath.Join(baseDir, fmt.Sprintf("%d", tc.ID))
		if err := os.MkdirAll(tcDir, 0755); err != nil {
			return fmt.Errorf("create testcase dir: %w", err)
		}

		if err := ps.writeContentIfAny(filepath.Join(tcDir, "input.txt"), tc.Input); err != nil {
			return err
		}
		if err := ps.writeContentIfAny(filepath.Join(tcDir, "output.txt"), tc.Output); err != nil {
			return err
		}
	}
	return nil
}

func (ps *ProblemStore) writeContentIfAny(path string, content string) error {
	if content == "" {
		return nil
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write testcase file %s: %w", path, err)
	}
	return nil
}

func cloneProblem(p *models.Problem) *models.Problem {
	if p == nil {
		return nil
	}
	cp := *p
	if len(p.Testcases) > 0 {
		cp.Testcases = cloneTestcases(p.Testcases)
	}
	return &cp
}

func cloneTestcases(src []models.Testcase) []models.Testcase {
	if len(src) == 0 {
		return nil
	}
	dst := make([]models.Testcase, len(src))
	copy(dst, src)
	return dst
}
