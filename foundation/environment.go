package foundation

import (
	"os"
	"sync"
)

var (
	envStr  string
	envOnce sync.Once
)

func GetEnvironment() string {
	envOnce.Do(func() {
		envStr = os.Getenv("ENV_STR")
		if len(envStr) == 0 {
			envStr = "local"
		}
	})
	return envStr
}

func IsProd() bool {
	return envStr == "prod"
}
