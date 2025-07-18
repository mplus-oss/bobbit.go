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

func GenerateJobDataFilename(c config.BobbitConfig, p payload.JobDetailMetadata, extFile DaemonFileTypeEnum) string {
	return filepath.Join(
		c.DataDir,
		fmt.Sprintf("%d-%s-%s.%s", p.Timestamp.UnixMilli(), p.ID, p.JobName, extFile),
	)
}

func SplitFilenameFromExtfile(filename string) string {
	var file string
	if filesplit := strings.Split(filename, "."); len(filesplit) > 1 {
		file = strings.Join(filesplit[:len(filesplit)-1], ".")
	} else {
		file = filename
	}
	return file
}

func ParseJobDataFilename(filename string) (p payload.JobDetailMetadata, err error) {
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

func ParseExitCode(c config.BobbitConfig, job payload.JobDetailMetadata) payload.JobStatusEnum {
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
