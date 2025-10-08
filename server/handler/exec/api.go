package exec

import (
	"github.com/joshjms/castletown/job"
	"github.com/joshjms/castletown/sandbox"
)

type Request struct {
	ID    string        `json:"id"`
	Files []job.File    `json:"files"`
	Procs []job.Process `json:"steps"`
}

type Response struct {
	ID      string           `json:"id"`
	Reports []sandbox.Report `json:"reports"`
}
