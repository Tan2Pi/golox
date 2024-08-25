package test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

var (
	expectedOutputPattern       = regexp.MustCompile(`// expect: ?(.*)`)
	expectedErrorPattern        = regexp.MustCompile(`// (Error.*)`)
	errorLinePattern            = regexp.MustCompile(`// \[((java|c) )?line (\d+)\] (Error.*)`)
	expectedRuntimeErrorPattern = regexp.MustCompile(`// expect runtime error: (.+)`)
	syntaxErrorPattern          = regexp.MustCompile(`\[.*line (\d+)\] (Error.+)`)
	stackTracePattern           = regexp.MustCompile(`\[line (\d+)\]`)
	nonTestPattern              = regexp.MustCompile(`// nontest`)

	//nolint:gochecknoglobals // static map
	skipGlob = []string{
		"loxtest/benchmark/*",
		"loxtest/limit/loop_too_large.lox",
		"loxtest/limit/no_reuse_constants.lox",
		"loxtest/limit/too_many_constants.lox",
		"loxtest/limit/too_many_locals.lox",
		"loxtest/limit/too_many_upvalues.lox",
		"loxtest/limit/stack_overflow.lox",
		// JVM doesn't correctly implement IEEE equality on boxed doubles. (not sure for Go)
		"loxtest/number/nan_equality.lox",
		// These are just for earlier chapters.
		"loxtest/scanning/*",
		"loxtest/expressions/*",
	}
)

type TestSet struct {
	a            *assert.Assertions
	t            *testing.T
	Name         string
	Path         string
	ExpectedOuts []ExpectedOutput
	// ExpectedErrors is the set of expected compile error messages.
	ExpectedErrors map[string]struct{}
	// ExpectedRuntimeError is an expected error or the empty string
	ExpectedRuntimeError string
	// RuntimeErrorLine is the line # where the error occurs
	RuntimeErrorLine int
	// ExpectedExitCode of the interpreter based on expectations.
	ExpectedExitCode int
}

type ExpectedOutput struct {
	Line   int
	Output string
}

//nolint:gochecknoglobals // prob not needed, but oh whale
var (
	tests []TestSet
)

func TestMain(m *testing.M) {
	tests = loadTests()
	_, err := os.Stat("../build/glox")
	if os.IsNotExist(err) {
		panic(fmt.Errorf("could not find glox binary: %w", err))
	}

	os.Exit(m.Run())
}

func TestLoxSuite(t *testing.T) {
	t.Parallel()
	for _, test := range tests {
		t.Run(test.Name, func(tt *testing.T) {
			tt.Parallel()
			test.a = assert.New(tt)
			test.t = tt
			outs, errLines, err := runBinary(test.Path)
			test.validateExitCode(err)
			if test.ExpectedRuntimeError != "" {
				test.validateRuntimeError(errLines)
			} else {
				test.validateCompileErrors(errLines)
			}

			test.validateOutput(outs)
		})
	}
}

func (ts *TestSet) validateExitCode(err error) {
	if ts.ExpectedExitCode == 0 {
		ts.a.NoError(err)
	} else if ts.a.Error(err) {
		exitErr := err.(*exec.ExitError) //nolint:errcheck,errorlint // cmd returns errors with this type
		ts.a.Equalf(ts.ExpectedExitCode, exitErr.ExitCode(),
			"Expected return code %v and got %v", ts.ExpectedExitCode, exitErr.ExitCode())
	}
}

func (ts *TestSet) validateRuntimeError(errorLines []string) {
	if ts.ExpectedRuntimeError == "" {
		return
	}

	if len(errorLines) < 2 {
		ts.t.Errorf("Expected runtime error '%s' and got: %v", ts.ExpectedRuntimeError, errorLines)
	}

	if ts.ExpectedRuntimeError != errorLines[0] {
		ts.t.Errorf("Expected runtime error '%s' and got: '%s'", ts.ExpectedRuntimeError, errorLines[0])
		return
	}

	stackLines := errorLines[1:]
	var match []string
	for _, line := range stackLines {
		match = stackTracePattern.FindStringSubmatch(line)
		if match != nil {
			break
		}
	}

	if match == nil {
		ts.t.Errorf("Expected stack trace and got: %v", stackLines)
		return
	}

	stackLine, err := strconv.Atoi(match[1])
	if ts.a.NoError(err) {
		ts.a.Equalf(ts.RuntimeErrorLine, stackLine,
			"Expected runtime error on line %v but was on line %v",
			ts.RuntimeErrorLine, stackLine)
	}
}

