//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	mg.Deps(Tidy)
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", "build/glox", "./cmd/...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Test() error {
	mg.Deps(Build)
	fmt.Println("Running tests")
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Lint() error {
	mg.Deps(Tidy)
	fmt.Println("Linting...")
	cmd := exec.Command("golangci-lint", "run", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Manage your deps, or running package managers.
func Tidy() error {
	fmt.Println("Installing Deps...")
	cmd := exec.Command("go", "mod", "tidy")
	return cmd.Run()
}

// Clean up after yourself
func Clean() error {
	fmt.Println("Cleaning...")
	os.RemoveAll("build/glox")
	return exec.Command("go", "clean", "-testcache").Run()
}
