package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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

type ErrorResponse struct {
	Error string `json:"error"`
}

type MultiErrorResponse struct {
	Errors []string `json:"errors"`
}

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)

const passwordLenMin = 8
const sessionAge = 60 * 60 * 24 * 7

func main() {
	db, err := sql.Open("postgres", fmt.Sprintf("user=gud password=%s", os.Getenv("PQ_PASS")))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userExistsStmt, err := db.Prepare("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);")
	if err != nil {
		log.Fatal(err)
	}
	defer userExistsStmt.Close()
	newUserStmt, err := db.Prepare(
		"INSERT INTO users (username, email, password, created_at) VALUES ($1, $2, $3, NOW());")
	if err != nil {
		log.Fatal(err)
	}
	defer newUserStmt.Close()
	getUserStmt, err := db.Prepare("SELECT id, password FROM users WHERE username = $1")
	if err != nil {
		log.Fatal(err)
	}
	defer getUserStmt.Close()

	http.HandleFunc("/api/v1/signup", func(w http.ResponseWriter, r *http.Request) {
		req, errs, intErr := validateSignUp(r, userExistsStmt)
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
		_, err = newUserStmt.Exec(req.Username, req.Email, hash)
		if err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/v1/login", func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{err.Error()})
			return
		}

		res, err := getUserStmt.Query(req.Username)
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

		var id uint
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
	})

	http.Handle("/api/v1/logout", verifySession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})))

	http.Handle("/api/v1/projects/create", verifySession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf("user id is %d", context.Get(r, "user_id").(uint))))
	})))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func validateSignUp(r *http.Request, userExistsStmt *sql.Stmt) (*SignUpRequest, []string, error) {
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

func createSession(w http.ResponseWriter, r *http.Request, id uint, remember bool) error {
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
				context.Set(r, "user_id", id)
				next.ServeHTTP(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
	})
}

func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Print(err)
}
