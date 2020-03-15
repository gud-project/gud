package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gitlab.com/magsh-2019/2/gud/gud"
)

func createIssue(w http.ResponseWriter, r *http.Request) {
	var req gud.CreateIssueRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		reportError(w, http.StatusBadRequest, "failed to receive issue data")
		return
	}

	_, err = createIssueStmt.Exec(
		req.Title, req.Content, r.Context().Value(KeySelectedUserId), r.Context().Value(KeyProjectId))
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getIssues(w http.ResponseWriter, r *http.Request) {
	rows, err := getIssuesStmt.Query(r.Context().Value(KeyProjectId))
	if err != nil {
		handleError(w, err)
		return
	}
	defer rows.Close()

	issues := make([]gud.Issue, 0)
	for rows.Next() {
		var issue gud.Issue
		if err := rows.Scan(&issue.Title, &issue.Author, &issue.Content, &issue.Id, &issue.Status); err != nil {
			handleError(w, err)
			return
		}
		issues = append(issues, issue)
	}

	if err := rows.Err(); err != nil {
		handleError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(gud.GetIssuesResponse{Issues: issues})
	if err != nil {
		handleError(w, err)
		return
	}
}

func getIssue(w http.ResponseWriter, r *http.Request) {
	var issue gud.Issue
	id := mux.Vars(r)["issue"]
	var err error
	issue.Id, err = strconv.Atoi(id)
	if err != nil {
		reportError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = getIssueStmt.QueryRow(issue.Id).Scan(&issue.Author, &issue.Title, &issue.Content, &issue.Status)
	if err == sql.ErrNoRows {
		reportError(w, http.StatusNotFound, fmt.Sprintf("issue #%s not found", id))
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(issue)
	if err != nil {
		handleError(w, err)
		return
	}
}
