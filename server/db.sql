CREATE TABLE users (
    id         serial PRIMARY KEY,
    username   varchar   NOT NULL,
    email      varchar   NOT NULL,
    password   bytea(60) NOT NULL,
    created_at timestamp NOT NULL,
    active     bool      NOT NULL
);

CREATE TABLE projects (
    id         serial PRIMARY KEY,
    name       varchar   NOT NULL,
    owner_id   int       NOT NULL REFERENCES users(id),
    created_at timestamp NOT NULL
);
