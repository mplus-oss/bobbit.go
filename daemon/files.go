package daemon

import (
	"fmt"
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
	var file string
	if filesplit := strings.Split(filename, "."); len(filesplit) > 2 {
		file = strings.Join(filesplit[0:2], ".")
	} else {
		file = filename
	}

	p.ID = strings.Join(strings.Split(file, "-")[3:4], "-")
	p.JobName = strings.Join(strings.Split(file, "-")[4:], "-")
	timestamp, err := time.Parse(time.RFC3339Nano, strings.Join(strings.Split(file, "-")[:3], "-"))
	if err != nil {
		return p, err
	}
	p.Timestamp = timestamp

	return p, nil
}
