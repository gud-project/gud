package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/magsh-2019/2/gud/gud"
)

func main() {
	defer closeDB()

	root := mux.NewRouter()
	api := root.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/signup", signUp)
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

	log.Fatal(http.ListenAndServe(":8080", root))
}

// reportError reports an error to the client side (e.g. invalid input, unauthorized)
func reportError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(gud.ErrorResponse{Error: message})
}

// handleError handles a server-side error without reporting to the user (e.g. SQL errors)
func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Println(err)
}
