CREATE TABLE IF NOT EXISTS users (
    user_id    serial PRIMARY KEY,
    username   varchar   NOT NULL,
    email      varchar   NOT NULL,
    password   bytea     NOT NULL,
    created_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
    project_id serial PRIMARY KEY,
    name       varchar   NOT NULL,
    user_id    int       NOT NULL REFERENCES users(user_id),
    created_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS members (
    member_id  serial PRIMARY KEY,
    user_id    int NOT NULL REFERENCES users(user_id),
    project_id int NOT NULL REFERENCES projects(project_id)
);

CREATE TYPE issue_status AS ENUM ('open', 'in_progress', 'done', 'closed');

CREATE TABLE IF NOT EXISTS issues (
    issue_id   serial PRIMARY KEY,
    title      varchar      NOT NULL,
    user_id    int          NOT NULL REFERENCES users(user_id),
    content    varchar      NOT NULL,
    project_id int          NOT NULL REFERENCES projects(project_id),
    status     issue_status NOT NULL,
    created_at timestamp    NOT NULL
);

CREATE TYPE pr_status AS ENUM ('open', 'merged', 'closed');

CREATE TABLE IF NOT EXISTS prs (
    pr_id      serial PRIMARY KEY,
    title      varchar   NOT NULL,
    user_id    int       NOT NULL REFERENCES users(user_id),
    content    varchar   NOT NULL,
    project_id int       NOT NULL REFERENCES projects(project_id),
    "from"     varchar   NOT NULL,
    "to"       varchar   NOT NULL,
    status     pr_status NOT NULL,
    created_at timestamp NOT NULL
);

CREATE TYPE job_status AS ENUM ('pending', 'success', 'failure');

CREATE TABLE jobs (
    job_id     serial PRIMARY KEY,
    project_id int        NOT NULL REFERENCES projects(project_id),
    "version"  char(40)   NOT NULL,
    status     job_status NOT NULL,
    logs       varchar    NOT NULL
);
