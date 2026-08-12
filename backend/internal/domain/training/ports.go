package training

import (
	"context"
	"math/rand"
	"time"
)

type Repository interface {
	CatalogUpdatedAt(ctx context.Context) (*time.Time, error)
	ReplaceCatalog(ctx context.Context, contests []Contest, problems []Problem, syncedAt time.Time) error
	ListEligibleProblems(ctx context.Context, userID string, now time.Time, excludeLatest int) ([]Problem, error)
	GetUserSyncState(ctx context.Context, userID string) (SyncState, error)
	SaveSubmissionSync(ctx context.Context, userID string, submissions []Submission, state SyncState) error
	CreateSession(ctx context.Context, session Session) error
	GetActiveSession(ctx context.Context, userID string, now time.Time) (*Session, error)
	GetSession(ctx context.Context, id string, now time.Time) (Session, error)
	ListSessions(ctx context.Context, userID string, now time.Time, page, pageSize int) (SessionPage, error)
	ApplySessionAcceptances(ctx context.Context, sessionID string, submissions []Submission, now time.Time) error
	AbortSession(ctx context.Context, sessionID string, now time.Time) error
}

type AtCoderClient interface {
	FetchCatalog(ctx context.Context) ([]Contest, []Problem, error)
	FetchSubmissions(ctx context.Context, userID string, fromSecond int64) ([]Submission, error)
}

type RandomSource interface { Shuffle(n int, swap func(i, j int)) }

func NewRandomSource(seed int64) RandomSource { return rand.New(rand.NewSource(seed)) }

