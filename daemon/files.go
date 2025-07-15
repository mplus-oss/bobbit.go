package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mplus-oss/bobbit.go/payload"
)

func (d *DaemonStruct) GenerateJobDataFilename(p payload.JobRequestMetadata, extFile DaemonFileTypeEnum) string {
	return filepath.Join(
		d.DataDir,
		fmt.Sprintf("%s-%s-%s.%s", p.Timestamp.Format(time.RFC3339Nano), p.ID, p.JobName, extFile),
	)
}

func (d *DaemonStruct) SplitFilenameFromExtfile(filename string) string {
	var file string
	if filesplit := strings.Split(filename, "."); len(filesplit) > 2 {
		file = strings.Join(filesplit[0:2], ".")
	} else {
		file = filename
	}
	return file
}

func (d *DaemonStruct) ParseJobDataFilename(filename string) (p payload.JobRequestMetadata, err error) {
	file := d.SplitFilenameFromExtfile(filename)

	p.ID = strings.Join(strings.Split(file, "-")[3:4], "-")
	p.JobName = strings.Join(strings.Split(file, "-")[4:], "-")
	timestamp, err := time.Parse(time.RFC3339Nano, strings.Join(strings.Split(file, "-")[:3], "-"))
	if err != nil {
		return p, err
	}
	p.Timestamp = timestamp

	return p, nil
}

func (d *DaemonStruct) FindJobDataFilename(s payload.JobSearchMetadata) (p payload.JobRequestMetadata, err error) {
	files, err := os.ReadDir(d.DataDir)
	if err != nil {
		return p, err
	}

	for _, file := range files {
		if !strings.Contains(file.Name(), s.Search) {
			continue
		}

		jobParser, err := d.ParseJobDataFilename(file.Name())
		if err != nil {
			return p, err
		}
		if jobParser.ID == s.Search || jobParser.JobName == s.Search {
			p = jobParser
		}
	}

	return p, nil
}

func (d *DaemonStruct) ParseExitCode(job payload.JobRequestMetadata) payload.JobStatusEnum {
	exitCodeBytes, err := os.ReadFile(d.GenerateJobDataFilename(job, DAEMON_EXITCODE))
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
