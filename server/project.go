package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"gitlab.com/magsh-2019/2/gud/gud"
)

const projectsPath = "projects"
const dirPerm = 0755

func createProject(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Name == "" {
		reportError(w, http.StatusBadRequest, "missing name")
		return
	}

	name := req.Name
	userId := context.Get(r, KeyUserId).(int)

	projectExists, err := checkExists(projectExistsStmt, name, userId)
	if err != nil {
		handleError(w, err)
		return
	}
	if projectExists {
		reportError(w, http.StatusBadRequest, "project already exists")
		return
	}

	res, err := createProjectStmt.Exec(name, userId)
	if err != nil {
		handleError(w, err)
		return
	}

	projectId, err := res.LastInsertId()
	if err != nil {
		handleError(w, err)
		return
	}

	dir := filepath.Join(projectsPath, strconv.Itoa(userId), strconv.Itoa(int(projectId)))
	err = os.Mkdir(dir, dirPerm)
	if err != nil {
		handleError(w, err)
		return
	}

	_, err = gud.Start(dir)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func verifyProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		username := vars["user"]
		projectName := vars["project"]

		userId := context.Get(r, KeyUserId).(uint)
		matches, err := checkExists(userIdMatchesNameStmt, userId, username)
		if err != nil {
			handleError(w, err)
			return
		}
		if !matches {
			reportError(w, http.StatusUnauthorized, "incorrect user")
			return
		}

		res := projectByNameStmt.QueryRow(projectName)
		var projectId uint
		err = res.Scan(&projectId)
		if err == sql.ErrNoRows {
			reportError(w, http.StatusNotFound, "project not found")
			return
		}
		if err != nil {
			handleError(w, err)
			return
		}

		context.Set(r, KeyProjectId, projectId)
		next.ServeHTTP(w, r)
	})
}
