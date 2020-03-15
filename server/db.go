package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB
var newUserStmt,
	userExistsStmt,
	getUserStmt,
	userByNameStmt,
	userIdMatchesNameStmt,
	createProjectStmt,
	projectExistsStmt,
	getProjectStmt,
	projectByNameStmt,
	userProjectsStmt,
	hasMemberStmt,
	inviteMemberStmt,
	createIssueStmt,
	getIssuesStmt,
	getIssueStmt,
	createPrStmt,
	getPrsStmt,
	getPrStmt* sql.Stmt

func init() {
	var err error
	db, err = sql.Open("postgres", "") // set PGUSER and PGPASSWORD
	if err != nil {
		log.Fatal(err)
	}

	// language=PostgreSQL
	{
		newUserStmt = mustPrepare(
			"INSERT INTO users (username, email, password, created_at) VALUES ($1, $2, $3, NOW()) RETURNING user_id;")

		userExistsStmt = mustPrepare(
			"SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);")

		getUserStmt = mustPrepare(
			"SELECT username FROM users WHERE user_id = $1;")

		userByNameStmt = mustPrepare(
			"SELECT user_id, password FROM users WHERE username = $1;")

		userIdMatchesNameStmt = mustPrepare(
			"SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1 AND username = $2);")

		createProjectStmt = mustPrepare(
			"INSERT INTO projects (name, user_id, created_at) VALUES ($1, $2, NOW()) RETURNING project_id;")

		projectExistsStmt = mustPrepare(
			"SELECT EXISTS(SELECT 1 FROM projects WHERE name = $1 AND user_id = $2);")

		getProjectStmt = mustPrepare(
			"SELECT user_id FROM projects WHERE project_id = $1;")

		projectByNameStmt = mustPrepare(`
			SELECT project_id, user_id FROM projects JOIN users USING (user_id)
			WHERE projects.name = $2 AND users.username = $1;`)

		userProjectsStmt = mustPrepare(
			"SELECT name FROM projects WHERE user_id = $1;")

		hasMemberStmt = mustPrepare(
			"SELECT EXISTS(SELECT 1 FROM members WHERE user_id = $1 AND project_id = $2);")

		inviteMemberStmt = mustPrepare(
			"INSERT INTO members (user_id, project_id) VALUES ($1, $2);")

		createIssueStmt = mustPrepare(
			"INSERT INTO issues (title, content, state, user_id, project_id) VALUES ($1, $2, $3, $4, $5);")

		getIssuesStmt = mustPrepare(
			"SELECT issue_id, user_id, title, content, status FROM issues WHERE project_id = $1")

		getIssueStmt = mustPrepare(
			"SELECT user_id, title, content, status FROM issues WHERE issue_id = $1")

		createPrStmt = mustPrepare(
			"INSERT INTO issues (title, content, user_id, project_id, \"from\", \"to\") VALUES ($1, $2, $3, $4, $5, $6);")

		getPrsStmt = mustPrepare(
			"SELECT issue_id, user_id, title, content, \"from\", \"to\" FROM prs WHERE project_id = $1")

		getPrStmt = mustPrepare(
			"SELECT user_id, title, content, \"from\", \"to\" FROM prs WHERE issue_id = $1")
	}
}

func checkExists(stmt *sql.Stmt, args ...interface{}) (bool, error) {
	row := stmt.QueryRow(args...)

	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

func execReturningId(stmt *sql.Stmt, args ...interface{}) (int, error) {
	var id int
	err := stmt.QueryRow(args...).Scan(&id)
	return id, err
}

func mustPrepare(query string) *sql.Stmt {
	stmt, err := db.Prepare(query)
	if err != nil {
		closeDB()
		log.Fatal(err)
	}

	return stmt
}

func closeDB() error {
	for _, stmt := range []*sql.Stmt{
		newUserStmt,
		userExistsStmt,
		userByNameStmt,
		getUserStmt,
		userIdMatchesNameStmt,
		createProjectStmt,
		projectExistsStmt,
		getProjectStmt,
		projectByNameStmt,
		userProjectsStmt,
		hasMemberStmt,
		inviteMemberStmt,
	} {
		if stmt != nil {
			err := stmt.Close()
			if err != nil {
				return err
			}
		}
	}

	return db.Close()
}
