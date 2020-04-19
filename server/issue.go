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

	id, err := execReturningId(createIssueStmt,
		req.Title, req.Content, r.Context().Value(KeySelectedUserId), r.Context().Value(KeyProjectId))
	if err != nil {
		handleError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(gud.CreateIssueResponse{Id: id})
	if err != nil {
		handleError(w, err)
		return
	}
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
		issue, err := scanIssue(rows)
		if err != nil {
			handleError(w, err)
		}
		issues = append(issues, *issue)
	}

	if err := rows.Err(); err != nil {
		handleError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(issues)
	if err != nil {
		handleError(w, err)
		return
	}
}

func getIssue(w http.ResponseWriter, r *http.Request) {
	issueId, err := strconv.Atoi(mux.Vars(r)["issue"])
	if err != nil {
		reportError(w, http.StatusBadRequest, err.Error())
		return
	}

	issue, err := scanIssue(getIssueStmt.QueryRow(issueId))
	if err == sql.ErrNoRows {
		reportError(w, http.StatusNotFound, fmt.Sprintf("issue #%d not found", issueId))
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

func setIssueStatus(w http.ResponseWriter, r *http.Request) {
	issueId, err := strconv.Atoi(mux.Vars(r)["issue"])
	if err != nil {
		reportError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req gud.UpdateIssueRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		reportError(w, http.StatusBadRequest, err.Error())
		return
	}

	status := req.Status
	if _, ok := map[string]struct{}{
		"open":        {},
		"in_progress": {},
		"done":        {},
		"closed":      {},
	}[status]; !ok {
		reportError(w, http.StatusBadRequest, fmt.Sprintf("unrecognized issue status: %s", status))
		return
	}

	res, err := setIssueStatusStmt.Exec(issueId, status)
	if err != nil {
		handleError(w, err)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		handleError(w, err)
		return
	}

	if rows == 0 {
		reportError(w, http.StatusNotFound, fmt.Sprintf("issue #%d not found", issueId))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type scanner interface {
	Scan(ret ...interface{}) error
}

func scanIssue(row scanner) (*gud.Issue, error) {
	var issue gud.Issue
	var authorId int
	err := row.Scan(&issue.Id, &authorId, &issue.Title, &issue.Content, &issue.Status, &issue.Created)
	if err != nil {
		return nil, err
	}
	err = getUserStmt.QueryRow(authorId).Scan(&issue.Author)
	if err != nil {
		return nil, err
	}

	return &issue, nil
}
