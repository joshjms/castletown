package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/joshjms/castletown/sandbox"
)

type Request struct {
	Steps []Process
	Files []File
}

type Process struct {
	Image         string   `json:"image"`
	Cmd           []string `json:"cmd"`
	Stdin         string   `json:"stdin"`
	MemoryLimitMB int64    `json:"memoryLimitMB"` // Memory limit in megabytes
	TimeLimitMs   uint64   `json:"timeLimitMs"`   // Time limit in milliseconds
	ProcLimit     int64    `json:"procLimit"`

	Files   []string `json:"files"`
	Persist []string `json:"persist"`
}

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Response []*sandbox.Report

type ExecHandler struct {
	OverlayfsDir    string
	FilesDir        string
	ImagesDir       string
	LibcontainerDir string

	m *sandbox.Manager
}

func NewExecHandler(overlayfsDir, filesDir, imagesDir, libcontainerDir string, manager *sandbox.Manager) *ExecHandler {
	return &ExecHandler{
		OverlayfsDir:    overlayfsDir,
		FilesDir:        filesDir,
		ImagesDir:       imagesDir,
		LibcontainerDir: libcontainerDir,
		m:               manager,
	}
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

	id := uuid.NewString()
	filesDir := filepath.Join(h.FilesDir, id)

	if err := os.MkdirAll(filesDir, 0755); err != nil {
		http.Error(w, fmt.Sprintf("cannot create files directory: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(filesDir)

	if err := createFiles(filesDir, req.Files); err != nil {
		http.Error(w, fmt.Sprintf("cannot create files: %v", err), http.StatusInternalServerError)
		return
	}

	reports := make(Response, len(req.Steps))

	for i, process := range req.Steps {
		report, err := handleProcess(r.Context(), h.m, process, h.OverlayfsDir, filesDir, h.ImagesDir)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot handle process: %v", err), http.StatusInternalServerError)
			return
		}
		reports[i] = report
	}

	reportsJson, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot marshal reports: %v", err), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(reportsJson)
}

func handleProcess(ctx context.Context, manager *sandbox.Manager, process Process, overlayfsDir, filesDir, imageDir string) (*sandbox.Report, error) {
	rootfsImageDir, err := getRootfsDir(imageDir, process.Image)
	if err != nil {
		return nil, fmt.Errorf("error getting rootfs directory: %w", err)
	}

	copyFiles := getCopyFiles(filesDir, process.Files)
	saveFiles := getSaveFiles(filesDir, process.Persist)

	sandboxConfig := &sandbox.Config{
		RootfsImageDir: rootfsImageDir,
		Args:           process.Cmd,
		Stdin:          process.Stdin,
		Cwd:            "/box",
		Env:            []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		ContainerUID:   65534,
		ContainerGID:   65534,
		UserNamespace: &sandbox.UserNamespaceConfig{
			RootUID:     uint32(os.Getuid()),
			UIDMapStart: 100000,
			UIDMapCount: 65534,
			RootGID:     uint32(os.Getgid()),
			GIDMapStart: 100000,
			GIDMapCount: 65534,
		},
		TimeLimitMs: int64(process.TimeLimitMs),
		Cgroup: &sandbox.CgroupConfig{
			CpuQuota: 100000,
			Memory:   process.MemoryLimitMB * 1024 * 1024,
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
		Copy: copyFiles,
		Save: saveFiles,
	}

	sandbox, err := manager.NewSandbox(
		uuid.NewString(),
		sandboxConfig,
		overlayfsDir,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating sandbox: %w", err)
	}

	return sandbox.Run(ctx)
}
