package models

import "time"

type Verdict string

const (
	VerdictPending             Verdict = "PENDING"
	VerdictJudging             Verdict = "JUDGING"
	VerdictAccepted            Verdict = "AC"
	VerdictWrongAnswer         Verdict = "WA"
	VerdictTimeLimitExceeded   Verdict = "TLE"
	VerdictMemoryLimitExceeded Verdict = "MLE"
	VerdictRuntimeError        Verdict = "RE"
	VerdictCompilationError    Verdict = "CE"
	VerdictSystemError         Verdict = "SE"
	VerdictInternalError       Verdict = "IE"
)

type Problem struct {
	ID          int64      `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Difficulty  int        `json:"difficulty" db:"difficulty"`
	TimeLimit   int64      `json:"time_limit" db:"time_limit"`
	MemoryLimit int64      `json:"memory_limit" db:"memory_limit"`
	Testcases   []Testcase `json:"test_cases" db:"test_cases"`
	Tags        []string   `json:"tags" db:"tags"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type Submission struct {
	ID            int64     `json:"id" db:"id"`
	ProblemID     int64     `json:"problem_id" db:"problem_id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	Code          string    `json:"code" db:"code"`
	Language      string    `json:"language" db:"language"`
	Verdict       Verdict   `json:"verdict" db:"verdict"`
	Score         float64   `json:"score" db:"score"`
	ExecutionTime int64     `json:"execution_time" db:"execution_time"`
	MemoryUsed    int64     `json:"memory_used" db:"memory_used"`
	Message       string    `json:"message" db:"message"`
	TestsPassed   int       `json:"tests_passed" db:"tests_passed"`
	TestsTotal    int       `json:"tests_total" db:"tests_total"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Testcase represents a single test case for a problem
type Testcase struct {
	ID         int64  `json:"id" db:"id"`
	ProblemID  int64  `json:"problem_id" db:"problem_id"`
	Input      string `json:"input" db:"input"`
	InputFile  string `json:"input_file" db:"input_file"`
	Output     string `json:"output" db:"output"`
	OutputFile string `json:"output_file" db:"output_file"`
	IsHidden   bool   `json:"is_hidden" db:"is_hidden"`
	Points     int    `json:"points" db:"points"`
}

// TestcaseResult represents the result of running a single test case
type TestcaseResult struct {
	TestcaseID     int64   `json:"testcase_id"`
	Verdict        Verdict `json:"verdict"`
	ExecutionTime  int64   `json:"execution_time"`
	MemoryUsed     int64   `json:"memory_used"`
	Input          string  `json:"input,omitempty"`
	ExpectedOutput string  `json:"expected_output,omitempty"`
	ActualOutput   string  `json:"actual_output,omitempty"`
	ErrorMessage   string  `json:"error_message,omitempty"`
}

// ExecutionResult represents the result of code execution
type ExecutionResult struct {
	Verdict       Verdict          `json:"verdict"`
	TestResults   []TestcaseResult `json:"test_results"`
	Score         float64          `json:"score"`
	ExecutionTime int64            `json:"execution_time"`
	MemoryUsed    int64            `json:"memory_used"`
	Message       string           `json:"message,omitempty"`
}

// User represents a user in the system
type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Language represents a supported programming language
type Language struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Extension        string  `json:"extension"`
	CompileCommand   string  `json:"compile_command"`
	ExecuteCommand   string  `json:"execute_command"`
	Version          string  `json:"version"`
	TimeMultiplier   float64 `json:"time_multiplier"`
	MemoryMultiplier float64 `json:"memory_multiplier"`
}
