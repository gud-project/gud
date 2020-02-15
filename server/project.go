package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"gitlab.com/magsh-2019/2/gud/gud"
)

const projectsPath = "projects"
const dirPerm = 0755

func createProject(w http.ResponseWriter, r *http.Request) {
	dir, msg, err := createProjectDir(r)
	if err != nil {
		handleError(w, err)
		return
	}
	if msg != "" {
		reportError(w, http.StatusBadRequest, msg)
		return
	}

	_, err = gud.Start(dir)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func importProject(w http.ResponseWriter, r *http.Request) {
	dir, msg, err := createProjectDir(r)
	if err != nil {
		handleError(w, err)
		return
	}
	if msg != "" {
		reportError(w, http.StatusBadRequest, msg)
		return
	}

	project, err := gud.StartHeadless(dir)
	if err != nil {
		handleError(w, err)
		return
	}

	err = project.PullBranch(gud.FirstBranchName, r.Body, r.Header.Get("Content-Type"))
	if err != nil {
		if inputErr, ok := err.(gud.InputError); ok {
			reportError(w, http.StatusBadRequest, inputErr.Error())
		} else {
			handleError(w, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func projectBranch(w http.ResponseWriter, r *http.Request) {
	project, err := gud.Load(contextProjectPath(r.Context()))
	if err != nil {
		handleError(w, err)
		return
	}

	hash, err := project.GetBranch(mux.Vars(r)["branch"])
	if err != nil {
		handleError(w, err)
		return
	}
	if hash == nil {
		reportError(w, http.StatusNotFound, "branch not found")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(hash[:])
}

func pushProject(w http.ResponseWriter, r *http.Request) {
	branchArr := r.URL.Query()["branch"]
	if len(branchArr) == 0 || branchArr[0] == "" {
		reportError(w, http.StatusBadRequest, "missing branch")
		return
	}

	project, err := gud.Load(contextProjectPath(r.Context()))
	if err != nil {
		handleError(w, err)
		return
	}

	err = project.PullBranch(branchArr[0], r.Body, r.Header.Get("Content-Type"))
	if err != nil {
		if inputErr, ok := err.(gud.InputError); ok {
			reportError(w, http.StatusBadRequest, inputErr.Error())
		} else {
			handleError(w, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func pullProject(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	branches := query["branch"]
	if len(branches) == 0 || branches[0] == "" {
		reportError(w, http.StatusBadRequest, "missing branch")
		return
	}

	var start *gud.ObjectHash
	var startHash gud.ObjectHash
	starts := query["start"]
	if len(starts) != 0 {
		n, err := hex.Decode(startHash[:], []byte(starts[0]))
		if err != nil || n != len(startHash) {
			reportError(w, http.StatusBadRequest, "invalid start hash")
			return
		}

		start = &startHash
	}

	project, err := gud.Load(contextProjectPath(r.Context()))
	if err != nil {
		handleError(w, err)
		return
	}

	boundary, err := project.PushBranch(w, branches[0], start)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", boundary))
}

func createProjectDir(r *http.Request) (dir string, errMsg string, err error) {
	nameArr := r.URL.Query()["name"]

	if len(nameArr) == 0 {
		return "", "missing project name", nil
	}

	name := nameArr[0]
	if !namePattern.MatchString(name) {
		return "", "invalid project name", nil
	}

	userId := r.Context().Value(KeyUserId).(int)

	projectExists, err := checkExists(projectExistsStmt, name, userId)
	if err != nil {
		return
	}
	if projectExists {
		return "", "project already exists", nil
	}

	res, err := createProjectStmt.Exec(name, userId)
	if err != nil {
		return
	}

	projectId, err := res.LastInsertId()
	if err != nil {
		return
	}

	dir = projectPath(userId, int(projectId))
	err = os.Mkdir(dir, dirPerm)

	return
}

func verifyProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		username := vars["user"]
		projectName := vars["project"]

		userId := r.Context().Value(KeyUserId).(uint)
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

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), KeyProjectId, projectId)))
	})
}

func contextProjectPath(ctx context.Context) string {
	return projectPath(ctx.Value(KeyUserId).(int), ctx.Value(KeyProjectId).(int))
}

func projectPath(userId, projectId int) string {
	return filepath.Join(projectsPath, strconv.Itoa(userId), strconv.Itoa(projectId))
}
