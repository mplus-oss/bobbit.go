package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mplus-oss/bobbit.go/internal/dblib"
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
	Metadata  string    `db:"metadata"` // JSON string representation of PayloadRegularMetadata
	PID       int       `db:"pid"`
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

	// MetadataFilter allows filtering jobs based on their metadata.
	// Keys are metadata field names, and values are the desired values.
	MetadataFilter map[string]string

	// GeneralKeywordSearch applies to JobFilter.ID OR JobFilter.Keyword
	GeneralKeywordSearch string

	// HideCommand prevents the command from being exposed in the job response.
	HideCommand bool

	DBGetFilter
}

// buildSelectQuery generates the raw SQL query for the jobs table.
func (j *JobModel) buildSelectQuery(filter *JobFilter) string {
	commandCol := "command"
	if filter.HideCommand {
		commandCol = "'[]' as command"
	}

	return fmt.Sprintf(`
		SELECT
			id, job_name, %s, status, exit_code,
			metadata, pid, created_at, updated_at
		FROM jobs
	`, commandCol)
}

// applyCriteria appends the WHERE logic to the query and returns the updated query and args.
func (j *JobModel) applyCriteria(query string, args []any, filter *JobFilter) (string, []any) {
	whereClauses := []string{}
	whereArgs := []any{}

	if filter.ID != "" {
		whereClauses = append(whereClauses, "id LIKE ?")
		whereArgs = append(whereArgs, filter.ID+"%")
	}
	if filter.Keyword != "" {
		whereClauses = append(whereClauses, "job_name LIKE ?")
		whereArgs = append(whereArgs, filter.Keyword)
	}

	if filter.GeneralKeywordSearch != "" {
		whereClauses = append(whereClauses, "(id LIKE ? OR job_name LIKE ?)")
		whereArgs = append(whereArgs, filter.GeneralKeywordSearch+"%", filter.GeneralKeywordSearch)
	}

	// Filter for active jobs
	if filter.ActiveOnly {
		whereClauses = append(whereClauses, "status = ?")
		whereArgs = append(whereArgs, payload.JOB_RUNNING)
	}

	// Filter for finished jobs
	if filter.FinishOnly {
		whereClauses = append(whereClauses, "(status = ? OR status = ?)")
		whereArgs = append(whereArgs, payload.JOB_FINISH, payload.JOB_FAILED)
	}

	// Add metadata filtering
	if len(filter.MetadataFilter) > 0 {
		if j.SupportsJSONFunctions {
			for k, v := range filter.MetadataFilter {
				sanitizedK := dblib.SanitizeJsonKey(k) // Trying to sanitize key...
				if sanitizedK == "" {
					log.Printf("[WARNING] Skipping metadata filter for empty or invalid key after sanitization: original='%s'", k)
					continue
				}

				// Using `json_extract(col, $.key)` keyword
				whereClauses = append(whereClauses, fmt.Sprintf("json_extract(metadata, '$.%s') LIKE ?", k))
				whereArgs = append(whereArgs, v)
			}
		} else {
			for k, v := range filter.MetadataFilter {
				// Fallback to LIKE for old SQLite version.
				// Example: v:"test%" = %"key":"test%"%
				whereClauses = append(whereClauses, "metadata LIKE ?")
				whereArgs = append(whereArgs, fmt.Sprintf("%%\"%s\":\"%s\"%%", k, v))
			}
		}
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + join(whereClauses, " AND ")
		args = append(args, whereArgs...)
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

// Get fetch the job data from the table and return it as a array of JobModel.
func (j *JobModel) Get(filter *JobFilter) ([]*JobModel, error) {
	query := j.buildSelectQuery(filter)
	args := []any{}
	query, args = j.applyCriteria(query, args, filter)

	var jobs []*JobModel
	if err := j.DB.Select(&jobs, query, args...); err != nil {
		return nil, err
	}

	for i := range jobs {
		jobs[i].BaseModel = j.BaseModel
	}

	return jobs, nil
}

// WaitJob will waiting the filter until it finished.
func (j *JobModel) WaitJob(ctx context.Context, cancel context.CancelFunc, filter *JobFilter) (*JobModel, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var finalJob *JobModel
	var finalErr error
	jobFound := false

	for {
		select {
		case <-ctx.Done():
			return finalJob, finalErr

		case <-ticker.C:
			query := j.buildSelectQuery(filter)
			args := []any{}
			filter.Limit = 1
			query, args = j.applyCriteria(query, args, filter)

			var p JobModel
			if err := j.DB.GetContext(ctx, &p, query, args...); err != nil {
				if err == sql.ErrNoRows {
					// If job was previously found but now not found,
					// it means the job record was deleted or is no longer matching the filter.
					if jobFound {
						finalErr = fmt.Errorf("Job was found but is no longer available")
						cancel()
					}
					continue
				}

				finalErr = err
				cancel()
				continue
			}

			// Mark that we've found the job at least once
			jobFound = true

			p.BaseModel = j.BaseModel
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
// NOTE: Ensure that 'Command', 'PID', and 'Metadata' fields (strings) are updated
// with the latest JSON content before calling this method.
func (j *JobModel) Update() error {
	query := `
		UPDATE jobs 
		SET 
			job_name = :job_name,
			command = :command,
			status = :status,
			exit_code = :exit_code,
			pid = :pid,
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

	supportsJSON, err := dblib.CheckSQLiteJSONFunctions(db)
	if err != nil {
		log.Printf("[WARNING] Failed to check SQLite JSON function support: %v", err)
		supportsJSON = false
	}

	return &JobModel{
		ID:        job.ID,
		JobName:   job.JobName,
		Command:   string(cmdBytes),
		Status:    int(job.Status),
		ExitCode:  job.ExitCode,
		Metadata:  metaString,
		BaseModel: BaseModel{DB: db, SupportsJSONFunctions: supportsJSON},
	}, nil
}
