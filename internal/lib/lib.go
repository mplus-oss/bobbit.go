package lib

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"os"
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
