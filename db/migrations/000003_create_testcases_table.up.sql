CREATE TABLE IF NOT EXISTS testcases (
    id BIGSERIAL PRIMARY KEY,
    problem_id BIGINT NOT NULL,
    input TEXT,
    input_file VARCHAR(512),
    output TEXT,
    output_file VARCHAR(512),
    is_hidden BOOLEAN NOT NULL DEFAULT false,
    points INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE,
    CHECK (
        (input IS NOT NULL AND input_file IS NULL) OR
        (input IS NULL AND input_file IS NOT NULL)
    ),
    CHECK (
        (output IS NOT NULL AND output_file IS NULL) OR
        (output IS NULL AND output_file IS NOT NULL)
    )
);

CREATE INDEX idx_testcases_problem_id ON testcases(problem_id);
CREATE INDEX idx_testcases_is_hidden ON testcases(is_hidden);
