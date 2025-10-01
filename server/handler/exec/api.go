package exec

import "github.com/joshjms/castletown/sandbox"

type Request struct {
	ID    string    `json:"id"`
	Files []File    `json:"files"`
	Procs []Process `json:"steps"`
}

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Process struct {
	Image         string   `json:"image"`
	Cmd           []string `json:"cmd"`
	Stdin         string   `json:"stdin"`
	MemoryLimitMB int64    `json:"memoryLimitMB"`
	TimeLimitMs   uint64   `json:"timeLimitMs"`
	ProcLimit     int64    `json:"procLimit"`
	Files         []string `json:"files"`
	Persist       []string `json:"persist"`
}

type Response []sandbox.Report
