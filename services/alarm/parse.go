package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var (
	commentsRegexp  = regexp.MustCompile(`#.*`)
	emptyLineRegexp = regexp.MustCompile(`^[\t ]*$`)
	rowRegexp       = regexp.MustCompile(`(\d{1,2}):(\d{2})[\t ]+(.*)[\t ]*`)
	newlineRegexp   = regexp.MustCompile(`\r?\n`)
)

type Task struct {
	Time        time.Duration
	Description string
}

func ParseSchedule(schedule string) ([]Task, error) {
	var tasks []Task
	for i, line := range newlineRegexp.Split(schedule, -1) {
		line = commentsRegexp.ReplaceAllString(line, "")

		if emptyLineRegexp.MatchString(line) {
			continue
		}

		row := rowRegexp.FindStringSubmatch(line)

		if len(row) != 4 {
			return nil, fmt.Errorf("formatting error on line %d near %q", i, line)
		}

		hour, _ := strconv.Atoi(row[1])
		min, _ := strconv.Atoi(row[2])
		description := row[3]

		when := time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute

		tasks = append(tasks, Task{
			Time:        when,
			Description: description,
		})
	}

	return tasks, nil
}

func FormatSchedule(tasks []Task) string {
	var str string

	for _, task := range tasks {
		hours := int(task.Time / time.Hour)
		mins := int((task.Time % time.Hour) / time.Minute)
		description := task.Description

		str += fmt.Sprintf("%02d:%02d %s\n", hours, mins, description)
	}

	return str
}
