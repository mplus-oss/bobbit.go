package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mplus.software/oss/bobbit.go/payload"
)

func (d *DaemonStruct) GenerateJobDataFilename(p payload.JobRequestMetadata, extFile string) string {
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
