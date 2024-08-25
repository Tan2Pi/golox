package lox

import (
	"fmt"
	"os"
)

func Debug(msg string, a ...any) {
	if os.Getenv(EnvDebug) == "" {
		return
	}

	fmt.Printf(msg, a...)
}
