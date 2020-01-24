package main

import (
	"encoding/json"
	"log"
	"net/http"

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

func main() {
	defer closeDB()

	api := mux.NewRouter()
	api.HandleFunc("/signup", signUp).Methods(http.MethodPost)
	api.HandleFunc("/login", login).Methods(http.MethodPost)
	api.Handle("/logout", verifySession(http.HandlerFunc(logout))).Methods(http.MethodPost)

	projects := api.PathPrefix("/projects").Subrouter()
	projects.Use(verifySession)

	projects.HandleFunc("/create", createProject).Methods(http.MethodPost)

	project := api.PathPrefix("/project/{user}/{project}").Subrouter()
	project.Use(verifySession, verifyProject)

	http.Handle("/api/v1", api)
	log.Fatal(http.ListenAndServe(":8080", nil))
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
