package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mplus-oss/bobbit.go/payload"
)

// JobModel represents a single row in the 'jobs' database table.
// Complex fields (such as slices or maps) are serialized and stored as JSON strings.
type JobModel struct {
	ID        string    `db:"id"`
	JobName   string    `db:"job_name"`
	Command   string    `db:"command"` // JSON string representation of []string
	Status    int       `db:"status"`  // Integer cast from JobStatusEnum
	ExitCode  int       `db:"exit_code"`
	Metadata  string    `db:"metadata"`   // JSON string representation of PayloadRegularMetadata
	CreatedAt time.Time `db:"created_at"` // Generated automatically (current_timestamp)
	UpdatedAt time.Time `db:"updated_at"` // Generated automatically (TRIGGER jobs_update_updated_at)
	BaseModel
}

// Save create new data in the table
func (j *JobModel) Save() error {
	query := `
		INSERT INTO jobs (id, job_name, command, status, exit_code, metadata)
		VALUES (:id, :job_name, :command, :status, :exit_code, :metadata)
	`
	_, err := j.DB.NamedExec(query, j)
	return err
}

// Get fetch the job data from the table and return it as a array of JobModel.
func (j *JobModel) Get(filter *DBGetFilter) ([]*JobModel, error) {
	query := "SELECT * FROM jobs"
	args := []any{}

	if filter.ID != "" || filter.Keyword != "" {
		query += " WHERE id LIKE ? OR job_name LIKE ?"
		args = append(args, "%"+filter.ID+"%")
		args = append(args, "%"+filter.Keyword+"%")
	}

	if filter.SortDesc {
		query += " ORDER BY created_at DESC"
	} else {
		query += " ORDER BY created_at ASC"
	}

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)

		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	var jobs []*JobModel
	if err := j.DB.Select(&jobs, query, args...); err != nil {
		return nil, err
	}

	return jobs, nil
}

// Update persists the current state of the JobModel to the database.
//
// NOTE: Ensure that 'Command' and 'Metadata' fields (strings) are updated
// with the latest JSON content before calling this method.
func (j *JobModel) Update(db *sqlx.DB) error {
	query := `
		UPDATE jobs 
		SET 
			job_name = :job_name,
			command = :command,
			status = :status,
			exit_code = :exit_code,
			metadata = :metadata
		WHERE id = :id
	`

	_, err := db.NamedExec(query, j)
	if err != nil {
		return fmt.Errorf("failed to update job %s: %w", j.ID, err)
	}

	return nil
}

// MarkFinished updates only the status and exit code of a job.
//
// To use this function, the required property is `JobModel.ID` and `JobModel.ExitCode`.
func (j *JobModel) MarkJobFinished() error {
	status := payload.JOB_RUNNING
	if j.ExitCode == 0 {
		status = payload.JOB_FINISH
	} else {
		status = payload.JOB_FAILED
	}

	query := `UPDATE jobs SET status = ?, exit_code = ? WHERE id = ?`
	if _, err := j.DB.Exec(query, status, j.ExitCode, j.ID); err != nil {
		return fmt.Errorf("failed to mark job %s as finished: %w", j.ID, err)
	}

	return nil
}

// ToPayload converts the raw database model back into a JobResponse struct.
// It handles the deserialization of JSON fields to ensure the data is ready for the CLI.
func (j *JobModel) ToPayload() (*payload.JobResponse, error) {
	var cmd []string
	if err := json.Unmarshal([]byte(j.Command), &cmd); err != nil {
		return nil, err
	}

	var meta payload.PayloadRegularMetadata
	if j.Metadata != "" {
		if err := json.Unmarshal([]byte(j.Metadata), &meta); err != nil {
			return nil, err
		}
	}

	return &payload.JobResponse{
		Status:   payload.JobStatusEnum(j.Status),
		ExitCode: j.ExitCode,
		JobDetailMetadata: payload.JobDetailMetadata{
			ID:       j.ID,
			JobName:  j.JobName,
			Command:  cmd,
			Metadata: meta,
		},
	}, nil
}

// NewJobModel creates a database-ready JobModel from a JobResponse object.
// It serializes complex fields into JSON strings to prepare for database insertion.
func NewJobModel(db *sqlx.DB, job payload.JobResponse) (*JobModel, error) {
	cmdBytes, err := json.Marshal(job.Command)
	if err != nil {
		return nil, err
	}

	metaString := ""
	if job.Metadata != nil {
		metaBytes, err := json.Marshal(job.Metadata)
		if err != nil {
			return nil, err
		}
		metaString = string(metaBytes)
	}

	return &JobModel{
		ID:        job.ID,
		JobName:   job.JobName,
		Command:   string(cmdBytes),
		Status:    int(job.Status),
		ExitCode:  job.ExitCode,
		Metadata:  metaString,
		BaseModel: BaseModel{DB: db},
	}, nil
}
