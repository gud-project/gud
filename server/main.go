package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type SignUpRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type CreateProjectRequest struct {
	Name string `json:"name"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MultiErrorResponse struct {
	Errors []string `json:"errors"`
}

const projectsPath = "projects"
const dirPerm = 0755

func main() {
	defer closeDB()

	api := mux.NewRouter()
	api.HandleFunc("/signup", signUp).Methods("POST")
	api.HandleFunc("/login", login).Methods("POST")
	api.Handle("/logout", verifySession(http.HandlerFunc(logout))).Methods("POST")

	projects := api.PathPrefix("/projects").Subrouter()
	projects.Use(verifySession)

	projects.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		var req CreateProjectRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{"missing name"})
			return
		}

		name := req.Name
		userId := context.Get(r, KeyUserID).(int)

		row, err := projectExistsStmt.Query(name, userId)
		if err != nil {
			handleError(w, err)
			return
		}
		defer row.Close()

		var projectExists bool
		row.Next()
		err = row.Scan(&projectExists)
		if err != nil {
			handleError(w, err)
			return
		}
		if projectExists {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{"project already exists"})
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

		err = os.Mkdir(filepath.Join(projectsPath, strconv.Itoa(userId), strconv.Itoa(int(projectId))), dirPerm)
		if err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	http.Handle("/api/v1", api)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Print(err)
}
