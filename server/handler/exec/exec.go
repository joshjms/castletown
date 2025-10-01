package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/joshjms/castletown/sandbox"
)

type ExecHandler struct {
	m *sandbox.Manager
}

func NewExecHandler(m *sandbox.Manager) *ExecHandler {
	return &ExecHandler{m: m}
}

func (h *ExecHandler) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid json: %v", err), http.StatusBadRequest)
		return
	}

	req.ID = uuid.NewString()

	reports, err := h.handleRequest(r.Context(), req)
	if err != nil {
		http.Error(w, fmt.Sprintf("error running processes: %v", err), http.StatusInternalServerError)
		return
	}

	reportsJson, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot marshal reports: %v", err), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(reportsJson)
}

func (h *ExecHandler) handleRequest(ctx context.Context, req Request) (Response, error) {
	if len(req.Procs) == 0 {
		return nil, fmt.Errorf("no processes specified")
	}

	if err := verifyImages(req.Procs); err != nil {
		return nil, fmt.Errorf("invalid images: %w", err)
	}

	prepareFileDirs(req.ID, req.Procs)
	fileDeps, err := getFileDependencies(req.ID, req.Procs, req.Files)
	if err != nil {
		return nil, fmt.Errorf("error getting file dependencies: %w", err)
	}

	var reports Response

	for i, proc := range req.Procs {
		cfg := getDefaultConfig()
		cfg.Args = proc.Cmd
		cfg.RootfsImageDir = getImageDir(proc.Image)
		cfg.BoxDir = getProcFileDir(req.ID, i)
		cfg.Files = fileDeps[i]

		if proc.TimeLimitMs > 0 {
			cfg.TimeLimitMs = int64(proc.TimeLimitMs)
		}
		if proc.MemoryLimitMB > 0 {
			cfg.Cgroup.Memory = int64(proc.MemoryLimitMB) * 1024 * 1024
		}
		cfg.Stdin = proc.Stdin

		containerId := fmt.Sprintf("%s-%d", req.ID, i)
		s, err := h.m.NewSandbox(containerId, cfg)
		if err != nil {
			return nil, fmt.Errorf("cannot create sandbox for process %d: %v", i, err)
		}
		defer h.m.DestroySandbox(containerId)

		report, err := s.Run(ctx)
		if err != nil {
			return nil, fmt.Errorf("error running process %d: %v", i, err)
		}

		reports = append(reports, *report)
	}

	return reports, nil
}

func getDefaultConfig() *sandbox.Config {
	return &sandbox.Config{
		Env: []string{
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
		Cwd:         "/box",
		TimeLimitMs: 1000,
		Cgroup: &sandbox.CgroupConfig{
			CpuShares:  100000,
			CpuQuota:   100000,
			Memory:     256 * 1024 * 1024,
			PidsLimit:  100,
			CpusetCpus: "0",
			CpusetMems: "0",
		},
		Rlimit: &sandbox.RlimitConfig{
			Core: &sandbox.Rlimit{
				Hard: 0,
				Soft: 0,
			},
			Fsize: &sandbox.Rlimit{
				Hard: 1 * 1024 * 1024,
				Soft: 1 * 1024 * 1024,
			},
			NoFile: &sandbox.Rlimit{
				Hard: 64,
				Soft: 64,
			},
		},
	}
}
