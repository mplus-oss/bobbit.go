package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mplus-oss/bobbit.go/config"
	"github.com/mplus-oss/bobbit.go/payload"
)

// GenerateJobDataFilename generates a filename for job-related data based on the provided configuration, job metadata, and file type.
func GenerateJobDataFilename(c config.BobbitConfig, p payload.JobDetailMetadata, extFile DaemonFileTypeEnum) string {
	return filepath.Join(
		c.DataDir,
		fmt.Sprintf("%d-%s-%s.%s", p.Timestamp.UnixMilli(), p.ID, p.JobName, extFile),
	)
}

// SplitFilenameFromExtfile splits the filename from its extension.
// It returns the filename without the extension.
func SplitFilenameFromExtfile(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// ParseJobDataFilename parses the job detail metadata from a given filename.
// It returns a JobDetailMetadata struct and an error if parsing fails.
func ParseJobDataFilename(filename string) (p payload.JobDetailMetadata, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Failed to parsing filename, format not supported: filename=%v err=%v", filename, r)
		}
	}()

	file := SplitFilenameFromExtfile(filename)
	fileSplit := strings.Split(file, "-")

	p.ID = strings.Join(fileSplit[1:2], "-")
	p.JobName = strings.Join(fileSplit[2:], "-")

	time64, err := strconv.ParseInt(fileSplit[0], 10, 64)
	if err != nil {
		return p, err
	}
	p.Timestamp = time.UnixMilli(time64)

	return p, nil
}

// FindJobDataFilename searches for a job data filename based on the provided configuration and search metadata.
// It returns the found JobDetailMetadata and an error if an issue occurs during directory reading or parsing.
func FindJobDataFilename(c config.BobbitConfig, s payload.JobSearchMetadata) (p payload.JobDetailMetadata, err error) {
	files, err := os.ReadDir(c.DataDir)
	if err != nil {
		return p, err
	}

	for _, file := range files {
		if !strings.Contains(file.Name(), s.Search) {
			continue
		}

		jobParser, err := ParseJobDataFilename(file.Name())
		if err != nil {
			return p, err
		}
		if jobParser.ID == s.Search || jobParser.JobName == s.Search {
			p = jobParser
		}
	}

	return p, nil
}

// ParseExitCode reads the exit code for a job from a file and updates the job's status accordingly.
// It returns an error if reading or parsing the exit code fails.
func ParseExitCode(c config.BobbitConfig, job *payload.JobResponse) error {
	exitCodeBytes, err := os.ReadFile(GenerateJobDataFilename(c, job.JobDetailMetadata, DAEMON_EXITCODE))
	if err != nil {
		lockfile := GenerateJobDataFilename(c, job.JobDetailMetadata, DAEMON_LOCKFILE)
		job.ExitCode = -1

		if _, err := os.Stat(lockfile); err == nil {
			job.Status = payload.JOB_RUNNING
		} else {
			job.Status = payload.JOB_NOT_RUNNING
		}
		return nil
	}

	code, err := strconv.Atoi(strings.TrimSpace(string(exitCodeBytes)))
	if err != nil {
		return err
	}

	job.ExitCode = code
	switch job.ExitCode {
	case 0:
		job.Status = payload.JOB_FINISH
	default:
		job.Status = payload.JOB_FAILED
	}

	return nil
}
