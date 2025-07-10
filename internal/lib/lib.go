package lib

import "os"

func GetDefaultEnv(envVar string, value string) string {
	envVar = os.Getenv(envVar)	
	if envVar == "" {
		envVar = value
	}
	return envVar
}
