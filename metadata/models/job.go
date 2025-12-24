package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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

type JobFilter struct {
	// ActiveOnly filters results to include only active jobs.
	// Cannot be combined with FinishOnly.
	ActiveOnly bool

	// FinishOnly filters results to include only finished jobs whether job is success or failed.
	// Cannot be combined with ActiveOnly.
	FinishOnly bool

	DBGetFilter
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

func (j *JobModel) Delete() {
	_, err := j.DB.NamedExec("DELETE FROM jobs WHERE id = :id", j)
	if err != nil {
		log.Printf("[WARNING] Failed to delete %s job: %v", j.ID, err)
	}
}

// applyCriteria appends the WHERE logic to the query and returns the updated query and args.
func (j *JobModel) applyCriteria(query string, args []any, filter *JobFilter) (string, []any) {
	if filter.ID != "" || filter.Keyword != "" {
		query += " WHERE id LIKE ? OR job_name LIKE ?"
		args = append(args, filter.ID+"%")
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

	return query, args
}

// Get fetch the job data from the table and return it as a array of JobModel.
func (j *JobModel) Get(filter *JobFilter) ([]*JobModel, error) {
	query := "SELECT * FROM jobs"
	args := []any{}
	query, args = j.applyCriteria(query, args, filter)

	var jobs []*JobModel
	if err := j.DB.Select(&jobs, query, args...); err != nil {
		return nil, err
	}

	return jobs, nil
}

// WaitJob will waiting the filter until it finished.
func (j *JobModel) WaitJob(ctx context.Context, cancel context.CancelFunc, filter *JobFilter) (*JobModel, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var finalJob *JobModel
	var finalErr error

	for {
		select {
		case <-ctx.Done():
			return finalJob, finalErr

		case <-ticker.C:
			query := "SELECT * FROM jobs"
			args := []any{}
			filter.Limit = 1
			query, args = j.applyCriteria(query, args, filter)

			var p JobModel
			if err := j.DB.GetContext(ctx, &p, query, args...); err != nil {
				if err == sql.ErrNoRows {
					continue
				}

				finalErr = err
				cancel()
				continue
			}

			if p.Status != int(payload.JOB_RUNNING) {
				finalJob = &p
				cancel()
			}
		}
	}
}

// Get fetch the job data from the table and return it as a int of JobModel size.
func (j *JobModel) Count(filter *JobFilter) (int, error) {
	query := "SELECT COUNT(*) FROM jobs"
	args := []any{}
	query, args = j.applyCriteria(query, args, filter)

	var count int
	if err := j.DB.Get(&count, query, args...); err != nil {
		return 0, err
	}

	return count, nil
}

// Update persists the current state of the JobModel to the database.
//
// NOTE: Ensure that 'Command' and 'Metadata' fields (strings) are updated
// with the latest JSON content before calling this method.
func (j *JobModel) Update() error {
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

	_, err := j.DB.NamedExec(query, j)
	if err != nil {
		return fmt.Errorf("failed to update job %s: %w", j.ID, err)
	}

	return nil
}

// MarkFinished updates only the status and exit code of a job.
//
// To use this function, the required property is `JobModel.ID` and `JobModel.ExitCode`.
func (j *JobModel) MarkJobFinished() error {
	if j.ExitCode == 0 {
		j.Status = int(payload.JOB_FINISH)
	} else {
		j.Status = int(payload.JOB_FAILED)
	}

	query := `UPDATE jobs SET status = :status, exit_code = :exit_code WHERE id = :id`
	if _, err := j.DB.NamedExec(query, j); err != nil {
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
			ID:        j.ID,
			JobName:   j.JobName,
			Command:   cmd,
			Metadata:  meta,
			CreatedAt: j.CreatedAt,
			UpdatedAt: j.UpdatedAt,
		},
	}, nil
}

// BulkToPayload converts the bulk of raw database model back into a bulk of JobResponse struct.
func (j *JobModel) BulkToPayload(jm []*JobModel) (p []*payload.JobResponse, err error) {
	if len(jm) == 0 {
		return []*payload.JobResponse{}, nil
	}

	p = make([]*payload.JobResponse, 0, len(jm))
	for _, v := range jm {
		if v == nil {
			continue
		}

		pv, err := v.ToPayload()
		if err != nil {
			return nil, err
		}

		p = append(p, pv)
	}

	return p, nil
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
