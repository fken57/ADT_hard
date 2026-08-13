package training

import (
	"errors"
	"time"
)

const (
	StatusActive   = "ACTIVE"
	StatusFinished = "FINISHED"
	StatusAborted  = "ABORTED"
)

var (
	ErrActiveSessionExists   = errors.New("an active training session already exists")
	ErrAbortCooldown         = errors.New("a new training cannot start before the aborted session deadline")
	ErrProblemSetUnavailable = errors.New("could not generate a valid problem set")
	ErrSessionNotFound       = errors.New("training session not found")
	ErrSessionNotActive      = errors.New("training session is not active")
	ErrExternalDataStale     = errors.New("required AtCoder data could not be refreshed")
)

type SlotConfig struct {
	Name             string   `json:"name"`
	AllowedIndexes   []string `json:"allowedIndexes"`
	TargetDifficulty int      `json:"targetDifficulty"`
	Tolerance        int      `json:"tolerance"`
}

type DifficultyProfile struct {
	Name   string
	Weight int
	Slots  []SlotConfig
}

type Config struct {
	UserID                string
	Duration              time.Duration
	ExcludeLatestContests int
	CatalogTTL            time.Duration
	PollInterval          time.Duration
	Slots                 []SlotConfig
	DifficultyProfiles    []DifficultyProfile
}

func DefaultConfig() Config {
	standard := difficultySlots(0)
	return Config{
		UserID: "fken_prime_57", Duration: 75 * time.Minute,
		ExcludeLatestContests: 10, CatalogTTL: 24 * time.Hour,
		PollInterval: 15 * time.Second,
		Slots:        standard,
		DifficultyProfiles: []DifficultyProfile{
			{Name: "STANDARD", Weight: 70, Slots: standard},
			{Name: "LIGHT", Weight: 15, Slots: difficultySlots(-100)},
			{Name: "HEAVY", Weight: 15, Slots: difficultySlots(100)},
		},
	}
}

func difficultySlots(offset int) []SlotConfig {
	return []SlotConfig{
		{Name: "Warmup", AllowedIndexes: []string{"D", "E"}, TargetDifficulty: 900 + offset, Tolerance: 125},
		{Name: "Stable", AllowedIndexes: []string{"D", "E"}, TargetDifficulty: 1150 + offset, Tolerance: 125},
		{Name: "Main", AllowedIndexes: []string{"E"}, TargetDifficulty: 1300 + offset, Tolerance: 125},
		{Name: "Stretch", AllowedIndexes: []string{"E", "F"}, TargetDifficulty: 1500 + offset, Tolerance: 125},
		{Name: "Challenge", AllowedIndexes: []string{"E", "F"}, TargetDifficulty: 1700 + offset, Tolerance: 125},
	}
}

type Contest struct {
	ID             string    `json:"id" db:"id"`
	Title          string    `json:"title" db:"title"`
	StartTime      time.Time `json:"startTime" db:"start_time"`
	DurationSecond int64     `json:"durationSecond" db:"duration_second"`
	ProblemCount   int       `json:"problemCount" db:"problem_count"`
}

func (contest Contest) EndTime() time.Time {
	return contest.StartTime.Add(time.Duration(contest.DurationSecond) * time.Second)
}

type Problem struct {
	ID           string    `json:"problemId" db:"problem_id"`
	ContestID    string    `json:"contestId" db:"contest_id"`
	Index        string    `json:"problemIndex" db:"problem_index"`
	Title        string    `json:"title" db:"title"`
	Difficulty   *int      `json:"difficulty" db:"difficulty"`
	ContestStart time.Time `json:"-" db:"contest_start_time"`
	ProblemCount int       `json:"-" db:"problem_count"`
}

type SessionProblem struct {
	ID           string     `json:"id" db:"id"`
	SessionID    string     `json:"-" db:"session_id"`
	Slot         string     `json:"slot" db:"slot"`
	ContestID    string     `json:"contestId" db:"contest_id"`
	ProblemID    string     `json:"problemId" db:"problem_id"`
	Index        string     `json:"problemIndex" db:"problem_index"`
	Title        string     `json:"title" db:"title"`
	Difficulty   *int       `json:"difficulty,omitempty" db:"difficulty"`
	AcceptedAt   *time.Time `json:"acceptedAt,omitempty" db:"accepted_at"`
	PenaltyCount int        `json:"penaltyCount" db:"penalty_count"`
	URL          string     `json:"url" db:"-"`
}

func (problem SessionProblem) Link() string {
	return "https://atcoder.jp/contests/" + problem.ContestID + "/tasks/" + problem.ProblemID
}

type Session struct {
	ID                string           `json:"id" db:"id"`
	AtCoderUserID     string           `json:"atcoderUserId" db:"atcoder_user_id"`
	StartedAt         time.Time        `json:"startedAt" db:"started_at"`
	DurationSeconds   int              `json:"durationSeconds" db:"duration_seconds"`
	EndedAt           *time.Time       `json:"endedAt,omitempty" db:"ended_at"`
	Status            string           `json:"status" db:"status"`
	FallbackLevel     int              `json:"fallbackLevel" db:"fallback_level"`
	DifficultyProfile string           `json:"difficultyProfile" db:"difficulty_profile"`
	CreatedAt         time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time        `json:"updatedAt" db:"updated_at"`
	Problems          []SessionProblem `json:"problems"`
}

func (session Session) Deadline() time.Time {
	return session.StartedAt.Add(time.Duration(session.DurationSeconds) * time.Second)
}

func (session Session) AcceptedCount() int {
	count := 0
	for _, problem := range session.Problems {
		if problem.AcceptedAt != nil {
			count++
		}
	}
	return count
}

type Submission struct {
	ID          int64
	EpochSecond int64
	ProblemID   string
	Result      string
}

func AcceptedDuringSession(session Session, submission Submission) bool {
	return submission.Result == "AC" && SubmissionDuringSession(session, submission)
}

func SubmissionDuringSession(session Session, submission Submission) bool {
	submittedAt := time.Unix(submission.EpochSecond, 0).UTC()
	cutoff := session.Deadline()
	if session.Status == StatusAborted && session.EndedAt != nil && session.EndedAt.Before(cutoff) {
		cutoff = *session.EndedAt
	}
	return !submittedAt.Before(session.StartedAt) && submittedAt.Before(cutoff)
}

type SyncState struct {
	LastSubmissionEpoch int64
	LastSuccessfulAt    *time.Time
}

type SessionPage struct {
	Sessions []Session `json:"sessions"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
	Total    int       `json:"total"`
}
