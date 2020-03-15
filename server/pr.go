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

func createPr(w http.ResponseWriter, r *http.Request) {
	user := mux.Vars(r)["user"]
	project := mux.Vars(r)["project"]

	var userId, projectId int
	err := getUserStmt.QueryRow(user).Scan(&userId)
	if err != nil {
		handleError(w, err)
		return
	}

	err = getProjectStmt.QueryRow(user, project).Scan(&projectId)
	if err != nil {
		handleError(w, err)
		return
	}

	var req gud.CreatePrRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		reportError(w, http.StatusBadRequest, "failed to receive pr data")
		return
	}

	_, err = createPrStmt.Exec(req.Title, req.Content, userId, projectId, req.From, req.To)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getPrs(w http.ResponseWriter, r *http.Request) {
	user := mux.Vars(r)["user"]
	project := mux.Vars(r)["project"]

	var projectId int
	err := getProjectStmt.QueryRow(user, project).Scan(&projectId)
	if err != nil {
		handleError(w, err)
		return
	}

	rows, err := getPrsStmt.Query(projectId)
	if err != nil {
		handleError(w, err)
		return
	}
	defer rows.Close()

	prs := make([]gud.PullRequest, 0)
	for rows.Next() {
		var pr gud.PullRequest
		if err := rows.Scan(&pr.Title, &pr.Author, &pr.Content, &pr.Id, &pr.From, &pr.To); err != nil {
			handleError(w, err)
			return
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		handleError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(gud.GetPrsResponse{Prs: prs})
	if err != nil {
		handleError(w, err)
		return
	}
}

func getPr(w http.ResponseWriter, r *http.Request) {
	var pr gud.PullRequest
	id := mux.Vars(r)["pr"]
	var err error
	pr.Id, err = strconv.Atoi(id)
	if err != nil {
		reportError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = getPrStmt.QueryRow(pr.Id).Scan(&pr.Author, &pr.Title, &pr.Content, &pr.From, &pr.To)
	if err == sql.ErrNoRows {
		reportError(w, http.StatusNotFound, fmt.Sprintf("pull request !%s not found", id))
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(pr)
	if err != nil {
		handleError(w, err)
		return
	}
}
