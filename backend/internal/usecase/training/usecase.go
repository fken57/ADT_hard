package trainingusecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"atcoder_shojin/backend/internal/domain/training"
	"github.com/google/uuid"
)

const HistoryPageSize = 10

type Usecase struct {
	repository training.Repository
	client     training.AtCoderClient
	config     training.Config
	now        func() time.Time
	random     func() training.RandomSource
	syncMu     sync.Mutex
}

type SyncInfo struct {
	Status           string     `json:"status"`
	LastSuccessfulAt *time.Time `json:"lastSuccessfulAt,omitempty"`
	Message          string     `json:"message,omitempty"`
}

type SessionResponse struct {
	Session        *training.Session `json:"session"`
	ServerNow      time.Time         `json:"serverNow"`
	SubmissionSync SyncInfo          `json:"submissionSync"`
}

type HistoryResponse struct {
	training.SessionPage
	ServerNow time.Time `json:"serverNow"`
}

func New(repository training.Repository, client training.AtCoderClient, config training.Config) *Usecase {
	usecase := &Usecase{repository: repository, client: client, config: config, now: func() time.Time { return time.Now().UTC() }}
	usecase.random = func() training.RandomSource { return training.NewRandomSource(usecase.now().UnixNano()) }
	return usecase
}

func (usecase *Usecase) Start(ctx context.Context) (SessionResponse, error) {
	now := usecase.now().UTC()
	if err := usecase.refreshCatalog(ctx, now); err != nil {
		return SessionResponse{}, err
	}
	if _, err := usecase.syncUser(ctx, now); err != nil {
		return SessionResponse{}, training.ErrExternalDataStale
	}
	candidates, err := usecase.repository.ListEligibleProblems(ctx, usecase.config.UserID, now, usecase.config.ExcludeLatestContests)
	if err != nil {
		return SessionResponse{}, err
	}
	selected, level, err := training.SelectProblemSet(usecase.config, candidates, usecase.random())
	if err != nil {
		return SessionResponse{}, err
	}
	session := training.Session{ID: uuid.NewString(), AtCoderUserID: usecase.config.UserID, StartedAt: now, DurationSeconds: int(usecase.config.Duration / time.Second), Status: training.StatusActive, FallbackLevel: level, CreatedAt: now, UpdatedAt: now, Problems: make([]training.SessionProblem, 0, len(selected))}
	for _, item := range selected {
		problem := training.SessionProblem{ID: uuid.NewString(), SessionID: session.ID, Slot: item.Slot.Name, ContestID: item.Problem.ContestID, ProblemID: item.Problem.ID, Index: item.Problem.Index, Title: item.Problem.Title, Difficulty: item.Problem.Difficulty}
		problem.URL = problem.Link()
		session.Problems = append(session.Problems, problem)
	}
	if err := usecase.repository.CreateSession(ctx, session); err != nil {
		return SessionResponse{}, err
	}
	state, _ := usecase.repository.GetUserSyncState(ctx, usecase.config.UserID)
	return SessionResponse{Session: &session, ServerNow: now, SubmissionSync: SyncInfo{Status: "OK", LastSuccessfulAt: state.LastSuccessfulAt}}, nil
}

func (usecase *Usecase) Active(ctx context.Context) (SessionResponse, error) {
	now := usecase.now().UTC()
	session, err := usecase.repository.GetActiveSession(ctx, usecase.config.UserID, now)
	if err != nil {
		return SessionResponse{}, err
	}
	state, _ := usecase.repository.GetUserSyncState(ctx, usecase.config.UserID)
	return SessionResponse{Session: session, ServerNow: now, SubmissionSync: SyncInfo{Status: "OK", LastSuccessfulAt: state.LastSuccessfulAt}}, nil
}

func (usecase *Usecase) Get(ctx context.Context, id string) (SessionResponse, error) {
	now := usecase.now().UTC()
	session, err := usecase.repository.GetSession(ctx, id, now)
	if err != nil {
		return SessionResponse{}, err
	}
	state, _ := usecase.repository.GetUserSyncState(ctx, usecase.config.UserID)
	return SessionResponse{Session: &session, ServerNow: now, SubmissionSync: SyncInfo{Status: "OK", LastSuccessfulAt: state.LastSuccessfulAt}}, nil
}

