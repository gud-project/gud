package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB
var newUserStmt,
	userExistsStmt,
	userByNameStmt,
	createProjectStmt,
	projectExistsStmt *sql.Stmt

func init() {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("user=gud password=%s", os.Getenv("PQ_PASS")))
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
			"SELECT user_id, password FROM users WHERE username = $1")

		createProjectStmt = mustPrepare(
			"INSERT INTO projects (name, user_id, created_at) VALUES ($1, $2, NOW());")

		projectExistsStmt = mustPrepare(
			"SELECT EXISTS(SELECT 1 FROM projects WHERE name = $1 AND user_id = $2);")
	}
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
		createProjectStmt,
		projectExistsStmt,
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
