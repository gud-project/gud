package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func createUserRouter(route *mux.Route) *mux.Router {
	r := route.Subrouter()
	r.HandleFunc("/projects", userProjects).Methods(http.MethodGet)

	return r
}

func userProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := userProjectsStmt.Query(r.Context().Value(KeySelectedUserId))
	if err != nil {
		handleError(w, err)
		return
	}

	names := make([]string, 0)
	for projects.Next() {
		var name string
		err = projects.Scan(&name)
		if err != nil {
			handleError(w, err)
			return
		}

		names = append(names, name)
	}
	err = projects.Err()
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(names)
}

func selectSelf(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), KeySelectedUserId, r.Context().Value(KeyUserId))))
	})
}

func selectUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := mux.Vars(r)["user"]

		var userId int
		var _password string
		err := userByNameStmt.QueryRow(username).Scan(&userId, &_password)

		if err == sql.ErrNoRows {
			reportError(w, http.StatusNotFound, fmt.Sprintf("user %s not found", username))
		} else if err != nil {
			handleError(w, err)
		} else {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), KeySelectedUserId, userId)))
		}
	})
}
