package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gitlab.com/magsh-2019/2/gud/gud"
)

func createJob(projectId int, project gud.Project, hash gud.ObjectHash) error {
	hasDockerfile, err := project.HasFile("Dockerfile", hash)
	if err != nil {
		return err
	}
	if hasDockerfile {
		jobId, err := execReturningId(createJobStmt, projectId, hash.String())
		if err != nil {
			return err
		}

		go runJob(jobId, project, hash)
	}

	return nil
}

func runJob(id int, project gud.Project, hash gud.ObjectHash) {
	var code int
	var logs []byte
	var err error
	defer func() {
		if err != nil {
			log.Println(err)
			code = -1
			logs = []byte("internal error")
		}

		var status string
		if code == 0 {
			status = "success"
		} else {
			status = "failure"
		}
		_, _ = finishJobStmt.Exec(id, status, logs)
	}()

	var tar bytes.Buffer
	err = project.Tar(&tar, hash)
	if err != nil {
		return
	}

	code, logs, err = execJob(&tar)
	if err != nil {
		return
	}
}

func getJobs(w http.ResponseWriter, r *http.Request) {
	rows, err := getJobsStmt.Query(r.Context().Value(KeyProjectId))
	if err != nil {
		handleError(w, err)
		return
	}
	defer rows.Close()

	jobs := make([]gud.Job, 0)
	for rows.Next() {
		var job gud.Job
		err = rows.Scan(&job.Id, &job.Version, &job.Status)
		if err != nil {
			handleError(w, err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		handleError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(jobs)
	if err != nil {
		handleError(w, err)
		return
	}
}

func getJob(w http.ResponseWriter, r *http.Request) {
	jobId, err := strconv.Atoi(mux.Vars(r)["job"])
	if err != nil {
		reportError(w, http.StatusBadRequest, err.Error())
		return
	}

	var job gud.Job
	job.Id = jobId
	err = getJobStmt.QueryRow(jobId).Scan(&job.Version, &job.Status, &job.Logs)
	if err == sql.ErrNoRows {
		reportError(w, http.StatusNotFound, fmt.Sprintf("job #%d not found", jobId))
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(job)
	if err != nil {
		handleError(w, err)
		return
	}
}
