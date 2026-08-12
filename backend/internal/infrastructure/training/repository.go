package trainingrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"atcoder_shojin/backend/internal/domain/training"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Repository struct{ db *sqlx.DB }

func NewRepository(db *sqlx.DB) *Repository { return &Repository{db: db} }

func (repository *Repository) CatalogUpdatedAt(ctx context.Context) (*time.Time, error) {
	var value time.Time
	err := repository.db.GetContext(ctx, &value, "SELECT synced_at FROM catalog_sync_states WHERE name='atcoder'")
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (repository *Repository) ReplaceCatalog(ctx context.Context, contests []training.Contest, problems []training.Problem, syncedAt time.Time) error {
	tx, err := repository.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, contest := range contests {
		_, err = tx.ExecContext(ctx, `INSERT INTO atcoder_contests(id,title,start_time,duration_second,problem_count,updated_at)
			VALUES(?,?,?,?,?,?) ON DUPLICATE KEY UPDATE title=VALUES(title),start_time=VALUES(start_time),duration_second=VALUES(duration_second),problem_count=VALUES(problem_count),updated_at=VALUES(updated_at)`, contest.ID, contest.Title, contest.StartTime, contest.DurationSecond, contest.ProblemCount, syncedAt)
		if err != nil {
			return fmt.Errorf("save contest %s: %w", contest.ID, err)
		}
	}
	for _, problem := range problems {
		_, err = tx.ExecContext(ctx, `INSERT INTO atcoder_problems(problem_id,contest_id,problem_index,title,difficulty,updated_at)
			VALUES(?,?,?,?,?,?) ON DUPLICATE KEY UPDATE contest_id=VALUES(contest_id),problem_index=VALUES(problem_index),title=VALUES(title),difficulty=VALUES(difficulty),updated_at=VALUES(updated_at)`, problem.ID, problem.ContestID, problem.Index, problem.Title, problem.Difficulty, syncedAt)
		if err != nil {
			return fmt.Errorf("save problem %s: %w", problem.ID, err)
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO catalog_sync_states(name,synced_at) VALUES('atcoder',?) ON DUPLICATE KEY UPDATE synced_at=VALUES(synced_at)`, syncedAt)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (repository *Repository) ListEligibleProblems(ctx context.Context, userID string, now time.Time, excludeLatest int) ([]training.Problem, error) {
	items := []training.Problem{}
	err := repository.db.SelectContext(ctx, &items, `WITH ranked AS (
		SELECT id,start_time,problem_count,ROW_NUMBER() OVER (ORDER BY start_time DESC) AS recent_rank
		FROM atcoder_contests
		WHERE id REGEXP '^abc[0-9]+$' AND DATE_ADD(start_time, INTERVAL duration_second SECOND) <= ?
	)
	SELECT p.problem_id,p.contest_id,p.problem_index,p.title,p.difficulty,r.start_time AS contest_start_time,r.problem_count
	FROM atcoder_problems p JOIN ranked r ON r.id=p.contest_id
	LEFT JOIN accepted_problems a ON a.atcoder_user_id=? AND a.problem_id=p.problem_id
	WHERE r.recent_rank>? AND r.problem_count IN (7,8) AND p.problem_index IN ('D','E','F') AND a.problem_id IS NULL`, now, userID, excludeLatest)
	return items, err
}

func (repository *Repository) GetUserSyncState(ctx context.Context, userID string) (training.SyncState, error) {
	var row struct {
		LastSubmissionEpoch int64     `db:"last_submission_epoch"`
		LastSuccessfulAt    time.Time `db:"last_successful_at"`
	}
	err := repository.db.GetContext(ctx, &row, `SELECT last_submission_epoch,last_successful_at FROM atcoder_user_sync_states WHERE atcoder_user_id=?`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return training.SyncState{}, nil
	}
	if err != nil {
		return training.SyncState{}, err
	}
	return training.SyncState{LastSubmissionEpoch: row.LastSubmissionEpoch, LastSuccessfulAt: &row.LastSuccessfulAt}, nil
}

func (repository *Repository) SaveSubmissionSync(ctx context.Context, userID string, submissions []training.Submission, state training.SyncState) error {
	tx, err := repository.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, item := range submissions {
		if item.Result != "AC" {
			continue
		}
		acceptedAt := time.Unix(item.EpochSecond, 0).UTC()
		_, err = tx.ExecContext(ctx, `INSERT INTO accepted_problems(atcoder_user_id,problem_id,accepted_at) VALUES(?,?,?) ON DUPLICATE KEY UPDATE accepted_at=LEAST(accepted_at,VALUES(accepted_at))`, userID, item.ProblemID, acceptedAt)
		if err != nil {
			return err
		}
	}
	when := time.Now().UTC()
	if state.LastSuccessfulAt != nil {
		when = *state.LastSuccessfulAt
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO atcoder_user_sync_states(atcoder_user_id,last_submission_epoch,last_successful_at) VALUES(?,?,?) ON DUPLICATE KEY UPDATE last_submission_epoch=GREATEST(last_submission_epoch,VALUES(last_submission_epoch)),last_successful_at=VALUES(last_successful_at)`, userID, state.LastSubmissionEpoch, when)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (repository *Repository) CreateSession(ctx context.Context, session training.Session) error {
	tx, err := repository.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = finishExpired(ctx, tx, session.StartedAt); err != nil {
		return err
	}
	var count int
	err = tx.GetContext(ctx, &count, `SELECT COUNT(*) FROM training_sessions WHERE atcoder_user_id=? AND status='ACTIVE' FOR UPDATE`, session.AtCoderUserID)
	if err != nil {
		return err
	}
	if count > 0 {
		return training.ErrActiveSessionExists
	}
	err = tx.GetContext(ctx, &count, `SELECT COUNT(*) FROM training_sessions WHERE atcoder_user_id=? AND status='ABORTED' AND DATE_ADD(started_at,INTERVAL duration_seconds SECOND)>? FOR UPDATE`, session.AtCoderUserID, session.StartedAt)
	if err != nil {
		return err
	}
	if count > 0 {
		return training.ErrAbortCooldown
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO training_sessions(id,atcoder_user_id,started_at,duration_seconds,ended_at,status,fallback_level,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, session.ID, session.AtCoderUserID, session.StartedAt, session.DurationSeconds, session.EndedAt, session.Status, session.FallbackLevel, session.CreatedAt, session.UpdatedAt)
	if err != nil {
		var mysqlError *mysql.MySQLError
		if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
			return training.ErrActiveSessionExists
		}
		return err
	}
	for _, problem := range session.Problems {
		_, err = tx.ExecContext(ctx, `INSERT INTO training_problems(id,session_id,slot,contest_id,problem_id,problem_index,title,difficulty,accepted_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, problem.ID, session.ID, problem.Slot, problem.ContestID, problem.ProblemID, problem.Index, problem.Title, problem.Difficulty, problem.AcceptedAt, session.CreatedAt, session.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func finishExpired(ctx context.Context, executor sqlx.ExtContext, now time.Time) error {
	_, err := executor.ExecContext(ctx, `UPDATE training_sessions SET status='FINISHED',ended_at=DATE_ADD(started_at,INTERVAL duration_seconds SECOND),updated_at=? WHERE status='ACTIVE' AND DATE_ADD(started_at,INTERVAL duration_seconds SECOND)<=?`, now, now)
	return err
}

func (repository *Repository) GetActiveSession(ctx context.Context, userID string, now time.Time) (*training.Session, error) {
	if _, err := repository.db.ExecContext(ctx, `UPDATE training_sessions SET status='FINISHED',ended_at=DATE_ADD(started_at,INTERVAL duration_seconds SECOND),updated_at=? WHERE status='ACTIVE' AND DATE_ADD(started_at,INTERVAL duration_seconds SECOND)<=?`, now, now); err != nil {
		return nil, err
	}
	var id string
	err := repository.db.GetContext(ctx, &id, `SELECT id FROM training_sessions WHERE atcoder_user_id=? AND status='ACTIVE'`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	session, err := repository.loadSession(ctx, id)
	return &session, err
}

func (repository *Repository) GetSession(ctx context.Context, id string, now time.Time) (training.Session, error) {
	if _, err := repository.db.ExecContext(ctx, `UPDATE training_sessions SET status='FINISHED',ended_at=DATE_ADD(started_at,INTERVAL duration_seconds SECOND),updated_at=? WHERE id=? AND status='ACTIVE' AND DATE_ADD(started_at,INTERVAL duration_seconds SECOND)<=?`, now, id, now); err != nil {
		return training.Session{}, err
	}
	return repository.loadSession(ctx, id)
}

func (repository *Repository) loadSession(ctx context.Context, id string) (training.Session, error) {
	var session training.Session
	err := repository.db.GetContext(ctx, &session, `SELECT id,atcoder_user_id,started_at,duration_seconds,ended_at,status,fallback_level,created_at,updated_at FROM training_sessions WHERE id=?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return training.Session{}, training.ErrSessionNotFound
	}
	if err != nil {
		return training.Session{}, err
	}
	err = repository.db.SelectContext(ctx, &session.Problems, `SELECT tp.id,tp.session_id,tp.slot,tp.contest_id,tp.problem_id,tp.problem_index,tp.title,tp.difficulty,tp.accepted_at,
		CASE WHEN tp.accepted_at IS NULL THEN 0 ELSE (
			SELECT COUNT(*) FROM training_problem_submissions tps
			WHERE tps.session_id=tp.session_id AND tps.problem_id=tp.problem_id
			AND tps.result<>'AC' AND tps.submitted_at<tp.accepted_at
		) END AS penalty_count
		FROM training_problems tp WHERE tp.session_id=? ORDER BY FIELD(tp.slot,'D1','E1','E2','E3','F1')`, id)
	for index := range session.Problems {
		session.Problems[index].URL = session.Problems[index].Link()
	}
	return session, err
}

func (repository *Repository) ListSessions(ctx context.Context, userID string, now time.Time, page, pageSize int) (training.SessionPage, error) {
	if _, err := repository.db.ExecContext(ctx, `UPDATE training_sessions SET status='FINISHED',ended_at=DATE_ADD(started_at,INTERVAL duration_seconds SECOND),updated_at=? WHERE status='ACTIVE' AND DATE_ADD(started_at,INTERVAL duration_seconds SECOND)<=?`, now, now); err != nil {
		return training.SessionPage{}, err
	}
	var total int
	if err := repository.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM training_sessions WHERE atcoder_user_id=?`, userID); err != nil {
		return training.SessionPage{}, err
	}
	ids := []string{}
	if err := repository.db.SelectContext(ctx, &ids, `SELECT id FROM training_sessions WHERE atcoder_user_id=? ORDER BY started_at DESC LIMIT ? OFFSET ?`, userID, pageSize, (page-1)*pageSize); err != nil {
		return training.SessionPage{}, err
	}
	result := training.SessionPage{Sessions: make([]training.Session, 0, len(ids)), Page: page, PageSize: pageSize, Total: total}
	for _, id := range ids {
		session, err := repository.loadSession(ctx, id)
		if err != nil {
			return training.SessionPage{}, err
		}
		result.Sessions = append(result.Sessions, session)
	}
	return result, nil
}

func (repository *Repository) ApplySessionAcceptances(ctx context.Context, sessionID string, submissions []training.Submission, now time.Time) error {
	tx, err := repository.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var session training.Session
	err = tx.GetContext(ctx, &session, `SELECT id,started_at,duration_seconds,ended_at,status FROM training_sessions WHERE id=? FOR UPDATE`, sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return training.ErrSessionNotFound
	}
	if err != nil {
		return err
	}
	for _, item := range submissions {
		if !training.SubmissionDuringSession(session, item) {
			continue
		}
		submittedAt := time.Unix(item.EpochSecond, 0).UTC()
		_, err = tx.ExecContext(ctx, `INSERT IGNORE INTO training_problem_submissions(session_id,problem_id,submission_id,submitted_at,result)
			SELECT session_id,problem_id,?,?,? FROM training_problems WHERE session_id=? AND problem_id=?`, item.ID, submittedAt, item.Result, sessionID, item.ProblemID)
		if err != nil {
			return err
		}
		if item.Result == "AC" {
			_, err = tx.ExecContext(ctx, `UPDATE training_problems SET accepted_at=CASE WHEN accepted_at IS NULL OR ? < accepted_at THEN ? ELSE accepted_at END,updated_at=? WHERE session_id=? AND problem_id=?`, submittedAt, submittedAt, now, sessionID, item.ProblemID)
			if err != nil {
				return err
			}
		}
	}
	if err = finishExpired(ctx, tx, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (repository *Repository) AbortSession(ctx context.Context, sessionID string, now time.Time) error {
	result, err := repository.db.ExecContext(ctx, `UPDATE training_sessions SET status='ABORTED',ended_at=?,updated_at=? WHERE id=? AND status='ACTIVE' AND DATE_ADD(started_at,INTERVAL duration_seconds SECOND)>?`, now, now, sessionID, now)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		session, loadErr := repository.GetSession(ctx, sessionID, now)
		if loadErr != nil {
			return loadErr
		}
		if session.Status != training.StatusActive {
			return training.ErrSessionNotActive
		}
	}
	return nil
}

func IsDuplicateError(err error) bool {
	var mysqlError *mysql.MySQLError
	return errors.As(err, &mysqlError) && mysqlError.Number == 1062
}
func IsConnectionError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "database"))
}
