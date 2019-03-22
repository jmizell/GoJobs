package jobs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

const (
	red  = 31
	blue = 34
)

// LogFilename is the path to write LogEntry messages
var LogFilename = ""

// exitCode will be set to the highest exit value any job exits with
var ExitCode int

// JobSpec defines a job as the shell command,
// and it's environment
type JobSpec struct {

	// Tag is the string that will identify a job in LogEntry
	Tag string `json:"tag"`

	// Command is the shell command, including all parameters to run
	Command string `json:"command"`

	// Shell is the shell environment to pass the command to
	Shell string `json:"shell"`

	// Dir is the working directory to run the command in
	Dir string `json:"dir"`

	// Env is the environment variables that will be set for the job
	Env []string `json:"env"`

	job      *exec.Cmd     `json:"-"`
	exitCode int           `json:"-"`
	duration time.Duration `json:"-"`
}

// Config is a list of JobSpecs to run
type Config []JobSpec

// LogEntry stores the line output of a job for
// both standard error, and standard out combined.
type LogEntry struct {

	// Tag is the Tag value specified in the JobSpec
	Tag string `json:"tag"`

	// Message is the standard out, or standard error output of a job
	Message string `json:"message"`

	// Timestamp is the record of when a message was emitted by a job
	Timestamp string `json:"timestamp"`

	// Color is the terminal color that a job was logged as,
	// color is used to highlight log levels
	Color int `json:"color"`
}

// LogFatalcreates a red LogEntry with a string,
// and calls os.exit(1)
func LogFatal(job *JobSpec, message string) {
	LogFormat(red, job, message)
	os.Exit(1)
}

// LogError creates a red LogEntry with a string
func LogError(job *JobSpec, message string) {
	LogFormat(red, job, message)
}

// LogErrorf creates a red LogEntry with a format string
func LogErrorf(job *JobSpec, format string, a ...interface{}) {
	LogFormat(red, job, format, a...)
}

// LogInfo creates a blue LogEntry with a string
func LogInfo(job *JobSpec, message string) {
	LogFormat(blue, job, message)
}

// LogInfo creates a blue LogEntry with a format string
func LogInfof(job *JobSpec, format string, a ...interface{}) {
	LogFormat(blue, job, format, a...)
}

// LogFormat creates a LogEntry, with the supplied color,
// and format string
func LogFormat(color int, job *JobSpec, format string, a ...interface{}) {
	fmt.Printf("[ \x1b[%dm%s\x1b[0m ] %s\n", color, job.Tag, fmt.Sprintf(format, a...))

	if LogFilename != "" {
		f, err := os.OpenFile(LogFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatal(err.Error())
			}
		}()

		msg, err := json.Marshal(&LogEntry{
			Tag:       job.Tag,
			Message:   fmt.Sprintf(format, a...),
			Timestamp: time.Now().String(),
			Color:     color,
		})

		if _, err = f.WriteString(fmt.Sprintf("%v\n", string(msg))); err != nil {
			log.Fatal()
		}
	}
}

// Status logs the status of all jobs. This is intended to be
// called after all jobs are complete.
func (j *Config) Status(timeStart time.Time) {
	for _, j := range *j {
		if j.exitCode > 0 {
			LogErrorf(&j, "exit=%d, duration=%s", j.exitCode, j.duration)
		} else {
			LogInfof(&j, "exit=%d, duration=%s", j.exitCode, j.duration)
		}
	}
	fmt.Printf("total run time %s\n", time.Now().Sub(timeStart))
}

// Shutdown call Process.Kill() on all jobs
func (j *Config) Shutdown() {
	for _, j := range *j {
		if j.job != nil {
			_ = j.job.Process.Kill()
		}
	}
}

// Run prepares a exec.Cmd, configures logging for the job,
// starts it, and waits for it to exit.
func Run(wg *sync.WaitGroup, job *JobSpec) {
	defer wg.Done()

	job.job = exec.Command(job.Shell, "-c", job.Command)
	job.job.Dir = job.Dir
	job.job.Env = job.Env

	LogInfo(job, "job started")

	stdout, err := job.job.StdoutPipe()
	if err != nil {
		LogFatal(job, err.Error())
	}
	stderr, err := job.job.StderrPipe()
	if err != nil {
		LogFatal(job, err.Error())
	}

	go func() {
		reader := bufio.NewReader(stdout)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			LogInfo(job, scanner.Text())
		}
	}()

	timeStart := time.Now()
	if err := job.job.Start(); err != nil {
		LogFatal(job, err.Error())
	}

	reader := bufio.NewReader(stderr)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		LogInfo(job, scanner.Text())
	}

	if err := job.job.Wait(); err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			if s, ok := e.Sys().(syscall.WaitStatus); ok {
				LogErrorf(job, "job error, exit=%d", s.ExitStatus())
				job.exitCode = s.ExitStatus()

				if s.ExitStatus() > ExitCode {
					ExitCode = s.ExitStatus()
				}
			}
		} else {
			LogError(job, "job error, exit=1")
			job.exitCode = 1

			if 1 > ExitCode {
				ExitCode = 1
			}
		}
	} else {
		LogInfo(job, "job complete, exit=0")
	}

	job.duration = time.Now().Sub(timeStart)
}
