package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"time"

	"github.com/jmizell/GoJobs/jobs"
)

var Jobs = jobs.Config{}

func main() {

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	tag := flag.String("tag", "", "tag to be applied to job output")
	command := flag.String("command", "", "command to run")
	shell := flag.String("shell", "/bin/bash", "the shell to run the command in")
	dir := flag.String("dir", cwd, "directory to run the command in")
	add := flag.Bool("add", false, "adds a command to the jobs file")
	env := flag.String("env", "", "environment variables to set, formatted as a json object. \n"+
		"(default to shells environment variables)")
	run := flag.Bool("run", false, "runs the commands in the jobs file")
	file := flag.String("file", "jobs.json", "file to use for jobs")
	LogFilename := flag.String("logfile", "", "path to where logs are to be written or read from")
	logs := flag.Bool("logs", false, "output logs")
	filter := flag.String("filter", "", "filter log output using this regex")
	flag.Parse()

	// read logs
	if *logs {

		var msgFilter *regexp.Regexp

		if *filter != "" {
			msgFilter, err = regexp.Compile(*filter)
			if err != nil {
				log.Fatal("error compiling regex ", err.Error())
			}
		}

		if _, err := os.Stat(*LogFilename); os.IsNotExist(err) {
			log.Fatal("error finding file ", err.Error())
		}

		f, err := os.Open(*LogFilename)
		if err != nil {
			log.Fatal("error opening file ", err.Error())
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {

			msg := &jobs.LogEntry{}
			if err := json.Unmarshal(scanner.Bytes(), msg); err != nil {
				log.Fatal("error unmarshaling json ", err.Error())
			}

			if *tag != "" && msg.Tag != *tag {
				continue
			}

			if *filter != "" && !msgFilter.MatchString(msg.Message) {
				continue
			}

			jobs.LogFormat(msg.Color, &jobs.JobSpec{Tag: msg.Tag}, msg.Message)
		}
		if err := scanner.Err(); err != nil {
			log.Fatal("error scanning lines for messages ", err.Error())
		}

		os.Exit(0)
	}

	// read in job file
	if _, err := os.Stat(*file); !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(*file)
		if err != nil {
			log.Fatal(err.Error())
		}

		err = json.Unmarshal(b, &Jobs)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// add a job to the config
	if *add {
		if *tag == "" {
			log.Fatal("job tag is required")
		}

		environmentVariables := os.Environ()
		if *env != "" {
			err := json.Unmarshal([]byte(*env), &environmentVariables)
			if err != nil {
				log.Fatal(err.Error())
			}
		}

		Jobs = append(Jobs, jobs.JobSpec{
			Tag:     *tag,
			Command: *command,
			Shell:   *shell,
			Dir:     *dir,
			Env:     environmentVariables,
		})

		b, err := json.MarshalIndent(Jobs, "", "  ")
		if err != nil {
			log.Fatal(err.Error())
		}

		if err := ioutil.WriteFile(*file, b, 0600); err != nil {
			log.Fatalf("couldn't write file %s, %s", *file, err.Error())
		}

		log.Printf("wrote job %s to %s", *tag, *file)
	}

	if *run {
		jobs.LogFilename = *LogFilename

		// Create empty log file
		if jobs.LogFilename != "" {
			err := ioutil.WriteFile(jobs.LogFilename, []byte(""), 0600)
			if err != nil {
				log.Fatal(err)
			}
		}

		wg := &sync.WaitGroup{}
		timeStart := time.Now()

		for i := range Jobs {
			wg.Add(1)
			go jobs.Run(wg, &Jobs[i])
		}

		// make sure we kill child processes
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, os.Kill)
		go func() {
			<-c
			Jobs.Shutdown()
			Jobs.Status(timeStart)
			if 1 > jobs.ExitCode {
				jobs.ExitCode = 1
			}
			os.Exit(jobs.ExitCode)
		}()

		wg.Wait()
		Jobs.Status(timeStart)
		os.Exit(jobs.ExitCode)
	}
}
