package trainingusecase

import (
	"atcoder_shojin/backend/internal/domain/training"
	"context"
	"testing"
	"time"
)

type fakeClient struct {
	catalogProblems []training.Problem
	submissions     []training.Submission
	err             error
}

func (client *fakeClient) FetchCatalog(context.Context) ([]training.Contest, []training.Problem, error) {
	return []training.Contest{{ID: "abc001"}}, client.catalogProblems, client.err
}
func (client *fakeClient) FetchSubmissions(context.Context, string, int64) ([]training.Submission, error) {
	return client.submissions, client.err
}

type fakeRepository struct {
	updated    *time.Time
	candidates []training.Problem
	state      training.SyncState
	session    *training.Session
	created    training.Session
	applied    []training.Submission
}

func (repo *fakeRepository) CatalogUpdatedAt(context.Context) (*time.Time, error) {
	return repo.updated, nil
}
func (repo *fakeRepository) ReplaceCatalog(_ context.Context, _ []training.Contest, _ []training.Problem, at time.Time) error {
	repo.updated = &at
	return nil
}
func (repo *fakeRepository) ListEligibleProblems(context.Context, string, time.Time, int) ([]training.Problem, error) {
	return repo.candidates, nil
}
func (repo *fakeRepository) GetUserSyncState(context.Context, string) (training.SyncState, error) {
	return repo.state, nil
}
func (repo *fakeRepository) SaveSubmissionSync(_ context.Context, _ string, _ []training.Submission, state training.SyncState) error {
	repo.state = state
	return nil
}
func (repo *fakeRepository) CreateSession(_ context.Context, session training.Session) error {
	repo.created = session
	repo.session = &session
	return nil
}
func (repo *fakeRepository) GetActiveSession(context.Context, string, time.Time) (*training.Session, error) {
	return repo.session, nil
}
func (repo *fakeRepository) GetSession(context.Context, string, time.Time) (training.Session, error) {
	if repo.session == nil {
		return training.Session{}, training.ErrSessionNotFound
	}
	return *repo.session, nil
}
func (repo *fakeRepository) ListSessions(context.Context, string, time.Time, int, int) (training.SessionPage, error) {
	return training.SessionPage{}, nil
}
func (repo *fakeRepository) ApplySessionAcceptances(_ context.Context, _ string, items []training.Submission, _ time.Time) error {
	repo.applied = items
	return nil
}
func (repo *fakeRepository) AbortSession(context.Context, string, time.Time) error {
	repo.session.Status = training.StatusAborted
	return nil
}

type fixedRandom struct{}

func (fixedRandom) Intn(_ int) int                  { return 0 }
func (fixedRandom) Shuffle(_ int, _ func(i, j int)) {}
func difficulty(value int) *int                     { return &value }

func TestStartRefreshesDataAndCreatesFiveProblemSession(t *testing.T) {
	now := time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepository{candidates: []training.Problem{{ID: "d", ContestID: "a", Index: "D", Difficulty: difficulty(1000)}, {ID: "e1", ContestID: "b", Index: "E", Difficulty: difficulty(1200)}, {ID: "e2", ContestID: "c", Index: "E", Difficulty: difficulty(1400)}, {ID: "e3", ContestID: "d", Index: "E", Difficulty: difficulty(1600)}, {ID: "f", ContestID: "e", Index: "F", Difficulty: difficulty(1700)}}}
	client := &fakeClient{catalogProblems: repo.candidates}
	usecase := New(repo, client, training.DefaultConfig())
	usecase.now = func() time.Time { return now }
	usecase.random = func() training.RandomSource { return fixedRandom{} }
	response, err := usecase.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if response.Session == nil || len(response.Session.Problems) != 5 || repo.created.StartedAt != now {
		t.Fatalf("response=%#v", response)
	}
}

func TestSyncReturnsStaleSessionWithoutFailingWhenExternalAPIIsDown(t *testing.T) {
	now := time.Now().UTC()
	session := &training.Session{ID: "session", AtCoderUserID: "fken_prime_57", StartedAt: now, DurationSeconds: 4500, Status: training.StatusActive}
	repo := &fakeRepository{session: session}
	client := &fakeClient{err: context.DeadlineExceeded}
	usecase := New(repo, client, training.DefaultConfig())
	usecase.now = func() time.Time { return now }
	response, err := usecase.Sync(context.Background(), "session")
	if err != nil {
		t.Fatal(err)
	}
	if response.SubmissionSync.Status != "STALE" || response.Session == nil {
		t.Fatalf("response=%#v", response)
	}
}
