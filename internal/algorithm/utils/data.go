package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateTaskID() string {
	return fmt.Sprintf("task_%d_%d", time.Now().UnixNano(), rand.Int63())
}
