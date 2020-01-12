package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type ErrorResponse struct {
	Errors []string `json:"errors"`
}

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)

const passwordLenMin = 8

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

	http.HandleFunc("/api/v1/signup", func(w http.ResponseWriter, req *http.Request) {
		username, email, password, errs, intErr := validateSignUp(req, userExistsStmt)
		if intErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{errs})
			return
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, err = newUserStmt.Exec(username, email, hash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func validateSignUp(
	req *http.Request, userExists *sql.Stmt) (username, email, password string, errs []string, intErr error) {
	errs = make([]string, 0, 3)

	err := req.ParseForm()
	if err != nil {
		errs = append(errs, "invalid form input")
		return
	}

	username = req.FormValue("username")
	if username == "" {
		errs = append(errs, "missing username")
	} else {
		var res *sql.Rows
		res, intErr = userExists.Query(username)
		if intErr != nil {
			return
		}

		var userExists bool
		res.Next()
		intErr = res.Scan(&userExists)
		if intErr != nil {
			return
		}
		if userExists {
			errs = append(errs, "username already exists")
		}
	}

	email = req.FormValue("email")
	if email == "" {
		errs = append(errs, "missing email")
	} else if !emailPattern.MatchString(email) {
		errs = append(errs, "invalid email")
	}

	password = req.FormValue("password")
	if password == "" {
		errs = append(errs, "missing password")
	} else if len(password) < passwordLenMin {
		errs = append(errs, fmt.Sprintf("password must be at least %d characters long", passwordLenMin))
	}

	return
}
