package grader_test

import (
	"testing"

	"github.com/joshjms/castletown/internal/cache"
	"github.com/joshjms/castletown/internal/config"
	"github.com/joshjms/castletown/internal/grader"
	"github.com/joshjms/castletown/internal/models"
	"github.com/joshjms/castletown/internal/repository"
	"github.com/joshjms/castletown/internal/sandbox"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestGrader_HandleSubmission(t *testing.T) {
	sandbox.Init()

	cfg := config.Load()
	cfg.ProblemCacheDir = t.TempDir()
	sm := sandbox.NewManager(&cfg)
	repo, err := repository.NewPostgresRepository(cfg.Database.DSN)
	if err != nil {
		t.Fatal(err)
	}

	log := zerolog.New(nil).With().Timestamp().Logger()

	store := cache.NewProblemStore(repo, 16, cfg.ProblemCacheDir)
	g := grader.NewGrader(log, cfg, sm, repo, store)
	sub := &models.Submission{
		ID:        1,
		ProblemID: 1,
		Language:  "cpp",
		Code: `
#include <bits/stdc++.h>
using namespace std;

int main() {
	int n, q; cin >> n >> q;
	vector <int> a(n + 5, 0);

	for(int i = 1; i <= n; i++)
		cin >> a[i];
	
	vector <int> l(q + 5, 0), r(q + 5, 0);
	for(int t = 1; t <= q; t++) {
		cin >> l[t] >> r[t];
	}

	for(int t = 1; t <= q; t++) {
		set <int> st;
		for(int i = l[t]; i <= r[t]; i++) {
			st.insert(a[i]);
		}
		int sz = st.size();
		cout << sz << "\n";
	}
}
		`,
	}

	assert.NoError(t, g.Handle(t.Context(), sub))
	assert.Equal(t, models.VerdictAccepted, sub.Verdict)
}
