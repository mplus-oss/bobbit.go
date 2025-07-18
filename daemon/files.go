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

func GenerateJobDataFilename(c config.BobbitConfig, p payload.JobRequestMetadata, extFile DaemonFileTypeEnum) string {
	return filepath.Join(
		c.DataDir,
		fmt.Sprintf("%s-%s-%s.%s", p.Timestamp.Format(time.RFC3339Nano), p.ID, p.JobName, extFile),
	)
}

func SplitFilenameFromExtfile(filename string) string {
	var file string
	if filesplit := strings.Split(filename, "."); len(filesplit) > 2 {
		file = strings.Join(filesplit[0:2], ".")
	} else {
		file = filename
	}
	return file
}

func ParseJobDataFilename(filename string) (p payload.JobRequestMetadata, err error) {
	file := SplitFilenameFromExtfile(filename)

	p.ID = strings.Join(strings.Split(file, "-")[3:4], "-")
	p.JobName = strings.Join(strings.Split(file, "-")[4:], "-")
	timestamp, err := time.Parse(time.RFC3339Nano, strings.Join(strings.Split(file, "-")[:3], "-"))
	if err != nil {
		return p, err
	}
	p.Timestamp = timestamp

	return p, nil
}

func FindJobDataFilename(c config.BobbitConfig, s payload.JobSearchMetadata) (p payload.JobRequestMetadata, err error) {
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

func ParseExitCode(c config.BobbitConfig, job payload.JobRequestMetadata) payload.JobStatusEnum {
	exitCodeBytes, err := os.ReadFile(GenerateJobDataFilename(c, job, DAEMON_EXITCODE))
	if err != nil {
		return payload.JOB_NOT_RUNNING
	}

	code, _ := strconv.Atoi(strings.TrimSpace(string(exitCodeBytes)))
	if code == 0 {
		return payload.JOB_FINISH
	} else {
		return payload.JOB_FAILED
	}
}
