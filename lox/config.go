package lox

import (
	"fmt"
	"os"
	"runtime/pprof"
)

const (
	EnvNoPanicReturn  = "NO_PANIC_RETURN"
	EnvProfileEnabled = "COLLECT_PROFILE"
	EnvDebug          = "LOG_LEVEL"
)

func ShouldReturn(res any) bool {
	isRetVal := func() bool {
		_, ok := res.(ReturnValue)
		return ok
	}
	return res != nil && isRetVal()
}

//nolint:nonamedreturns // named return for readability
func CollectProfile() (stop func()) {
	path, ok := os.LookupEnv(EnvProfileEnabled)
	if !ok {
		return func() {}
	}

	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating profile file %s: %v\n", path, err)
		os.Exit(1)
	}

	_ = pprof.StartCPUProfile(f)
	return func() {
		pprof.StopCPUProfile()
	}
}
