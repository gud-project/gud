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
	setIssueStatusStmt,
	createPrStmt,
	getPrsStmt,
	getPrStmt,
	mergePrStmt,
	closePrStmt,
	createJobStmt,
	finishJobStmt,
	getJobsStmt,
	getJobStmt *sql.Stmt

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
			SELECT project_id, user_id FROM projects WHERE projects.name = $2 AND user_id = $1;`)

		userProjectsStmt = mustPrepare(
			"SELECT name FROM projects WHERE user_id = $1;")

		hasMemberStmt = mustPrepare(
			"SELECT EXISTS(SELECT 1 FROM members WHERE user_id = $1 AND project_id = $2);")

		inviteMemberStmt = mustPrepare(
			"INSERT INTO members (user_id, project_id) VALUES ($1, $2);")

		createIssueStmt = mustPrepare(`
			INSERT INTO issues (title, content, user_id, project_id, status, created_at)
			VALUES ($1, $2, $3, $4, 'open', NOW())
			RETURNING issue_id;`)

		getIssuesStmt = mustPrepare(
			"SELECT issue_id, user_id, title, content, status, created_at FROM issues WHERE project_id = $1")

		getIssueStmt = mustPrepare(
			"SELECT issue_id, user_id, title, content, status, created_at FROM issues WHERE issue_id = $1")

		setIssueStatusStmt = mustPrepare(
			"UPDATE issues SET status = $2 WHERE issue_id = $1;")

		createPrStmt = mustPrepare(`
			INSERT INTO prs (title, content, user_id, project_id, "from", "to", status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, 'open', NOW())
			RETURNING pr_id;`)

		getPrsStmt = mustPrepare(
			`SELECT pr_id, user_id, title, content, "from", "to", status, created_at FROM prs WHERE project_id = $1`)

		getPrStmt = mustPrepare(
			`SELECT pr_id, user_id, title, content, "from", "to", status, created_at FROM prs WHERE pr_id = $1`)

		mergePrStmt = mustPrepare(
			"UPDATE prs SET status = 'merged' WHERE pr_id = $1;")

		closePrStmt = mustPrepare(
			"UPDATE prs SET status = 'closed' WHERE pr_id = $1;")

		createJobStmt = mustPrepare(
			`INSERT INTO jobs (project_id, "version", status, logs) VALUES ($1, $2, 'pending', '') RETURNING job_id;`)

		finishJobStmt = mustPrepare(
			"UPDATE jobs SET status = $2, logs = $3 WHERE job_id = $1;")

		getJobsStmt = mustPrepare(
			`SELECT job_id, "version", status FROM jobs WHERE project_id = $1;`)

		getJobStmt = mustPrepare(
			`SELECT "version", status, logs FROM jobs WHERE job_id = $1;`)
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
		createIssueStmt,
		getIssuesStmt,
		getIssueStmt,
		setIssueStatusStmt,
		createPrStmt,
		getPrsStmt,
		getPrStmt,
		mergePrStmt,
		closePrStmt,
		createJobStmt,
		finishJobStmt,
		getJobsStmt,
		getJobStmt,
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
