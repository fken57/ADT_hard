ALTER TABLE training_sessions
    ADD COLUMN difficulty_profile VARCHAR(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT 'LEGACY'
    AFTER fallback_level;
