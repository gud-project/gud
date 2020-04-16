package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/gorilla/sessions"
	"gitlab.com/magsh-2019/2/gud/gud"
	"golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)
var namePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type ContextKey int
const (
	KeyUserId ContextKey = iota
	KeySelectedUserId
	KeyProjectId
)

const passwordLenMin = 8
const sessionAge = 60 * 60 * 24 * 7

func signUp(w http.ResponseWriter, r *http.Request) {
	req, errs, intErr := validateSignUp(r)
	if intErr != nil {
		handleError(w, intErr)
		return
	}
	if errs != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(gud.MultiErrorResponse{Errors: errs})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	id, err := execReturningId(newUserStmt, req.Username, req.Email, hash)
	if err != nil {
		handleError(w, err)
		return
	}

	err = os.Mkdir(filepath.Join(projectsPath, strconv.Itoa(id)), dirPerm)
	if err != nil {
		handleError(w, err)
		return
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	var req gud.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		reportError(w, http.StatusBadRequest, err.Error())
		return
	}

	row := userByNameStmt.QueryRow(req.Username)
	var id int
	var hash []byte

	err = row.Scan(&id, &hash)
	if err == sql.ErrNoRows {
		reportError(w, http.StatusUnauthorized, "user not found")
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(req.Password))
	if err != nil {
		if err != bcrypt.ErrMismatchedHashAndPassword {
			handleError(w, err)
			return
		}
		reportError(w, http.StatusUnauthorized, "user/password combination does not match")
		return
	}

	err = createSession(w, r, id, req.Remember)
	if err != nil {
		handleError(w, err)
		return
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "session")
	if sess.IsNew {
		return
	}

	_, inProd := os.LookupEnv("PROD")
	sess.Options = &sessions.Options{
		MaxAge:   -1, // delete
		Secure:   inProd,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	err := sess.Save(r, w)
	if err != nil {
		handleError(w, err)
		return
	}
}

func validateSignUp(r *http.Request) (*gud.SignUpRequest, []string, error) {
	const maxErrs = 3
	errs := make([]string, 0, maxErrs)

	var req gud.SignUpRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errs = append(errs, err.Error())
		return nil, errs, nil
	}

	if req.Username == "" {
		errs = append(errs, "missing username")
	} else if !namePattern.MatchString(req.Username) {
		errs = append(errs, "invalid username")
	} else {
		userExists, intErr := checkExists(userExistsStmt, req.Username)
		if intErr != nil {
			return nil, nil, intErr
		}
		if userExists {
			errs = append(errs, "username already exists")
		}
	}

	if !emailPattern.MatchString(req.Email) {
		errs = append(errs, "invalid email")
	}

	if len(req.Password) < passwordLenMin {
		errs = append(errs, fmt.Sprintf("password must be at least %d characters long", passwordLenMin))
	}

	if len(errs) > 0 {
		return nil, errs, nil
	}
	return &req, nil, nil
}

func createSession(w http.ResponseWriter, r *http.Request, id int, remember bool) error {
	sess, _ := store.Get(r, "session")
	sess.Values["id"] = id

	_, inProd := os.LookupEnv("PROD")
	options := sessions.Options{
		MaxAge:   0, // deletes when session ends
		Secure:   inProd,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	if remember {
		options.MaxAge = sessionAge
	}

	sess.Options = &options
	return sess.Save(r, w)
}

func verifySession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, _ := store.Get(r, "session")

		if sess.IsNew {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			id := sess.Values["id"].(int)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), KeyUserId, id)))
		}
	})
}
