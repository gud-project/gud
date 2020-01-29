package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
)

var illegalNameChars = []int{' ', '@'}

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

type ErrorResponse struct {
	Error string `json:"error"`
}

type MultiErrorResponse struct {
	Errors []string `json:"errors"`
}

func main() {
	defer closeDB()
	sort.Ints(illegalNameChars)

	api := mux.NewRouter()
	api.HandleFunc("/signup", signUp).Methods(http.MethodPost)
	api.HandleFunc("/login", login).Methods(http.MethodPost)
	api.Handle("/logout", verifySession(http.HandlerFunc(logout))).Methods(http.MethodPost)

	projects := api.PathPrefix("/projects").Subrouter()
	projects.Use(verifySession)

	projects.HandleFunc("/create", createProject).Methods(http.MethodPost)
	projects.HandleFunc("/import", importProject).Methods(http.MethodPost)

	project := api.PathPrefix("/project/{user}/{project}").Subrouter()
	project.Use(verifySession, verifyProject)
	project.HandleFunc("/branch/{branch}", projectBranch).Methods(http.MethodGet)
	project.HandleFunc("/push", pushProject).Methods(http.MethodPost)
	project.HandleFunc("/pull", pullProject).Methods(http.MethodGet)

	http.Handle("/api/v1", api)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func validName(name string) bool {
	for _, r := range name {
		if !strconv.IsPrint(r) {
			return false
		}

		ind := sort.SearchInts(illegalNameChars, int(r))
		if ind < len(illegalNameChars) && int(r) == illegalNameChars[ind] {
			return false
		}
	}

	return true
}

func reportError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{message})
}

func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Println(err)
	log.Writer()
}
