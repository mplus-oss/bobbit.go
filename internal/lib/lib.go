package lib

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"time"
)

func GetDefaultEnv(envVar string, value string) string {
	envVar = os.Getenv(envVar)
	if envVar == "" {
		envVar = value
	}
	return envVar
}

func GenerateRandomHash(size int) (string, error) {
	randomBytes := make([]byte, size)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	hasher := md5.New()
	hasher.Write(randomBytes)
	fullHash := hex.EncodeToString(hasher.Sum(nil))
	return fullHash[:size], nil
}

func HumanizeDuration(duration time.Duration) string {
	if duration.Seconds() < 60.0 {
		return fmt.Sprintf("%d seconds", int64(duration.Seconds()))
	}
	if duration.Minutes() < 60.0 {
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%d minutes %d seconds", int64(duration.Minutes()), int64(remainingSeconds))
	}
	if duration.Hours() < 24.0 {
		remainingMinutes := math.Mod(duration.Minutes(), 60)
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%d hours %d minutes %d seconds",
			int64(duration.Hours()), int64(remainingMinutes), int64(remainingSeconds),
		)
	}

	remainingHours := math.Mod(duration.Hours(), 24)
	remainingMinutes := math.Mod(duration.Minutes(), 60)
	remainingSeconds := math.Mod(duration.Seconds(), 60)
	return fmt.Sprintf("%d days %d hours %d minutes %d seconds",
		int64(duration.Hours()/24), int64(remainingHours),
		int64(remainingMinutes), int64(remainingSeconds),
	)
}
