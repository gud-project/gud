package main

import (
	"encoding/json"
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

	title := req.Title
	content := req.Content

	user := mux.Vars(r)["user"]
	project := mux.Vars(r)["project"]

	var userId, projectId int
	err = getUserStmt.QueryRow(user).Scan(&userId)
	if err != nil {
		handleError(w, err)
		return
	}

	err = getProjectStmt.QueryRow(user, project).Scan(&projectId)
	if err != nil {
		handleError(w, err)
		return
	}

	_, err = createIssueStmt.Exec(title, content, userId, projectId)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getIssues(w http.ResponseWriter, r *http.Request) {
	user := mux.Vars(r)["user"]
	project := mux.Vars(r)["project"]

	var projectId int
	err := getProjectStmt.QueryRow(user, project).Scan(&projectId)
	if err != nil {
		handleError(w, err)
		return
	}

	rows, err := getIssuesStmt.Query(projectId)
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
