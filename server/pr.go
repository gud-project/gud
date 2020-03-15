package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.com/magsh-2019/2/gud/gud"
	"net/http"
	"strconv"
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

	createPrStmt.QueryRow(req.Title, req.Content, userId, projectId, req.From, req.To)
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

	prs := make([]gud.Pr, 0)
	for rows.Next() {
		var pr gud.Pr
		if err := rows.Scan(pr.Title, pr.Author, pr.Content, pr.Id, pr.From, pr.To); err != nil {
			handleError(w, err)
			return
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		handleError(w, err)
		return
	}

	var res gud.GetPrsResponse
	res.Prs = prs

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(res)
	if err != nil {
		handleError(w, err)
		return
	}

	_, err = w.Write(buf.Bytes())
	if err != nil {
		handleError(w, err)
		return
	}
}

func getPr(w http.ResponseWriter, r *http.Request) {
	var pr gud.Pr
	id := mux.Vars(r)["pr"]
	var err error
	pr.Id, err = strconv.Atoi(id)
	if err != nil {
		reportError(w, http.StatusBadRequest, err.Error())
	}

	err = getPrStmt.QueryRow(pr.Id).Scan(pr.Author, pr.Title, pr.Content, pr.From, pr.To)
	if err != nil {
		handleError(w, err)
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(pr)
	if err != nil {
		handleError(w, err)
		return
	}

	_, err = w.Write(buf.Bytes())
	if err != nil {
		handleError(w, err)
		return
	}
}