func (usecase *Usecase) Sync(ctx context.Context, id string) (SessionResponse, error) {
	now := usecase.now().UTC()
	session, err := usecase.repository.GetSession(ctx, id, now)
	if err != nil {
		return SessionResponse{}, err
	}
	if session.AtCoderUserID != usecase.config.UserID {
		return SessionResponse{}, training.ErrSessionNotFound
	}
	items, syncErr := usecase.syncFrom(ctx, session.StartedAt.Unix(), now)
	if syncErr == nil {
		syncErr = usecase.repository.ApplySessionAcceptances(ctx, id, items, now)
	}
	updated, getErr := usecase.repository.GetSession(ctx, id, now)
	if getErr != nil {
		return SessionResponse{}, getErr
	}
	state, _ := usecase.repository.GetUserSyncState(ctx, usecase.config.UserID)
	info := SyncInfo{Status: "OK", LastSuccessfulAt: state.LastSuccessfulAt}
	if syncErr != nil {
		info.Status = "STALE"
		info.Message = "Submission status could not be updated. Retrying..."
	}
	return SessionResponse{Session: &updated, ServerNow: now, SubmissionSync: info}, nil
}

func (usecase *Usecase) Abort(ctx context.Context, id string) (SessionResponse, error) {
	now := usecase.now().UTC()
	if err := usecase.repository.AbortSession(ctx, id, now); err != nil {
		return SessionResponse{}, err
	}
	return usecase.Get(ctx, id)
}

func (usecase *Usecase) History(ctx context.Context, page int) (HistoryResponse, error) {
	if page < 1 {
		return HistoryResponse{}, fmt.Errorf("page must be at least 1")
	}
	now := usecase.now().UTC()
	result, err := usecase.repository.ListSessions(ctx, usecase.config.UserID, now, page, HistoryPageSize)
	return HistoryResponse{SessionPage: result, ServerNow: now}, err
}

func (usecase *Usecase) refreshCatalog(ctx context.Context, now time.Time) error {
	updated, err := usecase.repository.CatalogUpdatedAt(ctx)
	if err != nil {
		return err
	}
	if updated != nil && now.Sub(*updated) < usecase.config.CatalogTTL {
		return nil
	}
	contests, problems, err := usecase.client.FetchCatalog(ctx)
	if err != nil {
		return training.ErrExternalDataStale
	}
	if len(contests) == 0 || len(problems) == 0 {
		return training.ErrExternalDataStale
	}
	if err := usecase.repository.ReplaceCatalog(ctx, contests, problems, now); err != nil {
		return err
	}
	return nil
}

func (usecase *Usecase) syncUser(ctx context.Context, now time.Time) ([]training.Submission, error) {
	state, err := usecase.repository.GetUserSyncState(ctx, usecase.config.UserID)
	if err != nil {
		return nil, err
	}
	from := state.LastSubmissionEpoch
	if from > 0 {
		from--
	}
	return usecase.syncFrom(ctx, from, now)
}

func (usecase *Usecase) syncFrom(ctx context.Context, from int64, now time.Time) ([]training.Submission, error) {
	usecase.syncMu.Lock()
	defer usecase.syncMu.Unlock()
	state, err := usecase.repository.GetUserSyncState(ctx, usecase.config.UserID)
	if err != nil {
		return nil, err
	}
	items, err := usecase.client.FetchSubmissions(ctx, usecase.config.UserID, from)
	if err != nil {
		return nil, err
	}
	cursor := state.LastSubmissionEpoch
	for _, item := range items {
		if item.EpochSecond > cursor {
			cursor = item.EpochSecond
		}
	}
	state.LastSubmissionEpoch = cursor
	state.LastSuccessfulAt = &now
	if err := usecase.repository.SaveSubmissionSync(ctx, usecase.config.UserID, items, state); err != nil {
		return nil, err
	}
	return items, nil
}
