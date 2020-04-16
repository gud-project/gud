package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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

func init() {
	err := os.MkdirAll(projectsPath, 0755)
	if err != nil {
		panic(err)
	}
}

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

	_, err = gud.StartHeadless(dir)
	if err != nil {
		handleError(w, err)
		return
	}
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

	pullProjectFrom(w, r, *project, branchArr[0])
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

	var buf bytes.Buffer
	boundary, err := project.PushBranch(&buf, branches[0], start)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", boundary))
	_, err = buf.WriteTo(w)
	if err != nil {
		handleError(w, err)
	}
}

func inviteMember(w http.ResponseWriter, r *http.Request) {
	var req gud.InviteMemberRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		reportError(w, http.StatusBadRequest, err.Error())
	}

	var memberId int
	var _password string
	err = userByNameStmt.QueryRow(req.Name).Scan(&memberId, &_password)
	if err == sql.ErrNoRows {
		reportError(w, http.StatusBadRequest, fmt.Sprintf("user %s not found", req.Name))
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}

	var ownerId int
	projectId := r.Context().Value(KeyProjectId).(int)
	err = getProjectStmt.QueryRow(projectId).Scan(&ownerId)
	if err != nil {
		handleError(w, err)
		return
	}

	isMember, err := checkExists(hasMemberStmt, memberId, projectId)
	if err != nil {
		handleError(w, err)
		return
	}
	if isMember {
		reportError(w, http.StatusBadRequest, fmt.Sprintf("user %s is already a member", req.Name))
	}

	_, err = inviteMemberStmt.Exec(memberId, projectId)
	if err != nil {
		handleError(w, err)
		return
	}
}

func projectBranches(w http.ResponseWriter, r *http.Request) {
	p, err := gud.Load(contextProjectPath(r.Context()))
	if err != nil {
		handleError(w, err)
		return
	}

	branches := make(map[string]string)
	err = p.ListBranches(func(branch string) error {
		hash, err := p.GetBranch(branch)
		if err != nil {
			return err
		}

		branches[branch] = hash.String()
		return nil
	})
	if err != nil {
		handleError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(branches)
}

func createProjectDir(r *http.Request) (dir string, errMsg string, err error) {
	var req gud.CreateProjectRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return
	}

	name := req.Name
	if name == "" {
		return "", "missing project name", nil
	}

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

	projectId, err := execReturningId(createProjectStmt, name, userId)
	if err != nil {
		return
	}

	dir = projectPath(userId, projectId)
	err = os.Mkdir(dir, dirPerm)

	return
}

func pullProjectFrom(w http.ResponseWriter, r *http.Request, project gud.Project, branch string) {
	var username string
	err := getUserStmt.QueryRow(r.Context().Value(KeyUserId)).Scan(&username)
	if err != nil {
		handleError(w, err)
		return
	}

	hash, err := project.PullBranchFrom(branch, r.Body, r.Header.Get("Content-Type"), username)
	if err != nil {
		if inputErr, ok := err.(gud.InputError); ok {
			reportError(w, http.StatusBadRequest, inputErr.Error())
		} else {
			handleError(w, err)
		}
		return
	}

	err = createJob(r.Context().Value(KeyProjectId).(int), project, *hash)
	if err != nil {
		handleError(w, err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func verifyProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ownerName := vars["user"]
		projectName := vars["project"]
		userId := r.Context().Value(KeyUserId).(int)

		var projectId, ownerId int
		err := projectByNameStmt.QueryRow(ownerName, projectName).Scan(&projectId, &ownerId)
		if err == sql.ErrNoRows {
			reportError(w, http.StatusNotFound, "project not found")
			return
		}
		if err != nil {
			handleError(w, err)
			return
		}

		if ownerId != userId {
			isMember, err := checkExists(hasMemberStmt, userId, projectId)
			if err != nil {
				handleError(w, err)
				return
			}
			if !isMember {
				reportError(w, http.StatusUnauthorized, "you do not have access to this project")
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(context.WithValue(r.Context(),
			KeyProjectId, projectId),
			KeySelectedUserId, ownerId)))
	})
}

func contextProjectPath(ctx context.Context) string {
	return projectPath(ctx.Value(KeySelectedUserId).(int), ctx.Value(KeyProjectId).(int))
}

func projectPath(userId, projectId int) string {
	return filepath.Join(projectsPath, strconv.Itoa(userId), strconv.Itoa(projectId))
}
