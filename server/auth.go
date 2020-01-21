package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)

type ContextKey int

const (
	KeyUserID ContextKey = iota
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
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(MultiErrorResponse{errs})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	res, err := newUserStmt.Exec(req.Username, req.Email, hash)
	if err != nil {
		handleError(w, err)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		handleError(w, err)
	}

	err = os.Mkdir(filepath.Join(projectsPath, strconv.Itoa(int(id))), dirPerm)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{err.Error()})
		return
	}

	res, err := userByNameStmt.Query(req.Username)
	if err != nil {
		handleError(w, err)
		return
	}
	defer res.Close()

	if !res.Next() {
		err = res.Err()
		if err != nil {
			handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{"user not found"})
		return
	}

	var id int
	var hash []byte
	err = res.Scan(&id, &hash)
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
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{"user/password combination does not match"})
		return
	}

	err = createSession(w, r, id, req.Remember)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "session")
	if sess.IsNew {
		return
	}

	sess.Options = &sessions.Options{
		MaxAge:   -1, // delete
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	err := sess.Save(r, w)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func validateSignUp(r *http.Request) (*SignUpRequest, []string, error) {
	const maxErrs = 3
	errs := make([]string, 0, maxErrs)

	var req SignUpRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errs = append(errs, err.Error())
		return nil, errs, nil
	}

	if req.Username == "" {
		errs = append(errs, "missing username")
	} else if strings.ContainsRune(req.Username, '@') {
		errs = append(errs, "username cannot contain @")
	} else {
		var res *sql.Rows
		res, intErr := userExistsStmt.Query(req.Username)
		if intErr != nil {
			return nil, nil, intErr
		}
		defer res.Close()

		var userExists bool
		res.Next()
		intErr = res.Scan(&userExists)
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

	options := sessions.Options{
		MaxAge:   0, // deletes when session ends
		Secure:   true,
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
		if !sess.IsNew {
			id, ok := sess.Values["id"].(uint)
			if ok {
				context.Set(r, KeyUserID, id)
				next.ServeHTTP(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
	})
}
