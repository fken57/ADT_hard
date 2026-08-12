CREATE TABLE IF NOT EXISTS atcoder_contests (
    id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    start_time DATETIME(6) NOT NULL,
    duration_second BIGINT NOT NULL,
    problem_count INT NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    KEY atcoder_contests_start_idx (start_time)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS atcoder_problems (
    problem_id VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
    contest_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    problem_index VARCHAR(8) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    title VARCHAR(255) NOT NULL,
    difficulty INT NULL,
    updated_at DATETIME(6) NOT NULL,
    KEY atcoder_problems_contest_idx (contest_id),
    KEY atcoder_problems_index_difficulty_idx (problem_index, difficulty),
    CONSTRAINT atcoder_problems_contest_fk FOREIGN KEY (contest_id)
        REFERENCES atcoder_contests(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS catalog_sync_states (
    name VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
    synced_at DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS accepted_problems (
    atcoder_user_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    problem_id VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    accepted_at DATETIME(6) NOT NULL,
    PRIMARY KEY (atcoder_user_id, problem_id),
    KEY accepted_problems_problem_idx (problem_id)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS atcoder_user_sync_states (
    atcoder_user_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
    last_submission_epoch BIGINT NOT NULL,
    last_successful_at DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS training_sessions (
    id CHAR(36) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
    atcoder_user_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    started_at DATETIME(6) NOT NULL,
    duration_seconds INT NOT NULL,
    ended_at DATETIME(6) NULL,
    status VARCHAR(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    fallback_level INT NOT NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    active_user_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin
        GENERATED ALWAYS AS (CASE WHEN status = 'ACTIVE' THEN atcoder_user_id ELSE NULL END) PERSISTENT,
    UNIQUE KEY training_sessions_one_active_user (active_user_id),
    KEY training_sessions_user_started_idx (atcoder_user_id, started_at),
    CONSTRAINT training_sessions_status_check CHECK (status IN ('ACTIVE', 'FINISHED', 'ABORTED'))
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS training_problems (
    id CHAR(36) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
    session_id CHAR(36) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    slot VARCHAR(8) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    contest_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    problem_id VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    problem_index VARCHAR(8) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    title VARCHAR(255) NOT NULL,
    difficulty INT NULL,
    accepted_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    UNIQUE KEY training_problems_session_slot (session_id, slot),
    UNIQUE KEY training_problems_session_contest (session_id, contest_id),
    UNIQUE KEY training_problems_session_problem (session_id, problem_id),
    KEY training_problems_problem_idx (problem_id),
    CONSTRAINT training_problems_session_fk FOREIGN KEY (session_id)
        REFERENCES training_sessions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