func (ts *TestSet) validateCompileErrors(errorLines []string) {
	foundErrors := map[string]struct{}{}
	unexpectedCount := 0

	// ts.t.Logf("test = %s, expectedErrors = %#v, errorLines = %#v", ts.t.Name(), ts.ExpectedErrors, errorLines)

	for _, line := range errorLines {
		// ts.t.Logf("err line: %s", line)
		match := syntaxErrorPattern.FindStringSubmatch(line)
		//nolint:nestif // not too complex
		if match != nil {
			err := fmt.Sprintf("[line %v] %s", match[1], match[2])
			if _, ok := ts.ExpectedErrors[err]; ok {
				foundErrors[err] = struct{}{}
			} else {
				if unexpectedCount < 10 {
					ts.a.Fail("Unexpected error", line)
				}
				unexpectedCount++
			}
		} else if line != "" {
			if unexpectedCount < 10 {
				ts.a.Fail("Unexpected output on stderr", line)
			}
			unexpectedCount++
		}
	}

	if unexpectedCount > 10 {
		ts.a.Fail("(truncated) more...)", unexpectedCount-10)
	}

	for e := range ts.ExpectedErrors {
		ts.a.Containsf(foundErrors, e, "Missing expected error: %s", e)
	}
}

func (ts *TestSet) validateOutput(outputLines []string) {
	if len(outputLines) > 0 && outputLines[len(outputLines)-1] == "" {
		outputLines = outputLines[:len(outputLines)-1]
	}

	index := 0
	for ; index < len(outputLines); index++ {
		line := outputLines[index]
		if index >= len(ts.ExpectedOuts) {
			ts.t.Errorf("Got output '%s' when none was expected.", line)
			continue
		}

		expected := ts.ExpectedOuts[index]
		ts.a.Equalf(expected.Output, line,
			"Expected output '%s' on line %v and got '%s'.",
			expected.Output, expected.Line, line,
		)
	}

	for index < len(ts.ExpectedOuts) {
		expected := ts.ExpectedOuts[index]
		ts.a.Failf("Missing expected output '%s' on line %v.",
			expected.Output, expected.Line)
		index++
	}
}

func runBinary(path string) ([]string, []string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("../build/glox", path)
	cmd.Env = append(os.Environ(), "GOCOVERDIR=covdata")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.Split(stdout.String(), "\n"), strings.Split(stderr.String(), "\n"), err
}

//nolint:gocognit // too complex but I'm too lazy to refactor
func loadTests() []TestSet {
	tests = []TestSet{}
	eg, ctx := errgroup.WithContext(context.Background())
	sendCh := make(chan TestSet)
	workqueue := make(chan string)
	go func() {
		for ts := range sendCh {
			tests = append(tests, ts)
		}
	}()
	workers := 100

	for range workers {
		eg.Go(func() error {
			for path := range workqueue {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					err := loadTest(path, sendCh)
					if err != nil {
						return err
					}
				}
			}
			return nil
		})
	}
	_ = filepath.WalkDir("loxtest", func(path string, d fs.DirEntry, _ error) error {
		slog.Debug("checking", slog.String("path", path))
		if d.IsDir() {
			slog.Debug("skipping", slog.String("path", path))
			return nil
		}

		for _, pattern := range skipGlob {
			match, err := filepath.Match(pattern, path)
			if err != nil {
				return fmt.Errorf("bad pattern %s for file %s", pattern, path)
			}

			if match {
				slog.Debug("skipping path", "path", path)
				return nil
			}
		}

		workqueue <- path

		return nil
	})

	close(workqueue)

	if err := eg.Wait(); err != nil {
		slog.Error("error while reading test files", "error", err)
		panic(err)
	}

	close(sendCh)

	return tests
}

func loadTest(path string, ch chan TestSet) error {
	f, err := os.Open(path)
	defer func() { _ = f.Close() }()
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	s := bufio.NewScanner(f)
	lineNum := 0

	testSet := TestSet{
		Name:           path,
		Path:           path,
		ExpectedOuts:   []ExpectedOutput{},
		ExpectedErrors: map[string]struct{}{},
	}

	for s.Scan() {
		line := s.Text()
		lineNum++
		match := nonTestPattern.FindStringSubmatch(line)
		if match != nil {
			return nil // exit early, not a test
		}

		match = expectedOutputPattern.FindStringSubmatch(line)
		if match != nil {
			testSet.ExpectedOuts = append(testSet.ExpectedOuts, ExpectedOutput{
				Line:   lineNum,
				Output: match[1],
			})
			continue
		}

		match = expectedErrorPattern.FindStringSubmatch(line)
		if match != nil {
			testSet.ExpectedErrors[fmt.Sprintf("[line %v] %s", lineNum, match[1])] = struct{}{}
			testSet.ExpectedExitCode = 65
			continue
		}

		match = errorLinePattern.FindStringSubmatch(line)
		if match != nil {
			lang := match[2]
			if lang == "" || lang == "java" {
				testSet.ExpectedErrors[fmt.Sprintf("[line %s] %s", match[3], match[4])] = struct{}{}
				testSet.ExpectedExitCode = 65
			}
			continue
		}

		match = expectedRuntimeErrorPattern.FindStringSubmatch(line)
		if match != nil {
			testSet.RuntimeErrorLine = lineNum
			testSet.ExpectedRuntimeError = match[1]
			testSet.ExpectedExitCode = 70
		}

		if len(testSet.ExpectedErrors) > 0 && testSet.ExpectedRuntimeError != "" {
			slog.Error("Invalid test file", "path", path)
			slog.Error("Cannot expect both compile and runtime errors")
			return nil
		}
	}

	slog.Debug("Adding test to set", "path", path)
	ch <- testSet

	return nil
}
