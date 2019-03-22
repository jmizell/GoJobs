# GoJobs
GoJobs is a management tool for running concurrent bash commands in 
the same terminal.  It logs the output of each command and tracks
the final state. 

It's intended as a simple way of monitoring the output of multiple 
services while testing, without having to run multiple terminals, 
background processes, or manually track PIDs.

# Usage

```
Usage of gojobs:
  -add
    	adds a command to the jobs file
  -command string
    	command to run
  -dir string
    	directory to run the command in  
  -env string
    	environment variables to set, formatted as a json list in the form of ["key=value"]. 
    	(default to shells environment variables)
  -file string
    	file to use for jobs (default "jobs.json")
  -filter string
    	filter log output using this regex
  -logfile string
    	path to where logs are to be written or read from
  -logs
    	output logs
  -run
    	runs the commands in the jobs file
  -shell string
    	the shell to run the command in (default "/bin/bash")
  -tag string
    	tag to be applied to job output

```

### Adding a job

You can add a command with only ``gojobs -add -command="[SHELL COMMAND]" -tag="[JOB TAG]"``. 
All other arguments are optional. 

**Warning**: gojobs will by default copy the current environment variables 
as part of the job spec into the config. To override this behavior, use the 
-env flag to specify values as a json list in the form of ["key=value"].

For example, to pass only home, and user, the flag would look like 

```-env='["HOME=/home/user_name","USER=user_name"]'```

##### example

```
gojobs \
    -add \
    -command="[SHELL COMMAND]" \
    -dir="[DIRECTORY TO RUN COMMAND]" \
    -shell="[PATH TO SHELL]" \
    -tag="[JOB TAG]" \
    -env="[JSON OF ENVIRONMENT VARIABLES]" \
    -file="[CONFIG FILE]"
```

### Running jobs

After adding a job to your config, you can immediately run it with ``gojobs -run``. If not log 
file is specified, then the output of the jobs will be written to the terminal only.

##### example

```
gojobs \
    -run \
    -file="[CONFIG FILE]" \
    -logfile="[LOG FILE]"
```

Alternatively, you can add a job, and immediately run it by specifying a new job in the 
arguments. This will add the job to the config, and execute the config in one command.

##### example

```
gojobs \
    -run \
    -command="[SHELL COMMAND]" \
    -dir="[DIRECTORY TO RUN COMMAND]" \
    -shell="[PATH TO SHELL]" \
    -tag="[JOB TAG]" \
    -env="[JSON OF ENVIRONMENT VARIABLES]" \
    -file="[CONFIG FILE]" \
    -logfile="[LOG FILE]"
```

### View the log

You can replay the log to the terminal with ``gojobs -logs -logfile="[LOG FILE]"``. This will
dump all command output to the terminal. You can filter logs by specifying a single tag to 
output, by using a regex pattern to filter the log message field, or both.

##### example

```
gojobs \
    -logs \
    -tag="[FILTER ON THIS TAG]" \
    -filter="[LOG FILTER REGEX]" \
    -logfile="[LOG FILE]"
```

# Install

To install, make sure you have go installed, and your GOBIN is in your path. Then run

```
go get github.com/jmizell/GoJobs
go install github.com/jmizell/GoJobs/cmd/gojobs
```
