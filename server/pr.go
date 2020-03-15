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
	var req gud.CreatePrRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		reportError(w, http.StatusBadRequest, "failed to receive pr data")
		return
	}

	_, err = createPrStmt.Exec(
		req.Title, req.Content, r.Context().Value(KeySelectedUserId), r.Context().Value(KeyProjectId), req.From, req.To)
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
	prId, err := strconv.Atoi(mux.Vars(r)["pr"])
	if err != nil {
		reportError(w, http.StatusBadRequest, err.Error())
		return
	}

	pr, err := scanPr(getPrStmt.QueryRow(prId))
	if err == sql.ErrNoRows {
		reportError(w, http.StatusNotFound, fmt.Sprintf("pull request !%d not found", prId))
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

func scanPr(row scanner) (*gud.PullRequest, error) {
	var pr gud.PullRequest
	var authorId int
	err := row.Scan(&pr.Id, &authorId, &pr.Title, &pr.Content, &pr.From, &pr.To)
	if err != nil {
		return nil, err
	}
	err = getUserStmt.QueryRow(authorId).Scan(&pr.Author)
	if err != nil {
		return nil, err
	}

	return &pr, nil
}
