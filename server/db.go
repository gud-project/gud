package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB
var newUserStmt,
	userExistsStmt,
	userByNameStmt,
	userIdMatchesNameStmt,
	createProjectStmt,
	projectExistsStmt,
	projectByNameStmt *sql.Stmt

func init() {
	var err error
	db, err = sql.Open("postgres", "") // set PGUSER and PGPASSWORD
	if err != nil {
		log.Fatal(err)
	}

	// language=PostgreSQL
	{
		newUserStmt = mustPrepare(
			"INSERT INTO users (username, email, password, created_at) VALUES ($1, $2, $3, NOW());")

		userExistsStmt = mustPrepare(
			"SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);")

		userByNameStmt = mustPrepare(
			"SELECT user_id, password FROM users WHERE username = $1;")

		userIdMatchesNameStmt = mustPrepare(
			"SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1 AND username = $2);")

		createProjectStmt = mustPrepare(
			"INSERT INTO projects (name, user_id, created_at) VALUES ($1, $2, NOW());")

		projectExistsStmt = mustPrepare(
			"SELECT EXISTS(SELECT 1 FROM projects WHERE name = $1 AND user_id = $2);")

		projectByNameStmt = mustPrepare(
			"SELECT project_id FROM projects WHERE name = $1 AND user_id = $2;")
	}
}

func checkExists(stmt *sql.Stmt, args ...interface{}) (bool, error) {
	row := stmt.QueryRow(args...)

	var exists bool
	err := row.Scan(&exists)
	return exists, err
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
		userIdMatchesNameStmt,
		createProjectStmt,
		projectExistsStmt,
		projectByNameStmt,
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
