package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mplus-oss/bobbit.go/daemon"
	"github.com/mplus-oss/bobbit.go/internal/lib"
	"github.com/mplus-oss/bobbit.go/payload"
)

// Compatible version: v0.1.0 - v0.2.0
func handleFixLogFile() error {
	jobs := make(map[string]bool, 0)
	timeNow := time.Now()

	files, err := os.ReadDir(c.DataDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		_, err := daemon.ParseJobDataFilename(file.Name())
		if err == nil {
			log.Printf("Skipped: %v\n", file.Name())
			continue
		}

		jobName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		jobs[jobName] = true
	}

	for job, _ := range jobs {
		log.Printf("Fixing: %v\n", job)

		hash, err := lib.GenerateRandomHash(16)
		if err != nil {
			log.Println(err)
			continue
		}

		jobDetail := payload.JobDetailMetadata{
			ID:        hash,
			JobName:   job,
			Timestamp: timeNow,
		}

		checkingPath := func (ext daemon.DaemonFileTypeEnum) error {
			checkedPath := filepath.Join(c.DataDir, job+"."+string(ext))
			if _, err := os.Stat(checkedPath); !errors.Is(err, os.ErrNotExist) {
				newfilepath := daemon.GenerateJobDataFilename(c, jobDetail, ext)
				if err := os.Rename(checkedPath, newfilepath); err != nil {
					return err
				}
				log.Printf("Fixed: %s\n", newfilepath)
			}
			return nil
		}
		
		if err := checkingPath(daemon.DAEMON_LOGFILE); err != nil {
			log.Println(err)
			continue
		}
		if err := checkingPath(daemon.DAEMON_EXITCODE); err != nil {
			log.Println(err)
			continue
		}
		if err := checkingPath(daemon.DAEMON_LOCKFILE); err != nil {
			log.Println(err)
			continue
		}

		timeNow = timeNow.Add(1 * time.Second)
	}

	return nil
}
