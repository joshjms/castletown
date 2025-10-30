package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/joshjms/castletown/job"
	"github.com/joshjms/castletown/sandbox"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid json: %v", err), http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		req.ID = uuid.NewString()
	}

	reports, err := handleRequest(r.Context(), req)
	if err != nil {
		http.Error(w, fmt.Sprintf("error running processes: %v", err), http.StatusInternalServerError)
		return
	}

	response := Response{
		ID:      req.ID,
		Reports: reports,
	}

	responseJson, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot marshal reports: %v", err), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJson)
}

func handleRequest(ctx context.Context, req Request) ([]sandbox.Report, error) {
	j := job.Job{
		ID:    req.ID,
		Files: req.Files,
		Procs: req.Procs,
	}

	jp := job.GetJobPool()
	_job := jp.AddOrAppendJob(&j)

	if err := _job.Prepare(); err != nil {
		return nil, fmt.Errorf("error preparing job: %w", err)
	}

	reports, err := _job.ExecuteAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing job: %w", err)
	}

	return reports, nil
}
