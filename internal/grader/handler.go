package grader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joshjms/castletown/internal/checker"
	"github.com/joshjms/castletown/internal/models"
	"github.com/joshjms/castletown/internal/sandbox"
)

func (g *Grader) handleCpp(ctx context.Context, sub *models.Submission, problem *models.Problem, testcases []models.Testcase, subDir string) error {
	g.log.Info().Int64("submission_id", sub.ID).Msg("Handling C++ submission")

	dir := filepath.Join(subDir, "compile")
	if err := os.Mkdir(dir, 0700); err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "main.cpp"), []byte(sub.Code), 0600); err != nil {
		return err
	}

	compileSandboxId := fmt.Sprintf("compile_%d", sub.ID)
	cfg := sandbox.GetDefaultConfig()
	cfg.Args = []string{"g++", "main.cpp", "-o", "main", "-O2", "-std=c++20"}
	cfg.RootfsImageDir = filepath.Join(g.cfg.ImagesDir, imagesMap["cpp"]["20"])
	cfg.BoxDir = dir
	cfg.TimeLimitMs = 3000
	cfg.Cgroup.Memory = 512 * 1024 * 1024

	err := g.sm.NewSandbox(compileSandboxId, cfg)
	if err != nil {
		return fmt.Errorf("cannot create compile sandbox: %v", err)
	}

	compileReport, err := g.sm.RunSandbox(ctx, compileSandboxId)
	if err != nil {
		return fmt.Errorf("compilation failed: %v", err)
	}

	g.log.Info().Any("compile_report", compileReport).Msg("Compilation report")

	if compileReport.ExitCode != 0 {
		return g.submissionCompileError(sub, compileReport)
	}

	checker := checker.NewChecker(checker.WithTokenComparison())

	maxTimeUsec := uint64(0)
	maxMemoryBytes := uint64(0)

	g.sm.DestroySandbox(compileSandboxId)

	for i, tc := range testcases {
		execSandboxId := fmt.Sprintf("exec_%d_%d", sub.ID, tc.ID)
		cfg := sandbox.GetDefaultConfig()
		cfg.Args = []string{"./main"}
		cfg.RootfsImageDir = filepath.Join(g.cfg.ImagesDir, imagesMap["cpp"]["20"])
		cfg.BoxDir = dir
		cfg.TimeLimitMs = problem.TimeLimit
		cfg.Cgroup.Memory = problem.MemoryLimit * 1024 * 1024
		cfg.Cgroup.PidsLimit = 1
		cfg.Stdin = tc.Input

		err := g.sm.NewSandbox(execSandboxId, cfg)
		if err != nil {
			g.sm.DestroySandbox(execSandboxId)
			return fmt.Errorf("cannot create exec sandbox: %v", err)
		}

		execReport, err := g.sm.RunSandbox(ctx, execSandboxId)
		if err != nil {
			g.sm.DestroySandbox(execSandboxId)
			return fmt.Errorf("execution failed: %v", err)
		}

		g.log.Info().Any("exec_report", execReport).Msgf("Execution report for testcase %d", i)

		if execReport.ExitCode != 0 {
			g.sm.DestroySandbox(execSandboxId)
			return g.submissionRuntimeError(sub, execReport, fmt.Sprintf("Runtime error on testcase %d", i))
		}

		correct, err := checker.Check(execReport.Stdout, tc.Output)
		if err != nil {
			g.sm.DestroySandbox(execSandboxId)
			return fmt.Errorf("error during checking: %v", err)
		}
		if !correct {
			g.sm.DestroySandbox(execSandboxId)
			return g.submissionWrongAnswer(sub, execReport, fmt.Sprintf("Wrong answer on testcase %d", i))
		}

		maxTimeUsec = max(maxTimeUsec, execReport.CPUTime)
		maxMemoryBytes = max(maxMemoryBytes, execReport.Memory)

		g.sm.DestroySandbox(execSandboxId)
	}

	return g.submissionAccepted(sub, maxTimeUsec, maxMemoryBytes)
}

func (g *Grader) submissionCompileError(sub *models.Submission, report sandbox.Report) error {
	g.log.Info().Int64("submission_id", sub.ID).Msg("Compilation error")
	return nil
}

func (g *Grader) submissionRuntimeError(sub *models.Submission, report sandbox.Report, msg string) error {
	g.log.Info().Int64("submission_id", sub.ID).Msg(msg)
	return nil
}

func (g *Grader) submissionWrongAnswer(sub *models.Submission, report sandbox.Report, msg string) error {
	g.log.Info().Int64("submission_id", sub.ID).Msg(msg)
	sub.Verdict = models.VerdictWrongAnswer
	sub.ExecutionTime = int64(report.CPUTime)
	sub.MemoryUsed = int64(report.Memory)
	return nil
}

func (g *Grader) submissionAccepted(sub *models.Submission, maxTimeUsec, maxMemoryBytes uint64) error {
	g.log.Info().Int64("submission_id", sub.ID).Msg("Submission accepted")
	sub.Verdict = models.VerdictAccepted
	sub.ExecutionTime = int64(maxTimeUsec)
	sub.MemoryUsed = int64(maxMemoryBytes)
	return nil
}
