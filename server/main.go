package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"gitlab.com/magsh-2019/2/gud/gud"
)

func main() {
	defer closeDB()

	api := mux.NewRouter()
	api.HandleFunc("/signup", signUp)
	api.HandleFunc("/login", login).Methods(http.MethodPost)
	api.Handle("/logout", verifySession(http.HandlerFunc(logout))).Methods(http.MethodPost)

	createUserRouter(api.PathPrefix("/me"), selectSelf)
	createUserRouter(api.PathPrefix("/user/{user}"), selectUser)

	projects := api.PathPrefix("/projects").Subrouter()
	projects.Use(verifySession)
	projects.HandleFunc("/create", createProject).Methods(http.MethodPost)
	projects.HandleFunc("/import", importProject).Methods(http.MethodPost)

	http.Handle("/", http.FileServer(http.Dir("./front/dist/")))
	http.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	if _, inProd := os.LookupEnv("PROD"); inProd {
		log.Println("server running on https")
		log.Fatal(http.ListenAndServeTLS(":443", os.Getenv("TLS_CERT"), os.Getenv("TLS_KEY"), nil))
	} else {
		log.Println("server running on 8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}

// reportError reports an error to the client side (e.g. invalid input, unauthorized)
func reportError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(gud.ErrorResponse{Error: message})
}

// handleError handles a server-side error without reporting to the user (e.g. SQL errors)
func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Println(err)
}
