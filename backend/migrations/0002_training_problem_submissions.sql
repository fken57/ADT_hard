CREATE TABLE IF NOT EXISTS training_problem_submissions (
    session_id CHAR(36) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    problem_id VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    submission_id BIGINT NOT NULL,
    submitted_at DATETIME(6) NOT NULL,
    result VARCHAR(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    PRIMARY KEY (session_id, submission_id),
    KEY training_problem_submissions_problem_idx (session_id, problem_id, submitted_at),
    CONSTRAINT training_problem_submissions_session_fk FOREIGN KEY (session_id)
        REFERENCES training_sessions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
