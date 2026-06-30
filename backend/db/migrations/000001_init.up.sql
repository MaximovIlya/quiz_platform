CREATE TYPE role AS ENUM ('ORGANIZER', 'PARTICIPANT');
CREATE TYPE question_type AS ENUM ('SINGLE', 'MULTIPLE');
CREATE TYPE session_status AS ENUM ('WAITING', 'ACTIVE', 'FINISHED');

CREATE TABLE users (
    id         TEXT PRIMARY KEY,
    email      TEXT NOT NULL UNIQUE,
    password   TEXT,
    name       TEXT NOT NULL,
    role       role NOT NULL DEFAULT 'PARTICIPANT',
    google_id  TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE quizzes (
    id                  TEXT PRIMARY KEY,
    title               TEXT NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    category            TEXT NOT NULL,
    time_per_question   INT  NOT NULL DEFAULT 30,
    points_per_question INT  NOT NULL DEFAULT 1000,
    scoring             TEXT NOT NULL DEFAULT 'standard',
    difficulty          TEXT NOT NULL DEFAULT 'Средне',
    tags                TEXT[] NOT NULL DEFAULT '{}',
    cover_image_url     TEXT,
    archived            BOOLEAN NOT NULL DEFAULT FALSE,
    author_id           TEXT NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE questions (
    id         TEXT PRIMARY KEY,
    quiz_id    TEXT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    text       TEXT NOT NULL,
    image_url  TEXT,
    type       question_type NOT NULL DEFAULT 'SINGLE',
    "order"    INT  NOT NULL,
    tags       TEXT[] NOT NULL DEFAULT '{}',
    time_limit INT  NOT NULL DEFAULT 30,
    points     INT  NOT NULL DEFAULT 1000
);

CREATE TABLE answers (
    id          TEXT PRIMARY KEY,
    question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    text        TEXT NOT NULL,
    is_correct  BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE quiz_sessions (
    id         TEXT PRIMARY KEY,
    quiz_id    TEXT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    room_code  TEXT NOT NULL UNIQUE,
    status     session_status NOT NULL DEFAULT 'WAITING',
    started_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE session_players (
    id         TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES quiz_sessions(id) ON DELETE CASCADE,
    user_id    TEXT NOT NULL REFERENCES users(id),
    score      INT  NOT NULL DEFAULT 0,
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE player_answers (
    id                TEXT PRIMARY KEY,
    session_player_id TEXT NOT NULL REFERENCES session_players(id) ON DELETE CASCADE,
    question_id       TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answered_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_correct        BOOLEAN NOT NULL,
    points            INT NOT NULL DEFAULT 0
);

CREATE TABLE player_answer_answers (
    player_answer_id TEXT NOT NULL REFERENCES player_answers(id) ON DELETE CASCADE,
    answer_id        TEXT NOT NULL REFERENCES answers(id) ON DELETE CASCADE,
    PRIMARY KEY (player_answer_id, answer_id)
);

CREATE TABLE password_reset_tokens (
    id         TEXT PRIMARY KEY,
    email      TEXT NOT NULL,
    token      TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
